package gcd

func HCF(a, b int) int {
	// a = bt + r, then hcf(a,b) = hcf(b,r)
	var r, a1, b1 int

	if a < b {
		a1 = b
		b1 = a
	} else {
		a1 = a
		b1 = b
	}
	for {
		r = a1 % b1 // % is MOD operator
		a1 = b1
		b1 = r
		if r == 0 {
			break
		}
	}
	return a1
} // HCF

func GCD(a, b int) int {
	// a = bt + r, then hcf(a,b) = hcf(b,r)

	var r int

	if a < b {
		a, b = b, a
	}

	for {
		r = a % b
		if r == 0 {
			break
		}
		a = b
		b = r
	}
	return b
}
