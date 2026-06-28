package main

import (
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
	curURI = storage.NewFileURI(workingDir)

	//fmt.Printf("workingDir is %s, current URI path is %s, URI name is %s\n\n", workingDir, curURI.Path(), curURI.Name())  Don't need this anymore, as I've found the issue.
	//dialog.ShowInformation("current URI path", curURI.Path(), parent)  My curURI was being reassigned in code below, that I've commented out.

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

	// pick a sensible starting directory (caller could set something else)  I've set this above to the current working directory.
	//if home, err := os.UserHomeDir(); err == nil {
	//	curURI = storage.NewFileURI(home)
	//} else {
	//	curURI = storage.NewFileURI(".")
	//}

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

/*

This code defines a custom Fyne file-open dialog that filters what the user sees before they pick a file.

What it does, in order:

1. Defines `fileEntry`
   - A small struct for one row in the list.
   - It stores:
     - `name`: display name
     - `uri`: the file or directory URI
     - `isDir`: whether the entry is a directory

2. Defines `NewOpenFileDialogWithPrefix(...)`
   - Inputs:
     - `parent`: the window to attach the dialog to
     - `prefix`: only show files whose base name starts with this text
     - `exts`: optional allowed extensions like `[]string{".png", ".jpg"}`
     - `onOpen`: callback fired when a file is chosen

3. Builds two filters
   - `startsWith(base string)`:
     - If `prefix` is empty, everything passes.
     - Otherwise it strips the file extension and checks whether the remaining name starts with `prefix`, case-insensitively.
   - `extAllowed(base string)`:
     - If `exts` is empty, everything passes.
     - Otherwise it keeps only files whose extension matches one of the allowed extensions, case-insensitively.

4. Creates the UI
   - `pathLabel` shows the current directory path.
   - `list` is a `widget.List` that displays the current directory contents.
   - Directory names are shown in brackets like `[src]`.
   - Files are shown as plain names.

5. Sets the starting directory
   - It calls `os.Getwd()` and turns that into a file URI.
   - That becomes the current folder the dialog opens in.

6. Implements `refreshList()`
   - Reads the current directory using `storage.ListerForURI(curURI)`.
   - Gets all children with `lister.List()`.
   - Splits them into:
     - `dirs`: all directories
     - `files`: only files that match the prefix and extension filters
   - Combines them as `dirs` first, then `files`.
   - Refreshes the list widget and updates the path label.

7. Creates buttons and selection behavior
   - `Open`:
     - Only enabled when a file is selected.
     - Calls `onOpen(selected.uri)`.
   - `Up`:
     - Moves to the parent directory.
     - Clears selection and refreshes the list.
   - List selection:
     - If the user clicks a directory, the dialog navigates into it immediately.
     - If the user clicks a file, it becomes the selected item and enables `Open`.

8. Builds and shows the dialog
   - Uses `dialog.NewCustomConfirm(...)` to make a custom modal dialog.
   - The `Open` confirmation also triggers `onOpen` if a file is selected.
   - Resizes the dialog and shows it.
   - Calls `refreshList()` once before display so it has initial content.

A few important details:

- Directories are always shown, even if they do not match the prefix or extension filter.
- The filters apply only to files.
- Selection of a directory is treated as navigation, not as a final choice.
- The file list is rebuilt every time you change folders.

Two caveats in the snippet:

- `dialog.ShowError(err, w)` uses `w`, but this function takes `parent`. That looks like a leftover bug unless `w` exists in an outer scope.
- `width` and `height` are also not defined in this snippet, so they must come from elsewhere or this code will not compile.

In plain terms: this is a custom “open file” dialog with folder navigation, a prefix filter, and optional extension filtering, built on top of Fyne’s storage and widget APIs.

*/
