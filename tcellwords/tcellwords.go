package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell"
	runewidth "github.com/mattn/go-runewidth"
)

type keyStructType struct {
	r    rune
	name string
}

var row = 0
var style = tcell.StyleDefault

func putln(scrn tcell.Screen, str string) {
	puts(scrn, style, 1, row, str)
	row++
}

func putfln(scrn tcell.Screen, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	puts(scrn, style, 1, row, s)
	row++
}

func putf(scrn tcell.Screen, style tcell.Style, x, y int, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	puts(scrn, style, x, y, s)
}

func deleol(scrn tcell.Screen, style tcell.Style, x, y int) {
	width, _ := scrn.Size() // don't need height for this calculation.
	empty := width - x      // don't care if this is off by 1.
	blanks := make([]byte, empty)
	for i := range blanks {
		blanks[i] = ' '
	}
	blankstring := string(blanks)
	puts(scrn, style, x, y, blankstring)
}

func puts(scrn tcell.Screen, style tcell.Style, x, y int, str string) {
	i := 0
	var deferred []rune
	dwidth := 0
	zwj := false
	for _, r := range str {
		if r == '\u200d' {
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
			deferred = append(deferred, r)
			zwj = true
			continue
		}
		if zwj {
			deferred = append(deferred, r)
			zwj = false
			continue
		}
		switch runewidth.RuneWidth(r) {
		case 0:
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
		case 1:
			if len(deferred) != 0 {
				scrn.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 1
		case 2:
			if len(deferred) != 0 {
				scrn.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 2
		}
		deferred = append(deferred, r)
	}
	if len(deferred) != 0 {
		scrn.SetContent(x+i, y, deferred[0], deferred[1:], style)
		i += dwidth
	}
}

func main() {
	scrn, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	//	encoding.Register()   Don't know why this is here.  This code works without this line, so I'll keep it out.

	if e = scrn.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	defer scrn.Fini()

	plain := tcell.StyleDefault
	bold := style.Bold(true)
	reverse := style.Reverse(true)
	reversedBlue := style.Background(tcell.ColorWhite).Foreground(tcell.ColorBlue)
	Blue := style.Foreground(tcell.ColorBlue)
	Green := style.Foreground(tcell.ColorGreen)
	AquaBlue := style.Foreground(tcell.ColorAqua)
	Yellow := style.Foreground(tcell.ColorYellow)
	Maroon := style.Foreground(tcell.ColorMaroon)
	NavyBlue := style.Foreground(tcell.ColorNavy)
	Purple := style.Foreground(tcell.ColorPurple)
	Red := style.Foreground(tcell.ColorRed)
	BoldBlue := Blue.Bold(true)
	BoldGreen := Green.Bold(true)
	BoldYellow := Yellow.Bold(true)
	BoldRed := Red.Bold(true)
	BoldNavy := NavyBlue.Bold(true)
	BoldMaroon := Maroon.Bold(true)
	BoldPurple := Purple.Bold(true)
	BoldAquaBlue := Blue.Bold(true)
	Color16 := style.Foreground(tcell.Color16)
	Color17 := style.Foreground(tcell.Color17)
	Color18 := style.Foreground(tcell.Color18)
	Color19 := style.Foreground(tcell.Color19)
	Color20 := style.Foreground(tcell.Color20)
	Color21 := style.Foreground(tcell.Color21)

	colorslice := make([]tcell.Style, 30)
	for i := 0; i < 30; i++ {
		colorslice[i] = plain
	}

	colorslice[0] = plain                                   // white
	colorslice[1] = Blue                                    // ok color but dark
	colorslice[2] = Green                                   // best color
	colorslice[3] = AquaBlue                                // cyan in other naming systems.  best color.
	colorslice[4] = Yellow                                  // best color
	colorslice[5] = Maroon                                  // bad color
	colorslice[6] = NavyBlue                                // bad color
	colorslice[7] = Purple                                  // bad color
	colorslice[8] = Red                                     // best color
	colorslice[9] = BoldBlue                                // ok color, but still dark.
	colorslice[10] = BoldGreen                              // best color
	colorslice[11] = BoldYellow                             // best color
	colorslice[12] = BoldRed                                // best color
	colorslice[13] = BoldNavy                               // bad color
	colorslice[14] = BoldMaroon                             // bad color
	colorslice[15] = BoldPurple                             // pad color
	colorslice[16] = BoldAquaBlue                           // bad color
	colorslice[17] = Color16                                // bad color
	colorslice[18] = Color17                                // bad color
	colorslice[19] = Color18                                // bad color
	colorslice[20] = Color19                                // bad color
	colorslice[21] = Color20                                // bad color
	colorslice[22] = Color21                                // bad color
	colorslice[23] = style.Foreground(tcell.ColorAliceBlue) // shows as white

	//	scrn.SetStyle(tcell.StyleDefault.
	//		Foreground(tcell.ColorBlack).
	//		Background(tcell.ColorWhite))
	scrn.SetStyle(tcell.StyleDefault. // reversing the colors
						Foreground(tcell.ColorWhite).
						Background(tcell.ColorBlack))
	scrn.Clear()

	style = bold
	putln(scrn, "Press ESC to Exit")
	putln(scrn, "Character set: "+scrn.CharacterSet())
	style = plain

	putln(scrn, "English:   October")
	putln(scrn, "Icelandic: október")
	putln(scrn, "Arabic:    أكتوبر")
	putln(scrn, "Russian:   октября")
	putln(scrn, "Greek:     Οκτωβρίου")
	putln(scrn, "Chinese:   十月 (note, two double wide characters)")
	putln(scrn, "Combining: A\u030a (should look like Angstrom)")
	putln(scrn, "Emoticon:  \U0001f618 (blowing a kiss)")
	putln(scrn, "Airplane:  \u2708 (fly away)")
	putln(scrn, "Command:   \u2318 (mac clover key)")
	putln(scrn, "Enclose:   !\u20e3 (should be enclosed exclamation)")
	putln(scrn, "ZWJ:       \U0001f9db\u200d\u2640 (female vampire)")
	putln(scrn, "ZWJ:       \U0001f9db\u200d\u2642 (male vampire)")
	putln(scrn, "Family:    \U0001f469\u200d\U0001f467\u200d\U0001f467 (woman girl girl)\n")
	putln(scrn, "Region:    \U0001f1fa\U0001f1f8 (USA! USA!)\n")
	putln(scrn, "")
	putln(scrn, "Box:")
	putln(scrn, string([]rune{
		tcell.RuneULCorner,
		tcell.RuneHLine,
		tcell.RuneTTee,
		tcell.RuneHLine,
		tcell.RuneURCorner,
	}))
	putln(scrn, string([]rune{
		tcell.RuneVLine,
		tcell.RuneBullet,
		tcell.RuneVLine,
		tcell.RuneLantern,
		tcell.RuneVLine,
	})+"  (bullet, lantern/section)")
	putln(scrn, string([]rune{
		tcell.RuneLTee,
		tcell.RuneHLine,
		tcell.RunePlus,
		tcell.RuneHLine,
		tcell.RuneRTee,
	}))
	putln(scrn, string([]rune{
		tcell.RuneVLine,
		tcell.RuneDiamond,
		tcell.RuneVLine,
		tcell.RuneUArrow,
		tcell.RuneVLine,
	})+"  (diamond, up arrow)")
	putln(scrn, string([]rune{
		tcell.RuneLLCorner,
		tcell.RuneHLine,
		tcell.RuneBTee,
		tcell.RuneHLine,
		tcell.RuneLRCorner,
	}))

	width, height := scrn.Size()
	putf(scrn, reversedBlue, 0, row, "from putfln.  Screen width: %d, height: %d", width, height)
	row++
	putf(scrn, reverse, 1, 17, " testing putf with a %s", "string I just typed.")

	colors := scrn.Colors()
	putfln(scrn, " Number of colors is %d", colors)

	scrn.Show()

	prompt := "enter a word:"

	for i := 0; i < 30; i++ {
		puts(scrn, style, 1, height-1, prompt)
		str := GetInputString(scrn, len(prompt)+4, height-1)
		if len(str) == 0 {
			break // return from main is same as os.Exit(0)
		}
		//		putln(scrn, str)   don't need to duplicate lines anymore
		style = colorslice[i]
		putfln(scrn, "from putfln: %s", str)
		scrn.Show()
		if row >= height {
			break
		}
	}

	//	scrn.Fini()  I don't need this because I already deferred a scrn.Fini() right after a successful init.
}

// --------------------------------------------------- GetInputString --------------------------------------

func GetInputString(scrn tcell.Screen, x, y int) string {

	deleol(scrn, style, x, y)
	scrn.ShowCursor(x, y)
	scrn.Show()
	donechan := make(chan bool)
	keychannl := make(chan rune)
	helpchan := make(chan bool)
	delchan := make(chan bool)
	upchan := make(chan bool)
	downchan := make(chan bool)
	homechan := make(chan bool)
	endchan := make(chan bool)
	leftchan := make(chan bool)
	rightchan := make(chan bool)

	pollevent := func() {
		for {
			event := scrn.PollEvent()
			switch event := event.(type) {
			case *tcell.EventKey:
				switch event.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					donechan <- true // I don't have to send true to quit.
					return
				case tcell.KeyCtrlL:
					scrn.Sync()
				case tcell.KeyF1, tcell.KeyF2:
					// help
					helpchan <- true
					return

				case tcell.KeyBackspace, tcell.KeyDEL, tcell.KeyDelete:
					delchan <- true
					// do not return after any of these keys are hit, as an entry is being edited.

				case tcell.KeyPgUp, tcell.KeyUp:
					upchan <- true
					return

				case tcell.KeyPgDn, tcell.KeyDown:
					downchan <- true
					return

				case tcell.KeyRight, tcell.KeyUpRight, tcell.KeyDownRight:
					rightchan <- true
					return

				case tcell.KeyLeft, tcell.KeyUpLeft, tcell.KeyDownLeft:
					leftchan <- true
					return

				case tcell.KeyHome:
					homechan <- true
					return

				case tcell.KeyEnd:
					endchan <- true
					return

				case tcell.KeyRune:
					r := event.Rune()
					keychannl <- r
					if r == ' ' {
						return
					}
				}
			case *tcell.EventResize:
				scrn.Sync()
			}
		}
	}

	go pollevent()

	bs := make([]byte, 0, 100) // byteslice to build up the string to be returned.
	for {
		select {
		case <-donechan: // reading from quitchan will block until its closed.
			return string(bs)

		case <-helpchan:
			putfln(scrn, "help message received.  %s", "enter key is delimiter")
			return "help"

		case <-delchan:
			if len(bs) > 0 {
				bs = bs[:len(bs)-1]
			}
			puts(scrn, style, x+len(bs), y, " ")
			scrn.ShowCursor(x+len(bs), y)
			scrn.Show()

		case <-upchan:
			return "up key, pgup, upleft, upright"

		case <-downchan:
			return "down key, pgdn, downleft, downright"

		case <-homechan:
			return "home key"

		case <-endchan:
			return "end key"

		case <-rightchan:
			return "right arrow"

		case <-leftchan:
			return "left arrow"

		case key := <-keychannl:
			if key == ' ' {
				if len(bs) > 0 {
					return string(bs)
				} else {
					go pollevent() // need to restart the go routine to fetch more keys.
					continue       // discard this extaneous space
				}
			}
			bs = append(bs, byte(key))
			puts(scrn, style, x, y, string(bs))

			scrn.ShowCursor(x+len(bs), y)
			scrn.Show()
		}
	}
} // GetInputString
