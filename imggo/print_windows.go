//go:build windows

package main

import (
	"bytes"
	"fmt"
	"image/jpeg"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/godoes/printers"
)

func showPrintDialog() {
	names, err := printers.ReadNames()
	if err != nil {
		dialog.ShowError(fmt.Errorf("read printers: %w", err), globalWindow)
		return
	}
	if len(names) == 0 {
		dialog.ShowInformation("PRINT", "No installed printers were found.", globalWindow)
		return
	}

	selectWidget := widget.NewSelect(names, nil)
	selectWidget.PlaceHolder = "Choose a printer"
	if defaultPrinter, err := printers.GetDefault(); err == nil {
		selectWidget.SetSelected(defaultPrinter)
	}
	if selectWidget.Selected == "" {
		selectWidget.SetSelected(names[0])
	}

	dialog.NewCustomConfirm("PRINT", "PRINT", "CANCEL",
		container.NewVBox(widget.NewLabel("Choose an installed printer:"), selectWidget),
		func(ok bool) {
			if !ok {
				return
			}
			if err := printCurrent(selectWidget.Selected); err != nil {
				dialog.ShowError(err, globalWindow)
			}
		},
		globalWindow,
	).Show()
}

func printCurrent(printerName string) error {
	if currentImage == nil {
		return fmt.Errorf("no image loaded")
	}
	if printerName == "" {
		return fmt.Errorf("no printer selected")
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, currentImage, &jpeg.Options{Quality: 95}); err != nil {
		return err
	}

	p, err := printers.Open(printerName)
	if err != nil {
		return err
	}
	defer p.Close()

	if err := p.StartRawDocument(printerName); err != nil {
		return err
	}
	if err := p.StartPage(); err != nil {
		return err
	}
	if _, err := p.Write(buf.Bytes()); err != nil {
		return err
	}
	if err := p.EndPage(); err != nil {
		return err
	}
	if err := p.EndDocument(); err != nil {
		return err
	}
	return nil
}
