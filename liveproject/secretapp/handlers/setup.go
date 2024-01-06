package handlers

import (
	"net/http"
)

func SetupHandlers(mux *http.ServeMux) {

	mux.HandleFunc("/healthcheck", healthCheckHandler)
	mux.HandleFunc("/", secretHandler)
}
