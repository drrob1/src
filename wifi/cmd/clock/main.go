package main

import (
	"fmt"
	"github.com/rivo/tview"
	"src/wifi"
)

func main() {
	app := tview.NewApplication()
	tv := tview.NewTextView()
	tv.SetBorder(true).SetTitle("Test Clock")
	ch := wifi.Clock()

	go func() {
		for {
			val := <-ch // this is blocking
			app.QueueUpdateDraw(func() {
				tv.Clear()
				fmt.Fprintf(tv, "%s ", val)
			})
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
