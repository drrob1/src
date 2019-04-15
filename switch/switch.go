package main

import (
	"fmt"
	"os"
)
import format "fmt"

func main() {
	MyString := " String test for output."
	fmt.Printf("Hello World. \n")
	fmt.Println("Hello World line 2.")
	format.Println("Called as format.Println")
	n, err := fmt.Printf("%s", MyString)
	switch {
	case err != nil:
		os.Exit(1)
	case n == 0:
		fmt.Println("\n No characters printed.  Don't know why.")
	case n != len(MyString):
		fmt.Printf("\n wrong count of characters output.  Don't know why. n=%d, len=%d \n", n, len(MyString))
	default:
		format.Printf("\nok.  Output %d number of characters", n)
	}

	format.Println("")

}
