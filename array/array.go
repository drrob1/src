package main

import (
         "fmt"
//         "os"
       );
import format "fmt"

func output(w []string) {
	for _, word := range w {
	  fmt.Printf("%s ",word);
	}
	fmt.Printf("\n");
}

func main() {
	words := [...]string {"the","quick","brown","fox"};  // array
	Words := [4] string {"the","quick","brown","fox"};   // array declared like in C
	wordslice := []string {"the","quick","brown","fox","jumped","over","the","lazy","dog"}; // slice
	fmt.Println(Words);
	fmt.Println(words[2]);
	fmt.Println(wordslice);
	output(words[0:2]);
	output(Words[:2]);
	output(wordslice[0:2]);
//	output(wordslice[]);  not allowed
	output(wordslice);
	output(wordslice[:]);
//	output(words);  type mismatch error because this is an array and the function needs a slice
//	output(Words);  same type mismatch error
	output(words[:]); // now this is a slice of an array
	format.Printf(" len of words is %d, len of Words is %d, len of wordslice is %d \n",len(words),
	  len(Words), len(wordslice));

//	dynamicwords := make([]string);  no initial elements
	dynamicwords := make([]string,4);  // 4 initial elements
	dynamicwords[0] = "the";
	dynamicwords[1] = "quick";
	dynamicwords[2] = "brown";
	dynamicwords[3] = "fox";
	output(dynamicwords);
//	dynamicwords[4] = "tilt -- runtime bounds check error"; // error shows this line #

	dynamicwords4 := make([]string,0,4);  // 0 initial elements, capacity of 4
	format.Println("Length of dynamicwords4 is ",len(dynamicwords4),", capacity is ", 
	  cap(dynamicwords4));
	dynamicwords4 = append(dynamicwords4,"jumped");
	dynamicwords4 = append(dynamicwords4,"over");
	dynamicwords4 = append(dynamicwords4,"the");
	dynamicwords4 = append(dynamicwords4,"lazy");
	output(dynamicwords4);
	format.Println("Length of dynamicwords4 is ",len(dynamicwords4),", capacity is ", 
	  cap(dynamicwords4));
	dynamicwords4 = append(dynamicwords4,"dog");
	format.Println("Length of dynamicwords4 is ",len(dynamicwords4),", capacity is ", 
	  cap(dynamicwords4));
// Copy
	wordscopy := make([]string,4);
//	copy(wordscopy,words);   // this is a type mismatch.  Does it work?  Nope
	copy(wordscopy,dynamicwords); 
	output(wordscopy);




}
