//go:build !windows

package main

import "fyne.io/fyne/v2/dialog"

func showPrintDialog() {
	dialog.ShowInformation("PRINT", "Printing is only supported on Windows.", globalWindow)
}
