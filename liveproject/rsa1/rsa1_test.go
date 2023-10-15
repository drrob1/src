package main

import "testing"

var testGCD = []struct {
	a, b, r int
}{
	{5, 3, 1},
	{3, 5, 1},
	{5, 15, 5},
	{11, 33, 11},
	{9, 11, 1},
	{11, 13, 1},
}

var testLCM = []struct {
	a, b, m int
}{
	{5, 3, 15},
	{15, 5, 15},
	{33, 11, 33},
	{11, 9, 99},
	{13, 11, 11 * 13},
}

func TestGCD(t *testing.T) {
	for _, n := range testGCD {
		r := gcd(n.a, n.b)
		if r != n.r {
			t.Errorf(" GCD of %d and %d should have been %d, but was %d instead.\n", n.a, n.b, n.r, r)
		}
		r = hcf(n.a, n.b)
		if r != n.r {
			t.Errorf(" HCF of %d and %d should have been %d, but was %d instead.\n", n.a, n.b, n.r, r)
		}
	}
}

func TestLCM(t *testing.T) {
	for _, n := range testLCM {
		multiple := lcm(n.a, n.b)
		if multiple != n.m {
			t.Errorf(" LCM of %d and %d should have been %d, but was %d instead.\n", n.a, n.b, n.m, multiple)
		}
	}
}
