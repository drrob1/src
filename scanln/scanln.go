package scanln

import (
	"fmt"
	"time"
)

/*
REVISION HISTORY
-------- -------
24 Sep 22 -- First version.
*/

const lastAltered = "Sep 24, 2022"
const maxTimeout = 10

// WithTimeout (prompt string, timeOut int) string
// If it times out, or <enter> is hit before the timeout, then it will return an empty string.
func WithTimeout(prompt string, timeOut int) string {
	var ans string
	strChannel := make(chan string, 1) // Note that the buffer size of 1 is necessary to avoid deadlock of goroutines and guarantee garbage collection of the timeout channel.
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

/*
func main() {
	var err error
	fmt.Printf(" scanlineWithTimeout test last altered %s, len(os.Args) = %d.\n", lastAltered, len(os.Args))
	timeout := 0
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
		} else {
			timeout, err = strconv.Atoi(toutStr)
			if err != nil {
				fmt.Printf(" Error from Atoi call is %v, timeout set to max of %d.\n", err, maxTimeout)
				timeout = maxTimeout
			}
		}
		fmt.Printf(" Entered %s, timeout = %d\n", toutStr, timeout)
	}

	returnedString := scanlnWithTimeout("enter something before it times out", timeout)
	fmt.Printf(" returnedString is %q\n", returnedString)
}

*/
