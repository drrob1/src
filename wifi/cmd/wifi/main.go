package main

import (
	"github.com/rivo/tview"
	"src/wifi"
)

func main() {
	app := tview.NewApplication()
	table := tview.NewTable().SetBorders(true)          // this means draw row and col lines
	table.SetBorder(true).SetTitle("Wifi Monitor v1.0") // refers to the container holding the table, and this means draw a border around the application, w/ a title at the top.

	wifi.NewPlugin(app, table, "Time", wifi.Clock)
	wifi.NewPlugin(app, table, "Ping", wifi.Ping, "www.google.com")
	wifi.NewPlugin(app, table, "Ping", wifi.Ping, "8.8.8.8")
	wifi.NewPlugin(app, table, "Ifconfig", wifi.Nifs)
	wifi.NewPlugin(app, table, "HTTP", wifi.HttpGet, "https://youtu.be")

	err := app.SetRoot(table, true).SetFocus(table).Run()
	if err != nil {
		panic(err)
	}
}
