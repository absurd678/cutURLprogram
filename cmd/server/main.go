package main

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	"github.com/absurd678/skill/cmd/config"
	"github.com/absurd678/skill/internal/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var mapURLmain = map[string]string{
	"sharaga": "https://mai.ru",
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
const shortURLsize int = 10

// ----------------------STRUCTURES----------------------------
type (
	Connection struct {
		mapURL map[string]string
	}

	responseData struct { // the field of logResponse
		code int
		size int
	}

	logResponse struct { // to log response data
		res  http.ResponseWriter
		data *responseData
	}
)

// ----------------------logResponse-------------------------------
func (lR *logResponse) Write(b []byte) (int, error) {
	size, err := lR.res.Write(b)
	if err != nil {
		return 0, err
	}
	lR.data.size += size
	return size, nil
}

func (lR *logResponse) WriteHeader(StatusCode int) {
	lR.res.WriteHeader(StatusCode)
	lR.data.code = StatusCode
}

func (lR *logResponse) Header() http.Header {
	return lR.res.Header()
}

//-------------------------------------------------------------------

// RandString generates a random string with the given length
func RandString(n int) string {
	// rand.Seed is deprecated, use NewSource instead :D
	r := rand.New(rand.NewSource(time.Now().Unix()))
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[r.Intn(len(letterBytes))]
	}
	return string(b)
}

func (c *Connection) GetHandler(res http.ResponseWriter, req *http.Request) {
	// take {id} and search for value in the map
	shortURL := chi.URLParam(req, "id")
	original, ok := c.mapURL[shortURL]
	if !ok {
		res.WriteHeader(http.StatusBadRequest) // DOESN'T WORK to fill code field for logResponse
		res.Write([]byte("Invalid URL for GET"))
		return
	}

	// Add the Location header with original URL
	res.Header().Add("Location", original) // No location actually sent. However the header is added.
	res.WriteHeader(http.StatusTemporaryRedirect)
	res.Write([]byte(""))
}

func (c *Connection) PostHandler(res http.ResponseWriter, req *http.Request) {

	// Get the URL from the body (and the new id also)
	original, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest) // to fill code field for logResponse
		res.Write([]byte("Invalid URL for POST"))
		return
	}
	// get the new id from the b flag
	c.mapURL[config.UrlID] = string(original)

	res.WriteHeader(http.StatusCreated)
	// Body answer: localhost:8080/{id}
	res.Write([]byte(req.URL.Path + config.UrlID))
}

func (c *Connection) PostHandlerJSON(res http.ResponseWriter, req *http.Request) {
	// get json: {"url": "some_url"}
	// return json: {"result": "short_url"}
	var some_url models.SomeURL
	var short_url models.ShortURL
	var buff []byte
	var err error

	if err = json.NewDecoder(req.Body).Decode(&some_url); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	short_url = models.ShortURL{URL: config.UrlID}
	c.mapURL[short_url.URL] = some_url.URL
	res.WriteHeader(http.StatusCreated)
	if buff, err = json.MarshalIndent(short_url, "", " "); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte("Unmarshable data"))
		return
	}
	res.Write(buff)
}

func checkURL(next http.Handler) http.Handler { // to avoid paths like localhost:8080/{id}/extrapath

	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// Logging setup
		middlewareLogger, err := zap.NewDevelopment()
		if err != nil {
			http.Error(res, "Logger error", http.StatusInternalServerError)
		}
		sugarLogger := middlewareLogger.Sugar() // for JSON-like messages
		// Logging request
		sugarLogger.Infow("Request parameters",
			"URI", req.RequestURI,
			"Method", req.Method,
		)
		// ResponseWriter implementation
		logRW := &logResponse{res, &responseData{code: 0, size: 0}}
		timeDuration := time.Now() // query duration

		// Handlers
		if req.Method == http.MethodGet && regexp.MustCompile(`^/[a-zA-Z0-9-]+$`).MatchString(req.URL.Path) {
			next.ServeHTTP(logRW, req)
		} else if req.Method == http.MethodPost && req.URL.Path == "/" {
			next.ServeHTTP(logRW, req)
		} else if req.Method == http.MethodPost && req.URL.Path == "/api/shorten" {
			next.ServeHTTP(logRW, req)
		} else {
			http.Error(res, "Invalid URL", http.StatusBadRequest)
			logRW.WriteHeader(http.StatusBadRequest)
			logRW.Write([]byte("Invalid URL"))
		}
		// Logging response
		sugarLogger.Infow(
			"Response parameters",
			"Status Code", logRW.data.code,
			"Size", logRW.data.size,
			"Duration", time.Since(timeDuration),
		)
	})
}

func LaunchMyRouter(c *Connection) chi.Router {
	myRouter := chi.NewRouter()
	myRouter.Use(checkURL)
	myRouter.Get("/{id}", c.GetHandler)
	myRouter.Post("/", c.PostHandler)
	myRouter.Post("/api/shorten", c.PostHandlerJSON)

	return myRouter
}

func main() {

	c := &Connection{mapURLmain}

	config.ParseFlags() // read a and b flags for host:port and {id} information

	err := http.ListenAndServe(config.HostFlags.String(), LaunchMyRouter(c))
	if err != nil {
		panic(err)
	}
}
