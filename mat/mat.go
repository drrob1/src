package mat

/********************************************************)
  (*                                                      *)
  (*                 Matrix arithmetic                    *)
  (*   We can handle matrices with up to 8191 elements    *)
  (*                                                      *)
  (*  Programmer:         P. Moylan                       *)
  (*  Last edited:        15 August 1995                  *)
  (*  Status:             OK                              *)
  (*                                                      *)
  (*      Portability problem: I've had to use an XDS     *)
  (*      language extension (open arrays) here; I        *)
  (*      haven't yet figured out how to do the job       *)
  (*      in ISO standard Modula-2.                       *)
  (*                                                      *)
  (********************************************************/

// REVISION HISTORY
// ================
// 19 Dec 16 -- Started conversion to Go from old Modula-2 source.  We'll see how long this takes.
// 24 Dec 16 -- Passed mattest.
// 25 Dec 16 -- Changed the code to use the Go swapping idiom

import (
	//  "os"
	//  "bufio"
	//  "path/filepath"
	//  "strings"
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"
	"strconv"
	"time"
	//
	"vec"
	//  "getcommandline"
	//  "timlibg"
	//  "tokenize"
)

const small = 1.0E-10
const SubscriptDim = 8192
const SizeFudgeFactor = 20 // decided to not use this in NewMatrix

// type EltType float64;  This is what it was
// type EltType vec.EltType;  // just imports and renames the type as defined in vec.go
// type Matrix2D [][]EltType;     // Array of Array of EltType;  Nevermind that it's a slice of slice
//  type VectorPtr []EltType;  Defined in vec.go
//  type ArrayPtr *EltType2Dim; //POINTER TO ARRAY OF ARRAY OF EltType;  Don't think I need this anyway.

type Matrix2D [][]float64
type Permutation []int

//TYPE subscript = [0..8191];
//type Permutation = POINTER TO ARRAY subscript OF subscript;  array [0..8191] OF subrange of integer.
//I'm going to ignore this subrange of integer called subscript, and just make it int.

type LongComplexSlice []complex128 // using Modula-2 name for this type as a userdefined type

func init() {
	rand.Seed(time.Now().UnixNano())
}

/************************************************************************)
(*                   CREATING AND DESTROYING MATRICES                   *)
(************************************************************************/

func NewMatrix(R, C int) Matrix2D { // I think row, column makes more sense than N x M
	// Creates an NxM matrix.
	// old code basically did this: NEW (result, N-1, M-1); RETURN result;

	if (R > SubscriptDim) || (C > SubscriptDim) || R < 1 || C < 1 {
		return nil
	}
	matrix := make(Matrix2D, R)
	for i := range matrix {
		matrix[i] = make([]float64, C)
	}
	return matrix
}

// PROCEDURE DisposeArray (VAR (*INOUT*) V: ArrayPtr;  N, M: CARDINAL);
// Deallocates an NxM matrix is not needed, because of the garbage collection.
// It just did DISPOSE (V);
/************************************************************************)
(*                          ASSIGNMENTS                                 *)
(************************************************************************/

//                                 PROCEDURE Zero (VAR (*OUT*) M: ARRAY OF ARRAY OF EltType;  r, c: CARDINAL);
func Zero(matrix Matrix2D) Matrix2D {
	// It zeros an already defined r by c matrix.  I'm not sure this is needed in Go, but here it is.

	for r := range matrix {
		for c := range matrix[r] {
			matrix[r][c] = 0
		}
	}
	return matrix
}

func BelowSmallMakeZero(matrix Matrix2D) Matrix2D {
	for r := range matrix {
		for c := range matrix[r] {
			if math.Abs(matrix[r][c]) < small {
				matrix[r][c] = 0
			}
		}
	}
	return matrix
}

//                                     PROCEDURE Unit (VAR (*OUT*) M: ARRAY OF ARRAY OF EltType;  N: CARDINAL);
func Unit(matrix Matrix2D) Matrix2D {
	// Creates an N by N identity matrix, with all zeros except along the main diagonal.
	matrix = Zero(matrix)
	for diag := range matrix {
		matrix[diag][diag] = 1
	}
	return matrix
}

//                                PROCEDURE Random (VAR (*OUT*) M: ARRAY OF ARRAY OF EltType;  r, c: CARDINAL);
func Random(matrix Matrix2D) Matrix2D {
	// Creates matrix with random integers from 0..100

	for r := range matrix {
		for c := range matrix[r] {
			matrix[r][c] = float64(rand.Intn(100))
		}
	}
	return matrix
}

//     PROCEDURE Copy (A: ARRAY OF ARRAY OF EltType;  r, c: CARDINAL; VAR (*OUT*) B: ARRAY OF ARRAY OF EltType);
func Copy2(Src, Dest Matrix2D) {
	// Copies an r x c matrix A to B, by doing an element by element copy.  I don't think just copying
	// pointers is correct.

	SrcRows := len(Src)
	SrcCols := len(Src[0])
	DestRows := len(Dest)
	DestCols := len(Dest[0])

	if (SrcRows != DestRows) || (SrcCols != DestCols) {
		fmt.Println(" Src and Dest are not same size.  Copy aborted.")
		Dest = Zero(Dest)
		return
	}

	for r := range Src {
		for c := range Src[r] {
			Dest[r][c] = Src[r][c]
		}
	}
}

