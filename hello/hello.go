package main

import (
	"fmt"
	"os"
)
import format "fmt"

func main() {
	fmt.Printf("Hello World. \n")
	fmt.Println("Hello World line 2.")
	format.Println("Called as format.Println")
	if n, err := fmt.Printf("Hello World. \n"); err != nil {
		os.Exit(1)
	} else {
		fmt.Printf(" Printed %d characters.\n", n)
	}
}
