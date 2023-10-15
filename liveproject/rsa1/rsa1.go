package main

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

}

/*

PROCEDURE HCF(a,b : CARDINAL) : CARDINAL;
(*     a = bt + r, then hcf(a,b) = hcf(b,r)          *)

VAR
  r : CARDINAL;

BEGIN
  IF a < b THEN
                                                                                      (* C1 := a; a := b; b := C1; *)
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
