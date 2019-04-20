package main

import (
	"fmt"
)

type ByteSize float64

// This works because 10 bits is KB, 20 bits is MB, etc.  And this is a shift op, not mult op, so it
// automatically multiplies by the respective power of 2.

const (
	_           = iota
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

func main() {
	fmt.Println(" KB ", KB, ", MB ", MB, ", GB ", GB, ", TB ", TB, ", PB ", PB)
	fmt.Println(" EB ", EB, ", ZB ", ZB, ", YB ", YB)
}
