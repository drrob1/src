package main;

import (
        "fmt"
       )

func main() {
  var x complex128 = complex(1,2); // 1 +2i
  var y complex128 = complex(3,4); // 3 +4i

  fmt.Println(" x*y is ",x*y,", x+y is ",x+y,", real part of product is ",real(x*y),
                                             ", imag part of product is ",imag(x*y));
}
