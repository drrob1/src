package main

import (
	"fmt"
	"unsafe"
)

func main() {
	var pi *int

	fmt.Println("vim-go")
	i := 10
	pi = &i
	p := unsafe.Pointer(&i)
	intPtr := (*[4]byte)(p)
	fmt.Println(" i is", i, "p is", p, "deref p is not allowed. pi is", pi, "intPtr is", intPtr)
	fmt.Println(" deref pi is", *pi)
}
