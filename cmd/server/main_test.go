package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type testRequestOptions struct {
	t            *testing.T
	ts           *httptest.Server
	method, path string
	body         io.Reader
}

// TestRequest is used instead of Client().Do() when need to check CheckRedirect
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
			require.Equal(t, tc.WantCode, resp.StatusCode)
			require.Equal(t, tc.WantLocation, resp.Header.Get("Location"))
		})
	}
}

// Test compression - no need, the body is always empty
/*
func Test_GzipGetHandler(t *testing.T) {
	tests := []struct {
		Name         string
		MapURL       map[string]string
		Path         string
		Method       string // Delete method no much need
		WantCode     int
		WantLocation string
	}{
		{
			Name: "accept_encoding",
			MapURL: map[string]string{
				"sharaga": "https://mai.ru",
			},
			Path:         "/sharaga",
			Method:       http.MethodGet,
			WantCode:     307,
			WantLocation: "https://mai.ru",
		},
	}
	for _, tc := range tests { // Accept compression
		t.Run(tc.Name, func(t *testing.T) {
			connection := &Connection{tc.MapURL}
			ts := httptest.NewServer(LaunchMyRouter(connection))

			req, err := http.NewRequest(
				tc.Method,
				ts.URL+tc.Path,
				nil,
			)
			req.Header.Set("Accept-Encoding", "gzip") // Accept compression

			require.NoError(t, err)
			ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
			require.NoError(t, err)
			resp, err := ts.Client().Do(req)

			// Decode the test body
			var bodyResp []byte
			resp.Body.Read(bodyResp)
			newBuffer := bytes.NewBuffer(bodyResp)
			reader, err := gzip.NewReader(newBuffer)
			require.NoError(t, err)
			_, err = reader.Read(bodyResp)
			require.NoError(t, err)

			require.Equal(t, tc.WantCode, resp.StatusCode)
			require.Equal(t, tc.WantLocation, resp.Header.Get("Location"))
		})
	}

}
*/

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
			require.NotEmpty(t, newBuffer) // original URL mustn't be empty
			testConnect := &Connection{tc.MapURL}
			ts := httptest.NewServer(LaunchMyRouter(testConnect))
			resp := testRequest(testRequestOptions{
				t:      t,
				ts:     ts,
				method: tc.Method,
				path:   tc.Path,
				body:   newBuffer,
			})
			require.Equal(t, tc.WantCode, resp.StatusCode)
		})
	}
}

// TESTING THE COMPRESSION

func Test_GzipPostHandler(t *testing.T) {
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
	}

	// Accept encoding
	// ! server sends compressed Body data
	// ! client decompresses data
	for _, tc := range tests {
		t.Run("accept_encoding", func(t *testing.T) {
			var bodyResp []byte
			newBuffer := bytes.NewBuffer([]byte(tc.Body))
			require.NotEmpty(t, newBuffer) // original URL mustn't be empty
			testConnect := &Connection{tc.MapURL}

			// Set request params
			ts := httptest.NewServer(LaunchMyRouter(testConnect))
			req, err := http.NewRequest(
				tc.Method,
				ts.URL+tc.Path,
				nil,
			)
			req.Header.Set("Accept-Encoding", "gzip") // Accept compression
			require.NoError(t, err)

			// do a request
			resp, err := ts.Client().Do(req)

			// Get a response
			// Decode the test body
			require.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))
			reader, err := gzip.NewReader(resp.Body)
			require.NoError(t, err)
			bodyResp, err = io.ReadAll(reader)

			require.NoError(t, err)
			require.Equal(t, tc.WantCode, resp.StatusCode)
			require.NotEmpty(t, bodyResp) // the shorturl can be anything so just make sure it has been decoded

			resp.Body.Close()
		})
	}

	// Content encoding
	// ! server gets compressed data
	// ! client sends compressed data
	for _, tc := range tests {
		t.Run("content_encoding", func(t *testing.T) {

			// check there is original URL at all
			newBuffer := bytes.NewBuffer([]byte(tc.Body))
			require.NotEmpty(t, newBuffer)

			// Encode the test body
			buf := bytes.NewBuffer(nil)
			writer := gzip.NewWriter(buf)
			_, err := writer.Write([]byte(tc.Body))
			require.NoError(t, err)
			err = writer.Close()
			require.NoError(t, err)

			// set request params
			testConnect := &Connection{tc.MapURL}
			ts := httptest.NewServer(LaunchMyRouter(testConnect))
			req, err := http.NewRequest(
				tc.Method,
				ts.URL+tc.Path,
				buf,
			)
			require.NoError(t, err)
			req.Header.Set("Content-Encoding", "gzip") // compression

			// do a request
			resp, err := ts.Client().Do(req)

			// after a response
			require.Equal(t, tc.WantCode, resp.StatusCode)
			resp.Body.Close()
		})
	}
}

