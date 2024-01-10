package handlers

import (
	"io"
	"net/http"
)

func healthCheckHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "ok")
}
