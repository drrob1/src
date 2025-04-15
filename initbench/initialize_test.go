package main

import "testing"

/*
13 Apr 25 -- From Chapter 13 of Mastering Go, 4th Ed.
*/

var t []int

func BenchmarkNew(b *testing.B) { // Win11 Desktop.  ~119 μs/op
	for i := 0; i < b.N; i++ {
		t = InitSliceNew(i)
	}
}

func BenchmarkAppend(b *testing.B) { // Win11 Desktop.  ~143 μs/op
	for i := 0; i < b.N; i++ {
		t = InitSliceAppend(i)
	}
}
