package main // queen2 that will take the problem to a 20 x 20 board.

import (
	"fmt"
	"strings"
	"time"
)

// The 8 Queens Puzzle from the Go project I'm doing now.  The goal is placing the 8 queens on a std 8 x 8 chessboard so that no queens threaten one another.
// For this size board, there are 12 distinct solutions.  The total of 92 solutions can be reduced by eliminating the duplicates that are derived from the fundamental solution by
// a rotation or reflection operation.
// Here we'll assume a square board.
// The key problem is to reduce the solution space to a manageable size.  PlaceQueen1 -> PlaceQueen2 did some of that, but this project will reduce the solution space much further.
// Each square has 2 states, queen or no queen.  So the solution space is 2^N*N.  He changes the problem so that a queen is placed only once in each column, so it becomes N^N.

// const numRowsCols = 5 // to start he wants 5.  This took 0.0865 sec to find a solution using method 1.
// const numRowsCols = 6 // This took 133.56 sec to find a solution using method 1.  This seems to be much more than a quadratic function.  Goes from 1E-2 to 1E2, which is 4 orders of magnitude.
// const numRowsCols = 5 // using method 2.  Took 0.0603 sec.
// const numRowsCols = 6 // using method 2.  Took 0.1104 sec.
// const numRowsCols = 7 // using method 2.  Took 4.19 sec.  Now using method 4 time = 0 on Windows11.
// const numRowsCols = 8 // using method 4.  Took ~513 us, or 0.000513 s
// const numRowsCols = 9 // using method 4.  Took ~500 us, or 0.0005 s.  Some runs on Win11 showed 0 sec.
// const numRowsCols = 10 // using method 4.  Took ~515 us, or 0.000515 s.  A few runs on Win11 showed 0 sec.
// const numRowsCols = 11 // using method 4.  Also took ~515 us, or 0.000515 s.  No runs on Win11 showed 0 sec.
// const numRowsCols = 12 // using method 4.  Took ~2.57 ms, or 0.00257 s.
// const numRowsCols = 13 // using method 4.  Took ~1.54 ms, or 0.00154 s.
// const numRowsCols = 14 // using method 4.  Took ~29 ms, or 0.029 s.
// const numRowsCols = 15 // using method 4.   Took ~25.5 ms, or 0.0255 s.
// const numRowsCols = 16 // using method 4.  Took ~237 ms, or 0.237 s.
// const numRowsCols = 17 // using method 4.  Took ~155 ms, or 0.155 s.
// const numRowsCols = 18 // using method 4.  Took ~1.33 s
// const numRowsCols = 19 // using method 4.  Took ~90 ms, or 0.09 s
const numRowsCols = 20 // using method 4.  Took ~8.5 s.

func makeBoard(rows int) [][]string { // I had to google this to get it right.
	aBoard := make([][]string, rows)
	for i := range aBoard {
		aBoard[i] = make([]string, rows)
	}

	for i := range aBoard[0] {
		for j := range aBoard[i] {
			aBoard[i][j] = "."
		}
	}
	return aBoard
}

func dumpBoard(board [][]string) {
	fmt.Println()
	for i := range board {
		for j := range board[i] {
			fmt.Printf("%s ", board[i][j])
		}
		fmt.Println()
	}
	fmt.Println()
}

func seriesIsLegal(board [][]string, r0, c0, dr, dc int) bool {
	numRows := len(board)
	var numQueens int

	r := r0
	c := c0

	for {
		if r >= numRows || c >= numRows || r < 0 || c < 0 {
			//fmt.Printf(" about to break in seriesIsLegal.  r = %d, c = %d\n", r, c)
			break
		}
		//fmt.Printf(" board[%d][%d] = %s\n", r, c, board[r][c])
		if strings.ToUpper(board[r][c]) == "Q" {
			numQueens++
		}
		r += dr
		c += dc
	}
	//fmt.Printf(" NumQueens = %d\n", numQueens)
	if numQueens < 2 {
		return true
	}
	return false
}

func boardIsLegal(board [][]string) bool {
	numRows := len(board)

	// check rows
	for r := range board {
		good := seriesIsLegal(board, r, 0, 0, 1)
		if !good {
			return false
		}
	}

	// check cols
	for c := range board {
		good := seriesIsLegal(board, 0, c, 1, 0)
		if !good {
			return false
		}
	}

	// check diagonals.  I have to check 4 diagonal loops
	for r := range board {
		good := seriesIsLegal(board, r, 0, 1, 1) // go down
		if !good {
			return false
		}
	}
	for c := range board {
		good := seriesIsLegal(board, 0, c, 1, 1) // go across
		if !good {
			return false
		}
	}
	for c := numRows - 1; c >= 0; c-- {
		good := seriesIsLegal(board, 0, c, 1, -1) // go across starting from the top right corner.
		if !good {
			return false
		}
	}
	for r := range board {
		good := seriesIsLegal(board, r, numRows-1, 1, -1) // go down starting from top right corner.
		if !good {
			return false
		}
	}

	return true
}

