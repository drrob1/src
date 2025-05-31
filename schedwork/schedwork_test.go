package main

import "testing"

func TestRemoveExt(t *testing.T) {
	var testString = []struct {
		in  string
		out string
	}{
		{"test.csv", "test"},
		{"test", "test"},
		{"", ""},
		{"schedule.csv", "schedule"},
	}

	for _, test := range testString {
		if test.out != removeExt(test.in) {
			t.Errorf("removeExt(%s) = %s, want %s", test.in, removeExt(test.in), test.out)
		}
	}
}

func TestIsDate(t *testing.T) {
	var testString = []struct {
		in  string
		out bool
	}{
		{"01/01/2020", true},
		{"01/01/20", true},
		{"1/1/2020", true},
		{"1/1/20", true},
		{"01/01/2020", true},
		{"green", false},
		{"", false},
	}

	for _, test := range testString {
		if test.out != isDate(test.in) {
			t.Errorf("isDate(%s) = %t, want %t", test.in, isDate(test.in), test.out)
		}
	}
}