func Copy(Src Matrix2D) Matrix2D {
	// Copies an r x c matrix A to B, by doing an element by element copy.  I don't think just copying
	// pointers is correct.

	SrcRows := len(Src)
	SrcCols := len(Src[0])

	Dest := NewMatrix(SrcRows, SrcCols)
	for r := range Src {
		for c := range Src[r] {
			Dest[r][c] = Src[r][c]
		}
	}
	return Dest
}

/************************************************************************)
(*                      THE BASIC MATRIX OPERATIONS                     *)
(************************************************************************/

// ----------------------------------------------------------------------------- Add -------------------

//  PROCEDURE Add (A, B : ARRAY OF ARRAY OF EltType;  r, c: CARDINAL; VAR (*OUT*) C: ARRAY OF ARRAY OF EltType);
func Add(A, B Matrix2D) Matrix2D {
	// Computes C = A + B.
	var C Matrix2D

	Arows := len(A)
	Acols := len(A[0])
	Brows := len(B)
	Bcols := len(B[0])
	if (Arows != Brows) || (Acols != Bcols) {
		return nil
	}

	C = NewMatrix(Brows, Acols) // Could have been either row and either col.  I chose those.

	for i := range A {
		for j := range A[i] {
			C[i][j] = A[i][j] + B[i][j]
		}
	}
	return C
}

// ----------------------------------------------------------------------------- Sub -------------------

//   PROCEDURE Sub (A, B: ARRAY OF ARRAY OF EltType;  r, c: CARDINAL; VAR (*OUT*) C: ARRAY OF ARRAY OF EltType);
func Sub(A, B Matrix2D) Matrix2D {
	// Computes C = A - B.
	//  var C Matrix2D;

	Arows := len(A)
	Acols := len(A[0])
	Brows := len(B)
	Bcols := len(B[0])
	if (Arows != Brows) || (Acols != Bcols) {
		return nil
	}

	C := NewMatrix(Arows, Bcols) // Could have been either row and either col.  I chose those.

	for i := range A {
		for j := range A[i] {
			C[i][j] = A[i][j] - B[i][j]
		}
	}
	return C
}

// -------------------------------------------------------------------------------- MUL -------------------

//   PROCEDURE Mul (A, B: ARRAY OF ARRAY OF EltType;  r, c1, c2: CARDINAL; VAR (*OUT*) C: ARRAY OF ARRAY OF EltType);
func Mul(A, B Matrix2D) Matrix2D {
	// Computes C = A x B.  Using std linear algebra rules.  In orig code A[r,c1], B[c1,c2]
	//    var C Matrix2D;
	var temp float64

	NumRowA := len(A)
	NumColA := len(A[0]) // all rows have same number of columns
	NumRowB := len(B)
	NumColB := len(B[0]) // all rows have same number of columns

	if NumColA != NumRowB {
		return nil
	}

	C := NewMatrix(NumRowA, NumColB)

	for i := range A { // ranging over number of rows of A
		for j := range B[0] { //  and also number of col of B
			temp = 0
			for k := range B { // ranging over number of rows of B
				temp += A[i][k] * B[k][j]
			}
			C[i][j] = temp
		}
	}
	return C
}

//  PROCEDURE ScalarMul (A: EltType;  B: ARRAY OF ARRAY OF EltType;  r, c: CARDINAL; VAR (*OUT*) C: ARRAY OF ARRAY OF EltType);
func ScalarMul(a float64, B Matrix2D) Matrix2D {
	// Computes C = a*B, where a is the scalar and B is the matrix.

	Brows := len(B)
	Bcols := len(B[0])

	C := NewMatrix(Brows, Bcols)

	for i := range B {
		for j := range B[i] {
			C[i][j] = a * B[i][j]
		}
	}
	return C
}

/************************************************************************)
(*                      SOLVING LINEAR EQUATIONS                        *)
(************************************************************************/

// -------------------------------------------------------------------------------- LUFactor ------------------------------------------------

//          PROCEDURE LUFactor (VAR (*INOUT*) A: ARRAY OF ARRAY OF EltType;  N: CARDINAL; perm: Permutation;  VAR (*OUT*) oddswaps: BOOLEAN);

