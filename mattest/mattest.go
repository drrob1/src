package main;

        /********************************************************)
        (*                                                      *)
        (*              Test of Matrices module                 *)
        (*                                                      *)
        (*  Programmer:         P. Moylan                       *)
        (*  Last edited:        15 August 1996                  *)
        (*  Status:             Working                         *)
        (*                                                      *)
        (********************************************************/

// REVISION HISTORY
// ================
// 21 Dec 2016 -- Started conversion to Go from old Modula-2 source.  We'll see how long this takes.

import (
  "fmt"
  "strconv"
  "math"
  "math/cmplx"
  "math/rand"
  "time"
//
  "vec"
  "mat"
)

// FROM Mat IMPORT (* proc *)  Zero, Write, Add, Sub, Mul, Random, Solve, GaussJ, Invert, Eigenvalues;


func BasicTest() {

    // Checks some simple matrix operations.

    const Arows = 2;
    const Acols = 3;
    const Brows = 3;
    const Bcols = 2;

//                       VAR A, D, E: ARRAY [1..Arows],[1..Acols] OF LONGREAL;
//                           B, C:    ARRAY [1..Brows],[1..Bcols] OF LONGREAL;
    var A, B, C, D, E [][]float64;

    A := make([][]float64,Arows);

    B := make([][]float64,Brows);
    C := make([][]float64,Brows);

    D := make([][]float64,Arows);
    E := make([][]float64,Arows);

    for i := range A {
      A[i] = make([]float64,Acols);
      D[i] = make([]float64,Acols);
      E[i] = make([]float64,Acols);
    }

    for i := range B {
      B[i] = make([]float64,Bcols);
      C[i] = make([]float64,Bcols);
    }
        fmt.Println("TEST OF SIMPLE MATRIX OPERATIONS");
        fmt.Println();
        fmt.Println();

//      Give a value to the A matrix.

        Random (A, Arows, Acols);
        WriteString ("Matrix A is");  WriteLn;
        Write (A, Arows, Acols, 5);

        (* Give a value to the B matrix. *)

        Random (B, Brows, Bcols);
        WriteString ("Matrix B is");  WriteLn;
        Write (B, Brows, Bcols, 5);

        (* Try an addition (it will fail). *)

        WriteString ("We can't compute A+B");  WriteLn;

        (* Try a multiplication (it should work). *)

        Mul (A, B, Arows, Acols, Bcols, C);
        WriteString ("C = A*B is");  WriteLn;
        Write (C, Arows, Bcols, 5);

        (* Give a value to the D matrix. *)

        Random (D, Arows, Acols);
        WriteString ("Matrix D is");  WriteLn;
        Write (D, Arows, Acols, 5);

        (* Try another addition (this one should work). *)

        Add (A, D, Arows, Acols, E);
        WriteString ("E = A+D is");  WriteLn;
        Write (E, Arows, Acols, 5);

        PressAnyKey;
        (*CloseWindow (w);*)
} //    END BasicTest;

(************************************************************************)

PROCEDURE SolveTest;

    (* Solution of a linear equation. *)

    CONST Arows = 4;  Acols = 4;
          Brows = 4;  Bcols = 2;

    VAR A: ARRAY [1..Arows],[1..Acols] OF LONGREAL;
        B, C, D, X: ARRAY [1..Brows],[1..Bcols] OF LONGREAL;
        (*w: Window;*)

    BEGIN
        (*
        OpenWindow (w, black, brown, 0, 24, 0, 79, simpleframe, nodivider);
        SelectWindow (w);
        *)
        Reset;
        WriteString ("SOLVING LINEAR ALGEBRAIC EQUATIONS");
        WriteLn;

        (* Give a value to the A matrix. *)

        Random (A, Arows, Acols);
        WriteString ("Matrix A is");  WriteLn;
        Write (A, Arows, Acols, 4);

        (* Give a value to the B matrix. *)

        Random (B, Brows, Bcols);
        WriteString ("Matrix B is");  WriteLn;
        Write (B, Brows, Bcols, 4);

        (* Solve the equation AX = B. *)

        Solve (A, B, X, Arows, Bcols);
        (*GaussJ (A, B, X, Arows, Bcols);*)

        (* Write the solution. *)

        WriteString ("The solution X to AX = B is");  WriteLn;
        Write (X, Brows, Bcols, 4);

        (* Check that the solution looks right. *)

        Mul (A, X, Arows, Acols, Bcols, C);
        Sub (B, C, Brows, Bcols, D);
        WriteString ("As a check, AX-B evaluates to");  WriteLn;
        Write (D, Brows, Bcols, 4);

        PressAnyKey;
        (*CloseWindow (w);*)

    END SolveTest;

