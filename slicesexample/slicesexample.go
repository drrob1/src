package main

import (
	"fmt"
	"os"
)
import format "fmt"

func main() {
	fmt.Printf("Hello World. \n")
	fmt.Println("Hello World line 2.")
	format.Println("Called as format.Println")
	if n, err := fmt.Printf("Hello World. \n"); err != nil {
		os.Exit(1)
	} else {
		fmt.Printf(" Printed %d characters.\n", n)
	}

	//added Jan 26, 2020, based on a post from gonuts group

	s := make([]int,3)
	s[0] = 100
	s[1] = 200
	s[2] = 300

	t := append(s,400)  // this has 4 elements now, so it's a different backing array than s
	u := append(t,500)

	fmt.Println("s:",s,", s.len",len(s), ", cap s", cap(s), ", t:", t, ", len, cap t",len(t),cap(t),", u, len, cap", u, len(u), cap(u) )

	u[0] = 0  // also changes t[0], as they have same backing array.

	fmt.Println("s:",s,", s.len",len(s), ", cap s", cap(s), ", t:", t, ", len, cap t",len(t),cap(t),", u, len, cap", u, len(u), cap(u) )

	t[2] = 2  // also changes u[2]

	fmt.Println("s:",s,", s.len",len(s), ", cap s", cap(s), ", t:", t, ", len, cap t",len(t),cap(t),", u, len, cap", u, len(u), cap(u) )

	s[1] = 1  // only changes s

	fmt.Println("s:",s,", s.len",len(s), ", cap s", cap(s), ", t:", t, ", len, cap t",len(t),cap(t),", u, len, cap", u, len(u), cap(u) )
}