func LUFactor(A Matrix2D, perm Permutation) (Matrix2D, bool) { // A is an InOut param.

	/*
	   LU decomposition of a square matrix.  We express A in the form P*L*U, where P is a permutation matrix,
	   L is lower triangular with unit diagonal elements, and U is upper triangular.  This is an in-place
	   computation, where on exit U occupies the upper triangle of A, and L (not including its diagonal entries)
	   is in the lower triangle.  The permutation information is returned in  perm.  Output parameter oddswaps
	   is TRUE iff an odd number of row interchanges were done by the permutation.
	   (We need to know this only if we are going to go on to calculate a determinant.)

	   The precise relationship between the implied permutation matrix P and the output parameter perm is somewhat
	   obscure.  The vector perm^ is not simply a permutation of the subscripts [0..N-1]; it does, however, have
	   the property that we can recreate P by walking through perm^ from start to finish, in the order used by
	   procedure LUSolve.
	   Sample use, as see in Solve below is: LU,s = LUFactor(LU,N,perm)
	*/

	var pivotrow int
	var sum, temp, maxval float64

	N := len(A)
	VV := vec.NewVector(N) // I anticipate this will return a slice, not a pointer.  So dereferencing syntax is not needed.
	oddswaps := false

	// Start by collecting (in VV), the maximum absolute value in each row; we'll use this for pivoting decisions.

	for row := range A {
		maxval = 0
		for col := range A[row] {
			if math.Abs(A[row][col]) > maxval {
				maxval = math.Abs(A[row][col])
			}
		}
		if maxval == 0 { // Go treats 0 as an abstract number that will match the type at runtime.
			// An all-zero row can never contribute pivot elements.
			VV[row] = 0 //              VV^[row] := 0.0;
		} else {
			VV[row] = 1 / maxval //        VV^[row] := 1.0/maxval;
		}
	}

	/* Crout's method: we work through one column at a time. */

	for col := range A {

		/* Upper triangular component of this column - except for the diagonal element, which we leave until after we've   selected a pivot from on or below the diagonal.          */

		if col > 0 {
			for row := 0; row < col; row++ { // FOR row 0 .. col-1
				sum = A[row][col]
				if row > 0 {
					for k := 0; k < row; k++ { // for k is 0 TO row-1 DO
						sum -= A[row][k] * A[k][col]
					} // END FOR k 0 to row-1;
				} // END IF;
				A[row][col] = sum
			} // END for row 0 to col-1
		} // END if col > 0

		// Lower triangular component in this column.  The results we get in this loop will not be correct until we've divided by the pivot; but we work out the pivot location as we go, and come back later for this division.

		maxval = 0
		pivotrow = col
		for row := col; row < N; row++ { // for row from col to N-1
			sum = A[row][col]
			if col > 0 {
				for k := 0; k < col; k++ { // for k from 0 to col-1
					sum -= A[row][k] * A[k][col]
				} //END FOR k from 0 to col-1
			} // END IF col>0
			A[row][col] = sum
			temp := VV[row] * math.Abs(sum) // temp := VV^[row] * ABS(sum);
			if temp >= maxval {
				maxval = temp
				pivotrow = row
			} // END IF temp>=maxval
		} // END FOR row from col to N-1

		// If pivot element was not already on the diagonal, do a row swap.

		if pivotrow != col {
			for k := 0; k < N; k++ { //  k from 0 to N-1
				// temp = A[pivotrow][k]; A[pivotrow][k] = A[col][k]; A[col][k] = temp;
				A[col][k], A[pivotrow][k] = A[pivotrow][k], A[col][k] // use the go idiom for swap
			} // END FOR k from 0 to N-1
			oddswaps = !oddswaps
			VV[pivotrow] = VV[col] // VV^[pivotrow] := VV^[col];

			// We don't bother updating VV^[col] here, because its value will never be used again.

		} // END IF pivotrow != col
		perm[col] = pivotrow //        perm^[col] := pivotrow;

		//* Finish off the calculation of the lower triangular part for this column by scaling by the pivot A[col,col].

		// Remark: if the pivot is still zero at this stage, then all the elements below it are also zero.  The LU
		// decomposition in this case is not unique - the original matrix is singular, therefore U will also be
		// singular -- but one solution is to leave all those elements zero.

		temp = A[col][col]
		if (col != N-1) && (temp != 0.0) {
			temp = 1 / temp
			for row := col + 1; row < N; row++ { // row from col+1 to N-1
				A[row][col] = temp * A[row][col]
			} // END FOR row from col+1 to N-1
		} // END IF col != N-1

	} // END FOR col range A

	return A, oddswaps
} // END LUFactor

/************************************************************************/

// ------------------------------------------------------------------------------------- LUSolve ---------------------------------------

//     PROCEDURE LUSolve (LU: ARRAY OF ARRAY OF EltType; VAR (*INOUT*) B: ARRAY OF ARRAY OF EltType; N, M: CARDINAL;  perm: Permutation);
//     func LUSolve (LU, B Matrix2D, N, M int, perm Permutation) Matrix2D {  // the syntax for this line is correct.

