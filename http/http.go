package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

/*
  9 Aug 25 -- This was originally written from "Black Hat Go".  I don't need this.  I'm going to change it to implement my idea towards lint updating itself.
               First, I have to see if I can get it to list directory contents.
               Turns out I did do this, in digest.go.  It uses a GitHub package called grab.  I'll use that, so I don't have to write my own code to do this.
*/

const urlRwsNet = "http://drrws.net/"
const urlRobSolomonName = "http://robsolomon.name/"
const urlHostGator = "http://drrws.com"

func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	dir := "." // or any directory you wish to list
	entries, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var files []string
	for _, entry := range entries {
		files = append(files, entry.Name())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func main() {

	http.HandleFunc("/", listFilesHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
