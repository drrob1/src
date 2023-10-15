package main

import (
	"math"
	"testing"
)

var testPow = []struct {
	a, b int
}{
	{5, 3},
	{3, 5},
	{5, 15},
	{9, 11},
	{11, 13},
	{8, 6},
	{7, 10},
	{9, 13},
	{213, 5},
}

var testPowMod = []struct {
	a, b, mod int
}{
	{8, 6, 10},
	{7, 10, 101},
	{9, 13, 283},
	{213, 5, 1000},
}

func TestFastExp(t *testing.T) {
	for _, n := range testPow {
		fastResult := fastExp(n.a, n.b)
		powResult := math.Pow(float64(n.a), float64(n.b))
		intPowResult := int(powResult)
		if fastResult != intPowResult {
			t.Errorf(" These should be equal but are not.  a=%d, b=%d, fastAns=%d, powAnsF=%.0f, powAns=%d", n.a, n.b, fastResult, powResult, intPowResult)
		}
	}
}

func TestFastExpMod(t *testing.T) {
	for _, n := range testPowMod {
		fastResult := fastExpMod(n.a, n.b, n.mod)
		powResultModF := math.Pow(float64(n.a), float64(n.b))
		powResultModI := int(powResultModF) % n.mod
		if fastResult != powResultModI {
			t.Errorf(" These should be equal but are not.  a=%d, b=%d, mod=%d, fastAnsMod=%d, powAnsModF=%.0f, powAnsModI=%d",
				n.a, n.b, n.mod, fastResult, powResultModF, powResultModI)
		}
	}
}