func LUSolve(LU, B Matrix2D, perm Permutation) Matrix2D { // B is an InOut param

	/* Solves the equation P*L*U*X = B, where P is a permutation        *)
	   (* matrix specified indirectly by perm; L is lower triangular; and  *)
	   (* U is upper triangular.  The "Matrix" LU is not a genuine matrix, *)
	   (* but rather a packed form of L and U as produced by procedure     *)
	   (* LUfactor above.  On return B is replaced by the solution X.      *)
	   (* Dimensions: left side is NxN, B is NxM.                          *)
	   (* Sample use: X = LUSolve(LU,X,N,M,perm)                           */

	var sum, scale float64

	/* Pass 1: Solve the equation L*Y = B (at the same time sorting *)
	   (* B in accordance with the specified permutation).  The        *)
	   (* solution Y overwrites the original value of B.               *)

	   (* Understanding how the permutations work is something of a    *)
	   (* black art.  It helps to know that (a) ip>=i for all i, and   *)
	   (* (b) in the summation over k below, we are accessing only     *)
	   (* those elements of B that have already been sorted into the   *)
	   (* correct order.                                               */

	N := len(B)
	//	M = len(B[0];  Not used, it turns out, after I use the range syntax.  Looks like this is translated from C, because even classic M-2 does not need the size of the array to be passed as a separate
	//	param.

	for i := range B { // for i from 0 to N-1
		ip := perm[i]         //    ip := perm^[i];
		for j := range B[i] { //     for j from 0 to M-1
			sum = B[ip][j]
			B[ip][j] = B[i][j]
			if i > 0 {
				for k := 0; k < i; k++ { // for k from 0 to i-1
					sum -= LU[i][k] * B[k][j]
				} // END FOR k from 0 to i-1
			} // END IF i>0
			B[i][j] = sum
		} // END FOR j from 0 to M-1
	} // END FOR i from 0 to N-1

	// Pass 2: solve the equation U*X = Y.

	for i := N - 1; i >= 0; i-- { // for i from N-1 to 0 by -1
		scale = LU[i][i]
		if scale == 0 {
			//  Matrix is singular.  Aborting.
			return nil
		} //END IF scale == 0
		for j := range B[i] { // for j from 0 to M-1
			sum = B[i][j]
			for k := i + 1; k < N; k++ { // for K from i+1 to N-1
				sum -= LU[i][k] * B[k][j]
			} // END FOR k from i+1 to N-1
			B[i][j] = sum / scale
		} // END FOR j from 0 to M-1
	} // END FOR i from N-1 to 0 by -1

	return B
} // END LUSolve;

// -------------------------------------------------------------------------------- GaussJ ----------------------------

//        PROCEDURE GaussJ (A, B: ARRAY OF ARRAY OF EltType; VAR (*OUT*) X: ARRAY OF ARRAY OF EltType; N, M: CARDINAL);

//        func GaussJ (A, B Matrix2D, N, M int) Matrix2D {  // X is the output matrix
func GaussJ(A, B Matrix2D) Matrix2D { // X is the output matrix

	/*
	   Solves the equation AX = B by Gauss-Jordan elimination.  In the present version A must be square and
	   nonsingular.
	   This approach to solving the equation is not the best available -- see below -- but is included here
	   anyway since it is popular.
	   Dimensions: A is NxN, B is NxM.
	*/

	//    VAR W: ArrayPtr;  i, j, k, prow: CARDINAL;
	//        pivot, scale, temp: EltType;
	//var W Matrix2D;
	var pivot float64
	var X Matrix2D

	N := len(A)
	//  M := len(B[0]);

	W := NewMatrix(N, N)
	Copy2(A, W) //        Copy (A, N, N, W^);
	X = Copy(B) //        Copy (B, N, M, X);

	/*
	   Remark: we are going to use elementary row operations to turn W into a unit matrix.  However we don't
	   bother to store the new 1.0 and 0.0 entries, because those entries will never be fetched again.
	   We simply base our calculations on the assumption that those values have been stored.

	   Dimensions: A is N x N, B is N x M.

	   Pass 1: by elementary row operations, make W into an upper triangular matrix.
	*/

	prow := 0
	for i := range W { // FOR i := 0 TO N-1 DO
		pivot = 0.0
		for j := i; j < N; j++ { // FOR j := i TO N-1 DO
			temp := W[j][i] // temp := W^[j,i];
			if math.Abs(temp) > math.Abs(pivot) {
				pivot = temp
				prow = j
			} // END IF temp > pivot
		} // END FOR j from i to N-1
		if math.Abs(pivot) < small { // Coefficient matrix is singular.  Aborting,
			return nil
		} // END IF pivot < small

		// Swap rows i and prow.

		for j := i; j < N; j++ { // FOR j := i TO N-1 DO
			// temp := W[i][j]; W[i][j] = W[prow][j]; W[prow][j] = temp;
			W[i][j], W[prow][j] = W[prow][j], W[i][j] // Go swapping idiom.
		} // END FOR j from i to N-1

		for j := range X[i] { // FOR j := 0 TO M-1 DO
			// temp := X[i][j]; X[i][j] = X[prow][j]; X[prow][j] = temp;
			X[i][j], X[prow][j] = X[prow][j], X[i][j] // Go swapping idiom
		} // END FOR j from 0 to M-1

		// Scale the i'th row of both W and X.

		for j := i + 1; j < N; j++ { // FOR j := i+1 TO N-1 DO
			W[i][j] = W[i][j] / pivot //  W^[i,j] := W^[i][j]/pivot;
		} // END FOR j from i+1 to N-1
		for j := range X[i] { // FOR j := 0 TO M-1 DO
			X[i][j] = X[i][j] / pivot
		} // END FOR j from 0 to M-1

		// Implicitly reduce the sub-column below W[i,i] to zero.

		for k := i + 1; k < N; k++ { // FOR k := i+1 TO N-1 DO
			scale := W[k][i]
			for j := i + 1; j < N; j++ { //  FOR j := i+1 TO N-1 DO
				W[k][j] -= scale * W[i][j] // W was dereferenced in M-2 code
			} // END FOR j from i+1 to N-1
			for j := range X[i] { // FOR j := 0 TO M-1 DO
				X[k][j] -= scale * X[i][j]
			} // END FOR j from 0 to M-1
		} // END FOR k from i+1 to N-1

	} // END FOR i from 0 to N-1

	// Pass 2: reduce the above-diagonal elements of W to zero.

	for i := N - 1; i > 0; i-- { // FOR i := N-1 TO 1 BY -1 DO

		// Implicitly reduce the sub-column above W[i,i] to zero.

		for k := 0; k < i; k++ { // FOR k := 0 TO i-1 DO
			scale := W[k][i]
			for j := range X[i] { // FOR j := 0 TO M-1 DO
				X[k][j] -= scale * X[i][j]
			} // END FOR j from 0 to M-1
		} // END FOR k from 0 to i-1

	} // END FOR i from N-1 to 1 BY -1

	// DisposeArray (W, N, N);  Not needed in Go.

	return X
} //    END GaussJ;

