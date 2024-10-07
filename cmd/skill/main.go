package main

import (
	"net/http"
	"regexp"
	"strings"
)

func postEndpoint(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		res.WriteHeader(http.StatusCreated)
		urlBody := strings.Join(req.Header.Values("Content-Type"), " ")
		res.Header().Add("Content-Type", urlBody)
	}
}

func getEndpoint(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		res.WriteHeader(http.StatusTemporaryRedirect)
		res.Header().Add("Location", req.URL.String())
	} else {
		http.Redirect(res, req, req.URL.String(), http.StatusCreated)
	}
}

var validPath = regexp.MustCompile("^/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.Error(w, "Bad URL path", http.StatusBadRequest)
			return
		}
		fn(w, r)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", makeHandler(postEndpoint))
	mux.HandleFunc("/get", makeHandler(getEndpoint))
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
