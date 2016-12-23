// IMPLEMENTATION MODULE Vec;
package vec;

        /********************************************************)
        (*                                                      *)
        (*                 Vector arithmetic                    *)
        (*                                                      *)
        (*  Programmer:         P. Moylan                       *)
        (*  Last edited:        15 August 1996                  *)
        (*  Status:             Seems to be working             *)
        (*                                                      *)
        (*      Portability note: this module contains some     *)
        (*      open array operations which are a language      *)
        (*      extension in XDS Modula-2 but not part of       *)
        (*      ISO standard Modula-2.  So far I haven't worked *)
        (*      out how to do this in the standard language.    *)
        (*                                                      *)
        (********************************************************/

// REVISION HISTORY
// ================
// 20 Dec 2016 -- Started conversion to Go from old Modula-2 source.  We'll see how long this takes.

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

// type EltType float64;
type VectorPtr []float64;                                       // TYPE VectorPtr = POINTER TO ARRAY OF EltType;

/************************************************************************/

/************************************************************************)
(*                  CREATING AND DESTROYING VECTORS                     *)
(************************************************************************/

//  PROCEDURE NewVector (N: CARDINAL): VectorPtr;
func NewVector (N int ) VectorPtr {

/* Creates a vector of N elements. */
// old code basically did this: VAR result: VetorcPtr; NEW (result, N-1); RETURN result;

    vector := make(VectorPtr,N);
    return vector;
}
/************************************************************************
Commented out as this is not needed in Go.  The garbage collector takes care of this.
PROCEDURE DisposeVector (VAR (*INOUT*) V: VectorPtr;  N: CARDINAL);

    (* Deallocates a vector of N elements. *)

    BEGIN
        DISPOSE (V);
    END DisposeVector;

(************************************************************************)
(*                          COPYING A VECTOR                            *)
(************************************************************************/

//                          PROCEDURE Copy (A: ARRAY OF EltType;  N: CARDINAL; VAR (*OUT*) B: ARRAY OF EltType);
func Copy(A VectorPtr) VectorPtr {

    // Copies an N-element vector A to B.
    var B VectorPtr;

        for i := range A {   // FOR i := 0 TO N-1 DO
            B[i] = A[i];
        } //END FOR
        return B;
} //    END Copy;

/************************************************************************)
(*                          VECTOR ARITHMETIC                           *)
(************************************************************************/

//                     PROCEDURE Add (A, B: ARRAY OF EltType;  elts: CARDINAL; VAR (*OUT*) C: ARRAY OF EltType);
func Add(A, B VectorPtr) VectorPtr {

    // Computes C = A + B.

    var C VectorPtr;

        if len(A) != len(B) {
          panic(" Vector Add and lengths of Vector A and Vector B are not the same.");
        }

        for i := range A {
            C[i] = A[i] + B[i];
        }
        return C;
}

/************************************************************************/

//                     PROCEDURE Sub (A, B: ARRAY OF EltType;  elts: CARDINAL; VAR (*OUT*) C: ARRAY OF EltType);
func Sub(A, B VectorPtr) VectorPtr {

    // Computes C = A - B.  All vectors have elts elements.

    var C VectorPtr;

        if len(A) != len(B) {
          panic(" Vector Sub and lengths of Vector A and Vector B are not the same.");
        }

        for i := range A {
            C[i] = A[i] - B[i];
        }
        return C;
}

/************************************************************************/

//PROCEDURE Mul(A:ARRAY OF ARRAY OF EltType;B:ARRAY OF EltType;N1,N2: CARDINAL;VAR(*OUT*)C: ARRAY OF EltType);
func Mul(A []VectorPtr, B VectorPtr) VectorPtr {

    // Computes C = A*B, where A is N1xN2 and B is N2x1.
    // C = A*B, where A is a 2D matrix, and B is a column vector, so result must also be a column vector

    var sum float64;
    var C VectorPtr;

        for i := range A {          // FOR i := 0 TO N1-1 DO  range over the 2D matrix
            sum = 0;
            for j := range B {      //   FOR j := 0 TO N2-1 DO  range over the column vector
                sum += A[i][j]*B[j];
            } // END FOR j
            C[i] = sum;
        } // END FOR i
        return C;
} //    END Mul;

/************************************************************************/

// PROCEDURE ScalarMul (A: EltType;  B: ARRAY OF EltType;  elts: CARDINAL; VAR (*OUT*) C: ARRAY OF EltType);
func ScalarMul(A float64, B VectorPtr) VectorPtr {

    // Computes C = A*B, where A is scalar and B is a vector

    var C VectorPtr;

        for i := range B { // FOR i := 0 TO elts-1 DO  range over the B vector
            C[i] = A * B[i];
        } // END FOR i
        return C;
} //    END ScalarMul;

/************************************************************************)
(*                              OUTPUT                                  *)
(************************************************************************/

//                                      PROCEDURE Write (V: ARRAY OF EltType;  N: CARDINAL;  places: CARDINAL);
func Write(V VectorPtr, places int) []string {

    /* Writes the N-element vector V to the screen, where each  *)
    (* column occupies a field "places" characters wide.        */

    // VAR i: CARDINAL;

        OutputStringSlice := make([]string,0,20);
        for i := range V {
            ss := strconv.FormatFloat(float64(V[i]),'G',places,64)
            OutputStringSlice = append(OutputStringSlice,fmt.Sprintf("  %s",ss));
        }
        OutputStringSlice = append(OutputStringSlice,"\n");
        return OutputStringSlice;
} //    END Write;

// END Vec.

