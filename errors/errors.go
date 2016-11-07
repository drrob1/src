package main

import (
         "fmt"
         "os"
	 "errors"
       );
// import format "fmt"

var (
	errorEmptyString = errors.New("Empty string error");
)


// panic and recover are basically exceptions in Go.  But Go does not use these casually, as does C++,
// and he advises againt them.  The error system is what he recommends using.
func outputWithPanic(msg string) error {
	if msg == "" {
	  panic(errorEmptyString);
	}
	_, err := fmt.Println(msg);
	return err;
}

func Output(msg string) error {
	if msg == "" {
	  return errorEmptyString;
	}
	_, err := fmt.Println(msg);
	return err;
}


func output(msg string) error {
	if msg == "" {
	  return fmt.Errorf("Empty string error");
	}
	_, err := fmt.Println(msg);
	return err;
}



func main() {
	if err := output("Hello World."); err != nil {
	  fmt.Println(" Output routine error: ",err);
          os.Exit(1);
        }
	if err := Output(""); err != nil {
	  if err == errorEmptyString {
	    fmt.Println("I don't like empty strings.");
	  }else{
	    fmt.Println(" Output routine error: ",err);
            os.Exit(1);
	  }
        }
	if err := outputWithPanic(""); err != nil {
	  fmt.Println(" Will never see this if the program panicked.");
	}
}