(************************************************************************)

PROCEDURE SingularTest;

    (* Linear equation with singular coefficient matrix. *)

    CONST Arows = 2;  Acols = 2;
          Brows = 2;  Bcols = 1;

    VAR A: ARRAY [1..Arows],[1..Acols] OF LONGREAL;
        B, X: ARRAY [1..Brows],[1..Bcols] OF LONGREAL;
        (*w: Window;*)

    BEGIN
        (*
        OpenWindow (w, black, brown, 0, 24, 0, 79, simpleframe, nodivider);
        SelectWindow (w);
        *)
        Reset;
        WriteString ("A SINGULAR PROBLEM");
        WriteLn;

        (* Give a value to the A matrix. *)

        A[1,1] := 1.0;
        A[1,2] := 2.0;
        A[2,1] := 2.0;
        A[2,2] := 4.0;
        WriteString ("Matrix A is");  WriteLn;
        Write (A, Arows, Acols, 4);

        (* Give a value to the B matrix. *)

        Random (B, Brows, Bcols);
        WriteString ("Matrix B is");  WriteLn;
        Write (B, Brows, Bcols, 4);

        (* Try to solve the equation AX = B. *)

        Solve (A, B, X, Arows, Bcols);

        WriteString ("The equation AX = B could not be solved");  WriteLn;

        PressAnyKey;
        (*CloseWindow (w);*)

    END SingularTest;

(************************************************************************)

PROCEDURE InversionTest;

    (* Inverting a matrix, also an eigenvalue calculation. *)

    CONST N = 5;

    VAR A, B, X: ARRAY [1..N],[1..N] OF LONGREAL;
        W: ARRAY [1..N] OF LONGCOMPLEX;
        (*w: Window;*)  j: CARDINAL;

    BEGIN
        (*
        OpenWindow (w, yellow, brown, 0, 24, 0, 79, simpleframe, nodivider);
        SelectWindow (w);
        *)
        Reset;
        WriteString ("INVERTING A SQUARE MATRIX");
        WriteLn;

        (* Give a value to the A matrix. *)

        Random (A, N, N);
        WriteString ("Matrix A is");  WriteLn;
        Write (A, N, N, 4);

        (* Invert it. *)

        Invert (A, X, N);

        (* Write the solution. *)

        WriteLn;
        WriteString ("The inverse of A is");  WriteLn;
        Write (X, N, N, 4);

        (* Check that the solution looks right. *)

        Mul (A, X, N, N, N, B);
        WriteLn;
        WriteString ("As a check, the product evaluates to");  WriteLn;
        Write (B, N, N, 4);

        PressAnyKey;
        CLS;
        
        WriteLn;  WriteString ("EIGENVALUES");  WriteLn;
        WriteString ("The eigenvalues of A are");  WriteLn;
        Eigenvalues (A, W, N);
        FOR j := 1 TO N DO
            WriteString ("    ");  WriteCx (W[j], 5);  WriteLn;
        END (*FOR*);

        PressAnyKey;
        WriteString ("The eigenvalues of its inverse are");  WriteLn;
        Eigenvalues (X, W, N);
        FOR j := 1 TO N DO
            WriteString ("    ");  WriteCx (W[j], 5);  WriteLn;
        END (*FOR*);

        PressAnyKey;
        (*CloseWindow (w);*)

    END InversionTest;

(************************************************************************)
(*                              MAIN PROGRAM                            *)
(************************************************************************)

BEGIN
(*
    BasicTest;
    SolveTest;
    SingularTest;
*)
    InversionTest;
END MatTest.
