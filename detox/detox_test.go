package main

/*
   2 May 23 -- I'm going to try and code table based testing function
*/

import (
	"flag"
	"os"
	"testing"
)

var testStrings = []struct {
	initFilename string
	newFilename  string
	changed      bool
}{
	{"this name has spaces and.too.many.dots.txt", "this_name_has_spaces_and-too-many-dots.txt", true},
	{"this name has spaces and right amount of dots.txt", "this_name_has_spaces_and_right_amount_of_dots.txt", true},
	{"this.name.txt", "this-name.txt", true},
	{"this name.txt.gpg", "this_name.txt.gpg", true},
	{"this_name.more.txt.gpg", "this_name-more.txt.gpg", true},
	{"this name.txt.gz", "this_name.txt.gz", true},
	{"this name.txt.xz", "this_name.txt.xz", true},
	{"this-name-has-nothing-to-do.txt", "this-name-has-nothing-to-do.txt", false},
	{"this-name-has-nothing-to-do.txt.gpg", "this-name-has-nothing-to-do.txt.gpg", false},
	{"txt", "txt", false},
	{".txt", ".txt", false},
	{".gpg", ".gpg", false},
	{".gz", ".gz", false},
	{".xz", ".xz", false},
}

func TestMain(m *testing.M) { // this example is in the docs of the testing package, that I was referred to by the golang nuts google group.
	flag.Parse()
	os.Exit(m.Run())
}

func TestDetoxFilenameNewWay(t *testing.T) {
	for _, f := range testStrings {
		newFN, changed := detoxFilenameNewWay(f.initFilename)
		if newFN != f.newFilename {
			t.Errorf(" Initial filename string is %q, and result should have been %q, but it was %q instead.",
				f.initFilename, f.newFilename, newFN)
		}
		if changed != f.changed {
			t.Errorf(" Expected value of changed was %t for %s, but got %t instead.\n", f.changed, f.initFilename, changed)
		}
	}
}
