// This code started life as termboxhw (termbox helloworld).  Now I have to write a func to
// GetInputString starting from the x, y passed in.

package main

import (
	"bufio"
	"fmt"
	"github.com/nsf/termbox-go"
	tb "github.com/nsf/termbox-go"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var fgBrightYellow, bkgrnd termbox.Attribute

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

// --------------------------------------------------- GetInputString --------------------------------------
func GetInputString(x, y int) string {
	bs := make([]byte, 0, 100) // byteslice to build up the string to be returned.
	termbox.SetCursor(x, y)

MainEventLoop:
	for {
		event := termbox.PollEvent()
		switch event.Type {
		case termbox.EventKey:
			ch := event.Ch
			key := event.Key
			if key == termbox.KeySpace {
				ch = ' '
				if len(bs) > 0 { // ignore spaces if there is no string yet
					break MainEventLoop
				}
			} else if ch == 0 { // need to process backspace and del keys
				printf_tb(x, y+15, fgBrightYellow, bkgrnd, "key = %q, %v, %x, %d", key, key, key, key)
				if key == termbox.KeyEnter {
					break MainEventLoop
				} else if key == termbox.KeyF1 || key == termbox.KeyF2 {
					bs = append(bs, "HELP"...)
					break MainEventLoop
				} else if key == termbox.KeyPgup || key == termbox.KeyArrowUp {
					bs = append(bs, "UP"...) // Code in C++ returned ',' here
					break MainEventLoop
				} else if key == termbox.KeyPgdn || key == termbox.KeyArrowDown {
					bs = append(bs, "DN"...) // Code in C++ returned '!' here
					break MainEventLoop
				} else if key == termbox.KeyArrowRight || key == termbox.KeyArrowLeft {
					bs = append(bs, '~') // Could return '<' or '>' or '<>' or '><' also
					break MainEventLoop
				} else if key == termbox.KeyEsc {
					bs = append(bs, 'Q')
					break MainEventLoop
					// this test must be last because all special keys above meet condition of key > '~'
				} else if (len(bs) > 0) && (key == termbox.KeyDelete || key > '~' || key == 8) { // key is 8 for <bs> on windows
					x--
					bs = bs[:len(bs)-1]
				}
			} else if ch == '=' {
				ch = '+'
			} else if ch == ';' {
				ch = '*'
			}
			termbox.SetCell(x, y, ch, fgBrightYellow, bkgrnd)
			if ch > 0 {
				x++
				bs = append(bs, byte(ch))
			}
			termbox.SetCursor(x, y)
			err := termbox.Flush()
			check(err)
		case termbox.EventResize:
			err := termbox.Sync()
			check(err)
			err = termbox.Flush()
			check(err)
		case termbox.EventError:
			panic(event.Err)
		case termbox.EventMouse:
		case termbox.EventInterrupt:
		case termbox.EventRaw:
		case termbox.EventNone:

		} // end switch-case on the Main Event  (Pun intended)

	} // MainEventLoop for ever

	return string(bs)
} // end GetInputString

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
	MaxCol, MaxRow := termbox.Size()
	fg = tb.ColorYellow
	fgBold = tb.ColorYellow | tb.AttrBold
	fgBrightYellow = fgBold
	fgcyan := tb.ColorCyan
	fgboldcyan := tb.ColorCyan | tb.AttrBold
	fgBrightCyan := fgboldcyan
	fgblue := tb.ColorBlue
	fgboldblue := tb.ColorBlue | tb.AttrBold
	bg = tb.ColorBlack
	bkgrnd = bg
	err = tb.Clear(fgBrightYellow, bkgrnd)
	check(err)
	err = tb.Flush()
	check(err)

	print_tb(x, y, fg, bg, "Hello World in Yellow")
	x++
	y++
	printf_tb(x, y, fgBold, bg, "Hello World in Bold Yellow.  MaxCol (x) = %d, MaxRow (y) = %d", MaxCol, MaxRow)
	x++
	y++
	print_tb(x, y, fgcyan, bg, "Hello World in Cyan")
	x++
	y++
	print_tb(x, y, fgboldcyan, bg, "Hello World in Bold Cyan")
	x++
	y++
	print_tb(x, y, fgBrightCyan, bg, "Hello World in Bright Cyan")
	x++
	y++
	print_tb(x, y, fgblue, bg, "Hello World in Blue")
	x++
	y++
	print_tb(x, y, fgboldblue, bg, "Hello World in Bold Blue, and then hit q to exit")
	x++
	y++

	//  x = startcol;
	tb.SetCursor(x, y)
	err = tb.Flush()
	check(err)

	for {
		s := GetInputString(x, y)
		print_tb(x, y+20, fgBrightYellow, bkgrnd, s)
		x++
		y++
		termbox.SetCursor(x, y)
		err = tb.Flush()
		check(err)
		printf_tb(x, y+30, fgBrightCyan, bkgrnd, "len of s is %d", len(s))
		if len(s) == 0 || s[0] == 'q' || s[0] == 'Q' {
			break
		}
	}
} // end Main func

func check(e error) {
	if e != nil {
		panic(e)
	}
}
