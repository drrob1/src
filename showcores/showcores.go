package main

import (
	"fmt"
	"runtime"
)

func main() {
	CPUs := runtime.NumCPU()
	gortn := runtime.NumGoroutine()
	fmt.Println(" NumCPU =", CPUs, ", Go Routines =", gortn)
}
