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
// 24 Dec 2016 -- Seems to work.

import (
  "fmt"
  "bufio"
  "os"
  "strings"
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
    var A, B, C, D, E, F [][]float64;

    A = make([][]float64,Arows);

    B = make([][]float64,Brows);
//    C = make([][]float64,Brows);

    D = make([][]float64,Arows);
    E = make([][]float64,Arows);

    for i := range A {
      A[i] = make([]float64,Acols);
      D[i] = make([]float64,Acols);
      E[i] = make([]float64,Acols);
    }

    for i := range B {
      B[i] = make([]float64,Bcols);
//      C[i] = make([]float64,Bcols);
    }
        F = mat.NewMatrix(Brows,Bcols);  // testing NewMatrix, not in original code
        G := mat.NewMatrix(Arows,Acols);//  testing NewMatrix, not in original code

        fmt.Println("Test of simple matrix operations.");
        fmt.Println();
        fmt.Println();

//      Give a value to the A matrix.

        A = mat.Random(A);                                            //        Random (A, Arows, Acols);
        fmt.Println(" Matrix A is:");
	ss := mat.Write (A, 5);                                       //        Write (A, Arows, Acols, 5);
	for _,s := range ss {
          fmt.Print(s);
	}
	fmt.Println();

//      Give a value to the B matrix.

        B = mat.Random (B);                                              // Random (B, Brows, Bcols);
        fmt.Println(" Matrix B is:");
	ss = mat.Write (B, 5);
	for _,s := range ss {
          fmt.Print(s);
	}
	fmt.Println();


//      Try an addition (it will fail).
        C = mat.Add(A,B);
	if C == nil {
          fmt.Println("We can't compute A+B");
	}else{
		fmt.Println(" Trying to add A+B, which should fail.  It seems to have worked.  C is:");
		ss = mat.Write(C,5);
	        for _,s := range ss {
                  fmt.Print(s);
	        }
	        fmt.Println();
	}

        // Try a multiplication (it should work).

        C = mat.Mul(A, B);
        fmt.Println("C = A*B is");
	ss = mat.Write (C, 5);
	for _,s := range ss {
          fmt.Print(s);
	}
	fmt.Println();


        // Give a value to the D matrix.

        D = mat.Random(D);
        fmt.Println("Matrix D is");
        ss = mat.Write(D, 5);
	for _,s := range ss {
          fmt.Print(s);
	}
	fmt.Println();


        // Try another addition (this one should work).

        E = mat.Add(A, D);
        fmt.Println("E = A+D is");
        ss = mat.Write (E, 5);
	for _,s := range ss {
          fmt.Print(s);
	}
	fmt.Println();

// My new test code
	F = mat.Add(D,E); //   should fail
	fmt.Println(" F = D + E;");
	if F != nil {
          ss = mat.Write (F, 5);
	  for _,s := range ss {
            fmt.Print(s);
	  }
	  fmt.Println();
        }else{
         fmt.Println(" F = D + E failed");
	 F = mat.Random(F);
	}

	G = mat.Sub(F,E)  //   should fail
        fmt.Println(" G = F - E;");
	if G != nil {
          ss = mat.Write (G, 5);
	  for _,s := range ss {
            fmt.Print(s);
	  }
	  fmt.Println();
	}else{
          fmt.Print(" E - F failed ");
          G = mat.Random(G);
        }

        ss = mat.Write(D,4);
        for _,s := range ss {
          fmt.Print(s);
        }
        fmt.Println();
        ss = mat.Write(E,4);
        for _,s := range ss {
          fmt.Print(s);
        }
        
	H := mat.Mul(D,B);  // should work
	fmt.Println( "H = G*F, well, now D*B:");

	if H != nil {
          ss := mat.Write (H, 5);
	  for _,s := range ss {
            fmt.Print(s);
	  }
	  fmt.Println();
        }else{
	  fmt.Println(" H = G*F did not work, well, now D*B.");
	}

        Q := mat.Sub(A,A);
        fmt.Println(" Q = A - A");
        if Q != nil {
          ss = mat.Write(Q,4);
          for _,s := range ss{
            fmt.Print(s);
          }
          fmt.Println();
        }else{
          fmt.Println(" Q = A - A did not work.");
        }
       

        K := mat.NewMatrix(2,2);
        K = mat.Random(K);
        L := mat.NewMatrix(2,2);
        L = mat.Random(L);
        fmt.Println();
        fmt.Println(" K and then L, and then K*L");
        ss = mat.Write(K,4);
        for _,s := range ss{
          fmt.Print(s);
        }
        ss = mat.Write(L,4);
        for _,s := range ss{
          fmt.Print(s);
        }

        L = mat.Mul(K,L);

        ss = mat.Write(L,4);
        for _,s := range ss{
          fmt.Print(s);
        }
} //    END BasicTest;

