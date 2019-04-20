package main

import (
	"fmt"
	"os"
)
import format "fmt"

func main() {
	var n int
	var err error // error is a built in type for errors
	fmt.Printf("Hello World. \n")
	fmt.Println("Hello World line 2.")
	format.Println("Called as format.Println")
	if n, err = fmt.Printf("Hello World. \n"); err != nil { // note that this combines 3 separate operations
		os.Exit(1)
	}
	fmt.Printf(" Printed %d characters.\n", n)

}
