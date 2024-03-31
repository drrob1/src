package mat

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"math"
	"math/cmplx"
	"math/rand/v2"
	"src/vec"
	"strconv"
	"strings"
)

//               Matrix arithmetic
//   Programmer:         P. Moylan
//   Last edited:        15 August 1995
//   Status:             OK
//    Modula-2 Portability problem: I've had to use an XDS language extension (open arrays) here.
//    I haven't yet figured out how to do the job in ISO standard Modula-2.

/*
 REVISION HISTORY
 ================
 19 Dec 16 -- Started conversion to Go from old Modula-2 source.  We'll see how long this takes.
 24 Dec 16 -- Passed mattest.
 25 Dec 16 -- Changed the code to use the Go swapping idiom
  1 Aug 20 -- Cleaning up some code.  I'm looking at this again because of adding gonum to solve.go --> gonumsolve.go
 13 Feb 22 -- Updated to modules
 21 Nov 22 -- static linter reported issues, so some of them are addressed, and others are ignored.
 31 Mar 23 -- StaticCheck reported that Copy2 won't work, because I used value semantics when I needed pointer semantics on the return var.  I took it out as it wasn't idiomatic anyway.
  3 Apr 23 -- Exporting Small.  And writing WriteZero, WriteZeroPair, and SolveInversion.
 10 Mar 24 -- Updated to Go 1.22, mostly in math/rand/v2.  And added Equal func.
 18 Mar 24 -- Adding WriteZeroln and WriteZeroPairLn, which just does the screen writes without returning anything.
 25 Mar 24 -- Added EqualApproximately.
 28 Mar 24 -- Added IsZeroApproximately, and IsZeroApprox.
*/

const LastAltered = "28 Mar 2024"
const Small = 1.0e-10
const SubscriptDim = 8192

type Matrix2D [][]float64
type Permutation []int

type LongComplexSlice []complex128 //

//   Creating matrices

func NewMatrix(R, C int) Matrix2D { // I think row, column makes more sense than N x M
	// Creates an NxM matrix as a slice of slices.  So it's a pointer that gets passed around.
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

//  ASSIGNMENTS

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
			if math.Abs(matrix[r][c]) < Small {
				matrix[r][c] = 0
			}
		}
	}
	return matrix
}

func Unit(matrix Matrix2D) Matrix2D {
	// Creates an N by N identity matrix, with all zeros except along the main diagonal.
	matrix = Zero(matrix)
	for diag := range matrix {
		matrix[diag][diag] = 1
	}
	return matrix
}

func Random(matrix Matrix2D) Matrix2D {
	// Creates matrix with random integers from 0..100

	for r := range matrix {
		for c := range matrix[r] {
			matrix[r][c] = float64(rand.IntN(100))
		}
	}
	return matrix
}

func Copy(Src Matrix2D) Matrix2D {
	// Copies an r x c matrix A to B, by doing an element by element copy.  I don't think just copying pointers is correct.

	SrcRows := len(Src)
	SrcCols := len(Src[0])

	Dest := NewMatrix(SrcRows, SrcCols)

	//copy(Dest, Src) //  Doesn't work so I did it wrong.
	for r := range Src {
		copy(Dest[r], Src[r]) // this works.
		//for c := range Src[r] { // golangci-lint recommended I use the copy() built in.
		//	Dest[r][c] = Src[r][c]
		//}
	}
	return Dest
}

//   THE BASIC MATRIX OPERATIONS

// ---------------------------------------------------- Add -------------------

func Add(A, B Matrix2D) Matrix2D {
	// Computes C = A + B.

	Arows := len(A)
	Acols := len(A[0])
	Brows := len(B)
	Bcols := len(B[0])
	if (Arows != Brows) || (Acols != Bcols) {
		return nil
	}

	C := NewMatrix(Brows, Acols) // Could have been either row and either col.  I chose those.

	for i := range A {
		for j := range A[i] {
			C[i][j] = A[i][j] + B[i][j]
		}
	}
	return C
}

// ---------------------------------------------------- Sub -------------------