// --------------------------------------------------------------------------------- Solve ---------------------------

//        PROCEDURE Solve (A, B: ARRAY OF ARRAY OF EltType; VAR (*OUT*) X: ARRAY OF ARRAY OF EltType; N, M: CARDINAL);
//        func Solve (A, B Matrix2D, N, M int) Matrix2D {  // return X

func Solve(A, B Matrix2D) Matrix2D { // return X

	// Solves the equation AX = B.  In the present version A must be square and nonsingular.

	// Dimensions: A is N x N, B is N x M.

	//                                            var s bool;   I don't know why they use s for a bool.  But they did.
	var X Matrix2D

	N := len(A)
	M := len(B[0])

	LU := NewMatrix(N, N)
	Copy2(A, LU)

	X = NewMatrix(N, M)
	X = Copy(B)
	//  Copy (A, N, N, LU^);  Copy (B, N, M, X);
	perm := make(Permutation, N*4) //  ALLOCATE (perm, N*SIZE(subscript));
	LU, _ = LUFactor(LU, perm)
	X = LUSolve(LU, X, perm) // X = LUSolve (LU, X, N, M, perm);

	//  For better accuracy, apply one step of iterative improvement.   Two or three steps might be better;
	//  but they might even make things worse, because we're still stuck with the rounding errors in LUFactor.

	if X != nil { // if the LUSolve failed, like because of a singular matrix, X is returned as nil
		ERROR := NewMatrix(N, M)
		product := NewMatrix(N, M)
		product = Mul(A, X)              // Mul (A, X, N, N, M, product);
		ERROR = Sub(B, product)          //  Sub (B, product^, N, M, error^);
		ERROR = LUSolve(LU, ERROR, perm) //   LUSolve (LU^, error^, N, M, perm);
		X = Add(X, ERROR)                //    Add (X, error^, N, M, X);
		//     DisposeArray (product, N, M);
		//      DisposeArray (error, N, M);
		//       DEALLOCATE (perm, N*SIZE(subscript));
		//        DisposeArray (LU, N, N);
	}
	return X // If X is nil, return it anyway as nil.
} //    END Solve;

// --------------------------------------------------------------------------------- Invert -------------------------

//    PROCEDURE Invert (A: ARRAY OF ARRAY OF EltType; VAR (*OUT*) X: ARRAY OF ARRAY OF EltType; N: CARDINAL);
//    func Invert (A Matrix2D, N int) Matrix2D { // Inverts an N x N nonsingular matrix.

func Invert(A Matrix2D) Matrix2D { // Inverts an N x N nonsingular matrix.

	var X Matrix2D
	// VAR I: ArrayPtr;

	N := len(A)
	I := NewMatrix(N, N)
	I = Unit(I)

	X = NewMatrix(N, N)
	X = Solve(A, I) // Solve (A, I^, X, N, N);
	// DisposeArray (I, N, N);
	return X
} // END Invert;

/************************************************************************)
(*                         EIGENPROBLEMS                                *)
(************************************************************************/

//                                    PROCEDURE Balance (VAR INOUT A: ARRAY OF ARRAY OF EltType;  N: CARDINAL);
func Balance(A Matrix2D) Matrix2D {

	/*
	   Replaces A by a better-balanced matrix with the same eigenvalues.  There is no effect on symmetrical matrices.
	   To minimize the effect of rounding, we scale only by a restricted set of scaling factors derived from the
	   machine's radix.
	*/

	//    VAR row, j: CARDINAL;
	//        done: BOOLEAN;

	const radix float64 = 2
	const radixsq = radix * radix

	var c, r, f, g, s float64

	//        N := len(A);

	for { // REPEAT
		done := true
		for row := range A { //  FOR row := 0 TO N-1 DO
			c = 0
			r = 0
			for j := range A { //  FOR j := 0 TO N-1 DO
				if j != row {
					c += math.Abs(A[j][row])
					r += math.Abs(A[row][j])
				} // END IF j != row
			} // END FOR j range A
			if (c != 0) && (r != 0) {
				g = r / radix
				f = 1
				s = c + r
				for c < g { // WHILE c < g DO
					f *= radix
					c *= radixsq
				} // END for c < g
				g = r * radix
				for c > g { //  WHILE c > g DO
					f /= radix
					c /= radixsq
				} // END for c > g
				if (c+r)/f < 0.95*s {
					done = false
					g = 1 / f

					// Here is the actual transformation: a scaling of this row and the corresponding column.

					for j := range A { // FOR j := 0 TO N-1 DO
						A[row][j] *= g
					} //END FOR j range A
					for j := range A { // FOR j := 0 TO N-1 DO
						A[j][row] *= f
					} // END FOR j range A
				} // END IF (c+r)/f < 0.95s

			} //END IF c and r nonzero

		} // END (* FOR row := 0 TO N-1 *);

		if done {
			break
		} // behaves as the UNTIL done statement in M-2
	} // This was the UNTIL done statement from M-2
	return A

} // END Balance;