func boardIsASolution(board [][]string) bool {
	var queens int
	legal := boardIsLegal(board)
	if !legal {
		return false
	}
	for i := range board {
		for j := range board[i] {
			if strings.ToUpper(board[i][j]) == "Q" {
				queens++
			}
		}
	}
	return queens == numRowsCols
}

func placeQueens1(board [][]string, numRows, r, c int) bool {
	// This function does not modify the board if the recursive calls find a solution.  So if a solution is found, it is sitting in the board as is.
	// This brute force method takes a long time.  The author doesn't try above a 6 x 6 board.  It's more than a quadratic function.
	if r >= numRows { // finished examining every position and have fallen off the board
		return boardIsASolution(board)
	}
	var nextR, nextC int
	nextR = r
	nextC = c + 1
	if nextC >= numRows {
		nextR = r + 1
		nextC = 0
	}
	//fmt.Printf(" r = %d, c = %d, nextR = %d, nextC = %d\n", r, c, nextR, nextC)
	//exit := pause()
	//if exit {
	//	os.Exit(1)
	//}
	test := placeQueens1(board, numRows, nextR, nextC) // test if don't place a queen at current (r,c).
	if test {
		return true
	}
	board[r][c] = "Q"
	test = placeQueens1(board, numRows, nextR, nextC) // test if do place a queen at current (r,c).
	if test {
		return true
	}
	// testing if don't place a queen at (r,c) and if do place a queen here.  Both returned false.  So there is no solution from this board.  Reset (r,c) and return false.
	board[r][c] = "."
	return false
}

func placeQueens2(board [][]string, numQueens, numRows, r, c int) bool {
	// This function does not modify the board if the recursive calls find a solution.  So if a solution is found, it is sitting in the board as is.
	// This brute force method takes a long time.  The author doesn't try above a 6 x 6 board.  It's more than a quadratic function.
	if numQueens == numRows {
		//fmt.Printf(" numQueens = %d, numRows = %d.  These should be equal.\n", numQueens, numRows)
		return boardIsASolution(board)
	}
	if r >= numRows { // finished examining every position and have fallen off the board
		return boardIsASolution(board)
	}
	var nextR, nextC int
	nextR = r
	nextC = c + 1
	if nextC >= numRows {
		nextR = r + 1
		nextC = 0
	}
	//fmt.Printf(" r = %d, c = %d, nextR = %d, nextC = %d\n", r, c, nextR, nextC)
	//exit := pause()
	//if exit {
	//	os.Exit(1)
	//}
	test := placeQueens2(board, numQueens, numRows, nextR, nextC) // test if don't place a queen at current (r,c).
	if test {
		return true
	}
	board[r][c] = "Q"
	numPlaced := numQueens + 1
	test = placeQueens2(board, numPlaced, numRows, nextR, nextC) // test if do place a queen at current (r,c).
	if test {
		return true
	}
	// testing if don't place a queen at (r,c) and if do place a queen here.  Both returned false.  So there is no solution from this board.  Reset (r,c) and return false.
	board[r][c] = "."
	return false
}

func placeQueens4(board [][]string, c int) bool {
	numRows := len(board)
	if c == numRows { // then a queen is assigned to every column
		return boardIsLegal(board)
	}
	if c < numRows {
		good := boardIsLegal(board)
		if !good {
			return false
		}
	}
	for r := 0; r < numRows; r++ {
		board[r][c] = "Q"
		success := placeQueens4(board, c+1)
		if success {
			return true
		}
		board[r][c] = "."
		continue
	}
	return false
}

func main() {
	// const numRows = 5
	board := makeBoard(numRowsCols)

	start := time.Now()
	//success := placeQueens1(board, numRowsCols, 0, 0)
	//success := placeQueens2(board, 0, numRowsCols, 0, 0)
	success := placeQueens4(board, 0)

	elapsed := time.Since(start)
	if success {
		fmt.Println("Success!")
		dumpBoard(board)
	} else {
		fmt.Println("No solution")
	}
	fmt.Printf("Elapsed: %f seconds, %s\n", elapsed.Seconds(), elapsed.String())
}

func pause() bool {
	fmt.Print(" Pausing.  Hit <enter> to continue.  Or 'n' to exit  ")
	var ans string
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	if strings.Contains(ans, "n") {
		return true
	}
	return false
}