func Sub(A, B Matrix2D) Matrix2D {
	// Computes C = A - B.

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

// ----------------------------------------------------- MUL -------------------

func Mul(A, B Matrix2D) Matrix2D {
	// Computes C = A x B.  Using std linear algebra rules.

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

//   SOLVING LINEAR EQUATIONS

// -------------------------------------------------------------------------------- LUFactor ------------------------------------------------

func LUFactor(A Matrix2D, perm Permutation) (Matrix2D, bool) { // A is an InOut param.

	/*
	       LU decomposition of a square matrix.  We express A in the form P*L*U, where P is a permutation matrix,
	       L is lower triangular with unit diagonal elements, and U is upper triangular.  This is an in-place
	       computation, where on exit U occupies the upper triangle of A, and L (not including its diagonal entries)
	       is in the lower triangle.  The permutation information is returned in perm.  Output parameter oddswaps
	       is TRUE iff an odd number of row interchanges were done by the permutation.
	       (We need to know this only if we are going to go on to calculate a determinant.)

	       The precise relationship between the implied permutation matrix P and the output parameter perm is somewhat
	       obscure.  The vector perm^ is not simply a permutation of the subscripts [0..N-1]; it does, however, have
	       the property that we can recreate P by walking through perm^ from start to finish, in the order used by
	       procedure LUSolve.
	   	This rtn returns a packed form of L and U into a single matrix called here LU.
	       Intended for use in Solve below.
	*/

	var pivotrow int
	var sum, temp, maxval float64

	N := len(A)
	VV := vec.NewVector(N)
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

	//	Crout's method: we work through one column at a time.

	for col := range A {

		//	Upper triangular component of this column - except for the diagonal element,
		//  which we leave until after we've selected a pivot from on or below the diagonal.

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

		// Lower triangular component in this column.  The results we get in this loop will not be correct until we've divided by the pivot;
		// but we work out the pivot location as we go, and come back later for this division.

		maxval = 0
		pivotrow = col
		for row := col; row < N; row++ {
			sum = A[row][col]
			if col > 0 {
				for k := 0; k < col; k++ {
					sum -= A[row][k] * A[k][col]
				}
			}
			A[row][col] = sum
			temp := VV[row] * math.Abs(sum)
			if temp >= maxval {
				maxval = temp
				pivotrow = row
			}
		}

		// If pivot element was not already on the diagonal, do a row swap.

		if pivotrow != col {
			for k := 0; k < N; k++ {
				A[col][k], A[pivotrow][k] = A[pivotrow][k], A[col][k]
			}
			oddswaps = !oddswaps
			VV[pivotrow] = VV[col]
		}
		perm[col] = pivotrow

		// Finish off the calculation of the lower triangular part for this column by scaling by the pivot A[col,col].

		// Remark: if the pivot is still zero at this stage, then all the elements below it are also zero.  The LU
		// decomposition in this case is not unique - the original matrix is singular, therefore U will also be
		// singular -- but one solution is to leave all those elements zero.

		temp = A[col][col]
		if (col != N-1) && (temp != 0.0) {
			temp = 1 / temp
			for row := col + 1; row < N; row++ {
				A[row][col] = temp * A[row][col]
			}
		}

	} // END FOR col range A

	return A, oddswaps
} // END LUFactor

// ------------------------------------------------------- LUSolve ---------------------------------------

func LUSolve(LU, B Matrix2D, perm Permutation) Matrix2D {
	/*
	   Solves the equation P*L*U * X = B, where P is a permutation matrix specified indirectly by perm; L is lower triangular; and
	   U is upper triangular.  The "Matrix" LU is not a genuine matrix, but rather a packed form of L and U as produced by procedure
	   LUfactor above.  Returns the solution X.
	   Dimensions: left side is NxN, B is NxM
	*/
	var sum, scale float64
	/*
	   Pass 1: Solve the equation L*Y = B (at the same time sorting B in accordance with the specified permutation).
	   The solution Y overwrites the original value of B.

	   Understanding how the permutations work is something of a black art.  It helps to know that (a) ip>=i for all i, and
	   (b) in the summation over k below, we are accessing only those elements of B that have already been sorted into the
	   correct order.
	*/

	N := len(B)

	for i := range B {
		ip := perm[i]
		for j := range B[i] {
			sum = B[ip][j]
			B[ip][j] = B[i][j]
			if i > 0 {
				for k := 0; k < i; k++ {
					sum -= LU[i][k] * B[k][j]
				}
			}
			B[i][j] = sum
		}
	}

	// Pass 2: solve the equation U*X = Y.

	for i := N - 1; i >= 0; i-- {
		scale = LU[i][i]
		if scale == 0 {
			//  Matrix is singular.  Aborting.
			return nil
		}
		for j := range B[i] {
			sum = B[i][j]
			for k := i + 1; k < N; k++ {
				sum -= LU[i][k] * B[k][j]
			}
			B[i][j] = sum / scale
		}
	}

	return B
} // END LUSolve;

// ---------------------------------- GaussJordan Elimination ----------------------------

func GaussJ(A, B Matrix2D) Matrix2D {

	// Solves the equation AX = B by Gauss-Jordan elimination.  In the present version A must be square and non-singular.
	// This approach to solving the equation is not the best available -- see below -- but is included here
	// anyway since it is popular.
	// Dimensions: A is NxN, B is NxM.

	var pivot float64
	var X Matrix2D

	N := len(A)

	W := Copy(A)
	X = Copy(B)

	// Remark: we are going to use elementary row operations to turn W into a unit matrix.  However we don't
	// bother to store the new 1.0 and 0.0 entries, because those entries will never be fetched again.
	// We simply base our calculations on the assumption that those values have been stored.

	// Dimensions: A is N x N, B is N x M.

	// Pass 1: by elementary row operations, make W into an upper triangular matrix.

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
		if math.Abs(pivot) < Small { // Coefficient matrix is singular.  Aborting,
			return nil
		} // END IF pivot < small

		// Swap rows i and prow.

		for j := i; j < N; j++ {
			W[i][j], W[prow][j] = W[prow][j], W[i][j]
		}

		for j := range X[i] {
			X[i][j], X[prow][j] = X[prow][j], X[i][j]
		}

		// Scale the i'th row of both W and X.

		for j := i + 1; j < N; j++ {
			W[i][j] = W[i][j] / pivot
		}
		for j := range X[i] {
			X[i][j] = X[i][j] / pivot
		}

		// Implicitly reduce the sub-column below W[i,i] to zero.
		for k := i + 1; k < N; k++ {
			scale := W[k][i]
			for j := i + 1; j < N; j++ {
				W[k][j] -= scale * W[i][j]
			}
			for j := range X[i] {
				X[k][j] -= scale * X[i][j]
			}
		}
	} // END FOR i range W

	// Pass 2: reduce the above-diagonal elements of W to zero.

	for i := N - 1; i > 0; i-- {
		// Implicitly reduce the sub-column above W[i,i] to zero.
		for k := 0; k < i; k++ {
			scale := W[k][i]
			for j := range X[i] {
				X[k][j] -= scale * X[i][j]
			}
		}
	}

	return X
} //    END GaussJ;

// --------------------------------- Solve ---------------------------

func Solve(A, B Matrix2D) Matrix2D {

	// Solves the equation A * X = B.  In the present version A must be square and nonsingular.
	// Dimensions: A is N x N, B is N x M, where M usually = 1.

	N := len(A)
	LU := Copy(A)
	X := Copy(B)

	perm := make(Permutation, N*4)
	LU, _ = LUFactor(LU, perm)
	X = LUSolve(LU, X, perm)

	//  For better accuracy, apply one step of iterative improvement.   Two or three steps might be better;
	//  but they might even make things worse, because we're still stuck with the rounding errors in LUFactor.

	if X != nil { // if the LUSolve failed, like because of a singular matrix, X is returned as nil
		product := Mul(A, X)
		ERROR := Sub(B, product)
		ERROR = LUSolve(LU, ERROR, perm)
		X = Add(X, ERROR)
	}
	return X // If X is nil, return it anyway as nil.
} // END Solve;

// ------------------------------- Invert -------------------------

func Invert(A Matrix2D) Matrix2D {
	N := len(A)
	u := NewMatrix(N, N)
	u = Unit(u)
	inv := Solve(A, u)
	return inv
} // END Invert;

// ---------------------------------- SolveInvert ------------------------------

func SolveInvert(a, b Matrix2D) Matrix2D {
	I := Invert(a)
	x := Mul(I, b)
	return x
}

// ---------------------------------- EqualApprox --------------------------------------

func EqualApprox(a, b Matrix2D) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a[0]) != len(b[0]) {
		return false
	}
	for i := range a {
		for j := range a[i] {
			if math.Abs(a[i][j]-b[i][j]) >= Small { // I can't compare floats using equal, it's too likely to fail due small differences in the numbers.
				return false
			}
		}
	}
	return true
}

