package main

import (
	"fmt"
	"time"
)

// For the liveproject.  Starting to write routines to factor numbers.

func findFactors(num int) []int {
	if num <= 0 {
		return nil
	}

	factors := make([]int, 0, 10) // I just plucked 10 out of the air.
	for num%2 == 0 && num > 0 {   // This loop stops when num is odd.
		if num%2 == 0 {
			factors = append(factors, 2)
			num = num / 2
		}
	}
	if num < 2 { // this means that the number is even.
		return factors
	}
	fmt.Printf(" in findFactors after check for factors of 2:  num=%d, factors so far is/are: %v\n", num, factors)

	// Now all factors of 2 have been handled.  Handle larger factors.  I don't understand the book instructions at all.
	// Problems so far, the last factor is not getting appended to the slice.  And my sqrt routine doesn't work for small numbers.  I have to see what's a small number by trial and error.

	factor := 3
	numSQRT := iSqrt(num)
	fmt.Printf(" in findFactors before loop: factor=%d, num = %d, numSQRT=%d\n", factor, num, numSQRT)
	for factor <= numSQRT && num > 1 { // only need to check factors < sqrt(num)
		if num%factor == 0 {
			factors = append(factors, factor)
			num = num / factor
			fmt.Printf(" in found Factors loop: factor=%d, num = %d\n", factor, num)
			continue
		}
		// factor is not a factor for num, so increment factor and try again.
		factor += 2
		fmt.Printf(" in findFactors loop: factor=%d, num = %d\n", factor, num)
	}

	// If num is prime, return original num as the only factor other than 1, which is not returned.
	if num > 1 {
		factors = append(factors, num)
	}

	return factors
}

func multiplyFactors(factors []int) int {
	product := 1
	for _, fac := range factors {
		product *= fac
	}
	return product
}

func main() {
	var num int

	for {
		fmt.Printf(" number: ")
		n, err := fmt.Scanln(&num)
		//fmt.Printf(" n=%d\n", n)
		if err != nil || n == 0 {
			break
		}

		start := time.Now()

		factors := findFactors(num)
		elapsed := time.Since(start)
		check := multiplyFactors(factors)
		fmt.Printf(" Number = %d, factors = [%v], in %s, check = %d\n", num, factors, elapsed.String(), check)
	}
}

func iSqrt(i int) int { // this uses dividing and averaging
	if i <= 0 {
		return 0
	}

	//if i < 100 { // the rounding of integer division doesn't work well for small numbers.
	//	return i
	//}

	sqrt := i / 2

	for j := 0; j < 10; j++ {
		guess := i / sqrt
		sqrt = (guess + sqrt) / 2
		if sqrt-guess <= 1 { // recall that this is not floating math.
			break
		}
	}

	return sqrt + 1 // to address an off by 1 problem.
}
