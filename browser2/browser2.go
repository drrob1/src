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
	fmt.Fprintf(w, `	<h1>index output message</h1>
	<p>paragraph about_handler output message</p>
	<p>new paragraph about_handler output message</p>
	`)
}

func about_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>about_handler output message</h1>")
	fmt.Fprintln(w, "<p>paragraph about_handler output message</p>")
	fmt.Fprintln(w, "<p>new paragraph about_handler output message</p>")
	fmt.Fprintf(w, "<p>You %s pass %s</p>", "can", "<strong>variables</strong>")
}
