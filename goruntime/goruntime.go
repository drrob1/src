/*
  From Mastering Concurrency in Go
*/

package main

import (
	"fmt"
	"runtime"
)

func listThreads() int {

	threads := runtime.GOMAXPROCS(0)
	return threads
}

func main() {
	threads := runtime.GOMAXPROCS(0)
	cpus := runtime.NumCPU()
	numofgortns := runtime.NumGoroutine()

	//  fmt.Printf("%d thread(s) available to Go.", listThreads())
	fmt.Println()
	fmt.Println(" Threads=", threads, ", CPUs=", cpus, ", Number of Go routines=", numofgortns)
	// answers were 12, 12, and 1 when I ran this 3/24/20.
	fmt.Println()
	fmt.Println()

}
