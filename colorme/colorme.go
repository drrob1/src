package main

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"time"
)

func main() {
	for i := 0; i < 2; i++ {
		ct.Foreground(ct.Yellow, false)
		fmt.Fprintln(ct.Writer, "I'm in yellow.")
		ct.ResetColor()
		time.Sleep(time.Second)

		fmt.Fprintln(ct.Writer, "Let's now do green.")
		time.Sleep(time.Second)
		ct.Foreground(ct.Green, false)
		fmt.Fprintln(ct.Writer, "I'm in green, not bright.")
		time.Sleep(time.Second)
		ctfmt.Println(ct.Green, true, "I'm in bright green.")
		time.Sleep(time.Second)

		ctfmt.Println(ct.Cyan, false, "I'm in cyan.")
		ctfmt.Println(ct.Cyan, true, "I'm in bright cyan.")
		time.Sleep(time.Second)

		fmt.Println(" I don't know what color I'll be in.  I'm using fmt.Println.  Maybe green?")

		ctfmt.Println(ct.Yellow, true, "I'm in bright yellow.")

		time.Sleep(time.Second)

		ctfmt.Println(ct.Red, false, "I'm in red.")
		ctfmt.Println(ct.Red, true, "I'm in bright red.")
		time.Sleep(time.Second)

		ctfmt.Println(ct.Blue, false, "I'm in blue.")
		ctfmt.Println(ct.Blue, true, "I'm in bright blue.")
		time.Sleep(time.Second)

		ctfmt.Println(ct.Magenta, false, "I'm in magenta.")
		ctfmt.Println(ct.Magenta, true, "I'm in bright magenta.")
		time.Sleep(2*time.Second)
	}
}
