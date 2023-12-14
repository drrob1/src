package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"src/wifi"
)

var app *tview.Application

func main() {
	app = tview.NewApplication()
	app.SetInputCapture(inputCap)
	table := tview.NewTable().SetBorders(true)          // this means draw row and col lines
	table.SetBorder(true).SetTitle("Wifi Monitor v1.0") // refers to the container holding the table, and this means draw a border around the application, w/ a title at the top.

	wifi.NewPlugin(app, table, "Time", wifi.Clock)
	wifi.NewPlugin(app, table, "Ping", wifi.Ping, "www.google.com")
	wifi.NewPlugin(app, table, "Ping", wifi.Ping, "8.8.8.8")
	wifi.NewPlugin(app, table, "Ping", wifi.Ping, "1.1.1.1")
	wifi.NewPlugin(app, table, "Ifconfig", wifi.Nifs)
	wifi.NewPlugin(app, table, "HTTP", wifi.HttpGet, "https://youtu.be")
	wifi.NewPlugin(app, table, "Average", wifi.AveFetchTime)

	err := app.SetRoot(table, true).SetFocus(table).Run()
	if err != nil {
		panic(err)
	}
}

// --------------------------------------------------- inputCap for tcell--------------------------------------

func inputCap(event *tcell.EventKey) *tcell.EventKey { // to be used for SetInputCapture
	//func (a *Application) SetInputCapture(capture func(event *tcell.EventKey) *tcell.EventKey) *Application

	pollevent := func() {
		for {
			if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyEnter {
				app.Stop()
			}
		}
	}

	go pollevent()

	return nil

} // inputCap for tcell
