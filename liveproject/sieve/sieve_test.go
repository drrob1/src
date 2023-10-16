package main

import "testing"

var testISqrt = []struct {
	i, sqr int
}{
	{4, 3},
	{5, 3},
	{6, 3},
	{7, 3},
	{8, 4},
	{9, 4},
	{10, 4},
	{15, 4},
	{16, 5},
	{24, 5},
	{25, 6},
	{26, 6},
	{35, 6},
	{36, 7},
	{47, 7},
	{48, 8},
	{49, 8},
	{50, 8},
	{99, 10},
	{100, 11},
	{101, 11},
}

func TestISqrt(t *testing.T) {
	for _, n := range testISqrt {
		sq := iSqrt(n.i)
		if sq != n.sqr {
			t.Errorf(" i = %d, square root should have been %d, but was %d instead.\n", n.i, n.sqr, sq)
		}
	}
}
