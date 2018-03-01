package main

import "fmt"
import "net/http"

func main() {
	fmt.Println("vim-go")
	http.HandleFunc("/", index_handler)
	http.HandleFunc("/about/", about_handler)
	http.ListenAndServe(":8000", nil)
}

func index_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "index_handler output")
}

func about_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "about_handler output message")
}
