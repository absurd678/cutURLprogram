package main

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

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
				"/api": "https://practicum.net",
			},
			inputPath:   "/api",
			methodInput: "GET",
			want: want{
				code:     307,
				location: "https://practicum.net",
			},
		},
		{
			name: "Incorrect test 1",
			mapURL: map[string]string{
				"/api": "https://practicum.net",
			},
			inputPath:   "/test",
			methodInput: "GET",
			want: want{
				code:     400,
				location: "", // ???
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
			req := httptest.NewRequest(tt.methodInput, tt.inputPath, nil)
			res := httptest.NewRecorder()

			h := setEndPoint(tt.mapURL)
			h(res, req)

			ans := res.Result()
			require.Equal(t, res.Code, tt.want.code)
			require.Equal(t, ans.Header.Get("Location"), tt.want.location)

			ans.Body.Close()
		})
	}

	// Post handler test
	for _, tt := range testPost {
		t.Run(tt.name, func(t *testing.T) {
			newBuffer := bytes.NewBuffer([]byte(tt.bodyInput))
			require.NotEmpty(t, newBuffer) // original URL mustn't be empty

			req := httptest.NewRequest(tt.methodInput, tt.inputPath, newBuffer)
			res := httptest.NewRecorder()

			h := setEndPoint(tt.mapURL)
			h(res, req)

			ans := res.Result()
			require.Equal(t, res.Code, tt.want)

			ans.Body.Close()
		})
	}
}
