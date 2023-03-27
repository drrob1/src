package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/term"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"
)

/*
REVISION HISTORY
======== =======
26 Mar 23 -- First copied here from Packtpub video course on concurrency in Go.  The course was released Dec 2021.  I want to add something about filling the screen or half screen w/ bad jokes.

*/

const LastAltered = "26 Mar 2023"
const defaultHeight = 40

type Response struct {
	ID     string `json:"id"`
	Joke   string `json:"joke"`
	Status int    `json:"status"`
}

func main() {
	execName, _ := os.Executable()
	execFI, _ := os.Stat(execName)
	lastLinkedTimeStamp := execFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")

	autoWidth, autoHeight, err := term.GetSize(int(os.Stdout.Fd())) // this now works on Windows, too
	if err != nil {
		autoHeight = defaultHeight
		//autoWidth = minWidth
	}
	_ = autoWidth
	fmt.Printf(" %s, last altered %s, compiled by %s, last linked on %s, autoheight = %d\n", os.Args[0],
		LastAltered, runtime.Version(), lastLinkedTimeStamp, autoHeight)

	start := time.Now()
	for i := 1; i < autoHeight/3; i++ {
		client := &http.Client{}
		request, err := http.NewRequest("GET", "https://icanhazdadjoke.com/", nil)
		if err != nil {
			fmt.Print(err.Error())
		}
		request.Header.Add("Accept", "application/json")
		request.Header.Add("Content-Type", "application/json")
		response, err := client.Do(request)

		if err != nil {
			fmt.Print(err.Error())
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				return
			}
		}(response.Body)

		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Print(err.Error())
		}
		var responseObject Response
		err = json.Unmarshal(bodyBytes, &responseObject)
		if err != nil {
			return
		}
		fmt.Println("\n", responseObject.Joke)

	}
	elapsed := time.Since(start)
	fmt.Printf("Processes took %s", elapsed)
}
