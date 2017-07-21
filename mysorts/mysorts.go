package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const LastAlteredDate = "21 July 17"

func StraightInsertion(input []string) []string {
	n := len(input)
	for i := 1; i < n; i++ {
		x := input[i]
		j := i
		for (j > 0) && (x < input[j-1]) {
			input[j] = input[j-1]
			j--
		} // while j > 0  and .LT. operator used
		input[j] = x
	} // for i := 1 TO n-1
	return input
} // END StraightInsertion

func BinaryInsertion(a []string) []string {
	n := len(a)
	for i := 1; i < n; i++ {
		x := a[i]
		L := 1
		R := i
		for L < R {
			m := (L + R) / 2
			if a[m] <= x {
				L = m + 1
			} else {
				R = m
			} // END if a[m] <= x
		} //END for L < R
		for j := i; j <= R+1; j-- {
			a[j] = a[j-1]
		} //END for j := i TO R+1 BY -1 DO
		a[R] = x
	} // END for i := 1 to n-1
	return a
} // END BinaryInsertion

func StraightSelection(a []string) []string {
	for i := 0; i <= n-2; i++ {
		k := i
		x := a[i]
		for j := i + 1; j <= n-1; j++ {
			if a[j] < x {
				k = j
				x = a[k]
			} // end if
		} // END for j := i+1 TO n-1
		a[k] = a[i]
		a[i] = x
	} // END for i := 0 to n-2
	return a
} // END StraightSelection

func ShellSort(a []string) []string {
	const T = 4
	var h [T]int

	h[0] = 9
	h[1] = 5
	h[2] = 3
	h[3] = 1
	n := len(a)
	for m := 0; m < T; m++ {
		k := h[m]
		for i := k + 1; i < n; i++ {
			x := a[i]
			j := i - k
			for (j >= k) & (x < a[j]) {
				a[j+k] = a[j]
				j = j - k
			} // END for/while (j >= k) & (x < a[j]) DO
			a[j+k] = x
		} // END FOR i := k+1 TO n-1 DO
	} // END FOR m := 0 TO T-1 DO
	return a
} //END ShellSort

// -----------------------------------------------------------
func sift(L, R int) {
	i := L
	j := 2*i + 1
	x := a[i]
	if (j < R) && (a[j] < a[j+1]) {
		j += 1
	} // end if
	for (j <= R) && (x < a[j]) { // was while in original code
		a[i] = a[j]
		i = j
		j = 2*j + 1
		if (j < R) && (a[j] < a[j+1]) {
			j += 1
		} // end if
	} //END for/while (j <= R) & (x < a[j]) DO
	a[i] = x
} // END sift;

func HeapSort(a []string) []string {

	L := n / 2
	R := n - 1
	for L > 0 {
		L--
		func(L, R) {
			sift(L, R)
		}(L, R)
	} // END for-while L>0
	for R > 0 {
		a[0], a[R] = a[R], a[0]
		R--
		func(L, R) {
			sift(L, R)
		}(L, R)
	} // END for-while R > 0
} // END HeapSort

//------------------------------------------------------------------------
func siftup(n int) { // items is global to this function which is called as an anonymous closure.
	i := n
	done := false
	for (i > 0) && !done { // Originally a while statement
		p := (i - 1) / 2
		if items[i] <= items[p] {
			done = true
		} else {
			items[i], items[p] = items[p], items[i]
			i = p
		} // END (* end if *)
	} // END (* end for-while *)
} // END siftup;

func siftdown(n int) { // items is global to this function which is called as an anonymous closure.
	i := 0
	c := 0
	done := false
	for (c < n) && !done { // originally a while statement
		c = 2*i + 1
		if c <= n {
			if (c < n) && (items[c] <= items[c+1]) {
				c++
			} // END if
			if items[c] <= items[i] {
				done := true
			} else {
				items[c], items[i] = items[i], items[c]
				i = c
			} // END if items[c] <= items[i]
		} // END if c <= n
	} // END (* for-while *);
	return items
} // END siftdown;

