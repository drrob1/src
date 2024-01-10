package handlers

import (
	"net/http"
	"net/http/httptest"
)

func testingHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":

	case "GET":

	case "PUT":
		w.WriteHeader(http.StatusMethodNotAllowed) // this is a 405 error

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func testMockServer() *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/xml")
	}

	return httptest.NewServer(http.HandlerFunc(f)) // when called, this returns the random port it uses.  When a browser is used to hit that port, the Go code executes.
	// HandlerFunc is a type conversion, not a function call.
}
