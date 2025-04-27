package main

import "testing"

func TestTryDateMDY(t *testing.T) {
	var testString = []struct {
		input  string
		output string
	}{
		{"3-4-25", "2025-03-04"},
		{"3-4-2025", "2025-03-04"},
		{"1/2/25", "2025-01-02"},
		{"1/2/2025", "2025-01-02"},
		{"1.2.2025", "2025-01-02"},
		{"1.2.25", "2025-01-02"},
		{"10.25.25", "2025-10-25"},
		{"10.25.2025", "2025-10-25"},
		{"25.10.2025", ""},
		{"11;11;25", ""},
		{"12.25/25", ""},
	}

	for _, sampletest := range testString {
		result := tryDateMDY(sampletest.input)
		if result != sampletest.output {
			t.Errorf("tryDateMDY(%s) = %s, want %s", sampletest.input, result, sampletest.output)
		}
	}
}
