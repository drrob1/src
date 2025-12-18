package main

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
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

func (f nameFilterType) Matches(u fyne.URI) bool { //  I'm going to add check against a directory
	// base name without directories
	name := u.Name()

	isDir, _ := storage.CanList(u) // this doesn't prevent the directories from being populated also
	if isDir {
		return false
	}

	// for case-insensitive substring match:
	return strings.Contains(
		strings.ToLower(name),
		strings.ToLower(f.search),
	)
}
