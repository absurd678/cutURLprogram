package main

import (
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"time"
)

var validPathget = regexp.MustCompile("^/([a-zA-Z0-9]+)$")
var validPathpost = regexp.MustCompile("^/$")
var mapURLmain = map[string]string{}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
const shortURLsize int = 10

// RandString генерирует случайную строку длины n
func RandString(n int) string {
	rand.Seed(time.Now().UnixNano()) // Инициализируем случайный генератор
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func GetHandler(res http.ResponseWriter, req *http.Request, mapURL map[string]string) {
	if m := validPathget.FindStringSubmatch(req.URL.Path); m == nil {
		http.Error(res, "Invalid URL for get", http.StatusBadRequest)
		return
	}
	// take {id} and search for value in the map
	shortURL := req.URL.Path
	original, ok := mapURL[shortURL]
	if !ok {
		http.Error(res, "Invalid URL for get", http.StatusBadRequest)
	}

	// Add the Location header with original URL
	res.Header().Add("Location", original)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

// TODO: return back random URL keys
func PostHandler(res http.ResponseWriter, req *http.Request, mapURL map[string]string) {
	if m := validPathpost.FindStringSubmatch(req.URL.Path); m == nil {
		http.Error(res, "Invalid URL for post", http.StatusBadRequest)
		return
	}

	// Get the URL from the body (and the new id also)
	original, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Generate the shortened URL
	newURL := RandString(shortURLsize)
	res.WriteHeader(http.StatusCreated)
	mapURL["/"+newURL] = string(original)

	// Body answer: localhost:8080/{id}
	res.Write([]byte(req.URL.Path + "/" + newURL))
}

func setEndPoint(mapURL map[string]string) func(res http.ResponseWriter, req *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			GetHandler(res, req, mapURL)
		case http.MethodPost:
			PostHandler(res, req, mapURL)
		default:
			http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", setEndPoint(mapURLmain))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
