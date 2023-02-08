package main

import (
	"fmt"
	"os"
	"strconv"
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
	s := "20.1"
	n, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf(" Conversion of 20.1 returned error of %s\n", err)
	} else {
		fmt.Printf(" Conversion of 20.1 returned no error and a result of %d\n", n)
	}
	s = "20."
	n, err = strconv.Atoi(s)
	if err != nil {
		fmt.Printf(" Conversion of 20. returned error of %s\n", err)
	} else {
		fmt.Printf(" Conversion of 20. returned no error and a result of %d\n", n)
	}
}
