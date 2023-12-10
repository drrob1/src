package wifi

/*
  From Go Network Diagnostics, Linux Magazine 275, Oct 2023, pp 58ff

  It is using unbuffered channels, so all channel read operations are blocking.  I took out the select statement in the code; this was also flagged by staticCheck.
*/

import (
	"fmt"
	"github.com/prometheus-community/pro-bing"
	"github.com/rivo/tview"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"
)

func NewPlugin(app *tview.Application, table *tview.Table, field string, fu func(...string) chan string, arg ...string) {
	if len(arg) > 0 {
		field += " " + strings.Join(arg, " ")
	}

	row := table.GetRowCount()
	table.SetCell(row, 0, tview.NewTableCell(field))

	ch := fu(arg...)

	go func() {
		for {
			val := <-ch // this is blocking.
			app.QueueUpdateDraw(func() {
				table.SetCell(row, 1, tview.NewTableCell(val))
			})
		}
	}()
	//go func() {  This is the way it's coded in the article.  Don't need select statement when there's only 1 channel to select.  I made that mistake also.  Flagged by staticCheck.
	//	for {
	//		select {
	//		case val := <-ch:
	//			app.QueueUpdateDraw(func() {
	//				table.SetCell(row, 1, tview.NewTableCell(val))
	//			})
	//		}
	//	}
	//}()
}
func Clock(arg ...string) chan string {

	// The article explains why this code uses time.Unix.  Go does not provide elegant formatting as a string for duration type, as returned from the time.Since() function.
	// Go does provide elegant formatting for absolute time values using the Format() function.  To get formatting for the duration type, the code converts a value of type duration
	// to absolute time by adding it to the beginning of time at zero Unix seconds.

	ch := make(chan string)
	start := time.Now()

	go func() {
		for {
			z := time.Unix(0, 0).UTC()
			ch <- z.Add(time.Since(start)).Format("15:04:05")
			time.Sleep(1 * time.Second)
		}
	}()

	return ch
}

func Clock2() chan string {
	// I want to play w/ just using the string interface of a duration and see what happens.  And I'll make it a buffered channel.

	chn := make(chan string, 1)
	start := time.Now()

	go func() {
		for {
			chn <- time.Since(start).String()
			time.Sleep(1 * time.Second)
		}
	}()

	return chn
}

func Ping(addr ...string) chan string {
	ch := make(chan string)
	firstTime := true

	go func() {
		for {
			pinger, err := probing.NewPinger(addr[0])
			pinger.Timeout, _ = time.ParseDuration("10s")

			if err != nil {
				ch <- err.Error()
				time.Sleep(10 * time.Second)
				continue
			}

			if firstTime {
				ch <- "Pinging ..."
				firstTime = false
			}

			pinger.Count = 3
			err = pinger.Run()

			if err != nil {
				ch <- err.Error()
				time.Sleep(10 * time.Second)
				continue
			}

			stats := pinger.Statistics()
			ch <- fmt.Sprintf("%v ", stats.Rtts)
			time.Sleep(10 * time.Second)
		}
	}()
	return ch
}

func Nifs(arg ...string) chan string {
	ch := make(chan string)

	go func() {
		for {
			eths, err := Ifconfig()

			if err != nil {
				ch <- err.Error()
				time.Sleep(10 * time.Second)
				continue
			}

			ch <- strings.Join(eths, ", ")
			time.Sleep(10 * time.Second)
		}
	}()

	return ch
}

func Ifconfig() ([]string, error) {
	var list []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return list, err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return list, err
		}

		if len(addrs) == 0 {
			continue
		}

		for _, addr := range addrs {
			ip := strings.Split(addr.String(), "/")[0]
			if net.ParseIP(ip).To4() != nil {
				list = append(list, iface.Name+" "+ip)
			}
		}
	}

	sort.Strings(list)
	return list, nil
}

func HttpGet(arg ...string) chan string {
	ch := make(chan string)

	firstTime := true
	go func() {
		for {
			if firstTime {
				ch <- "Fetching ..."
				firstTime = false
			}

			now := time.Now()
			_, err := http.Get(arg[0])
			if err != nil {
				ch <- err.Error()
				time.Sleep(10 * time.Second)
				continue
			}

			dur := time.Since(now)
			ch <- fmt.Sprintf("%.3f OK ", dur.Seconds())
			time.Sleep(10 * time.Second)
		}
	}()

	return ch
}