func anotherheapsort(items []string) []string {
	size := len(items)
	number := size - 1
	for index := 1; index <= number; index++ {
		func(idx int) {
			siftup(idx)
		}(index)
	} // END for

	for index := number; index >= 1; index-- {
		items[0], items[index] = items[index], items[0]
		func(idx int) {
			siftdown(idx)
		}(index - 1)
	} // END for
	return items
} // END anotherheapsort;

//-------------------------------------------------------------

func qsort(L, R int) {
	i := L
	j := R
	x := a[(L+R)/2]
	for i <= j { // REPEAT in original code
		for a[i] < x {
			i++
		} // END a sub i < x
		for x < a[j] {
			j--
		} // END x < a sub j
		if i <= j {
			a[i], a[j] = a[j], a[i]
			i++
			j--
		} // end if i <= j
	} // UNTIL i > j;
	if L < j {
		qsort(L, j)
	}
	if i < R {
		qsort(i, R)
	}
} // END qsort;

func QuickSort(a []string) []string {
	n := len(a) - 1
	func(a, b) {
		qsort(a, b)
	}(0, n)
	return a
} // END QuickSort

func NonRecursiveQuickSort(a []string) []string {
	const M = 12
	var low, high [M]int // index stack

	n := len(a)
	s := 0
	low[0] = 0
	high[0] = n - 1

	for { // REPEAT take top request from stack
		L := low[s]
		R := high[s]
		DEC(s)
		for { // REPEAT (*partition a[L] ...  a[R]*)
			i := L
			j := R
			x := a[(L+R)/2]
			for { // REPEAT
				for a[i] < x {
					i++
				}
				for x < a[j] {
					j--
				}
				if i <= j {
					a[i], a[j] = a[j], a[i]
					i++
					j--
				}
				if i > j { // UNTIL i > j;
					break
				}
			}
			if i < R { // stack request to sort right partition
				s++
				low[s] = i
				high[s] = R
			}
			R := j      // now L and R delimit the left partition
			if L >= R { // UNTIL L >= R
				break
			}
		}
		if s == 0 { // UNTIL s = 0
			break
		}
	}
} // END NonRecursiveQuickSort

//mergesort.go
func mergeSort(L []string) []string {
	if len(L) < 2 {
		return L
	} else {
		middle := len(L) / 2 // middle needs to be of type int
		left := mergeSort(L[:middle])
		right := mergeSort(L[middle:])
		return merge(left, right)
	} // end if else clause
}

func merge(left, right []string) []string {
	sum := len(left) + len(right)
	result := make([]string, 0, sum)
	i := 0
	j := 0
	for i < len(left) && j < len(right) {
		if left[i] < right[j] {
			result.append(left[i])
			i += 1
		} else {
			result.append(right[j])
			j += 1
		}
	} // end while

	for i < len(left) {
		result.append(left[i])
		i += 1
	}

	for j < len(right) {
		result.append(right[j])
		j += 1
	}

	return result
}

