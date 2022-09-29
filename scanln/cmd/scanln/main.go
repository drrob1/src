package main

import (
	"fmt"
	"os"
	"src/scanln"
	"strconv"
	"time"
)

/*
REVISION HISTORY
-------- -------
24 Sep 22 -- First version.
29 Sep 22 -- Need to test WithDuration.
*/

const lastAltered = "Sep 29, 2022"
const maxTimeout = 10

/*
// WithTimeout (prompt string, timeOut int) string
// If <enter> is hit before the timeout, then it will return an empty string.  If the timeout is reached before anything is entered then it returns a nil string.
func WithTimeout(prompt string, timeOut int) string {
	var ans string
	strChannel := make(chan string, 1)
	defer close(strChannel)
	ticker := time.NewTicker(1 * time.Second)
	if timeOut > maxTimeout {
		timeOut = maxTimeout
	}
	ticks := timeOut
	go func() {
		fmt.Printf(" %s \n", prompt)
		n, err := fmt.Scanln(&ans)
		if n == 0 || err != nil {
			strChannel <- ""
		}
		strChannel <- ans
	}()

	for {
		select {
		case <-ticker.C:
			if ticks < 1 {
				ticker.Stop()
				return ""
			}
			fmt.Printf("\r %d second(s)   ", ticks)
			ticks--

		case s := <-strChannel:
			return s
		}
	}
} // scanlnWithTimeout
*/

func main() {
	fmt.Printf(" scanln test last altered %s, len(os.Args) = %d.\n", lastAltered, len(os.Args))
	var err error
	var timeout int
	if len(os.Args) > 1 { // param is entered
		timeout, err = strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Printf(" Error from Atoi call is %v, timeout set to max of %d.\n", err, maxTimeout)
			timeout = maxTimeout
		}
	} else {
		var toutStr string // abbreviation for timeout string
		fmt.Printf(" Enter value for the timeout: ")
		n, er := fmt.Scanln(&toutStr)
		if n == 0 || er != nil {
			timeout = maxTimeout
			// fmt.Printf(" Hit <enter> or timed out, answer = %q\n", toutStr)  I don't need this since I show what was returned anyway.
		} else {
			timeout, err = strconv.Atoi(toutStr)
			if err != nil {
				fmt.Printf(" Error from Atoi call is %v, timeout set to max of %d.\n", err, maxTimeout)
				timeout = maxTimeout
			}
		}
		fmt.Printf(" Entered %q, timeout = %d\n", toutStr, timeout)
	}

	returnedString := scanln.WithTimeoutAndPrompt("enter something before it times out", timeout)
	fmt.Printf(" returnedString is %q\n", returnedString)

	returnedString = scanln.WithDuration(time.Duration(timeout) * time.Second)
	fmt.Printf(" After WithDuration and returnedString is %q\n", returnedString)
}
