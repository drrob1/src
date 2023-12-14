package wifi

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

/*
  From Go Network Diagnostics, Linux Magazine 275, Oct 2023, pp 58ff

  It is using unbuffered channels, so all channel read operations are blocking.  I took out the select statement in the code; this was also flagged by staticCheck.

  Each of these functions runs forever by starting a new goroutine that returns its results on the returned string channel.  So NewPlugin just needs to start the function passed to it once,
  and then it starts a separate goroutine to receive the channel strings and update the assigned table row forever.

  The original code uses blocking channels; I changed that to use buffered channels.

  If the Ifconfig rtn returns an IP in the range of 192.168.0.x, then the connection to the router is working.  If only the loopback interface appears, something is wrong w/ the
  assignment of IP addresses, so check the dhcp settings.

  The httpGet rtn provides an end-to-end test by loading the YouTube title page off of the web.  If this works, everything should be fine.  Because it also measures the time it takes
  to retrieve the page, you can estimate the speed of the ISP connection.  The article shows a time of 0.142 sec here, which it says is perfect.
  If the display in the table column gets stuck at Fetching ..., then something is wrong w/ the connection.  The other tests should give you some clues to the cause.

  If the hostname resolution fails due to incorrect DNS configuration, the error message is sent into the channel for display.

  Now I have to understand how these routines return their values to be displayed.
  All except Ifconfig returns its string value in a string channel.  But Ifconfig is called by Nifs.  Ifconfig is not initialized by NewPlugin.
  Nifs returns its string value on the string channel.

  tview is based on tcell.  So maybe I'll try a routine I wrote for tcell here.

  REVISION HISTORY
  ======== =======
  12 Dec 23 -- Starting to add the averaging code for the time it takes to retrieve the youtube header.  And I'm going to discard the first one, as that's always much too long.
               While the time is typically ~150 ms, the first one is 500-600 ms.
*/

const lastModified = "13 Dec 2023"
const timeToSleep = 10 * time.Second
const debugFile = "debug.out"

var sliceOfSeconds []float64

func init() {
	sliceOfSeconds = make([]float64, 0, 100)
}

