package main

import (
         "fmt"
         "os"
       );
import format "fmt"

func printthis(msg string) error {
	_, err := fmt.Printf("%s\n",msg);
	return err;
}

func printthat(msg string) (string,error) {
	msg += "\n";
	_, err := fmt.Printf(msg);
	return msg,err;
}

func printtofile(msg string) (e error) {          // e is declared to be of type error
	f, e := os.Create("helloworldfile.txt");
	if e != nil {
	  return;        // all named params will be returned here.  Don't need to list them.
	}
	defer f.Close();         // defer will wait and do this whenever the func returns, from any return point
	f.WriteString(msg);
	return;          // all named params will be automatically returned 
}

func print_this_variadic(msgs ...string) {
	for _,msg := range msgs {
	  fmt.Printf("%s\n",msg);
	}
}

func main() {
	fmt.Printf("Hello World. \n")
	fmt.Println("Hello World line 2.")
	format.Println("Called as format.Println")
        if n,err := fmt.Printf("Hello World. \n");  err != nil {
          os.Exit(1);
        }else{
          fmt.Printf(" Printed %d characters.\n",n);
        }

	PrintErr := printthis("uses the print_this function");
	fmt.Println(" The error code returned is ",PrintErr);

	appendedmsg,printerr := printthat("Uses the Print_That function");
	fmt.Printf(" The appended message is %q\n",appendedmsg);
        format.Println(" The error code returned is ",printerr);

	print_this_variadic("string"," messege ","printed ","to screen"," variadic");

	
}