func EqualApproximately(a, b Matrix2D, tol float64) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a[0]) != len(b[0]) {
		return false
	}
	tol = math.Abs(tol)
	for i := range a {
		for j := range a[i] {
			if math.Abs(a[i][j]-b[i][j]) >= tol { // I can't compare floats using equal, it's too likely to fail due small differences in the numbers.
				return false
			}
		}
	}
	return true
}

func IsZeroApprox(a Matrix2D) bool {
	return IsZeroApproximately(a, Small)
}

func IsZeroApproximately(a Matrix2D, tol float64) bool {
	tol = math.Abs(tol)
	for i := range a {
		for j := range a[i] {
			if math.Abs(a[i][j]) >= tol { // I can't compare floats using equal, it's too likely to fail due small differences in the numbers.
				return false
			}
		}
	}
	return true
}

//   EIGENPROBLEMS

func Balance(A Matrix2D) Matrix2D {
	/*
	   Replaces A by a better-balanced matrix with the same eigenvalues.  There is no effect on symmetrical matrices.
	   To minimize the effect of rounding, we scale only by a restricted set of scaling factors derived from the
	   machine's radix.
	*/

	const radix float64 = 2
	const radixsq = radix * radix

	var c, r, f, g, s float64

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

func Hessenberg(A Matrix2D) Matrix2D { // A is an InOut matrix.
	/*
		Transforms an NxN matrix into upper Hessenberg form, i.e. all entries below the diagonal zero except for the first subdiagonal.
		This is an "in-place" calculation, i.e. the answer replaces the original matrix.
	*/

	//  CONST small = 1.0E-15;  But for this Go translation, I made it 1e-10

	N := len(A)
	if N <= 2 {
		return nil
	}

	V := vec.NewVector(N)
	for pos := 1; pos < N-1; pos++ { // FOR pos := 1 TO N-2 DO

		/*
		   		At this point in the calculation, A has the form
		   		          A11     A12
		   		          A21     A22
		   		where A11 has (pos+1) rows and columns and is already in upper Hessenberg form; and A21 is zero except for its
		        last two columns.  This time around the loop, we are going to transform A such that column (pos-1) of A21 is
		   		reduced to zero.  The transformation will affect only the last column of A11, therefore will not alter its
		   		Hessenberg property.

		   		Step 1: we need A[pos,pos-1] to be nonzero.  To keep the calculations as well-conditioned as possible, we
		   		allow for a preliminary row and column swap.
		*/

		pivot := A[pos][pos-1]
		pivotrow := pos
		for i := pos + 1; i < N; i++ {
			temp := A[i][pos-1]
			if math.Abs(temp) > math.Abs(pivot) {
				pivot = temp
				pivotrow = i
			} // END IF temp > pivot
		}

		if math.Abs(pivot) < Small {

			/*
				The pivot is essentially zero, so we already have
				the desired property and no transformation is
				necessary this time.  We simply replace all of the
				"approximately zero" entries by 0.0.
			*/

			for i := pos; i < N; i++ { //  i := pos TO N-1 DO
				A[i][pos-1] = 0.0
			}

		} else {

			if pivotrow != pos {

				// Swap rows pos and pivotrow, and then swap the corresponding columns.

				for j := pos - 1; j < N; j++ {
					A[pos][j], A[pivotrow][j] = A[pivotrow][j], A[pos][j]
				}
				for i := range A {
					A[i][pos], A[i][pivotrow] = A[i][pivotrow], A[i][pos]
				}

			}

			/*
				Now we are going to replace A by T*A*Inverse(T), where T is a unit matrix except for column pos.
				That column is equal to a vector V, where V[i] = 0.0 for i < pos, and V[pos] = 1.0.  We don't bother
				storing those fixed elements explicitly.
			*/

			for i := pos + 1; i < N; i++ {
				V[i] = -A[i][pos-1] / pivot
			}

			/*
				    Premultiplication of A by T.  Because of the special structure of T, this affects only rows [pos+1..N].
					We also know that some of the results will be zero.
			*/
			for i := pos + 1; i < N; i++ {
				A[i][pos-1] = 0.0
				for j := pos; j < N; j++ {
					A[i][j] += V[i] * A[pos][j]
				}
			}

			// Postmultiplication by the inverse of T.  This affects only column pos.
			for i := range A {
				temp := 0.0
				for j := pos + 1; j < N; j++ {
					temp += A[i][j] * V[j]
				}
				A[i][pos] -= temp
			}

		} // END IF pivot < small
	} // END FOR pos from 2 to N-2
	return A
} // END Hessenberg;

/************************************************************************/

func QR(A Matrix2D) LongComplexSlice {
	/*
		Finds all the eigenvalues of an upper Hessenberg matrix.
		On return W contains the eigenvalues.

		Source: this is an adaption of code from "Numerical Recipes" by Press, Flannery, Teutolsky, and Vetterling.
	*/
	var shift, w, x, y, z, p, q, r, s float64

	// Compute matrix norm.

	anorm := math.Abs(A[0][0]) // first element in first row
	N := len(A)
	W := make(LongComplexSlice, N)
	for i := 1; i < N; i++ {
		for j := i - 1; j < N; j++ {
			anorm += math.Abs(A[i][j])
		}
	}

	last := N - 1
	shift = 0.0
	its := 0

MainOuterLOOP:
	for {
		/*
			Find, if possible, an L such that A[L,L-1] is zero to machine accuracy.  If we succeed then A is now block
			diagonal, and we can work independently on the final block (rows and columns L to last).
		*/
		L := last

	innerLOOP:
		for {
			if L == 0 {
				break innerLOOP
			}
			s := math.Abs(A[L-1][L-1]) + math.Abs(A[L][L])
			if s == 0.0 {
				s = anorm
			}
			if math.Abs(A[L][L-1])+s == s {
				break innerLOOP
			}
			L--
		}

		x = A[last][last]
		if L == last {
			// One eigenvalue found.

			W[last] = complex(x+shift, 0.0)
			if last > 0 {
				last--
				its = 0
			} else {
				break MainOuterLOOP
			}

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
					}
					W[last-1] = complex(x+z, 0.0)
					if z == 0.0 {
						W[last] = complex(x, 0.0)
					} else {
						W[last] = complex(x-w/z, 0.0)
					}
				} else {
					W[last-1] = complex(x+p, z)
					W[last] = cmplx.Conj(W[last-1])
				} // END IF q >= 0
				if last >= 2 {
					last -= 2
					its = 0
				} else {
					break MainOuterLOOP
				}

			} else {
				if its < 10 {
					its++
				} else {
					// If we're converging too slowly, modify the shift.

					shift += x
					for i := 0; i <= last; i++ {
						A[i][i] -= x
					}
					s = math.Abs(A[last][last-1]) + math.Abs(A[last-1][last-2])
					x = 0.75 * s
					y = x
					w = -0.4375 * s * s
					its = 0
				} // END IF its<10

				/*
				  We're now working on a sub-array [L..last] of size 3x3 or greater.  Our goal is to transform
				  the matrix so as to reduce the magnitudes of the elements on the first sub-diagonal, so that
				  after one or more iterations one of them will be zero to within machine accuracy.

				  Shortcut: if we can find two consecutive subdiagonal elements whose product is small, we're even better off.
				*/

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
					}
					u := math.Abs(A[m][m-1]) * (math.Abs(q) + math.Abs(r))
					v := math.Abs(p) * (math.Abs(A[m-1][m-1]) + math.Abs(z) + math.Abs(A[m+1][m+1]))
					if u+v == v {
						break anotherInnerLOOP
					}
					m--
				}

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
						}
						x = math.Abs(p) + math.Abs(q) + math.Abs(r)
						if x != 0.0 {
							p /= x
							q /= x
							r /= x
						}
					}
					s = math.Sqrt(p*p + q*q + r*r)
					if p < 0.0 {
						s = -s
					}
					if s != 0.0 {
						if k == m {
							if L != m {
								A[k][k-1] = -A[k][k-1]
							}
						} else {
							A[k][k-1] = -s * x
						}
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
							}
							A[k+1][j] -= p * y
							A[k][j] -= p * x
						}

						// Column transformation.

						imax := k + 3
						if last < imax {
							imax = last
						}
						for i := L; i <= imax; i++ {
							p = x*A[i][k] + y*A[i][k+1]
							if k != last-1 {
								p += z * A[i][k+2]
								A[i][k+2] -= p * r
							}
							A[i][k+1] -= p * q
							A[i][k] -= p
						}

					} // END IF s <> 0.0
				} //END FOR k := m TO last-1
			} // END IF test for 2x2 or bigger
		} // END IF test for 1x1
	} // END  main loop
	return W
} // END QR;

