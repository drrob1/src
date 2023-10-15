package main

import "fmt"

// This is the crypto section, using RSA algorithm.  Will start w/ GCD and LCM, ie, greatest common factor and lowest common multiple.
// GCD is also called HCF, ie, highest common factor.

func gcd(a, b int) int {
	if a <= 0 || b <= 0 {
		return 0
	}

	if a < b { // swap 'em
		b, a = a, b
	}

	for {
		r := a % b
		//fmt.Printf(" a = %d, b = %d, r = %d\n", a, b, r)
		if r == 0 {
			return b
		}
		a = b
		b = r
	}
}

func hcf(a, b int) int {
	return gcd(a, b)
}

func lcm(a, b int) int { // lowest common multiple
	intermed := b / gcd(a, b) // to not cause an overflow.  Intermed will not have a remainder as this is a constraint of gcd.
	return a * intermed
}

func main() {
	var a, b int
	for {
		fmt.Printf(" enter A: ")
		n, err := fmt.Scanln(&a)
		if a < 1 || err != nil || n == 0 {
			break
		}
		fmt.Printf(" Enter B: ")
		n, err = fmt.Scanln(&b)
		if b < 1 || err != nil || n == 0 {
			break
		}
		g := gcd(a, b)
		m := lcm(a, b)
		fmt.Printf(" A = %d, B = %d, gcd = %d, lcm = %d\n", a, b, g, m)
	}
}

/*

PROCEDURE HCF(a,b : CARDINAL) : CARDINAL;
(*     a = bt + r, then hcf(a,b) = hcf(b,r)          *)

VAR
  r : CARDINAL;

BEGIN
  IF a < b THEN
    a := a BXOR b;
    b := a BXOR b;
    a := a BXOR b;
  END;
  REPEAT
    r := a MOD b;
    a := b;
    b := r;
  UNTIL r = 0;
  RETURN a;
END HCF;


*/
