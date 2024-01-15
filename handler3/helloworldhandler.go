package main

import (
	"fmt"
	"net/http"
)

//type HelloHandler struct{}  not needed now

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

//type WorldHandler struct{}  not needed now

func world(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "World!")
}

//type helloWorldHandler struct{}  not needed now

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func main() {
	//hello := HelloHandler{}
	//world := WorldHandler{}
	//helloWorld := helloWorldHandler{}
	pauseChan := make(chan bool)

	server := http.Server{
		Addr: "127.0.0.1:8080",
	}

	http.HandleFunc("/hello/", hello) // will now catch more.  See text
	http.HandleFunc("/world/", world) // will now catch more.  See text
	http.HandleFunc("/", helloWorld)

	fmt.Printf(" About to start the server on localhost:8080\n")
	go func() {
		server.ListenAndServe() // this is blocking
		close(pauseChan)
	}()
	fmt.Printf(" Started the server on localhost:8080\n")

	// If I don't pause here, the pgm will just exit, which will stop the server, too.
	<-pauseChan

	fmt.Printf(" Fell off the edge of the pgm, stopping any goroutines, so the server is now stopped.") // Yep, that's what happened without the pause channel read op above.
}
