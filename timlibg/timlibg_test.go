package timlibg

import (
	"fmt"
	"testing"
)

func TestJulian(t *testing.T) { // these all pass
	var tests = []struct {
		m, d, y int
		julian  int
	}{
		{1, 1, 1970, 719163},
		{11, 29, 25, 739584},
		{11, 29, 2025, 739584},
		{1, 1, 2025, 739252},
		{1, 1, 25, 739252},
		{12, 1, 2025, 739586},
		{12, 1, 25, 739586},
	}
	fmt.Printf("Testing Julian() with %d tests\n", len(tests))

	for _, test := range tests {
		julian := JULIAN(test.m, test.d, test.y)
		if julian != test.julian {
			t.Errorf("Julian(%d, %d, %d) = %d, want %d", test.m, test.d, test.y, julian, test.julian)
		}
	}
}
