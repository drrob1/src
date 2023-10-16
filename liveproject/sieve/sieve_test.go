package main

import "testing"

var testISqrt = []struct {
	i, sqr int
}{
	{4, 2},
	{5, 2},
	{6, 2},
	{7, 2},
	{8, 3},
	{9, 3},
	{10, 3},
	{15, 3},
	{16, 4},
	{24, 4},
	{25, 5},
	{26, 5},
	{35, 5},
	{36, 6},
	{47, 6},
	{48, 7},
	{49, 7},
	{50, 7},
	{99, 9},
	{100, 10},
	{101, 10},
}

func TestISqrt(t *testing.T) {
	for _, n := range testISqrt {
		sq := isqrt(n.i)
		if sq != n.sqr {
			t.Errorf(" i = %d, square root should have been %d, but was %d instead.\n", n.i, n.sqr, sq)
		}
	}
}