//   EIGENVALUES

// ----------------------------------------------------------------------------- Eigenvalues -----------------------

func Eigenvalues(A Matrix2D) LongComplexSlice {
	// Finds all the eigenvalues of an NxN matrix.  This procedure does not modify A.

	var aCopy Matrix2D
	var W LongComplexSlice

	N := len(A)
	if N > 0 {
		aCopy = Copy(A)           //           Copy (A, N, N, Acopy^);
		aCopy = Balance(aCopy)    //           Balance (Acopy^, N);
		aCopy = Hessenberg(aCopy) //           Hessenberg (Acopy^, N);

		// W = make(LongComplexSlice, N)  This isn't used.  I don't know why it's here.  I can't debug this routine anyway, so I'll do what staticcheck says.
		W = QR(aCopy) //           QR (Acopy^, W, N);
		// not needed in Go            DisposeArray (Acopy, N, N);
	} // END IF
	return W
} // END Eigenvalues;

//   OUTPUT

// ------------------------------------------------------------------------------ Write ----------------------------

func Write(M Matrix2D, places int) []string {
	// Writes the r x c matrix M to a string slice, where each column occupies a field "places" characters wide.

	OutputStringSlice := make([]string, 0, 500)
	for i := range M {
		for j := range M[i] {
			ss := strconv.FormatFloat(M[i][j], 'G', places, 64)
			OutputStringSlice = append(OutputStringSlice, fmt.Sprintf("%10s", ss))
		} // END FOR j
		OutputStringSlice = append(OutputStringSlice, "\n")
	} // END FOR i
	OutputStringSlice = append(OutputStringSlice, "\n")

	return OutputStringSlice
} // END Write

