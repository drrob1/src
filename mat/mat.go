package mat;

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
// 19 Dec 2016 -- Started conversion to Go from old Modula-2 source.  We'll see how long this takes.


import (
  "os"
  "bufio"
  "fmt"
  "path/filepath"
  "strings"
  "strconv"
  "math/cmplx"
  "math/rand"
  "time"
//
  "getcommandline"
  "timlibg"
  "tokenize"
  "vec"
)

const small = 1.0E-15;
const SubscriptDim = 8192
//TYPE subscript = [0..8191];

type EltType float64;
type Matrix2D [][]EltType;     // Array of Array of EltType;  Again, nevermind that it's a slice of slice
//type ArrayPtr *EltType2Dim; //POINTER TO ARRAY OF ARRAY OF EltType;  Don't think I need this anyway.

type Permutation []int
//type Permutation = POINTER TO ARRAY subscript OF subscript;  array [0..8191] OF subrange of integer.  
//I'm going to ignore this subrange of integer called subscript, and just make it int.

func init() {
  rand.Seed(time.Now().UnixNano());
}
/************************************************************************)
(*                   CREATING AND DESTROYING MATRICES                   *)
(************************************************************************/

func NewArray (R, C int) Matrix2D { // I think row, column makes more sense than N x M 
    // Creates an NxM matrix.
// old code basically did this: NEW (result, N-1, M-1); RETURN result;

  matrix := make(Matrix2D,0,R);
  for i := 0,i < R, i++ {
    result[i] = make([]EltType,0,C)
  }
  return matrix;
}

// PROCEDURE DisposeArray (VAR (*INOUT*) V: ArrayPtr;  N, M: CARDINAL);
// Deallocates an NxM matrix is not needed, because of the garbage collection.

/************************************************************************)
(*                          ASSIGNMENTS                                 *)
(************************************************************************/

//                                 PROCEDURE Zero (VAR (*OUT*) M: ARRAY OF ARRAY OF EltType;  r, c: CARDINAL);
func Zero(matrix Matrix2D) Matrix2D {
    // It zeros an already defined r by c matrix.  I'm not sure this is needed in Go, but here it is.

  for r := range matrix {
    for c := range matrix(r) {
      matrix[r][c] = 0
    }
  }
  return matrix;
}

//                                     PROCEDURE Unit (VAR (*OUT*) M: ARRAY OF ARRAY OF EltType;  N: CARDINAL);
func Unit (matrix Matrix2D) Matrix2D {
// Creates an N by N identity matrix, with all zeros except along the main diagonal.
  matrix = Zero(matrix);
  for diag := range matrix {
    matrix[diag][diag] = 1
  }
  return matrix;
}

//                                PROCEDURE Random (VAR (*OUT*) M: ARRAY OF ARRAY OF EltType;  r, c: CARDINAL);
func Random (matrix Matrix2D) Matrix2D {
// Creates matrix with random integers from 0..100

  for r := range matrix {
    for c := range matrix[r] {
      matrix[r][c] = rand.Intn(100);
    }
  }
  return matrix;
}


//     PROCEDURE Copy (A: ARRAY OF ARRAY OF EltType;  r, c: CARDINAL; VAR (*OUT*) B: ARRAY OF ARRAY OF EltType);
func Copy (Src,Dest Matrix2D) Matrix2D {
// Copies an rxc matrix A to B, by doing an element by element copy.  I don't think just copying
// pointers is correct.

  for r := range Src {
    for c := range Src[r] {
      Dest[r][c] = Src[r][c];
    }
  }
  return Dest;
}

/************************************************************************)
(*                      THE BASIC MATRIX OPERATIONS                     *)
(************************************************************************/

//  PROCEDURE Add (A, B : ARRAY OF ARRAY OF EltType;  r, c: CARDINAL; VAR (*OUT*) C: ARRAY OF ARRAY OF EltType);
func Add(A,B,C Matrix2D) Matrix2D {
// Computes C := A + B.

  for i := range A {
    for j := range A[i] {
      C[i][j] = A[i][j] + B[i][j];
    }
  }
  return C;
}

