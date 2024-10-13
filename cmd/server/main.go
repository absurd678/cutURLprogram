package main

import (
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
)

var validPathget = regexp.MustCompile("^/([a-zA-Z0-9]+)$")
var validPathpost = regexp.MustCompile("^/$")
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
	rand.Seed(time.Now().UnixNano()) // Инициализируем случайный генератор
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
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
	res.WriteHeader(http.StatusTemporaryRedirect)
	res.Header().Add("Location", original) // No location actually sent. However the header is added.

}

func (c *Connection) PostHandler(res http.ResponseWriter, req *http.Request) {
	// Get the URL from the body (and the new id also)
	original, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Generate the shortened URL
	newURL := RandString(shortURLsize)
	res.WriteHeader(http.StatusCreated)
	c.mapURL[newURL] = string(original)

	// Body answer: localhost:8080/{id}
	res.Write([]byte(req.URL.Path + "/" + newURL))
}

func checkURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
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
	err := http.ListenAndServe(`:8080`, LaunchMyRouter(c))
	if err != nil {
		panic(err)
	}
}