func Writeln(M Matrix2D, places int) {
	ss := Write(M, places)
	printString(ss)
}

func printString(s []string) { // not exported, at the moment.
	for _, line := range s {
		ctfmt.Print(ct.Yellow, true, line)
	}
}

// ------------------------------------------------------------------------------ WriteZero ----------------------------

func WriteZero(M Matrix2D, places int) []string {
	// Writes the r x c matrix M to a string slice after making small values = 0, where each column occupies a field "places" characters wide.

	OutputStringSlice := make([]string, 0, 500)
	for i := range M {
		for j := range M[i] {
			v := M[i][j]
			if math.Abs(v) < Small {
				v = 0
			}
			ss := strconv.FormatFloat(v, 'G', places, 64)
			OutputStringSlice = append(OutputStringSlice, fmt.Sprintf("%10s", ss))
		} // END FOR j
		OutputStringSlice = append(OutputStringSlice, "\n")
	} // END FOR i
	OutputStringSlice = append(OutputStringSlice, "\n")

	return OutputStringSlice
} // END WriteZero

func WriteZeroln(M Matrix2D, places int) {
	ss := WriteZero(M, places)
	printString(ss)
}

// ------------------------------------------------------------------------------ WriteZeroPair ----------------------------

func WriteZeroPair(m1, m2 Matrix2D, places int) []string {
	// Writes the r x c matrix M to a string slice after making small values = 0, where each column occupies a field "places" characters wide.
	const padding = "               |"

	OutputStringSlice := make([]string, 0, 500)
	OutputStringSlice1 := make([]string, 0, 500)
	OutputStringSlice2 := make([]string, 0, 500)

	for i := range m1 {
		var line []string
		for j := range m1[i] {
			v := m1[i][j]
			if math.Abs(v) < Small {
				v = 0
			}
			ss := strconv.FormatFloat(v, 'G', places, 64)

			line = append(line, fmt.Sprintf("%10s", ss))
		} // END FOR j
		s := strings.Join(line, "")
		OutputStringSlice1 = append(OutputStringSlice1, s)
		OutputStringSlice1 = append(OutputStringSlice1, "\n")
	} // END FOR i
	OutputStringSlice1 = append(OutputStringSlice1, "\n")

	//fmt.Printf(" output string slice 1:\n")
	//for _, s := range OutputStringSlice1 {
	//	fmt.Print(s)
	//}
	//fmt.Println("--------------------- stringslice1")

	for i := range m2 {
		var line []string
		for j := range m2[i] {
			v := m2[i][j]
			if math.Abs(v) < Small {
				v = 0
			}
			ss := strconv.FormatFloat(v, 'G', places, 64)

			line = append(line, fmt.Sprintf("%10s", ss))
			//WriteLongReal (M[i,j], places);
		} // END FOR j
		s := strings.Join(line, "")
		OutputStringSlice2 = append(OutputStringSlice2, s)
		OutputStringSlice2 = append(OutputStringSlice2, "\n")
	} // END FOR i
	OutputStringSlice2 = append(OutputStringSlice2, "\n")

	//fmt.Printf(" output string slice 2:\n")
	//for _, s := range OutputStringSlice2 {
	//	fmt.Print(s)
	//}
	//fmt.Println("-------------------------- stringslice2")

	for i := range OutputStringSlice1 {
		ss := OutputStringSlice1[i]
		if ss != "\n" {
			ss = ss + padding + OutputStringSlice2[i]
		}
		OutputStringSlice = append(OutputStringSlice, ss)
	}

	return OutputStringSlice
} // END WriteZeroPair

