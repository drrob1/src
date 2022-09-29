package scanln

import (
	"fmt"
	"time"
)

/*
REVISION HISTORY
-------- -------
24 Sep 22 -- First version.
28 Sep 22 -- Expanded on the comment in WithTimeOut below.  And fixed an error in logic.
29 Sep 22 -- Added WithDuration
*/

const lastAltered = "Sep 29, 2022"
const maxTimeout = 10

// WithTimeout (timeOut int) string -- timeOut is in seconds.
// If it times out or <enter> is hit before the timeout, then it will return an empty string.
func WithTimeout(timeOut int) string {
	var ans string
	// Note that the buffer size of 1 is necessary to avoid deadlock of goroutines and guarantee garbage collection of the timeout channel.
	// On closer inspection, the go routine will send 2 strings down the channel if there's a timeout or I hit enter.  But there's only one read from the channel, so the
	// extra string sits in the buffer and is garbage collected when this routine returns to the caller.
	// Nevermind, I fixed this by using an else clause.
	strChannel := make(chan string, 1)
	defer close(strChannel)
	ticker := time.NewTicker(1 * time.Second)
	if timeOut > maxTimeout {
		timeOut = maxTimeout
	}
	ticks := timeOut
	go func() {
		n, err := fmt.Scanln(&ans)
		if n == 0 || err != nil {
			strChannel <- ""
		} else {
			strChannel <- ans
		}
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

// WithTimeoutAndPrompt (prompt string, timeOut int) string
// This uses Printf for the prompt, and then calls WithTimeout.
func WithTimeoutAndPrompt(prompt string, timeOut int) string {
	if len(prompt) > 0 {
		fmt.Printf(" %s \n", prompt)
	}
	return WithTimeout(timeOut)
}

// WithDuration (d Duration) string -- Duration from the time package.  Used in the time.After function call.
// If it times out or <enter> is hit before the timeout, then it will return an empty string.
func WithDuration(d time.Duration) string {
	var ans string
	// Note that the buffer size of 1 is necessary to avoid deadlock of goroutines and guarantee garbage collection of the timeout channel.
	// On closer inspection, the go routine will send 2 strings down the channel if there's a timeout or I hit enter.  But there's only one read from the channel, so the
	// extra string sits in the buffer and is garbage collected when this routine returns to the caller.
	// Nevermind, I fixed this by using an else clause.
	strChannel := make(chan string, 1)
	defer close(strChannel)
	if d > maxTimeout*time.Second {
		d = maxTimeout * time.Second
	}

	go func() {
		//n, err := fmt.Scanln(&ans)
		fmt.Scanln(&ans) // ignore number of characters entered or any errors
		strChannel <- ans
	}()

	for {
		select {
		case <-time.After(d): // this channel returns the current time at time of firing, but here that's ignored and discarded.
			return ""

		case s := <-strChannel:
			return s
		}
	}
} // WithDuration

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
