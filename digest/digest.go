package main

import (
	"fmt"
	"github.com/cavaliergopher/grab/v3"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"os"
	"src/timlibg"
	"strconv"
)

//import "github.com/cavaliercoder/grab"  Turns out that this URL is out of date.  When I switched to the above import, the code started working.

/*
  REVISION HISTORY
  ======== =======
  12 May 23 -- Got idea from Scott to write a Go program to d/l the TimesDigest every day.  Then I have to be able to email it to myself, and likely Karen and maybe Scott.
                 I'm typing this @10:30 pm, and this will take a bit of time.
   https://golangdocs.com/golang-download-files#:~:text=Go%20Download%20the%20File%20using%20the%20net%2Fhttp%20package,last%20part%20of%20the%20URL%20as%20the%20filename
*/

const lastModified = "13 May 23"
const td = "TimesDigest_"
const tail = ".pdf"
const urlBase = "http://s1.nyt.com/tdpdf/"

func main() {
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Printf(" %s to download the TimesDigest.  Last altered %s, %s last linked %s\n", execName, lastModified, os.Args[0], LastLinkedTimeStamp)

	todayString := dateStr()
	url := urlBase + td + todayString + tail
	resp, err := grab.Get(".", url)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " url = %s and did not download correctly.  Error returned from grab.Get is %s.  \n", url, err)
		os.Exit(1)
	}

	fmt.Println("Download saved to", resp.Filename)

}

/* ------------------------------------------- MakeDateStr ---------------------------------------------------* */

func dateStr() string {
	m, d, y := timlibg.TIME2MDY()
	MSTR := strconv.Itoa(m)
	DSTR := strconv.Itoa(d)
	YSTR := strconv.Itoa(y)

	if len(MSTR) == 1 {
		MSTR = "0" + MSTR
	}
	if len(DSTR) == 1 {
		DSTR = "0" + DSTR
	}
	datestr := YSTR + MSTR + DSTR
	return datestr
} // dateStr
