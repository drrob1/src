// Copyright 2019 The TCell Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use file except in compliance with the License.
// You may obtain a copy of the license at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// unicode just displays a Unicode test on your screen.
// Press ESC to exit the program.
package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
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

	encoding.Register()

	if e = scrn.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	plain := tcell.StyleDefault
	bold := style.Bold(true)

	scrn.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))
	scrn.Clear()

	quitchan := make(chan bool)
	keychan := make(chan keyStructType)

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

	scrn.Show()
	pollevent := func() {
		var keystruct keyStructType
		for {
			event := scrn.PollEvent()
			switch event := event.(type) {
			case *tcell.EventKey:
				switch event.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					//close(quitchan)
					quitchan <- true
					return
				case tcell.KeyCtrlL:
					scrn.Sync()
				case tcell.KeyRune:
					keystruct.r = event.Rune()
					keystruct.name = event.Name()
					keychan <- keystruct
				}
			case *tcell.EventResize:
				scrn.Sync()
			}
		}
	}

	go pollevent()

	for {
		select {
		case <-quitchan: // reading from quitchan will block until its closed.
			scrn.Fini()
			return
		case key := <-keychan:
			str := fmt.Sprintf("full name is %s, rune number is %d, rune quoted is %q, rune as string is %s", key.name, key.r, key.r, string(key.r))
			putln(scrn, str)
			scrn.Show()
		}
	}

}
