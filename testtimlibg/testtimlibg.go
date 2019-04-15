package main

import (
	"fmt"
	//        "time"
	"timlibg"
)

// need to test the TIME2MDY, GetDateTime, MDY2STR, JULIAN, GREGORIAN functions.

func main() {

	MM, DD, YY := timlibg.TIME2MDY()
	fmt.Println(" Testing TIME2MDY.  month is ", MM, ", day is ", DD, ", year is ", YY, ".")
	fmt.Println()

	dt := timlibg.GetDateTime()
	fmt.Printf(" GetDateTime returns dt \n %#v : ", dt)
	fmt.Println()

	mdystr := timlibg.MDY2STR(MM, DD, YY)
	fmt.Printf(" After calling MDY2STR on today as returned from TIME2MDY.  The string is %#v  and its type is %T\n", mdystr, mdystr)

	// still need to test JULIAN and GREGORIAN.  To follow.

	Today := timlibg.JULIAN(MM, DD, YY)
	fmt.Println(" Julian date for today is : ", Today)
	fmt.Println()

	mm, dd, yy := timlibg.GREGORIAN(Today)
	fmt.Println(" After GREGORIAN, today is : ", mm, "/", dd, "/", yy)
	fmt.Println()
}