//************************************************************************

func SolveTest() {

    // Solution of a linear equation. 

    const Arows = 4;
    const Acols = 4;
    const Brows = 4;
    const Bcols = 2;

    // var A [][]float64; // ARRAY [1..Arows],[1..Acols] OF LONGREAL;
    var B, C, D, X [][]float64;  // ARRAY [1..Brows],[1..Bcols] OF LONGREAL;

    A := make([][]float64,Arows);  // testing if create and assign works here.
    for i := range A {
      A[i] = make([]float64,Acols);
    }

    B = make([][]float64,Brows);
    C = make([][]float64,Brows);
    D = make([][]float64,Brows);
    X = make([][]float64,Brows);
    for i := range B {
      B[i] = make([]float64,Bcols);
      C[i] = make([]float64,Bcols);
      C[i] = make([]float64,Bcols);
      X[i] = make([]float64,Bcols);
    }


        fmt.Println ("SOLVING LINEAR ALGEBRAIC EQUATIONS");

        // Give a value to the A matrix.

        A = mat.Random (A);
        fmt.Println ("Matrix A is");
	ss := mat.Write (A, 4);
	for _,s := range ss {
          fmt.Print(s);
	}

        // Give a value to the B matrix.

        B = mat.Random(B);
        fmt.Println ("Matrix B is");
        ss = mat.Write (B, 4);
	for _,s := range ss {
          fmt.Print(s);
	}


        // Solve the equation AX = B.

        X = mat.Solve(A,B);   // X = mat.Solve(A, B, Arows, Bcols);
	Y := mat.GaussJ(A,B);  //  Y := mat.GaussJ(A, B, Arows, Bcols);

        // Write the solution.

	fmt.Println ("The solution X to AX = B is: X");
        ss = mat.Write (X, 4);
	for _,s := range ss {
          fmt.Print(s);
	}

	fmt.Println ("The solution X to AX = B is: Y");
        ss = mat.Write (Y, 4);
	for _,s := range ss {
          fmt.Print(s);
	}



        // Check that the solution looks right.

        C = mat.Mul(A,X);                            // Mul (A, X, Arows, Acols, Bcols, C);
        D = mat.Sub(B,C);                            // Sub (B, C, Brows, Bcols, D);
        fmt.Println ("As a check, AX-B evaluates to zero");
        ss = mat.Write(D,4);                         // Write (D, Brows, Bcols, 4);
	for _,s := range ss {
          fmt.Print(s);
	}
        fmt.Println();
        D = mat.BelowSmallMakeZero(D);
        ss = mat.Write(D,4);
	for _,s := range ss {
          fmt.Print(s);
	}
        fmt.Println();

} //    END SolveTest;

//************************************************************************

func SingularTest() {

    // Linear equation with singular coefficient matrix.

    const Arows = 2
    const Acols = 2;
    const Brows = 2;
    const Bcols = 1;

//    VAR A: ARRAY [1..Arows],[1..Acols] OF LONGREAL;
//        B, X: ARRAY [1..Brows],[1..Bcols] OF LONGREAL;

    A := mat.NewMatrix(Arows,Acols);
    B := mat.NewMatrix(Brows,Bcols);
    X := mat.NewMatrix(Brows,Bcols);

        if A == nil || B == nil || X == nil {
           fmt.Println(" Singular test failed in that a matrix came back nil from NewMatrix call.");
           return;
        }

        fmt.Println ("A SINGULAR PROBLEM.");

        // Give a value to the A matrix.

        A[0][0] = 1.0;
        A[0][1] = 2.0;
        A[1][0] = 2.0;
        A[1][1] = 4.0;
        fmt.Println ("Matrix A is:");
        ss := mat.Write (A, 4);
        for _,s := range ss {
          fmt.Print(s);
        }
        fmt.Println()

        // Give a value to the B matrix.

        B = mat.Random(B);
        fmt.Println ("Matrix B is:");
        ss = mat.Write(B, 4);
        for _,s := range ss {
          fmt.Print(s);
        }
        fmt.Println()


        // Try to solve the equation AX = B.

        X = mat.Solve(A,B);   // X = mat.Solve(A, B,Arows, Bcols);

        if X == nil { // it should be nil, as A is singular
          fmt.Println ("The equation AX = B could not be solved");
        }

} //    END SingularTest;

