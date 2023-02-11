package main

import (
	"fmt"
	"runtime"
	"strconv"
)
import format "fmt"

func main() {
	fmt.Printf("Hello World. \n")
	fmt.Println("Hello World line 2.")
	format.Println("Called as format.Println")
	//if n, err := fmt.Printf("Hello World. \n"); err != nil {
	//	os.Exit(1)
	//} else {
	//	fmt.Printf(" Printed %d characters.\n", n)
	//}
	//s := "20.1"
	//n, err := strconv.Atoi(s)
	//if err != nil {
	//	fmt.Printf(" Conversion of 20.1 returned error of %s\n", err)
	//} else {
	//	fmt.Printf(" Conversion of 20.1 returned no error and a result of %d\n", n)
	//}
	//s = "20."
	//n, err = strconv.Atoi(s)
	//if err != nil {
	//	fmt.Printf(" Conversion of 20. returned error of %s\n", err)
	//} else {
	//	fmt.Printf(" Conversion of 20. returned no error and a result of %d\n", n)
	//}
	goVersion := runtime.Version()
	goVersionS1 := goVersion[4:6] // this should be a string of characters 4 and 5, or the numerical digits after Go1.  At the time of writing this, it will be 20.
	goVersionS2 := goVersion[4:]  // this may have a "." and a revision number.
	goVersionInt, err := strconv.Atoi(goVersionS1)
	if err != nil {
		fmt.Printf(" Go version Atoi Err: %s\n", err)
	}
	goVersionFloat, err := strconv.ParseFloat(goVersionS2, 64)
	if err != nil {
		fmt.Printf(" Go version ParseFloat Err: %s\n", err)
	}
	fmt.Printf(" Go version int = %d, float = %f\n", goVersionInt, goVersionFloat)
}
