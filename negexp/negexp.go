package main

import (
	"fmt"
	"math"
)

/*                                                                                                -n      n
This is to see if math.Pow works as I expected that 1/n is same as -n.  Nope, I got that wrong.  a   = 1/a , ie, the reciprocal is not in the exponent but the entire expression.

 -n    1
a   = ----
       n
      a

9 Sep 24 -- First version.
*/

const base = 4.

func main() {
	rootBase := math.Pow(base, 1./2.)
	negBase := math.Pow(base, -2.)
	fmt.Printf(" %g to power of 1/2 is %f, and to power of -2 is %f\n", base, rootBase, negBase)

}