// ------------------------------------------------------------ InversionTest ------------------------

func InversionTest() {

    // Inverting a matrix, also an eigenvalue calculation.

    const N = 5;

//    VAR A, B, X: ARRAY [1..N],[1..N] OF LONGREAL;
//        W: ARRAY [1..N] OF LONGCOMPLEX;

        A := mat.NewMatrix(N,N);
        B := mat.NewMatrix(N,N);
        X := mat.NewMatrix(N,N);
        W := make([]complex128,N);

        fmt.Println ("INVERTING A SQUARE MATRIX");

        // Give a value to the A matrix.

        A = mat.Random(A);  // Random (A, N, N);
        fmt.Println ("Matrix A is");
        ss := mat.Write(A, 4);
        for _,s := range ss {
          fmt.Print(s);
        }
        fmt.Println();

        // Invert it.

        X = mat.Invert(A);   //  X = mat.Invert(A, N);

        // Write the solution.

        fmt.Println();
        fmt.Println ("The inverse of A is");
        ss = mat.Write (X, 4);
        for _,s := range ss {
          fmt.Print(s);
        }
        fmt.Println();


        // Check that the solution looks right.

        B = mat.Mul(A,X)            // Mul(A, X, N, N, N, B);
        fmt.Println();
        fmt.Println ("As a check, the product evaluates to the identity matrix");
        ss = mat.Write (B, 4);
        for _,s := range ss {
          fmt.Print(s);
        }
        fmt.Println();
        fmt.Println();
        B = mat.BelowSmallMakeZero(B);
        ss = mat.Write (B, 4);
        for _,s := range ss {
          fmt.Print(s);
        }
        fmt.Println();
        fmt.Println();

        pause();
        // CLS;

        fmt.Println();
        fmt.Println ("EIGENVALUES");
        fmt.Println();
        fmt.Println ("The eigenvalues of A are");
        W = mat.Eigenvalues (A);   // Eigenvalues (A, W, N);
        for j := range W {    // FOR j := 1 TO N DO
            fmt.Print ("    ");
            fmt.Print(W[j]);
            fmt.Println();
        }
        fmt.Println();
        for _,w := range W {  // just to see if this also works
          fmt.Printf("  %9G",w);
        }
        fmt.Println();

        pause();
        fmt.Println ("The eigenvalues of its inverse are");
        W = mat.Eigenvalues (X);  // Eigenvalues (X, W, N);
        for _,w := range W {   // FOR j := 1 TO N DO
          fmt.Printf("  %9G",w);       //  fmt.Println ("    ");  WriteCx (W[j], 5);  WriteLn;
        }
        fmt.Println();

        pause();

} //    END InversionTest;

/************************************************************************)
(*                              MAIN PROGRAM                            *)
(************************************************************************/

func main() {

//    scanner := bufio.NewScanner(os.Stdin)

    BasicTest();
    pause()
    SolveTest();
    pause();
    SingularTest();
    pause();
    InversionTest();
}

func pause() {
  scanner := bufio.NewScanner(os.Stdin)
  fmt.Print(" pausing ... hit <enter>");
  scanner.Scan();
  answer := scanner.Text();
  if err := scanner.Err(); err != nil {
    fmt.Fprintln(os.Stderr, "reading standard input:", err)
    os.Exit(1);
  }
  ans := strings.TrimSpace(answer);
  ans = strings.ToUpper(ans);
  fmt.Println(ans);
}



// END MatTest.