func main() {
	var filesize int64
	fmt.Println(" Sort a slice of strings, using the different algorithms.  Last altered", LastAlteredDate)
	fmt.Println()

	if len(os.Args) <= 1 {
		fmt.Println(" Usage: mysorts <filename>")
		os.Exit(0)
	}

	Ext1Default := ".dat"
	Ext2Default := ".txt"
	OutDefault := ".sorted"

	commandline := getcommandline.GetCommandLineString()
	BaseFilename := filepath.Clean(commandline)
	Filename := ""
	FileExists := false

	if strings.Contains(BaseFilename, ".") {
		Filename = BaseFilename
		FI, err := os.Stat(Filename)
		if err == nil {
			FileExists = true
			filesize = FI.Size()
		}
	} else {
		Filename = BaseFilename + Ext1Default
		FI, err := os.Stat(Filename)
		if err == nil {
			FileExists = true
			filesize = FI.Size()
		} else {
			Filename = BaseFilename + Ext2Default
			FI, err := os.Stat(Filename)
			if err == nil {
				FileExists = true
				filesize = FI.Size()
			}
		}
	}

	if !FileExists {
		fmt.Println(" File ", BaseFilename, " or ", Filename, " does not exist.  Exiting.")
		os.Exit(1)
	}

	byteslice := make([]byte, 0, filesize+5) // add 5 just in case
	byteslice, err := ioutil.ReadFile(Filename)
	if err != nil {
		fmt.Println(" Error from ioutil.ReadFile when reading ", Filename, ".  Exiting.")
		os.Exit(1)
	}

	bytesbuffer := bytes.NewBuffer(byteslice)

	OutFilename := BaseFilename + OutDefault
	OutputFile, err := os.Create(OutFilename)
	if err != nil {
		fmt.Println(" Error while opening OutputFile ", OutFilename, ".  Exiting.")
		os.Exit(1)
	}
	defer OutputFile.Close()
	OutBufioWriter := bufio.NewWriter(OutputFile)
	defer OutBufioWriter.Flush()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(" Enter number of words for this run.  0 means full file: ")
	scanner.Scan()
	requestedwordcount := strconv.Atoi(scanner.Text())
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}
	if len(INBUF) == 0 {
		os.Exit(0)
	}

	if requestedwordcount == 0 {
		requestedwordcount = filesize / 5
	}

	sliceofwords := make([]string, 0, requestedwordcount)

	for totalwords := 0; totalwords < requestedwordcount; totalwords++ { // Main processing loop
		word, err := bytesbuffer.ReadString(' ')
		if err != nil {
			break
		}
		word = strings.TrimSpace(word)
		if len(word) < 4 {
			continue
		}
		sliceofwords = append(sliceofwords, word)
	}

	for _, w := range sliceofwords {
		fmt.Print(w, " ")
	}

	t0 := time.Now()
	sortedsliceofwords := StraightSelection(sliceofwords)
	StraightSelectionTime := time.Since(t0)
	fmt.Println(" StraightSelection:", StraightSelectionTime)
	for _, w := range sortedsliceofwords {
		fmt.Print(w, " ")
	}
	fmt.Println()

	t1 := time.Now()
	sliceofsortedwords := StraightInsertion(sliceofwords)
	StaightInsertionTime := time.Since(t1)
	fmt.Println(" StraightInsertion:", StraightSelectionTime)
	for _, w := range sliceofsortedwords {
		fmt.Print(w, " ")
	}
	fmt.Println()

	t2 := time.Now()
	BinaryInsertionSortedWords := BinaryInsertion(sliceofwords)
	BinaryInsertionTime := time.Since(t2)
	fmt.Println(" BinaryInsertion:", BinaryInsertionTime)
	for _, w := range BinaryInsertionSortedWords {
		fmt.Print(w, " ")
	}
	fmt.Println()

	t3 := time.Now()
	ShellSortedWords := ShellSort(sliceofwords)
	ShellSortedTime := time.Since(t3)
	fmt.Println(" ShellSort:", ShellSortedTime)
	for _, w := range BinaryInsertionSortedWords {
		fmt.Print(w, " ")
	}
	fmt.Println()

	t4 := time.Now()
	HeapSortedWords := HeapSort(sliceofwords)
	HeapSortedTime := time.Since(t4)
	fmt.Println(" HeapSort:", HeapSortedTime)
	for _, w := range HeapSortedWords {
		fmt.Print(w, " ")
	}
	fmt.Println()

	t5 := time.Now()
	AnotherHeapSortedWords := anotherheapsort(sliceofwords)
	AnotherHeapTime := time.Since(t5)
	fmt.Println(" anotherheapsort:", AnotherHeapTime)
	for _, w := range AnotherHeapSortedWords {
		fmt.Print(w, " ")
	}
	fmt.Println()

	t6 := time.Now()
	QuickSortedWords := QuickSort(sliceofwords)
	QuickSortedTime := time.Since(t6)
	fmt.Println(" QuickSort:", QuickSortedTime)
	for _, w := range QuickSortedWords {
		fmt.Print(w, " ")
	}
	fmt.Println()

	t7 := time.Now()
	MergeSortedWords := mergeSort(sliceofwords)
	MergeSortTime := time.Since(t7)
	fmt.Println(" mergeSort:", MergeSortTime)
	for _, w := range QuickSortedWords {
		fmt.Print(w, " ")
	}
	fmt.Println()

}
