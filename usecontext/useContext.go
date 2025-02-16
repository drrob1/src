/*
I'm really trying to understand this.  So far, I think I've figured out that when the timeout or deadline passes, a done message is sent on the channel of that context.
But the goroutine runs to completion here.  That's due to the fact that they're not really doing anything, and no checking of the channel is done within the goroutines.
*/

package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// The f1 function creates and executes a goroutine
// The time.Sleep() call simulates the time it would take a real goroutine
// to do its job - in this case it is 4 seconds. If the c1 context calls
// the Done() function in less than 4 seconds, the goroutine will not have
// enough time to finish.

var wg sync.WaitGroup

func f1(t int) { // no timeout or deadline
	c1 := context.Background()
	// WithCancel returns a copy of parent context with a new Done channel
	c1, cancel := context.WithCancel(c1)
	defer func() {
		wg.Done()
		cancel()
	}()

	wg.Add(1)
	go func() { // still runs even after f1 exits
		fmt.Printf(" f1 work started\n")
		time.Sleep(4 * time.Second)
		fmt.Printf(" f1 work finished\n")
		cancel()
	}()

	select {
	case <-c1.Done():
		fmt.Println("f1() Done called:", c1.Err()) // here, there's no way to get a timeout
		return
	case r := <-time.After(time.Duration(t+1) * time.Second):
		fmt.Println("f1() time.After:", r)
		cancel()
	}
	return
}

func f2(t int) { // creates a timeout
	c2 := context.Background()
	c2, cancel := context.WithTimeout(c2, time.Duration(t)*time.Second)
	defer func() {
		wg.Done()
		cancel()
	}()

	wg.Add(1)
	go func() { // still runs even after f2 exits
		fmt.Printf(" f2 work started\n")
		time.Sleep(4 * time.Second)
		fmt.Printf(" f2 work finished\n")
		cancel()
	}()

	select {
	case <-c2.Done():
		fmt.Println("f2() Done called:", c2.Err())
		cancel()
		return
	case r := <-time.After(time.Duration(t+1) * time.Second): // want to see if the timeout ever triggers the cancel fcn.
		fmt.Println("f2() time.After:", r)
		cancel()
	}
	return
}

func f3(t int) { // creates a deadline
	c3 := context.Background()
	deadline := time.Now().Add(time.Duration(t) * time.Second)
	c3, cancel := context.WithDeadline(c3, deadline)
	defer func() {
		wg.Done()
		cancel()
	}()

	wg.Add(1)
	go func() { // still runs even after f3 exits
		fmt.Printf(" f3 work started\n")
		time.Sleep(4 * time.Second)
		fmt.Printf(" f3 work finished\n")
		cancel()
	}()

	select {
	case <-c3.Done():
		fmt.Println("f3() Done called:", c3.Err())
		cancel()
		return
	case r := <-time.After(time.Duration(t+1) * time.Second): // want to see if the deadline ever triggers the cancel fcn.
		fmt.Println("f3() time.After:", r)
		cancel()
	}
	return
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Need a delay in seconds.  Work takes 4 seconds.")
		return
	}
	executable, err := os.Executable()
	if err != nil {
		fmt.Printf("Cannot get executable path: %s\n", err)
		return
	}
	ExecFI, _ := os.Stat(executable)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")

	fmt.Printf(" useContext: %s seconds entered, compiled w/ %s on %s \n", os.Args[1], runtime.Version(), ExecTimeStamp)

	delay, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Delay:", delay)

	f1(delay)
	f2(delay)
	f3(delay)

	time.Sleep(time.Duration(delay) * time.Second)
	wg.Wait()
	m := min(delay, 4)
	time.Sleep(time.Duration(m) * time.Second)
}
