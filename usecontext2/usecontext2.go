/*
15 Feb 25 -- I'm really trying to understand this.  So far, I think I've figured out that when the timeout or deadline passes, a done message is sent on the channel of that context.
             But the goroutine runs to completion here.  That's due to the fact that they're not really doing anything, and no checking of the channel is done within the goroutines.
----------------------------------------------------------------------------------------------------
16 Feb 25 -- Now called usecontext2.go.  I want to explore canceling a goroutine.  The book version of this routine does not do that.  I'll create a channel that will send
             something after 4 sec; the f2 and f3 will wait for this channel in a goroutine, and f1 will use the After func to timeout.  We'll see if I can figure this out.
			Looks like the context must be used within the goroutine it's needed for, else it doesn't work as expected.  IE, the timeout or deadline was never triggered.
			When I created the contexts in the goroutines, then they would timeout when the time given was 1 sec, but not for 2+ sec.  Somewhat strange when the work is set for 4 sec,
			but I'm stopping now.
*/

package main

import (
	"context"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const gosleep = 4 * time.Second

var wg sync.WaitGroup
var boolChan1 = make(chan bool, 1)
var boolChan2 = make(chan bool, 1)
var boolChan3 = make(chan bool, 1)

func f1(t int) { // no timeout or deadline
	// WithCancel returns a copy of parent context with a new Done channel
	_, cancel := context.WithCancel(context.Background())
	defer func() {
		wg.Done()
		cancel()
	}()

	wg.Add(1)
	go func() { // still runs even after f1 exits
		fmt.Printf(" f1 work started\n")
		select {
		case <-boolChan1:
			fmt.Printf(" f1 work finished\n")
			return
		case r := <-time.After(time.Duration(t) * time.Second):
			fmt.Println("f1() time.After triggered, returning:", r)
			return
		}
	}()
}

func f2(t int) { // creates a timeout
	wg.Add(1)
	c2 := context.Background()
	go func() { // still runs even after f2 exits
		defer wg.Done()
		c2, cancel := context.WithTimeout(c2, time.Duration(t)*time.Second)
		defer cancel()
		fmt.Printf(" f2 work started\n")
		select {
		case <-boolChan2:
			ctfmt.Printf(ct.Green, false, " f2 work finished\n")
			return
		case <-c2.Done():
			fmt.Println("f2() timeout, returning:", c2.Err())
			return
		case r := <-time.After(time.Duration(t+1) * time.Second): // want to see if the timeout ever triggers the cancel fcn.
			fmt.Println("f2() time.After triggered, returning:", r)
			return
		}
	}()
}

func f3(t int) { // creates a deadline
	wg.Add(1)
	c3 := context.Background()
	go func() { // still runs even after f3 exits
		defer wg.Done()
		deadline := time.Now().Add(time.Duration(t) * time.Second)
		c3, cancel := context.WithDeadline(c3, deadline)
		defer cancel()
		fmt.Printf(" f3 work started\n")
		select {
		case <-boolChan3:
			ctfmt.Printf(ct.Cyan, true, " f3 work finished\n")
			return
		case <-c3.Done():
			fmt.Println("f3() deadline reached, returning:", c3.Err())
			return
		case r := <-time.After(time.Duration(t+1) * time.Second): // want to see if the deadline ever triggers the cancel fcn.
			fmt.Println("f3() time.After reached:", r)
			return
		}
	}()
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Need a delay in seconds.  Work takes ~3 seconds.")
		return
	}
	executable, err := os.Executable()
	if err != nil {
		fmt.Printf("Cannot get executable path: %s\n", err)
		return
	}
	ExecFI, _ := os.Stat(executable)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")

	delay, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf(" useContext: delay of %d seconds entered, compiled w/ %s on %s using %s\n", delay, runtime.Version(), ExecTimeStamp, runtime.Compiler)

	t0 := time.Now()
	go func() { // every 4 sec send a value down the boolChan
		for {
			fmt.Printf(" sleeping for %s before sending down the boolChans\n", gosleep.String())
			time.Sleep(gosleep)
			boolChan1 <- true
			boolChan2 <- true
			boolChan3 <- true
			fmt.Printf(" sent true down the channels after %s elapsed.\n", time.Since(t0))
		}
	}()

	time.Sleep(2 * time.Second)
	f1(delay)
	f2(delay)
	f3(delay)

	ctfmt.Printf(ct.Magenta, false, " There are %d goroutines running after f3 just before wg.Wait.\n", runtime.NumGoroutine())

	wg.Wait()
	m := max(delay, 5)
	time.Sleep(time.Duration(m) * time.Second)
	ctfmt.Printf(ct.Yellow, false, " There are %d goroutines running after wg.Wait, elapsed %s.\n", runtime.NumGoroutine(), time.Since(t0))
}