/************************************************************************/

//                                PROCEDURE Hessenberg (VAR INOUT A: ARRAY OF ARRAY OF EltType;  N: CARDINAL);
func Hessenberg(A Matrix2D) Matrix2D { // A is an InOut matrix.

	/* Transforms an NxN matrix into upper Hessenberg form, i.e. all    *)
	   (* entries below the diagonal zero except for the first subdiagonal.*)
	   (* This is an "in-place" calculation, i.e. the answer replaces the  *)
	   (* original matrix.                                                 */

	//       already defined globally for this package    CONST small = 1.0E-15;

	//      var pos, i, j, pivotrow int;
	//      var pivot, temp: float64;
	//      var V vec.VectorPtr;

	N := len(A)
	if N <= 2 {
		return nil
	}

	V := vec.NewVector(N)
	for pos := 1; pos < N-1; pos++ { // FOR pos := 1 TO N-2 DO

		/* At this point in the calculation, A has the form         *)
		   (*          A11     A12                                     *)
		   (*          A21     A22                                     *)
		   (* where A11 has (pos+1) rows and columns and is already in *)
		   (* upper Hessenberg form; and A21 is zero except for its    *)
		   (* last two columns.  This time around the loop, we are     *)
		   (* going to transform A such that column (pos-1) of A21 is  *)
		   (* reduced to zero.  The transformation will affect only    *)
		   (* the last column of A11, therefore will not alter its     *)
		   (* Hessenberg property.                                     *)

		   (* Step 1: we need A[pos,pos-1] to be nonzero.  To keep     *)
		   (* the calculations as well-conditioned as possible, we     *)
		   (* allow for a preliminary row and column swap.             */

		pivot := A[pos][pos-1]
		pivotrow := pos
		for i := pos + 1; i < N; i++ { // FOR i := pos+1 TO N-1 DO
			temp := A[i][pos-1]
			if math.Abs(temp) > math.Abs(pivot) {
				pivot = temp
				pivotrow = i
			} // END IF temp > pivot
		} // END FOR i from pos+1 to N-1

		if math.Abs(pivot) < small {

			/* The pivot is essentially zero, so we already have    *)
			   (* the desired property and no transformation is        *)
			   (* necessary this time.  We simply replace all of the   *)
			   (* "approximately zero" entries by 0.0.                 */

			for i := pos; i < N; i++ { //  i := pos TO N-1 DO
				A[i][pos-1] = 0.0
			}

		} else {

			if pivotrow != pos {

				// Swap rows pos and pivotrow, and then swap the corresponding columns.

				for j := pos - 1; j < N; j++ { // FOR j := pos-1 TO N-1 DO
					// temp := A[pivotrow][j]; A[pivotrow][j] = A[pos][j]; A[pos][j] = temp;
					A[pos][j], A[pivotrow][j] = A[pivotrow][j], A[pos][j] // Go swapping idiom
				} //END FOR j from pos-1 to N-1
				for i := range A { // FOR i := 0 TO N-1 DO
					// temp := A[i][pivotrow]; A[i][pivotrow] = A[i][pos]; A[i][pos] = temp;
					A[i][pos], A[i][pivotrow] = A[i][pivotrow], A[i][pos] // Go swapping idiom
				} // END FOR i range A

			} // END IF pivotrow != pos

			/* Now we are going to replace A by T*A*Inverse(T),     *)
			   (* where T is a unit matrix except for column pos.      *)
			   (* That column is equal to a vector V, where V[i] = 0.0 *)
			   (* for i < pos, and V[pos] = 1.0.  We don't bother      *)
			   (* storing those fixed elements explicitly.             */

			for i := pos + 1; i < N; i++ { // FOR i := pos+1 TO N-1 DO
				V[i] = -A[i][pos-1] / pivot
			}

			/* Premultiplication of A by T.  Because of the special *)
			   (* structure of T, this affects only rows [pos+1..N].   *)
			   (* We also know that some of the results will be zero.  */

			for i := pos + 1; i < N; i++ { // FOR i := pos+1 TO N-1 DO
				A[i][pos-1] = 0.0
				for j := pos; j < N; j++ { //   FOR j := pos TO N-1 DO
					A[i][j] += V[i] * A[pos][j]
				}
			}

			/* Postmultiplication by the inverse of T.  This affects*)
			   (* only column pos.                                     */

			for i := range A { // FOR i := 0 TO N-1 DO
				temp := 0.0
				for j := pos + 1; j < N; j++ { //  FOR j := pos+1 TO N-1 DO
					temp += A[i][j] * V[j]
				}
				A[i][pos] -= temp
			} //END FOR range A

		} // END IF pivot < small

	} // END FOR pos from 2 to N-2

	// DisposeVector (V, N); not needed in Go.
	return A
} // END Hessenberg;

