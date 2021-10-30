package main

import "fmt"
import "time"

// I will now add code from Mastering Go, 2nd ed.

func main() {
	t := time.Now()
	fmt.Println(" default time format", t, ", t.String() is", t.String())
	fmt.Println(" t.GoString() [new for Go 1.17] is", t.GoString()) // compiles w/ Go 1.17 but is flagged as an error in Go 1.16.
	fmt.Println(" Unix format ", t.Format(time.UnixDate))
	fmt.Println(" ANSIC format ", t.Format(time.ANSIC))
	fmt.Println(" RFC3339 format ", t.Format(time.RFC3339))
	fmt.Println(" My format ", t.Format("Jan-02-2006 15:04:05"))

	fmt.Println("Epoch time:", time.Now().Unix())
	fmt.Println(t, t.Format(time.RFC3339))
	fmt.Println(t.Weekday(), t.Day(), t.Month(), t.Year())

	fmt.Println()
	time.Sleep(time.Second)
	t1 := time.Now()
	fmt.Println("Time difference:", t1.Sub(t))

	formatT := t.Format("01 January 2006")
	fmt.Println(formatT)
	parisLoc, _ := time.LoadLocation("Europe/Paris")
	parisTime := t.In(parisLoc)
	fmt.Println("Paris:", parisTime)
	fmt.Println("UTC:", t.UTC())

	londonLoc, _ := time.LoadLocation("Europe/London")
	londonTime := t.In(londonLoc)
	fmt.Println("London:", londonTime)


}
