package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRequestOptions struct {
	t            *testing.T
	ts           *httptest.Server
	method, path string
	body         io.Reader
}

func testRequest(opts testRequestOptions) *http.Response {
	req, err := http.NewRequest(
		opts.method,
		opts.ts.URL+opts.path,
		opts.body,
	)
	require.NoError(opts.t, err)
	opts.ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := opts.ts.Client().Do(req)
	require.NoError(opts.t, err)
	return resp
}

// Test the Get handler
func Test_GetHandler(t *testing.T) {
	tests := []struct {
		Name         string
		MapURL       map[string]string
		Path         string
		Method       string // Delete method no much need
		WantCode     int
		WantLocation string
	}{
		{
			Name: "OK",
			MapURL: map[string]string{
				"sharaga": "https://mai.ru",
			},
			Path:         "/sharaga",
			Method:       http.MethodGet,
			WantCode:     307,
			WantLocation: "https://mai.ru",
		},
		{
			Name: "Requested url not in map", // Test name need to be more informative
			MapURL: map[string]string{
				"api": "https://practicum.net",
			},
			Path:     "/test",
			Method:   http.MethodGet,
			WantCode: 400,
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			connection := &Connection{tc.MapURL}
			ts := httptest.NewServer(LaunchMyRouter(connection))
			resp := testRequest(testRequestOptions{
				t:      t,
				ts:     ts,
				method: tc.Method,
				path:   tc.Path,
				body:   nil,
			})
			assert.Equal(t, tc.WantCode, resp.StatusCode)
			assert.Equal(t, tc.WantLocation, resp.Header.Get("Location"))
		})
	}
}

// Test the Post handler
func Test_PostHandler(t *testing.T) {
	tests := []struct {
		Name     string
		MapURL   map[string]string
		Path     string
		Method   string
		Body     string
		WantCode int
	}{
		{
			Name:     "OK",
			MapURL:   map[string]string{},
			Path:     "/",
			Method:   http.MethodPost,
			Body:     "https://practicum.net",
			WantCode: 201,
		},
		{
			Name:     "Bad request", // Test name need to be more informative
			MapURL:   map[string]string{},
			Path:     "/unneededID",
			Method:   http.MethodPost,
			Body:     "https://practicum.net",
			WantCode: 400,
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			newBuffer := bytes.NewBuffer([]byte(tc.Body))
			assert.NotEmpty(t, newBuffer) // original URL mustn't be empty
			testConnect := &Connection{tc.MapURL}
			ts := httptest.NewServer(LaunchMyRouter(testConnect))
			resp := testRequest(testRequestOptions{
				t:      t,
				ts:     ts,
				method: tc.Method,
				path:   tc.Path,
				body:   newBuffer,
			})
			assert.Equal(t, tc.WantCode, resp.StatusCode)
		})
	}
}
