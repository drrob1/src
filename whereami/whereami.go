package main;

import "fmt"
import "runtime"

func main() {
  fmt.Println(" Where am I?  I'm on:     ",runtime.GOOS);
}

/*
This will say either linux or windows, as by my tests just now.
*/