func NewPlugin(app *tview.Application, table *tview.Table, field string, fu func(...string) chan string, arg ...string) {
	// this is to integrate the functions w/ the table rows.  Input is a pointer to the application, pointer to the table, field name to display in the first col,
	// a function that takes input of one or more strings and returns a string channel, and finally an
	// optional last param is a string that is also used in the field name cell in col 0 for that row.  IE, the first col.
	// Since app and table are both pointers, they are output params as well as input params.
	if len(arg) > 0 {
		field += " " + strings.Join(arg, " ")
	}

	// associate the function w/ the next avail row by appending a new row to the table w/ each call.
	row := table.GetRowCount()
	table.SetCell(row, 0, tview.NewTableCell(field)) // append a row to the table.

	// call the new function, which runs forever.
	ch := fu(arg...) // arg is input to this routine as the last input param

	// create a separate go routine (which runs forever) to receive the string in its channel, and update the correct row.
	go func() {
		for {
			val := <-ch // this does not need to be in a select statement
			setCellFunc := func() {
				table.SetCell(row, 1, tview.NewTableCell(val)) // update contents of row and 2nd col, ie, col 1.
			}
			app.QueueUpdateDraw(setCellFunc) // redraw the table whenever the pgm gets around to it durng the next refresh.  Hence the Queue in the function name.
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

func Clock(arg ...string) chan string { // signature has to match the others.

	// The article explains why this code uses time.Unix.  Go does not provide elegant formatting as a string for duration type, as returned from the time.Since() function.
	// Go does provide elegant formatting for absolute time values using the Format() function.  To get formatting for the duration type, the code converts a value of type duration
	// to absolute time by adding it to the beginning of time at zero Unix seconds.

	ch := make(chan string, 1)
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
	// Now it's working to my satisfaction.

	chn := make(chan string, 1)
	start := time.Now()

	go func() {
		for {
			dur := time.Since(start)
			hrs := int(dur.Hours())
			minutes := int(dur.Minutes()) % 60
			totalSec := int(dur.Seconds())
			sec := totalSec % 60
			chn <- fmt.Sprintf("%d:%d:%d", hrs, minutes, sec)
			time.Sleep(996 * time.Millisecond)
		}
	}()

	return chn
}

func Ping(addr ...string) chan string { // expects either a hostname or an IP address as its argument.  The returned channel is used to continuously return ping results to the caller.
	ch := make(chan string, 1)
	firstTime := true

	go func() {
		for {
			pinger, err := probing.NewPinger(addr[0])     // send ICMP packets to the addr.
			pinger.Timeout, _ = time.ParseDuration("10s") // sets a timeout of 10 sec.

			if err != nil {
				ch <- err.Error()
				time.Sleep(10 * time.Second)
				continue
			}

			if firstTime {
				ch <- "Pinging ..."
				firstTime = false
			}

			pinger.Count = 3 // send 3 pings for each run.
			err = pinger.Run()

			if err != nil {
				ch <- err.Error()
				time.Sleep(10 * time.Second)
				continue
			}

			stats := pinger.Statistics()
			ch <- fmt.Sprintf("%v ", stats.Rtts) // stats.Rtts is a slice of floats containing the response times, in seconds.
			time.Sleep(timeToSleep)
		}
	}()
	return ch
}

func Nifs(arg ...string) chan string {
	ch := make(chan string, 1)

	go func() {
		for {
			eths, err := Ifconfig()

			if err != nil {
				ch <- err.Error()
				time.Sleep(10 * time.Second)
				continue
			}

			ch <- strings.Join(eths, ", ")
			time.Sleep(timeToSleep)
		}
	}()

	return ch
}

func Ifconfig() ([]string, error) {
	var list []string
	ifaces, err := net.Interfaces() // std lib function that returns all of the computer's network interfaces.
	if err != nil {
		return list, err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs() // this fetches the IP addresses for the computer's network interfaces.  It will have an address only if the network op succeeded.
		if err != nil {
			return list, err
		}

		if len(addrs) == 0 {
			continue
		}

		for _, addr := range addrs { // check the IP addresses fetched from iface.Addrs().
			ip := strings.Split(addr.String(), "/")[0]
			if net.ParseIP(ip).To4() != nil { // filter out IPv6 addresses, as almost no one in the US has one.
				list = append(list, iface.Name+" "+ip) // append interface name and IP address without the subnet suffix.
			}
		}
	}

	sort.Strings(list) // sort by interface name before returning them.
	return list, nil
}

func HttpGet(arg ...string) chan string {
	ch := make(chan string, 1)

	firstTime := true
	go func() {
		for {
			if firstTime {
				ch <- "Fetching ..."
				firstTime = false
			}

			now := time.Now()
			_, err := http.Get(arg[0]) // function blocks until data arrives or server returns an error.
			if err != nil {
				ch <- err.Error()
				time.Sleep(10 * time.Second)
				continue
			}

			dur := time.Since(now) // measure how long the Get call took.
			sliceOfSeconds = append(sliceOfSeconds, dur.Seconds())
			ch <- fmt.Sprintf("%.3f OK ", dur.Seconds())
			time.Sleep(timeToSleep)
		}
	}()

	return ch
}

func AveFetchTime(arg ...string) chan string {

	//OutputFile, err := os.OpenFile(debugFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	//if err != nil {
	//	fmt.Printf(" Cannot open %s becuase ERROR is: %s\n", debugFile, err)
	//}
	//OutputFile.WriteString("--------------------------------------------------------------------------------------\n")
	//OutputFile.Close()

	ch := make(chan string, 1)

	firstTime := true

	go func() {
		for {
			if firstTime {
				ch <- "Fetching ..."
				firstTime = false
				time.Sleep(timeToSleep)
				continue
			}
			sliceCopy := make([]float64, len(sliceOfSeconds))
			copy(sliceCopy, sliceOfSeconds) // when I didn't copy the slice, this routine was clobbering the slice when I removed the first element.  So now I copy the slice and trim the copy.  Now it works.

			if len(sliceOfSeconds) > 2 {
				sliceCopy = sliceCopy[1:]
			}
			//OutputFile, _ := os.OpenFile(debugFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
			//s := fmt.Sprintf("AveFetchTime:  len of slice=%d, 1st slice=%v, len 2nd=%d, 2nd slice: %v\n", len(sliceOfSeconds), sliceOfSeconds, len(sliceCopy), sliceCopy)
			//OutputFile.WriteString(s)
			var sumT float64 // When I defined this above the goroutine, it wasn't being initialized each time, so the totals were wrong.
			for _, r := range sliceCopy {
				sumT += r
			}
			ave := sumT / float64(len(sliceOfSeconds))
			//s = fmt.Sprintf("AveFetchTime: sumT = %.3f, len(sliceOfSeconds)=%d, ave=%.3f\n", sumT, len(sliceOfSeconds), ave)
			//OutputFile.WriteString(s)
			//OutputFile.Close()

			ch <- fmt.Sprintf("%.3f ave", ave)
			time.Sleep(timeToSleep)
		}

	}()

	return ch
}
