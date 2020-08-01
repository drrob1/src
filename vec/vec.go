// IMPLEMENTATION MODULE Vec;
package vec

/*
                  Vector arithmetic

   Programmer:         P. Moylan
   Last edited:        15 August 1996
   Status:             Seems to be working

       Portability note: this module contains some
       open array operations which are a language
       extension in XDS Modula-2 but not part of
       ISO standard Modula-2.  So far I haven't worked
       out how to do this in the standard language.
*/

/*
 REVISION HISTORY
 ================
 20 Dec 2016 -- Started conversion to Go from old Modula-2 source.  We'll see how long this takes.
  1 Aug 2020 -- Cleaned up code and comments, esp comments.

*/

import (
	//  "os"
	//  "bufio"
	"fmt"
	//  "path/filepath"
	//  "strings"
	"strconv"
	//  "math"
	//  "math/cmplx"
	//  "math/rand"
	//  "time"
	//
	//  "getcommandline"
	//  "timlibg"
	//  "tokenize"
)

type VectorSlice []float64

/*
   CREATING VECTORS
*/

func NewVector(N int) VectorSlice {
	// Creates a vector of N elements.
	// old code basically did this: VAR result: VetorcPtr; NEW (result, N-1); RETURN result;

	vector := make(VectorSlice, N)
	return vector
}

// PROCEDURE DisposeVector (VAR (*INOUT*) V: VectorSlice;  N: CARDINAL);  BEGIN DISPOSE (V); END DisposeVector;

func Copy(A VectorSlice) VectorSlice {
	var B VectorSlice

	for i := range A {
		B[i] = A[i]
	}
	return B
}

/*
   VECTOR ARITHMETIC
*/

func Add(A, B VectorSlice) VectorSlice {
	// Computes C = A + B.

	var C VectorSlice

	if len(A) != len(B) {
		panic(" Vector Add and lengths of Vector A and Vector B are not the same.")
	}

	for i := range A {
		C[i] = A[i] + B[i]
	}
	return C
}

/************************************************************************/

func Sub(A, B VectorSlice) VectorSlice {
	// Computes C = A - B.

	var C VectorSlice

	if len(A) != len(B) {
		panic(" Vector Sub and lengths of Vector A and Vector B are not the same.")
	}

	for i := range A {
		C[i] = A[i] - B[i]
	}
	return C
}

/************************************************************************/

func Mul(A []VectorSlice, B VectorSlice) VectorSlice {
	// Computes C = A*B, where A is N1xN2 and B is N2x1.
	// C = A*B, where A is a 2D matrix, and B is a column vector, so result must also be a column vector

	var sum float64
	var C VectorSlice

	for i := range A { // FOR i := 0 TO N1-1 DO  range over the 2D matrix
		sum = 0
		for j := range B { //   FOR j := 0 TO N2-1 DO  range over the column vector
			sum += A[i][j] * B[j]
		} // END FOR j
		C[i] = sum
	} // END FOR i
	return C
} //    END Mul;

/************************************************************************/

func ScalarMul(A float64, B VectorSlice) VectorSlice {
	// Computes C = A*B, where A is scalar and B is a vector

	var C VectorSlice

	for i := range B { // FOR i := 0 TO elts-1 DO  range over the B vector
		C[i] = A * B[i]
	} // END FOR i
	return C
} //    END ScalarMul;

/*
   OUTPUT
*/

func Write(V VectorSlice, places int) []string {
	// Writes the N-element vector V to a string slice (formerly the screen), where each column occupies a field "places" characters wide.

	OutputStringSlice := make([]string, 0, 20)
	for i := range V {
		ss := strconv.FormatFloat(float64(V[i]), 'G', places, 64)
		OutputStringSlice = append(OutputStringSlice, fmt.Sprintf("  %s", ss))
	}
	OutputStringSlice = append(OutputStringSlice, "\n")
	return OutputStringSlice
} // END Write;

// END Vec.
