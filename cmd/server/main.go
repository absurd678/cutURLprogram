package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	"github.com/absurd678/skill/cmd/config"
	"github.com/go-chi/chi/v5"
)

var mapURLmain = map[string]string{
	"sharaga": "https://mai.ru",
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
const shortURLsize int = 10

type Connection struct {
	mapURL map[string]string
}

// RandString генерирует случайную строку длины n
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
		http.Error(res, "Invalid URL for get", http.StatusBadRequest)
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
		http.Error(res, "Error reading request body", http.StatusBadRequest)
		return
	}
	// get the new id from the b flag
	c.mapURL[config.UrlID] = string(original)

	res.WriteHeader(http.StatusCreated)
	// Body answer: localhost:8080/{id}
	res.Write([]byte(req.URL.Path + config.UrlID))
}

func checkURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		log.Println(req.URL.Path)
		if req.Method == http.MethodGet && regexp.MustCompile(`^/[a-zA-Z0-9-]+$`).MatchString(req.URL.Path) {
			next.ServeHTTP(res, req)
		} else if req.Method == http.MethodPost && req.URL.Path == "/" {
			next.ServeHTTP(res, req)
		} else {
			http.Error(res, "Invalid URL", http.StatusBadRequest)
		}
	})
}

func LaunchMyRouter(c *Connection) chi.Router {
	myRouter := chi.NewRouter()
	myRouter.Use(checkURL)
	myRouter.Get("/{id}", c.GetHandler)
	myRouter.Post("/", c.PostHandler)

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