// Test the JSON handler
func TestPostHandlerJSON(t *testing.T) {
	testBlock := []struct {
		Name     string
		MapURL   map[string]string // you can't use handler without content struct type so the map is needed :(
		Path     string
		Method   string
		Body     string
		WantCode int
	}{
		{
			Name:     "Positive test 1",
			MapURL:   map[string]string{},
			Path:     "/api/shorten",
			Method:   http.MethodPost,
			Body:     `{"url": "https://ilovebebra.com"}`,
			WantCode: http.StatusCreated,
		},
		{
			Name:     "Negative test 1", // incorrect path
			MapURL:   map[string]string{},
			Path:     "/api/path",
			Method:   http.MethodPost,
			Body:     `{"url": "https://ilovebebra.com"}`,
			WantCode: http.StatusBadRequest,
		},
		{
			Name:     "Negative test 2", // incorrect method
			MapURL:   map[string]string{},
			Path:     "/api/shorten",
			Method:   http.MethodPut,
			Body:     `{"url": "https://ilovebebra.com"}`,
			WantCode: http.StatusBadRequest,
		},
		{
			Name:     "Negative test 3", // incorrect json
			MapURL:   map[string]string{},
			Path:     "/api/shorten",
			Method:   http.MethodPost,
			Body:     `<"url": "https://ilovebebra.com">`,
			WantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testBlock {
		t.Run(tc.Name, func(t *testing.T) {
			newConnect := &Connection{mapURL: tc.MapURL} // connect having optional map
			newBody := bytes.NewBuffer([]byte(tc.Body))
			require.NotEmpty(t, newBody) // body must json, not empty

			ts := httptest.NewServer(LaunchMyRouter(newConnect))
			resp := testRequest(
				testRequestOptions{
					t:      t,
					ts:     ts,
					method: tc.Method,
					path:   tc.Path,
					body:   newBody,
				},
			)
			require.Equal(t, tc.WantCode, resp.StatusCode)
		})
	}
}

// CHECK THE COMPRESSION
func Test_GzipPostHandlerJSON(t *testing.T) {
	testBlock := []struct {
		Name     string
		MapURL   map[string]string // you can't use handler without content struct type so the map is needed :(
		Path     string
		Method   string
		Body     string
		WantCode int
	}{
		{
			Name:     "Positive test 1",
			MapURL:   map[string]string{},
			Path:     "/api/shorten",
			Method:   http.MethodPost,
			Body:     `{"url": "https://ilovebebra.com"}`,
			WantCode: http.StatusCreated,
		},
	}

	// Accept encoding
	for _, tc := range testBlock {
		t.Run("accept_encoding", func(t *testing.T) {
			var bodyResp []byte

			newBuffer := bytes.NewBuffer([]byte(tc.Body))
			require.NotEmpty(t, newBuffer) // original URL mustn't be empty
			testConnect := &Connection{tc.MapURL}

			// set request parameters
			ts := httptest.NewServer(LaunchMyRouter(testConnect))
			req, err := http.NewRequest(
				tc.Method,
				ts.URL+tc.Path,
				newBuffer,
			)
			req.Header.Set("Accept-Encoding", "gzip") // Accept compression
			require.NoError(t, err)

			// Do a request
			resp, err := ts.Client().Do(req)

			// Get a response
			// Decode the test body
			require.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

			reader, err := gzip.NewReader(resp.Body)
			require.NoError(t, err)
			bodyResp, err = io.ReadAll(reader)

			require.NoError(t, err)
			require.Equal(t, tc.WantCode, resp.StatusCode)
			require.NotEmpty(t, bodyResp) // the shorturl can be anything so just make sure it has been decoded

			resp.Body.Close()
		})
	}

	// Content encoding

	for _, tc := range testBlock {
		t.Run("content_encoding", func(t *testing.T) {

			// check there is original URL at all
			newBuffer := bytes.NewBuffer([]byte(tc.Body))
			require.NotEmpty(t, newBuffer)

			// Encode the test body
			buf := bytes.NewBuffer(nil)
			writer := gzip.NewWriter(buf)
			_, err := writer.Write([]byte(tc.Body))
			require.NoError(t, err)
			err = writer.Close()
			require.NoError(t, err)

			// Set a request
			testConnect := &Connection{tc.MapURL}
			ts := httptest.NewServer(LaunchMyRouter(testConnect))
			req, err := http.NewRequest(
				tc.Method,
				ts.URL+tc.Path,
				buf,
			)
			require.NoError(t, err)
			req.Header.Set("Content-Encoding", "gzip") // compression

			// do a request
			resp, err := ts.Client().Do(req)

			// assert response
			require.Equal(t, tc.WantCode, resp.StatusCode)
			resp.Body.Close()
		})
	}
}
