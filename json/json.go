// Based on jsonload, formerly called json.go.  This is jsonmap because it loads all the valid poems into
// a cache that is a map type and then reads them from the cache

package main;

import (
	"fmt"
	"poetry"
	"net/http"
	"encoding/json"
	"os"
	"strconv"
	"log"
	"flag"
	"time"
	)

var c config; // need this to be global so the functions can use it, just like main() can.

var cache map[string]poetry.Poem;

type config struct {
	Route string;  // the URL to respond to for poetry requests
	BindAddress string  `json:"addr"` ;  // port to bind on
	ValidPoems []string `json:"valid"`; 
}



type poemWithTitle struct {
	Title string;  // this is exported, but it could be private if wanted.  IE, call it title
	Body poetry.Poem;
	WordCount int;
	TheCount int;
	WordCountStrHex string;
}

func poemHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm();
	poemName := r.Form["name"][0]                // first element of a slice

	log.Printf(" In poemhandler.  User requested poemname= %s\n",poemName);
	p,ok := cache[poemName];

	if !ok {
//	  log.Printf(" problem in getting from cache.  PoemName: %s\n",poemName);
//	  http.Error(w, " Not found (invalid) ",http.StatusNotFound);
	  return;
	}


//	sort.Sort(p[0]); // sort the first stanza.  Not in the version of this code using a poem cache.

// poemName is passed in by the curl line param
	pwt := poemWithTitle{poemName,p, p.GetNumWords(), p.GetNumThe(),
                                                                strconv.FormatInt(int64(p.GetNumWords()),16)};
	enc := json.NewEncoder(w);
	enc.Encode(pwt);
}


func main() {
	var anotherConfigFilename string;
	startTime := time.Now();

	configFilename := flag.String("conf","config","Name of configuration file");
	flag.StringVar(&anotherConfigFilename,"c","config.cfg","Name of another (unused) configuration file");
	flag.Parse();
	log.Println(" configfilename is",*configFilename,", anotherconfigfilename is",anotherConfigFilename);

	f,err := os.Open(*configFilename);
	if err != nil {
	  fmt.Println(" Cannot find config file",*configFilename,".  Exiting");
	  os.Exit(1);
	}

// maps have to be initialized as the do not have an initial zero value.
	cache = make(map[string]poetry.Poem);

// read config file
	dec := json.NewDecoder(f);
	err = dec.Decode(&c);  // this modifies the structure, so need ADROF operator
	f.Close();
	if err != nil {
	  fmt.Println( " Cannot decode the config file.  Exiting");
	  os.Exit(1);
	}

	for _,name := range c.ValidPoems {
	  cache[name],err = poetry.LoadPoem(name);
	  log.Println(" name=",name,", err=",err);
	  if err != nil {
	    log.Fatalf(" Cannot load a poem in the ValidPoems list: %s\n  Exiting.",name);
	  }
	}

	log.Println("loaded cache:",cache);

	log.Println(" Will now start the web server, looking for cached poem.");

	elapsed1 := time.Since(startTime);
	elapsed2 := time.Now().Sub(startTime);
	log.Println(" elapsed time since startup",elapsed1,", and",elapsed2);
 
	stopTime := time.Now();
	timediff := stopTime.Sub(startTime);
	log.Println(" time difference is",timediff);

	http.HandleFunc(c.Route,poemHandler);
	http.ListenAndServe(c.BindAddress,nil);

// Once this is started, I have to test it.  Easiest way is to use curl.  So from another terminal I did:
//  curl -v http://127.0.0.1:8088/poem\?name=shortpoem.txt  
//  or whatever file in ~/gocode that I wanted displayed.
// And now that this is json, he had me pipe output thru json_pp like this:
// curl http://127.0.0.1:8088/poem\?name=shortpoem.txt | json_pp
// curl http://localhost:8088/poem\?name=shortpoem.txt | json_pp     this line also works 

}
