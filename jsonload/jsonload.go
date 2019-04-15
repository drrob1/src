package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"poetry"
	"sort"
	"strconv"
)

var c config // need this to be global so the functions can use it, just like main() can.

type config struct {
	Route       string   // the URL to respond to for poetry requests
	BindAddress string   `json:"addr"` // port to bind on
	ValidPoems  []string `json:"valid"`
}

type poemWithTitle struct {
	Title           string // this is exported, but it could be private if wanted.  IE, call it title
	Body            poetry.Poem
	WordCount       int
	TheCount        int
	WordCountStrHex string
}

func poemHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	poemName := r.Form["name"][0] // first element of a slice

	found := false
	for _, v := range c.ValidPoems {
		if v == poemName {
			found = true
			break
		}
	}

	if !found {
		http.Error(w, " Not found (invalid) ", http.StatusNotFound)
		return
	}

	p, err := poetry.LoadPoem(poemName)
	if err != nil {
		http.Error(w, " File not found.", http.StatusNotFound)
		return
	}

	sort.Sort(p[0]) // sort the first stanza

	// poemName is passed in by the curl line param
	pwt := poemWithTitle{poemName, p, p.GetNumWords(), p.GetNumThe(),
		strconv.FormatInt(int64(p.GetNumWords()), 16)}
	enc := json.NewEncoder(w)
	//	enc.Encode(p);    Now that we defined poemWithTitle, this line is changed to
	enc.Encode(pwt)
}

func main() {
	f, err := os.Open("config")
	if err != nil {
		fmt.Println(" Cannot find config file.  Exiting")
		os.Exit(1)
	}

	dec := json.NewDecoder(f)
	//	err = dec.Decode(c);  // this line is an error because of lack of ADROF operator
	err = dec.Decode(&c) // this modifies the structure, so need ADROF operator
	f.Close()
	if err != nil {
		fmt.Println(" Cannot decode the config file.  Exiting")
		os.Exit(1)
	}

	p, err := poetry.LoadPoem("shortpoem.txt")
	if err != nil {
		fmt.Println(" Error reading from shortpoem.txt", err)
	}
	fmt.Println(p)
	fmt.Println()
	fmt.Printf("%#v\n", p)

	//
	fmt.Println(" Will now start the web server.")
	//	http.HandleFunc("/poem", poemHandler);  these lines are replaced by the config stuff
	//	http.ListenAndServe(":8088", nil);           // there could be an IP adr here also, in addition to the port
	http.HandleFunc(c.Route, poemHandler)
	http.ListenAndServe(c.BindAddress, nil)

	// Once this is started, I have to test it.  Easiest way is to use curl.  So from another terminal I did:
	//  curl -v http://127.0.0.1:8088/poem\?name=shortpoem.txt
	//  or whatever file in ~/gocode that I wanted displayed.
	// And now that this is json, he had me pipe output thru json_pp like this:
	// curl http://127.0.0.1:8088/poem\?name=shortpoem.txt | json_pp
	// curl http://localhost:8088/poem\?name=shortpoem.txt | json_pp     this line also works

}