//   PROCEDURE Sub (A, B: ARRAY OF ARRAY OF EltType;  r, c: CARDINAL; VAR (*OUT*) C: ARRAY OF ARRAY OF EltType);
func Sub(A,B,C Matrix2D) Matrix2D {
// Computes C := A - B.

  for i := range A {
    for j := range A[i] {
      C[i][j] = A[i][j] - B[i][j];
    }
  }
  return C;
}


//   PROCEDURE Mul (A, B: ARRAY OF ARRAY OF EltType;  r, c1, c2: CARDINAL; VAR (*OUT*) C: ARRAY OF ARRAY OF EltType);
func Mul(A,B,C Matrix2D) Matrix2D {
// Computes C := A x B.  It will panic if the dimensions are not correct
  var temp float64;

  NumRowA := len(A);
  NumColA := len(A[0]); // all rows have same number of columns

  NumRowB := len(B)
  NumColB := len(B[0]); // all rows have same number of columns

  if NumColA != NumRowB { // error.  I guess I'll panic as I cannot think of anything better to do
    panic(" matrix mult panic because NumColA not equal to NumRowB");
  }

  for i := range A { // ranging over number of rows of A
    for j := range A[i] { // ranging over number of columns of A, and also number of rows of B
      temp = 0;
      for k := range B[i] { // ranging over number of columns of B
        temp += A[i][j] * B[j][k];
      }
      C[i][k] = temp;
    }
  }
  return C;
}

//  PROCEDURE ScalarMul (A: EltType;  B: ARRAY OF ARRAY OF EltType;  r, c: CARDINAL; VAR (*OUT*) C: ARRAY OF ARRAY OF EltType);
func ScalarMul(a EltType, B,C Matrix2D) Matrix2D {
// Computes C := a*B, where a is the scalar and B is the matrix.

  for i := range B {
    for j := range B[i] {
      C[i][j] = a * B[i][j];
    }
  }
  return C;
}



/************************************************************************)
(*                      SOLVING LINEAR EQUATIONS                        *)
(************************************************************************/

PROCEDURE GaussJ (A, B: ARRAY OF ARRAY OF EltType;
                     VAR (*OUT*) X: ARRAY OF ARRAY OF EltType;
                     N, M: CARDINAL);

    (* Solves the equation AX = B by Gauss-Jordan elimination.  In the  *)
    (* present version A must be square and nonsingular.                *)
    (* This approach to solving the equation is not the best available  *)
    (* - see below - but is included here anyway since it is popular.   *)
    (* Dimensions: A is NxN, B is NxM.                                  *)

PROCEDURE Solve (A, B: ARRAY OF ARRAY OF EltType;
                    VAR (*OUT*) X: ARRAY OF ARRAY OF EltType;
                    N, M: CARDINAL);

    (* Solves the equation AX = B.  In the present version A must be    *)
    (* square and nonsingular.                                          *)
    (* Dimensions: A is NxN, B is NxM.                                  *)

PROCEDURE Invert (A: ARRAY OF ARRAY OF EltType;
                     VAR (*OUT*) X: ARRAY OF ARRAY OF EltType;
                     N: CARDINAL);

    (* Inverts an NxN nonsingular matrix. *)

(************************************************************************)
(*                           EIGENVALUES                                *)
(************************************************************************)

PROCEDURE Eigenvalues (A: ARRAY OF ARRAY OF EltType;
                          VAR OUT W: ARRAY OF LONGCOMPLEX;
                          N: CARDINAL);

    (* Finds all the eigenvalues of an NxN matrix.    *)
    (* This procedure does not modify A.              *)

(************************************************************************)
(*                          SCREEN OUTPUT                               *)
(************************************************************************)

PROCEDURE Write (M: ARRAY OF ARRAY OF EltType;  r, c: CARDINAL;  places: CARDINAL);

    (* Writes the rxc matrix M to the screen, where each column *)
    (* occupies a field "places" characters wide.               *)

END Mat.