/************************************************************************/

//                  PROCEDURE QR ( A: ARRAY OF ARRAY OF EltType;  VAR OUT W: ARRAY OF LONGCOMPLEX; N: CARDINAL);
func QR(A Matrix2D) LongComplexSlice {
	/* Finds all the eigenvalues of an upper Hessenberg matrix.         *)
	   (* On return W contains the eigenvalues.                            *)

	   (* Source: this is an adaption of code from "Numerical Recipes"     *)
	   (* by Press, Flannery, Teutolsky, and Vetterling.                   */

	//    VAR last, m, j, k, L, its, i, imax: CARDINAL;
	//        z, y, x, w, v, u, shift, s, r, q, p, anorm: LONGREAL;

	/********************************************************************/
	var shift, w, x, y, z, p, q, r, s float64

	// Compute matrix norm.  This looks wrong to me, but it seems to be giving satisfactory results.

	anorm := math.Abs(A[0][0]) // first element in first row
	N := len(A)
	W := make(LongComplexSlice, N)
	for i := 1; i < N; i++ { //  FOR i := 1 TO N-1 DO
		for j := i - 1; j < N; j++ { //  FOR j := i-1 TO N-1 DO
			anorm += math.Abs(A[i][j])
		} // END FOR j
	} // END FOR i from 1 to N-1

	last := N - 1
	shift = 0.0
	its := 0

MainOuterLOOP:
	for {
		/* Find, if possible, an L such that A[L,L-1] is zero to    *)
		   (* machine accuracy.  If we succeed then A is now block     *)
		   (* diagonal, and we can work independently on the final     *)
		   (* block (rows and columns L to last).                      */

		L := last
	innerLOOP:
		for {
			if L == 0 {
				break innerLOOP
			} //ENDIF
			s := math.Abs(A[L-1][L-1]) + math.Abs(A[L][L])
			if s == 0.0 {
				s = anorm
			} // ENDIF
			if math.Abs(A[L][L-1])+s == s {
				break innerLOOP
			} //END IF
			L--
		} // END innerLOOP

		x = A[last][last]
		if L == last {

			// One eigenvalue found.

			W[last] = complex(x+shift, 0.0)
			if last > 0 {
				last--
				its = 0
			} else {
				break MainOuterLOOP
			} // END IF

		} else {
			y = A[last-1][last-1]
			w = A[last][last-1] * A[last-1][last]
			if L == last-1 {

				// We're down to a 2x2 submatrix, so can work out the eigenvalues directly.

				p = 0.5 * (y - x)
				q = p*p + w
				z = math.Sqrt(math.Abs(q))
				x += shift
				if q >= 0.0 {
					if p < 0.0 {
						z = p - z
					} else {
						z = p + z
					} // END IF
					W[last-1] = complex(x+z, 0.0)
					if z == 0.0 {
						W[last] = complex(x, 0.0)
					} else {
						W[last] = complex(x-w/z, 0.0)
					} // END IF z is 0
				} else {
					W[last-1] = complex(x+p, z)
					W[last] = cmplx.Conj(W[last-1])
				} // END IF q >=0
				if last >= 2 {
					last -= 2
					its = 0
				} else {
					break MainOuterLOOP
				} // END IF

			} else {

				if its < 10 {
					its++
				} else {
					// If we're converging too slowly, modify the shift.

					shift += x
					for i := 0; i <= last; i++ {
						A[i][i] -= x
					} // END FOR
					s = math.Abs(A[last][last-1]) + math.Abs(A[last-1][last-2])
					x = 0.75 * s
					y = x
					w = -0.4375 * s * s
					its = 0
				} // END IF its<10

				/* We're now working on a sub-array [L..last] of    *)
				   (* size 3x3 or greater.  Our goal is to transform   *)
				   (* the matrix so as to reduce the magnitudes of     *)
				   (* the elements on the first sub-diagonal, so that  *)
				   (* after one or more iterations one of them will be *)
				   (* zero to within machine accuracy.                 *)

				   (* Shortcut: if we can find two consecutive         *)
				   (* subdiagonal elements whose product is small,     *)
				   (* we're even better off.                           */

				m := last - 2
			anotherInnerLOOP:
				for {
					z = A[m][m]
					r = x - z
					s = y - z
					p = (r*s-w)/A[m+1][m] + A[m][m+1]
					q = A[m+1][m+1] - z - r - s
					r = A[m+2][m+1]
					s = math.Abs(p) + math.Abs(q) + math.Abs(r)
					p = p / s
					q = q / s
					r = r / s
					if m == L {
						break anotherInnerLOOP
					} // ENDIF
					u := math.Abs(A[m][m-1]) * (math.Abs(q) + math.Abs(r))
					v := math.Abs(p) * (math.Abs(A[m-1][m-1]) + math.Abs(z) + math.Abs(A[m+1][m+1]))
					if u+v == v {
						break anotherInnerLOOP
					} // ENDIF
					m--
				} // END anotherInnerLOOP

				A[m+2][m] = 0.0
				for i := m + 3; i <= last; i++ {
					A[i][i-2] = 0.0
					A[i][i-3] = 0.0
				} //END FOR

				// Apply row and column transformations that should reduce the magnitudes of subdiagonal elements.

				for k := m; k < last; k++ {
					if k != m {
						p = A[k][k-1]
						q = A[k+1][k-1]
						r = 0.0
						if k != last-1 {
							r = A[k+2][k-1]
						} // END IF
						x = math.Abs(p) + math.Abs(q) + math.Abs(r)
						if x != 0.0 {
							p /= x //  p = p/x;
							q /= x //  q = q/x;
							r /= x //  r = r/x;
						} // END IF
					} // END IF
					s = math.Sqrt(p*p + q*q + r*r)
					if p < 0.0 {
						s = -s
					} // ENDIF
					if s != 0.0 {
						if k == m {
							if L != m {
								A[k][k-1] = -A[k][k-1]
							} // ENDIF
						} else {
							A[k][k-1] = -s * x
						} // END IF
						p += s
						x = p / s
						y = q / s
						z = r / s
						q /= p
						r /= p

						// Row transformation.

						for j := k; j <= last; j++ {
							p = A[k][j] + q*A[k+1][j]
							if k != last-1 {
								p += r * A[k+2][j]
								A[k+2][j] -= p * z
							} // END IF
							A[k+1][j] -= p * y
							A[k][j] -= p * x
						} // END FOR

						// Column transformation.

						imax := k + 3
						if last < imax {
							imax = last
						} // ENDIF
						for i := L; i <= imax; i++ {
							p = x*A[i][k] + y*A[i][k+1]
							if k != last-1 {
								p += z * A[i][k+2]
								A[i][k+2] -= p * r
							} // END IF
							A[i][k+1] -= p * q
							A[i][k] -= p
						} // END FOR

					} // END IF s <> 0.0

				} //END FOR k := m TO last-1

			} // END IF test for 2x2 or bigger

		} // END IF test for 1x1

	} // END  main loop
	return W
} // END QR;

