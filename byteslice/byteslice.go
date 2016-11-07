package main

import (
         "fmt"
         "os"
       );
import format "fmt"

func main() {
	f,err := os.Open("test.txt");
	if err != nil {
	  format.Println(" File open error is ",err);
	  os.Exit(1);
	}
	defer f.Close();

	b := make([]byte, 1024);  // this is the definition of the byteslice

	n, err := f.Read(b);

	fmt.Printf("%d: % x \n",n,b[:n]);
	fmt.Printf("%d: % c \n",n,b[:n]);

	stringversion := string(b);
	fmt.Printf("%d: %s\n",n,stringversion[:n]);

// converting back to a byte slice for writing to a file.
// He doesn't go this far, but I am.
	outfile,err := os.Create("outfile.txt");  // Open only works for file that already exists.
	if err != nil {
	  format.Println(" File open error is ",err);
	  os.Exit(1);
	}
	defer outfile.Close();

	outString := "any string";
	outfile.Write( []byte(outString) );

}
