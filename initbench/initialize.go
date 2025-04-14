package main

import "fmt"

/*
14 Apr 25 -- From Mastering Go, 4th ed.
*/

func InitSliceNew(n int) []int {
	s := make([]int, n)
	for i := 0; i < n; i++ {
		s[i] = i
	}
	return s
}

func InitSliceAppend(n int) []int {
	s := make([]int, 0, n) // this is my edit
	for i := 0; i < n; i++ {
		s = append(s, i)
	}
	return s
}

func main() {
	fmt.Println(InitSliceNew(10))
	fmt.Println(InitSliceAppend(10))
}
