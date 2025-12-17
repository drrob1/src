package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

/*
  17 Dec 25 -- I separated out the AI-generated code for a custom menu to open files, so I can try to understand it better.

*/

// fileEntry is an item in the list
type fileEntry struct { // used by the custom file dialog
	name  string
	uri   fyne.URI
	isDir bool
}

// nameMatch returns true if the base name passes your rule.
// Replace this with your own logic (contains, equals, regex, etc.)
func nameMatch(base string, query string) bool {
	// Example: case-insensitive substring match against base name (without ext)
	b := strings.TrimSuffix(base, filepath.Ext(base)) // since I'm using strings.Contains, I don't have to strip off the extension.  Maybe it's faster to do this?
	return strings.Contains(strings.ToLower(b), strings.ToLower(query))
}

func refreshLst(dir fyne.URI, nameQuery string) ([]fileEntry, *widget.Label, *widget.List) { // I'm rewriting it to use return params
	var out []fileEntry
	var pathLabel *widget.Label
	var list *widget.List

	lister, err := storage.ListerForURI(dir)
	if err != nil {
		pathLabel.SetText(fmt.Sprintf("Error: %v", err))
		list.Refresh()
		return out, pathLabel, list
	}
	children, err := lister.List()
	if err != nil {
		pathLabel.SetText(fmt.Sprintf("Error: %v", err))
		list.Refresh()
		return out, pathLabel, list
	}

	var dirs, files []fileEntry // I don't want the directories.  I won't use this.
	for _, u := range children {
		isDir, _ := storage.CanList(u)
		base := filepath.Base(u.Path())
		if isDir {
			dirs = append(dirs, fileEntry{name: base, uri: u, isDir: true})
			continue
		}
		if nameQuery == "" || nameMatch(base, nameQuery) {
			files = append(files, fileEntry{name: base, uri: u, isDir: false})
		}
	}
	out = append(dirs, files...)
	pathLabel.SetText(dir.Path())
	list.Refresh()
	return out, pathLabel, list
}

// NewOpenFileDialogWithPrefix shows a custom Open dialog that lists only files
// whose base name starts with the provided prefix (case-insensitive).
// Pass an empty prefix to show all files. Optionally, you can pass extensions
// to restrict by type (e.g., []string{".png", ".jpg"}); pass nil or empty to ignore.
func NewOpenFileDialogWithPrefix(parent fyne.Window, prefix string, exts []string, onOpen func(fyne.URI)) {
	var (
		items    []fileEntry
		selected *fileEntry
		curURI   fyne.URI
	)

	startsWith := func(base string) bool {
		if prefix == "" {
			return true
		}
		b := strings.TrimSuffix(base, filepath.Ext(base))
		return strings.HasPrefix(strings.ToLower(b), strings.ToLower(prefix))
	}

	extAllowed := func(base string) bool {
		if len(exts) == 0 {
			return true
		}
		for _, e := range exts {
			if strings.EqualFold(filepath.Ext(base), e) {
				return true
			}
		}
		return false
	}
	pathLabel := widget.NewLabel("")

	list := widget.NewList(
		func() int { return len(items) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			lbl := o.(*widget.Label)
			it := items[i]
			if it.isDir {
				lbl.SetText("[" + it.name + "]")
			} else {
				lbl.SetText(it.name)
			}
		},
	)

	refreshList := func() {
		items = items[:0]
		lister, err := storage.ListerForURI(curURI)
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}
		children, err := lister.List()
		if err != nil {
			dialog.ShowError(err, parent)
			return
		}
		var dirs, files []fileEntry
		for _, u := range children {
			isDir, _ := storage.CanList(u)
			base := filepath.Base(u.Path())
			if isDir {
				dirs = append(dirs, fileEntry{name: base, uri: u, isDir: true})
				continue
			}
			if startsWith(base) && extAllowed(base) {
				files = append(files, fileEntry{name: base, uri: u, isDir: false})
			}
		}
		items = append(dirs, files...)
		list.Refresh()
		pathLabel.SetText(curURI.Path())
	}

	// pick a sensible starting directory (caller could set something else)
	if home, err := os.UserHomeDir(); err == nil {
		curURI = storage.NewFileURI(home)
	} else {
		curURI = storage.NewFileURI(".")
	}

	openBtn := widget.NewButton("Open", func() {
		if selected != nil && !selected.isDir {
			onOpen(selected.uri)
		}
	})
	openBtn.Disable()

	upBtn := widget.NewButton("Up", func() {
		p := filepath.Dir(curURI.Path())
		if p == curURI.Path() {
			return
		}
		curURI = storage.NewFileURI(p)
		selected = nil
		openBtn.Disable()
		refreshList()
	})
	list.OnSelected = func(i widget.ListItemID) {
		if i < 0 || i >= len(items) {
			return
		}
		it := items[i]
		if it.isDir {
			curURI = it.uri
			selected = nil
			openBtn.Disable()
			refreshList()
			list.UnselectAll()
			return
		}
		selected = &items[i]
		openBtn.Enable()
	}

	toolbar := container.NewHBox(upBtn)
	content := container.NewBorder(
		container.NewVBox(pathLabel, toolbar),
		container.NewHBox(openBtn),
		nil, nil,
		list,
	)

	dlg := dialog.NewCustomConfirm("Open File", "Open", "Cancel", content, func(ok bool) {
		if ok && selected != nil && !selected.isDir {
			onOpen(selected.uri)
		}
	}, parent)

	refreshList()
	dlg.Show()
}
