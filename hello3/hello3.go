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
	n, err = fmt.Printf("Hello World take 2! \n")
	if err != nil {
		os.Exit(1)
	}
	fmt.Printf(" Printed %d characters.\n", n)

}
