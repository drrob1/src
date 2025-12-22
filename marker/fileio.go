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
  21 Dec 25 -- Junie AI generated the following code, w/ some input from me to get it right.  In 2 places it used variables before it defined them.  I moved the definitions up in the code.
*/

// fileEntry is an item in the list
type fileEntry struct {
	name  string
	uri   fyne.URI
	isDir bool
}

// NewOpenFileDialogWithPrefix shows a custom Open dialog that lists only files
// whose base name starts with the provided prefix (case-insensitive).
// Pass an empty prefix to show all files. Optionally, you can pass extensions
// to restrict by type (e.g., []string{".png", ".jpg"}); pass nil or empty to ignore.
func NewOpenFileDialogWithPrefix(parent fyne.Window, prefix string, exts []string, onOpen func(fyne.URI)) {
	var items []fileEntry
	var selected *fileEntry
	var curURI fyne.URI

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

	workingDir, err := os.Getwd()
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	curURI, err = listableFromPath(workingDir)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	fmt.Printf("workingDir is %s, current URI path is %s, URI name is %s\n\n", workingDir, curURI.Path(), curURI.Name())
	dialog.ShowInformation("current URI path", curURI.Path(), parent)

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
	dlg.Resize(fyne.NewSize(width, height))

	refreshList()
	dlg.Show()
}

/*
Usage in your app (e.g., in highlight.go):
openBtn := widget.NewButton("Open…", func() {
	// Example: only show files whose base name starts with "img_" and are PNG/JPG
	NewOpenFileDialogWithPrefix(w, "img_", []string{".png", ".jpg", ".jpeg"}, func(u fyne.URI) {
		// handle selection; for example, open via storage.OpenFileFromURI
		f, err := storage.OpenFileFromURI(u)
		if err != nil { dialog.ShowError(err, w); return }
		defer f.Close()
		// ... decode image, etc.
	})
})
*/
