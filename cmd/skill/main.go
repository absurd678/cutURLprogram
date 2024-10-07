package main

import (
	"net/http"
	"regexp"
	"strings"
)

var validPathget = regexp.MustCompile("^/([a-zA-Z0-9]+)$")
var validPathpost = regexp.MustCompile("^/$")

func Endpoint(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		if m := validPathpost.FindStringSubmatch(req.URL.Path); m == nil {
			http.Error(res, "Invalid URL for post", http.StatusBadRequest)
			return
		}
		res.WriteHeader(http.StatusCreated)
		urlBody := strings.Join(req.Header.Values("Content-Type"), " ")
		res.Header().Add("Content-Type", urlBody)
	} else if req.Method == http.MethodGet {
		if m := validPathget.FindStringSubmatch(req.URL.Path); m == nil {
			http.Error(res, "Invalid URL for get", http.StatusBadRequest)
			return
		}
		res.WriteHeader(http.StatusTemporaryRedirect)
		res.Header().Add("Location", req.URL.String())
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", Endpoint)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
