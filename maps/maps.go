package main

import (
	"fmt"
	//         "os"
)
import format "fmt"

func main() {

	daymonths := make(map[string]int) // dynamic definition using make.

	daymonths["Jan"] = 31
	daymonths["Feb"] = 28
	daymonths["Mar"] = 31
	daymonths["Apr"] = 30
	daymonths["May"] = 31
	daymonths["Jun"] = 30
	daymonths["Jul"] = 31
	daymonths["Aug"] = 31
	daymonths["Sep"] = 30
	daymonths["Oct"] = 31
	daymonths["Nov"] = 30
	daymonths["Dec"] = 31
	daymonths["extra"] = 99

	fmt.Println(" Days in January are ", daymonths["Jan"])
	format.Println(" Days in February are ", daymonths["Feb"])
	fmt.Println(" There are ", len(daymonths), " elements in the daymonths map.")

	// comma ok syntax example
	days, ok := daymonths["January"]
	fmt.Println(" Days for January index is ", days, ".  ok value is ", ok)

	days, ok = daymonths["Mar"]
	fmt.Println(" Days for Mar index is ", days, ".  ok value is ", ok)

	// for loop outputs in random order.  To get a particular order, need to sort, to be covered
	// later.
	for month, day := range daymonths {
		format.Println(month, " has ", day, " days.")
	}

	var day31 int
	for _, days = range daymonths {
		if days == 31 {
			day31++
		}
	}
	fmt.Println(day31, " months have 31 days.")

	// the delete built in function.
	delete(daymonths, "extra")
	fmt.Println(" There are ", len(daymonths), " elements in the daymonths map.")

	// static definition of a map using a different syntax

	DayMonths := map[string]int{
		"Jan": 31,
		"Feb": 28,
		"Mar": 31,
		/* etc */
	}
	format.Println(" Jan has", DayMonths["Jan"], " days, using the static map.")

}
