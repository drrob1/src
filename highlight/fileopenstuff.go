package main

import (
	"strings"

	"fyne.io/fyne/v2"
)

/*
  17 Dec 25 -- I separated out the AI-generated code for a custom menu to open files, so I can try to understand it better.
				I still don't understand it well.  So I asked perplexity.  I think I understand it now.  It's all about the SetFilter function that must return a bool.
*/

type FileFilterI interface {
	Matches(fyne.URI) bool
}

type nameFilterType struct {
	search string
}

func (f nameFilterType) Matches(u fyne.URI) bool { // why does this need to be a pointer?
	// base name without directories
	name := u.Name()

	// for case-insensitive substring match:
	return strings.Contains(
		strings.ToLower(name),
		strings.ToLower(f.search),
	)
}
