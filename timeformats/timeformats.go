package main

import (
	"fmt"
	"time"
)

/*
_2 Mar 25 -- I saw a use of time formats in Mastering Go, 4th ed, and I want to understand them more.
*/
func main() {
	t850 := time.Now().Format(time.RFC850)
	t1123 := time.Now().Format(time.RFC1123)
	tAnsic := time.Now().Format(time.ANSIC)
	tUnix := time.Now().Format(time.UnixDate)
	tRuby := time.Now().Format(time.RubyDate)
	fmt.Printf(" RFC850 %s\n RFC1123 %s\n Ansic %s\n tUnix %s\n tRuby %s\n", t850, t1123, tAnsic, tUnix, tRuby)

	t822z := time.Now().Format(time.RFC822Z)
	t822 := time.Now().Format(time.RFC822)
	t1123z := time.Now().Format(time.RFC1123Z)
	t3339 := time.Now().Format(time.RFC3339)
	tKitchen := time.Now().Format(time.Kitchen)
	fmt.Printf(" RFC822z %s\n RFC822 %s\n RFC1123z %s\n RFC3339 %s\n Kitchen %s\n", t822z, t822, t1123z, t3339, tKitchen)

	tStamp := time.Now().Format(time.Stamp)
	tDateTime := time.Now().Format(time.DateTime)
	fmt.Printf(" Stamp %s\n DateTime %s\n", tStamp, tDateTime)
}
