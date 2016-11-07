package main

import (
        "fmt"
	"net/http"
	"io/ioutil"
//	"time"
//	"math/rand"
//	"sync/atomic"
       );

type webPage struct {
  url string;
  body []byte;
  err error;
}




func (w *webPage) get() {
  resp, err := http.Get(w.url);
  if err != nil {
    w.err = err;
    return;
  }
  defer resp.Body.Close();

  w.body, err = ioutil.ReadAll(resp.Body)
  if err != nil {
    w.err = err;
  }
}

func (w *webPage) isOK() bool {
  return w.err == nil
}

func main() {

	w := new(webPage) // this creates an empty variable of this struct type.
	w.url = "http://robsolomon.info/";

	w.get();
	if w.isOK() {
	  fmt.Println("URL:",w.url,", length:",len(w.body));
	}else{
	  fmt.Println("URL:",w.url,", error:",w.err);
	}
	fmt.Println();
}
