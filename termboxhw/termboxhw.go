package main

import (
	"bufio"
	"fmt"
	tb "github.com/nsf/termbox-go"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func Cap(c rune) rune {
	r, _, _, _ := strconv.UnquoteChar(strings.ToUpper(string(c)), 0)
	return r
} // Cap

func print_tb(x, y int, fg, bg tb.Attribute, msg string) {
	for _, c := range msg {
		tb.SetCell(x, y, c, fg, bg)
		x++
	}
	err := tb.Flush()
	if err != nil {
		panic(err)
	}
}

func printf_tb(x, y int, fg, bg tb.Attribute, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	print_tb(x, y, fg, bg, s)
}

func main() {
	var x, y int
	var fg, fgBold, bg tb.Attribute

	startrow := 0
	startcol := 0
	if runtime.GOOS == "windows" {
		startrow = 2 // starting at row 0 or 1 fails in tcc.  I don't know why.  Row 1 works under cmd.
		startcol = 0
	}
	fmt.Println(" On", runtime.GOOS, ", ARCH =", runtime.GOARCH, ".  Press <enter> to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	_ = scanner.Text()

	err := tb.Init()
	if err != nil {
		fmt.Println(" Init call to termbox-go failed with error:", err)
		os.Exit(1)
	}
	defer tb.Close()

	x = startcol
	y = startrow
	fg = tb.ColorYellow
	fgBold = tb.ColorYellow | tb.AttrBold
	fgcyan := tb.ColorCyan
	fgboldcyan := tb.ColorCyan | tb.AttrBold
	fgblue := tb.ColorBlue
	fgboldblue := tb.ColorBlue | tb.AttrBold
	bg = tb.ColorBlack
	err = tb.Clear(fg, bg)
	if err != nil {
		fmt.Println(" First Clear call to termbox-go Clear failed with error:", err)
		os.Exit(1)
	}
	err = tb.Flush()
	if err != nil {
		panic(err)
	}

	print_tb(x, y, fg, bg, "Hello World in Yellow")
	y++
	print_tb(x, y, fgBold, bg, "Hello World in Bold Yellow.")
	y++
	print_tb(x, y, fgcyan, bg, "Hello World in Cyan")
	y++
	print_tb(x, y, fgboldcyan, bg, "Hello World in Bold Cyan")
	y++
	print_tb(x, y, fgblue, bg, "Hello World in Blue")
	y++
	print_tb(x, y, fgboldblue, bg, "Hello World in Bold Blue, and then hit q to exit")
	y++

	tb.SetCursor(x, y)
	err = tb.Flush()
	if err != nil {
		panic(err)
	}

EventLoop:
	for {
		event := tb.PollEvent()
		switch event.Type {
		case tb.EventKey:
			ch := event.Ch
			chkey := event.Key
			if Cap(ch) == 'Q' {
				break EventLoop
			} else if chkey == tb.KeySpace {
				ch = ' '
			} else if ch == 0 {
				if chkey == tb.KeyEnter || chkey == tb.KeyCtrlM || chkey == tb.KeyCtrlJ {
					x = startcol
					y++
					s := fmt.Sprintf("%q", chkey)
					printf_tb(x, y+10, fgboldcyan, bg, "%q", chkey)
					print_tb(x, y+15, fgcyan, bg, s)
				} else if chkey == tb.KeyBackspace || chkey == tb.KeyDelete || chkey > '~' {
					x--
				} else if chkey < tb.KeySpace || chkey > '~' {
					printf_tb(x, y+20, fgboldblue, bg, "%q", chkey)
				} // end if chkey is something
			} // end if ch == 0
			tb.SetCell(x, y, ch, fgBold, bg)
			// If a special key was entered, then ch will be 0.  Only increment if have a regular key

			if ch > 0 || chkey == tb.KeySpace {
				x++
			}
			if x > 40 {
				y++
				x = startcol
			}
			tb.SetCursor(x, y)
			err = tb.Flush()
			if err != nil {
				panic(err)
			}
		case tb.EventResize:
			err := tb.Sync()
			if err != nil {
				panic(err)
			}
			err = tb.Flush()
			if err != nil {
				panic(err)
			}
		case tb.EventError:
			panic(event.Err)

		case tb.EventMouse:

		case tb.EventInterrupt:

		case tb.EventRaw:

		case tb.EventNone:
			//           ignore these for now
		} // end switch-case
	} // end for ever EventLoop loop

	err = tb.Flush() // ignore any possible error as we are about to exit anyway.
}
