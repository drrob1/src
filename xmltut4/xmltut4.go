package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type NewsAggPage struct {
	Title, News string
}

func main() {
	http.HandleFunc("/", index_handler)
	http.HandleFunc("/agg/", NewsAggHandler) // news aggregator
	http.ListenAndServe(":8000", nil)        // assumes localhost

}

func index_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "output message from index_handler")
}

func NewsAggHandler(w http.ResponseWriter, r *http.Request) {
	p := NewsAggPage{Title: "This can be anything I want", News: "Some News"}
	t, err := template.ParseFiles("basictemplate.html")
	check(err)
	t.Execute(w, p)
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
