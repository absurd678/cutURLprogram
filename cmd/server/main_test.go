package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testRequest(t *testing.T, ts *httptest.Server, method,
	path string, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, body)
	fmt.Println(ts.URL + path) //
	assert.NoError(t, err)

	resp, err := ts.Client().Do(req)
	assert.NoError(t, err)

	return resp
}

func TestHandlers(t *testing.T) {

	type want struct {
		code     int
		location string
	}
	type testsGetStruct struct {
		name        string
		mapURL      map[string]string
		inputPath   string
		methodInput string // Delete method no much need
		want        want
	}
	type testsPostStruct struct {
		name        string
		mapURL      map[string]string
		inputPath   string
		methodInput string
		bodyInput   string
		want        int // just answer code
	}

	testGet := []testsGetStruct{
		{
			name: "Correct test 1",
			mapURL: map[string]string{
				"sharaga": "https://mai.ru",
			},
			inputPath:   "/sharaga",
			methodInput: "GET",
			want: want{
				code:     307,
				location: "https://mai.ru",
			},
		},
		{
			name: "Incorrect test 1",
			mapURL: map[string]string{
				"api": "https://practicum.net",
			},
			inputPath:   "/test",
			methodInput: "GET",
			want: want{
				code:     400,
				location: "",
			},
		},
	}

	testPost := []testsPostStruct{
		{
			name:        "Correct test 1",
			mapURL:      map[string]string{},
			inputPath:   "/",
			methodInput: "POST",
			bodyInput:   "https://practicum.net",
			want:        201,
		},
		{
			name:        "Incorrect test 1",
			mapURL:      map[string]string{},
			inputPath:   "/unneededID",
			methodInput: "POST",
			bodyInput:   "https://practicum.net",
			want:        400,
		},
	}

	// Get handler test
	for _, tt := range testGet {
		t.Run(tt.name, func(t *testing.T) {

			testConnect := &Connection{tt.mapURL}
			ts := httptest.NewServer(LaunchMyRouter(testConnect))
			res := testRequest(t, ts, tt.methodInput, tt.inputPath, nil)

			assert.Equal(t, tt.want.code, res.StatusCode)                 // Получаем код ответа
			assert.Equal(t, tt.want.location, res.Header.Get("Location")) // Получаем заголовок "Location"

		})
	}

	// Post handler test
	for _, tt := range testPost {
		t.Run(tt.name, func(t *testing.T) {
			newBuffer := bytes.NewBuffer([]byte(tt.bodyInput))
			assert.NotEmpty(t, newBuffer) // original URL mustn't be empty

			testConnect := &Connection{tt.mapURL}
			ts := httptest.NewServer(LaunchMyRouter(testConnect))
			res := testRequest(t, ts, tt.methodInput, tt.inputPath, newBuffer)

			assert.Equal(t, tt.want, res.StatusCode)

		})
	}
}
