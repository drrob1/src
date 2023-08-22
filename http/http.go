package main

import (
	"fmt"
	"net/http"
	"src/poetry"
)

func poemHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	poemName := r.Form["name"][0] // first element of a slice

	p, err := poetry.LoadPoem(poemName)
	if err != nil {
		http.Error(w, " File not found.", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, " %s \n", p)
}

func main() {
	p, err := poetry.LoadPoem("shortpoem.txt")
	if err != nil {
		fmt.Println(" Error reading from shortpoem.txt", err)
	}
	fmt.Println(p)
	fmt.Println()
	fmt.Printf("%#v\n", p)

	//
	fmt.Println(" Will now start the web server.")
	http.HandleFunc("/poem", poemHandler)
	http.ListenAndServe(":8088", nil) // there could be an IP adr here also, in addition to the port

	// Once this is started, I have to test it.  Easiest way is to use curl.  So from another terminal I did:
	//  curl -v http://127.0.0.1:8088/poem\?name=shortpoem.txt
	//  or whatever file in ~/gocode that I wanted displayed.

}
