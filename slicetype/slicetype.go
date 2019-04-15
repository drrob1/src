package main

import (
	"fmt"
	//	"net/http"
	//	"io/ioutil"
)

type sumableSlice []int

func (s sumableSlice) sum() int {
	var sum int

	for _, i := range s {
		sum += i
	}
	return sum
}

func main() {
	var s sumableSlice = sumableSlice{1, 2, 3, 4, 5, 6}

	fmt.Println("Sum of 1..6 is", s.sum())
	fmt.Println()
}
