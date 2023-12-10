package main

import (
	"fmt"
	"github.com/rivo/tview"
	"src/wifi"
)

func main() {
	app := tview.NewApplication()
	tv := tview.NewTextView()
	tv.SetBorder(true).SetTitle("Test Clock2")
	ch := wifi.Clock2()

	go func() {
		for {
			val := <-ch     // this is no longer blocking
			clk := func() { // now I'm playing w/ the syntax I find more clear than what I usually see.  The conventional syntax is in clock/main.go
				tv.Clear()
				fmt.Fprintf(tv, "%s ", val)
			}
			app.QueueUpdateDraw(clk)
		}
	}()

	//go func() {  Coded this way in the article, but don't need select statement when there's only 1 channel to select from.
	//	for {
	//		select {
	//		case val := <-ch:
	//			app.QueueUpdateDraw(func() {
	//				tv.Clear()
	//				fmt.Fprintf(tv, "%s ", val)
	//			})
	//		}
	//	}
	//}()

	err := app.SetRoot(tv, true).Run()
	if err != nil {
		panic(err)
	}
}
