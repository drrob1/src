package main

import "fmt"
import "time"

// I will now add code from Mastering Go, 2nd ed.

func main() {
	t := time.Now()
	fmt.Println(" default time format ", t)
	fmt.Println(" Unix format ", t.Format(time.UnixDate))
	fmt.Println(" ANSIC format ", t.Format(time.ANSIC))
	fmt.Println(" RFC3339 format ", t.Format(time.RFC3339))
	fmt.Println(" My format ", t.Format("Jan-02-2006 15:04:05"))

	fmt.Println("Epoch time:", time.Now().Unix())
	fmt.Println(t, t.Format(time.RFC3339))
	fmt.Println(t.Weekday(), t.Day(), t.Month(), t.Year())

	time.Sleep(time.Second)
	t1 := time.Now()
	fmt.Println("Time difference:", t1.Sub(t))

	formatT := t.Format("01 January 2006")
	fmt.Println(formatT)
	loc, _ := time.LoadLocation("Europe/Paris")
	londonTime := t.In(loc)
	fmt.Println("Paris:", londonTime)

}
