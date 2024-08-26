package main

/*
  From Linux Magazine 277 Dec 2023
  25 Aug 2024 -- From listing 3
*/

import (
	"flag"
	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"strconv"
	"time"
)

var verboseFlag = flag.Bool("v", false, "Verbose mode")

func mkChart(symbol string, reslt qMap) *widgets.BarChart {
	bc := widgets.NewBarChart()
	bc.Data = []float64{}
	bc.Labels = []string{}
	bc.BarWidth = 1
	bc.BarGap = 0

	vals := reslt[symbol]
	var minFloat float64

	for i := len(vals) - 1; i >= 0; i-- {
		closingPrice, err := strconv.ParseFloat(vals[i].closingPrice, 64)
		if err != nil {
			panic(err)
		}
		bc.Data = append(bc.Data, closingPrice)

		if minFloat == 0 || closingPrice < minFloat {
			minFloat = closingPrice
		}

		bc.Labels = append(bc.Labels, weekday(vals[i].date))
	}

	for i := range bc.Data {
		bc.Data[i] -= minFloat
	}

	bc.NumFormatter = func(f float64) string {
		return ""
	}

	bc.Title = symbol
	bc.BarWidth = 1
	bc.BarColors = []termui.Color{termui.ColorRed, termui.ColorGreen}

	return bc
}

func weekday(date string) string {
	dt, _ := time.Parse("2006-01-02", date)
	return string(dt.Weekday().String()[0])
}

func main() {
	flag.Parse()
	symbols := "aapl,nflx,meta,amzn,tsla,goog"

	result, err := fetchQ(symbols)
	if err != nil {
		panic(err)
	}

	err = termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	charts := []*widgets.BarChart{}
	for s := range result {
		charts = append(charts, mkChart(s, result))
	}

	grid := termui.NewGrid()
	termWidth, termHeight := termui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)
	grid.Set(
		termui.NewRow(1.0/2,
			termui.NewCol(1.0/3, charts[0]),
			termui.NewCol(1.0/3, charts[1]),
			termui.NewCol(1.0/3, charts[2]),
		),
		termui.NewRow(1.0/2,
			termui.NewCol(1.0/3, charts[3]),
			termui.NewCol(1.0/3, charts[4]),
			termui.NewCol(1.0/3, charts[5]),
		),
	)

	termui.Render(grid)
	<-termui.PollEvents()
}
