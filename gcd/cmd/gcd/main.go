package main

import (
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"src/gcd"
	"src/misc"
)

// 7 Jul 24 -- Now to really test if the go test system caches random numbers, but compiling using go install does not.
//				Yep, it works here when compiled w/ go install.  I don't use go run much.

func main() {
	for range 10 {
		i := misc.RandRange(100, 20_000)
		j := misc.RandRange(100, 20_000)
		if gcd.HCF(i, j) == gcd.GCD(i, j) {
			ctfmt.Printf(ct.Green, true, " GCD(%d, %d) = %d\n", i, j, gcd.GCD(i, j))
		} else {
			ctfmt.Printf(ct.Red, true, " GCD(%d, %d) = %d\n", i, j, gcd.GCD(i, j))
		}
	}
}
