package main

import "fmt"
import "time"

func main() {
	t := time.Now()
	fmt.Println(" default time format ", t)
	fmt.Println(" Unix format ", t.Format(time.UnixDate))
	fmt.Println(" ANSIC format ", t.Format(time.ANSIC))
	fmt.Println(" RFC3339 format ", t.Format(time.RFC3339))
	fmt.Println(" My format ", t.Format("Jan-02-2006 15:04:05"))
}
