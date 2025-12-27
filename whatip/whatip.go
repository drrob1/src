package main // What is my IP?, from tutorial for fyne YouTube video #19
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

/*
  27 Dec 25 -- First version of whatip program.
*/

const lastModified = "27 Dec 25"

type IP struct {
	Query   string
	Country string
	City    string
}

func myIP() (string, string, string) {
	req, err := http.Get("http://ip-api.com/json/")
	if err != nil {
		return err.Error(), "", ""
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err.Error(), "", ""
	}

	var ip IP

	json.Unmarshal(body, &ip)

	return ip.Query, ip.City, ip.Country
}

func main() {
	a := app.New()
	s := fmt.Sprintf("What is my IP, last modified %s, compiled with %s", lastModified, runtime.Version())
	w := a.NewWindow(s)
	w.Resize(fyne.NewSize(400, 400))

	typedKey := func(ev *fyne.KeyEvent) {
		key := string(ev.Name)
		switch key {
		case "Q", "Escape", "X":
			os.Exit(0)
		}
	}
	w.Canvas().SetOnTypedKey(typedKey)

	labelTitle := widget.NewLabel("What is my IP?")
	labelIP := widget.NewLabel("IP is ...")
	labelValue := widget.NewLabel("...")
	labelCity := widget.NewLabel("...")
	labelCountry := widget.NewLabel("...")

	btnFunc := func() {
		labelValue.Text, labelCity.Text, labelCountry.Text = myIP()
		labelValue.Refresh()
		labelCity.Refresh()
		labelCountry.Refresh()

	}
	btn := widget.NewButton("Run", btnFunc)

	quitBtn := widget.NewButton("Quit", func() { os.Exit(0) })

	w.SetContent(container.NewVBox(labelTitle, labelIP, labelValue, labelCity, labelCountry, btn, quitBtn))

	w.ShowAndRun()
}
