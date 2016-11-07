package main;

import ("fmt");

func main () {
  atoz := "The Quick Brown Fox jumps over the lazy dog.";

  for i, r := range(atoz) {
    fmt.Printf("%d %c \n",i,r);
  }
  fmt.Printf(" Length of atoz is %d characters\n",len(atoz));
}

