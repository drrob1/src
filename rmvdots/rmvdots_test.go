package main

/*
   2 May 23 -- I'm going to try and code table based testing function
*/

import "testing"

var testStrings = []struct {
	initFilename string
	newFilename  string
	changed      bool
}{
	{"this_name_has_spaces_and.too.many.dots.txt", "this_name_has_spaces_and-too-many-dots.txt", true},
	{"this name has spaces and right amount of dots.txt", "this name has spaces and right amount of dots.txt", false},
	{"this.name.txt", "this-name.txt", true},
	{"this_name.txt.gpg", "this_name.txt.gpg", false},
	{"this_name.more.txt.gpg", "this_name-more.txt.gpg", true},
	{"this name.txt.gz", "this name.txt.gz", false},
	{"this_name.txt.xz", "this_name.txt.xz", false},
	{"this-name-has-nothing-to-do.txt", "this-name-has-nothing-to-do.txt", false},
	{"this-name-has-nothing-to-do.txt.gpg", "this-name-has-nothing-to-do.txt.gpg", false},
	{"txt", "txt", false},
	{".txt", ".txt", false},
	{".gpg", ".gpg", false},
	{".gz", ".gz", false},
	{".xz", ".xz", false},
}

func TestTooManyDots(t *testing.T) {
	for _, f := range testStrings {
		newFN, changed := tooManyDots(f.initFilename)
		if newFN != f.newFilename {
			t.Errorf(" Initial filename string is %q, and result should have been %q, but it was %q instead.",
				f.initFilename, f.newFilename, newFN)
		}
		if changed != f.changed {
			t.Errorf(" Expected value of changed was %t, but got %t instead.\n", f.changed, changed)
		}
	}
}