/************************************************************************)
(*                           EIGENVALUES                                *)
(************************************************************************/

// ----------------------------------------------------------------------------- Eigenvalues -----------------------

//           PROCEDURE Eigenvalues (A: ARRAY OF ARRAY OF EltType; VAR OUT W: ARRAY OF LONGCOMPLEX; N: CARDINAL);
func Eigenvalues(A Matrix2D) LongComplexSlice {

	// Finds all the eigenvalues of an NxN matrix.  This procedure does not modify A.

	var Acopy Matrix2D //            VAR Acopy: ArrayPtr;
	var W LongComplexSlice

	N := len(A)
	if N > 0 {
		Acopy = NewMatrix(N, N)
		Copy2(A, Acopy)           //           Copy (A, N, N, Acopy^);
		Acopy = Balance(Acopy)    //           Balance (Acopy^, N);
		Acopy = Hessenberg(Acopy) //           Hessenberg (Acopy^, N);

		W = make(LongComplexSlice, N)
		W = QR(Acopy) //           QR (Acopy^, W, N);
		// not needed in Go            DisposeArray (Acopy, N, N);
	} // END IF
	return W
} // END Eigenvalues;

/************************************************************************)
(*                          SCREEN OUTPUT                               *)
(************************************************************************/

// ------------------------------------------------------------------------------ Write ----------------------------

//  PROCEDURE Write (M: ARRAY OF ARRAY OF EltType;  r, c: CARDINAL;  places: CARDINAL);
func Write(M Matrix2D, places int) []string {

	// Writes the r x c matrix M to the screen, where each column occupies a field "places" characters wide.

	//    VAR i, j: CARDINAL;

	OutputStringSlice := make([]string, 0, 500)
	for i := range M { // FOR i := 0 TO r-1 DO
		for j := range M[i] { //   FOR j := 0 TO c-1 DO
			ss := strconv.FormatFloat(M[i][j], 'G', places, 64)
			OutputStringSlice = append(OutputStringSlice, fmt.Sprintf("%10s", ss))
			//WriteLongReal (M[i,j], places);
		} // END FOR j
		OutputStringSlice = append(OutputStringSlice, "\n")
	} // END FOR i
	OutputStringSlice = append(OutputStringSlice, "\n")
	/*
	   n := len(M);
	   m := len(M[0]);
	   fmt.Printf(" Matrix is %d x %d, rows x cols\n",n,m);
	   for i := range M {
	     for j := range M[0] {
	       fmt.Print("    ",M[i][j]);
	     }
	     fmt.Println();
	   }
	   fmt.Println();
	   fmt.Println();
	   for _,s := range OutputStringSlice {
	     fmt.Print(s);
	   }
	   fmt.Println(" end from mat.Write");
	   fmt.Println();
	*/
	return OutputStringSlice
} // END Write

// END Mat.
