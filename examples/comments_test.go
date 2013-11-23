package main

import (
	"net/http"
	"net/http/httptest"

	"testing"
)

func createServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(commentsHandler))
}

func TestRequest(t *testing.T) {
	server := createServer()
	defer server.Close()

	http.Get(server.URL)
}
