package main

import (
	"fmt"
	"net/http"
)

type HelloHandler struct{}

func (h *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

type WorldHandler struct{}

func (h *WorldHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "World!")
}

type helloWorldHandler struct{}

func (h *helloWorldHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func main() {
	hello := HelloHandler{}
	world := WorldHandler{}
	helloworld := helloWorldHandler{}

	server := http.Server{
		Addr: "127.0.0.1:8080",
	}

	http.Handle("/hello", &hello)
	http.Handle("/world", &world)
	http.Handle("/", &helloworld)

	fmt.Printf(" About to start the server on localhost:8080\n")
	server.ListenAndServe()
	fmt.Printf(" Started the server on localhost:8080\n") // The only way to stop the server is ^C, so this line isn't executed.

}
