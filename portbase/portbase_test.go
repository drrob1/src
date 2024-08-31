package main

import "testing"

// From within the portbase directory: go test -v .
// From the src directory: go test -v ./portbase\
// The full directory is first compiled before the test is run, so it took ~10 sec and I heard the fan spin up on Win11 Desktop.  But the code works, so there's that.

var testSymbols = []struct {
	input  string
	output string
}{
	{"abc, def, ghi, klm", "abc,def,ghi,klm"},
	{"nop, qrs, tuv, wxyz", "nop,qrs,tuv,wxyz"},
}

func TestRemoveAllSpaces(t *testing.T) {
	for _, tt := range testSymbols {
		sym := removeAllSpaces(tt.input)
		if sym != tt.output {
			t.Errorf("removeAllSpaces(%q) = %q, want %q", tt.input, sym, tt.output)
		}
	}
}