func WriteZeroPairln(m1, m2 Matrix2D, places int) {
	ss := WriteZeroPair(m1, m2, places)
	println(ss)
}

// END Mat.

/*
From https://www.developer.com/languages/matrix-go-golang/, added Oct 8, 2023.

package main

import (
	"fmt"
	"math/rand"
	"time"
)

func AddMatrix(matrix1 [][]int, matrix2 [][]int) [][]int {
	result := make([][]int, len(matrix1))  I didn't know this syntax is allowed and works.
	for i, a := range matrix1 {
		for j, _ := range a {
			result[i] = append(result[i], matrix1[i][j]+matrix2[i][j])
		}
	}
	return result
}

func SubMatrix(matrix1 [][]int, matrix2 [][]int) [][]int {
	result := make([][]int, len(matrix1))
	for i, a := range matrix1 {
		for j, _ := range a {
			result[i] = append(result[i], matrix1[i][j]-matrix2[i][j])
		}
	}
	return result
}

func populateRandomValues(size int) [][]int {

	m := make([][]int, size)
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			m[i] = append(m[i], rand.Intn(10)-rand.Intn(9))
		}
	}
	return m
}

func main() {
	rand.Seed(time.Now().Unix())
	var size int
	fmt.Println("Enter size of the square matrix: ")
	fmt.Scanln(&size)
	x1 := populateRandomValues(size)
	x2 := populateRandomValues(size)

	fmt.Println("matrix1:", x1)
	fmt.Println("matrix2:", x2)

	fmt.Println("ADD: matrix1 + matrix2: ", AddMatrix(x1, x2))
	fmt.Println("SUB: matrix1 - matrix2: ", SubMatrix(x1, x2))
}

func MulMatrix(matrix1 [][]int, matrix2 [][]int) [][]int {
	result := make([][]int, len(matrix1))
	for i := 0; i < len(matrix1); i++ {
		result[i] = make([]int, len(matrix1))
		for j := 0; j < len(matrix2); j++ {
			for k := 0; k < len(matrix2); k++ {
				result[i][j] += matrix1[i][k] * matrix2[k][j]
			}
		}
	}
	return result
}

func main() {
	//...
	fmt.Println("MUL: matrix1 * matrix2: ", MulMatrix(x1, x2))
}
*/
