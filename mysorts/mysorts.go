package main

import (
	"bufio"
	"bytes"
	"fmt"
	"getcommandline"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const LastAlteredDate = "15 May 2020"

/*
  REVISION HISTORY
  ----------------
  July 2017 -- First version
  26 July 17 -- Will try to learn delve (dlv) by using it to debug the routines here that don't work.
   7 Aug  17 -- Thinking about a mergeSort with an insertionshort below, maybe 5 elements.
   8 Nov  17 -- Added comparing to sort.Slice.  I need to remember how I did this, so it will take a day or so.
  10 July 19 -- Added better comments and output strings.
  28 July 19 -- Adding Stable, and SliceStable
  29 July 19 -- Changing some formating of the output.
  15 May  20 -- Fixed ShellSort after starting to read High Performance Go by Ron Stephen.  I then remembered that ShellSort
                  is a modified BubbleSort, so I coded it as such.  And now it works.
*/

// -----------------------------------------------------------
func StraightInsertion(input []string) []string {
	n := len(input)
	for i := 1; i < n; i++ {
		x := input[i]
		j := i
		for (j > 0) && (x < input[j-1]) {
			input[j] = input[j-1]
			j--
		}
		input[j] = x
	} // for i := 1 TO n-1
	return input
} // END StraightInsertion

// -----------------------------------------------------------
func BinaryInsertion(a []string) []string {
	n := len(a)
	for i := 1; i < n; i++ {
		x := a[i]
		L := 0 // I think the mistake was here, where I first set L to 1.
		R := i
		for L < R {
			m := (L + R) / 2
			if a[m] <= x {
				L = m + 1
			} else {
				R = m
			}
		} //END while L < R

		for j := i; j >= R+1; j-- {
			a[j] = a[j-1]
		} //END for j := i TO R+1 BY -1 DO
		a[R] = x
	} // END for i :=
	return a
} // END BinaryInsertion

// -----------------------------------------------------------
func StraightSelection(a []string) []string {
	n := len(a)
	for i := 0; i < n-1; i++ {
		k := i
		x := a[i]
		for j := i + 1; j <= n-1; j++ {
			if a[j] < x {
				k = j
				x = a[k]
			}
		} // END for j := i+1 TO n-1
		a[k] = a[i]
		a[i] = x
	} // END for i := 0 to n-2
	return a
} // END StraightSelection

// -----------------------------------------------------------
func BadShellSort(a []string) []string {
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
			for (j >= k) && (x <= a[j]) {
				a[j+k] = a[j]
				j = j - k
			} // END for/while (j >= k) & (x < a[j]) DO
			a[j+k] = x
		} // END FOR i := k+1 TO n-1 DO
	} // END FOR m := 0 TO T-1 DO
	return a
} //END BadShellSort

/* -----------------------------------------------------------
{{{
   Bubblesort from "Essential Algorithms" by Rod Stephens.  This is pseudo-code

Bubblesort(Data: values[])
    // Repeat until the array is sorted.
    Boolean: not_sorted = True
    While (not_sorted)
        // Assume we won't find a pair to swap.
        not_sorted = False
        // Search the array for adjacent items that are out of order.
        For i = 0 To <length of values> - 1
            // See if items i and i - 1 are out of order.
            If (values[i] < values[i - 1]) Then
                // Swap them.
                Data: temp = values[i]
                values[i] = values[i - 1]
                values[i - 1] = temp
                // The array isn't sorted after all.
                not_sorted = True
            End If
        Next i
    End While
End Bubblesort
}}}
*/

// -----------------------------------------------------------
func ShellSort(a []string) []string { // revisiting this as I'm reading "High Performance Go."
	var h = []int{9, 5, 3, 1}

	n := len(a)
    //	t0 := time.Now()

	for _, k := range h {
		//k := h[m] when m is the index into h.  I decided that this form of the for range loop made more sense, as I do not need the actual index into h.
        //		fmt.Println(" ShellSort:  k=", k, ", n =", n)
		if k >= n {
			continue
		}

		for { // loop until sorted
			sorted := true
			for i := k; i < n; i++ {
				if a[i] < a[i-k] {
					a[i], a[i-k] = a[i-k], a[i]
					sorted = false
					//fmt.Println("  ShellSort:  i =", i, ", sorted=", sorted)
				}
			} // END FOR i := k TO last item DO
			if sorted {
				break
			}
			//elapsed := time.Since(t0)
			//if elapsed > 30*time.Second { return a }
		} // end loop until sorted
	} // END FOR range h
	return a
} //END ShellSort

// -----------------------------------------------------------
func sift(a []string, L, R int) []string {
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
	return a
} // END sift;

// -----------------------------------------------------------
func HeapSort(a []string) []string {
	n := len(a)
	L := n / 2
	R := n - 1
	for L > 0 {
		L--
		a = sift(a, L, R)
	} // END for-while L>0
	for R > 0 {
		a[0], a[R] = a[R], a[0]
		R--
		a = sift(a, L, R)
	} // END for-while R > 0
	return a
} // END HeapSort

//------------------------------------------------------------------------
func siftup(items []string, n int) []string { // items is global to this function which is called as an anonymous closure.
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
	return items
} // END siftup;

// -----------------------------------------------------------
func siftdown(items []string, n int) []string { // items is global to this function which is called as an anonymous closure.
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
				done = true
			} else {
				items[c], items[i] = items[i], items[c]
				i = c
			} // END if items[c] <= items[i]
		} // END if c <= n
	} // END (* for-while *);
	return items
} // END siftdown;

// -----------------------------------------------------------
func anotherheapsort(items []string) []string {
	size := len(items)
	number := size - 1
	for index := 1; index <= number; index++ {
		items = siftup(items, index)
	} // END for

	for index := number; index >= 1; index-- {
		items[0], items[index] = items[index], items[0]
		items = siftdown(items, index)
	} // END for
	return items
} // END anotherheapsort;

//-------------------------------------------------------------

func qsort(a []string, L, R int) []string {
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
		a = qsort(a, L, j)
	}
	if i < R {
		a = qsort(a, i, R)
	}
	return a
} // END qsort;

// -----------------------------------------------------------
func QuickSort(a []string) []string {
	n := len(a) - 1
	a = qsort(a, 0, n)
	return a
} // END QuickSort

// -----------------------------------------------------------
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
		s--
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
	return a
} // END NonRecursiveQuickSort

// -----------------------------------------------------------
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

// -----------------------------------------------------------
func merge(left, right []string) []string {
	sum := len(left) + len(right)
	result := make([]string, 0, sum)
	i := 0
	j := 0
	for i < len(left) && j < len(right) {
		if left[i] < right[j] {
			result = append(result, left[i])
			i += 1
		} else {
			result = append(result, right[j])
			j += 1
		}
	} // end while

	for i < len(left) {
		result = append(result, left[i])
		i += 1
	}

	for j < len(right) {
		result = append(result, right[j])
		j += 1
	}

	return result
}

// -----------------------------------------------------------
// modified mergesort.go
func ModifiedMergeSort(L []string) []string {
	if len(L) < 6 {
		L = StraightInsertion(L)
		return L
	} else {
		middle := len(L) / 2 // middle needs to be of type int
		left := ModifiedMergeSort(L[:middle])
		right := ModifiedMergeSort(L[middle:])
		return merge(left, right)
	} // end if else clause
}

//-----------------------------------------------------------------------+
//                               MAIN PROGRAM                            |
//-----------------------------------------------------------------------+

func main() {
	var filesize int64
	fmt.Println(" Sort a slice of strings, using the different algorithms.  Last altered", LastAlteredDate)
	fmt.Println()

	// File I/O.  Construct filenames
	if len(os.Args) <= 1 {
		fmt.Println(" Usage: mysorts <filename>")
		os.Exit(0)
	}

	Ext1Default := ".dat"
	Ext2Default := ".txt"
	OutDefault := ".sorted"

	date := time.Now()
	datestring := date.Format("Mon Jan 2 2006 15:04:05 MST") // written to output file below.

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
	//	OutputFile, err := os.Create(OutFilename)
	OutputFile, err := os.OpenFile(OutFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println(" Error while opening OutputFile ", OutFilename, ".  Exiting.")
		os.Exit(1)
	}
	defer OutputFile.Close()
	OutBufioWriter := bufio.NewWriter(OutputFile)
	defer OutBufioWriter.Flush()
	_, err = OutBufioWriter.WriteString("------------------------------------------------------\n")
	_, err = OutBufioWriter.WriteString(datestring)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)

	// Read in the words to sort
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(" Enter number of words for this run.  0 means full file: ")
	scanner.Scan()
	answer := scanner.Text()
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}
	requestedwordcount, err := strconv.Atoi(answer)
	if err != nil {
		fmt.Println(" No valid answer entered.  Will assume 0.")
		requestedwordcount = 0
	}

	if requestedwordcount == 0 {
		requestedwordcount = int(filesize / 7)
	}

	s := fmt.Sprintf(" filesize = %d, requestedwordcount = %d \n", filesize, requestedwordcount)
	OutBufioWriter.WriteString(s)
	mastersliceofwords := make([]string, 0, requestedwordcount)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)

	for totalwords := 0; totalwords < requestedwordcount; totalwords++ { // Main processing loop
		word, err := bytesbuffer.ReadString('\n')
		if err != nil {
			break
		}
		word = strings.TrimSpace(word)
		//	word = strings.ToLower(strings.TrimSpace(word))
		if len(word) < 4 {
			continue
		}
		mastersliceofwords = append(mastersliceofwords, word)
	}

	numberofwords := len(mastersliceofwords)

	allowoutput := false
	if numberofwords < 50 {
		allowoutput = true
	}

	// make the sliceofwords
	if allowoutput {
		fmt.Println("master before:", mastersliceofwords)
	}
	sliceofwords := make([]string, numberofwords)

	fmt.Println()
	fmt.Println()

	// sort.StringSlice method
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("slice before first sort.StringSlice:", sliceofwords)
	}
	NativeWords := sort.StringSlice(sliceofwords)
	t9 := time.Now()
	NativeWords.Sort()
	NativeSortTime := time.Since(t9)
	NativeSortTimeNano := NativeSortTime.Nanoseconds()
	s = fmt.Sprintf(" after NativeSort: %s, %d ns \n", NativeSortTime.String(), NativeSortTimeNano)
	fmt.Println(s)
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	//	s = fmt.Sprintf("%v\n", NativeWords)
	//	_, err = OutBufioWriter.WriteString(s)
	//	check(err)
	if allowoutput {
		for _, w := range NativeWords {
			fmt.Print(w, " ")
		}
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// StraightSelection
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Print(" sliceofwords before StraightSelection: ")
		for _, w := range sliceofwords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}

	t0 := time.Now()
	sortedsliceofwords := StraightSelection(sliceofwords)
	StraightSelectionTime := time.Since(t0)
	StraightSelectionTimeNano := StraightSelectionTime.Nanoseconds()
	s = fmt.Sprintf(" After StraightSelection: %s, %d ns \n", StraightSelectionTime.String(), StraightSelectionTimeNano)
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
	//	s = fmt.Sprintf("%v\n", sortedsliceofwords)
	//	_, err = OutBufioWriter.WriteString(s)
	//	check(err)
	if allowoutput {
		for _, w := range sortedsliceofwords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// StraightInsertion
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before StraightInsertion:", sliceofwords)
	}
	t1 := time.Now()
	sliceofsortedwords := StraightInsertion(sliceofwords)
	StraightInsertionTime := time.Since(t1)
	s = fmt.Sprintf(" After StraightInsertion: %s, %d ns \n", StraightInsertionTime.String(), StraightInsertionTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
	//	s = fmt.Sprintf("%v \n", sliceofsortedwords)
	//	_, err = OutBufioWriter.WriteString(s)
	//	check(err)
	if allowoutput {
		for _, w := range sliceofsortedwords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// BinaryInsertion
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before BinaryInsertion:", sliceofwords)
	}
	t2 := time.Now()
	BinaryInsertionSortedWords := BinaryInsertion(sliceofwords)
	BinaryInsertionTime := time.Since(t2)
	s = fmt.Sprintf(" After BinaryInsertion: %s, %d ns \n", BinaryInsertionTime.String(), BinaryInsertionTime.Nanoseconds())
	fmt.Println(s)
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	if allowoutput {
		for _, w := range BinaryInsertionSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// ShellSort --   05/15/2020 1:01:16 PM will try again.  It works.
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before ShellSort:", sliceofwords)
	}
	t3 := time.Now()
	ShellSortedWords := ShellSort(sliceofwords)
	ShellSortedTime := time.Since(t3)
	s = fmt.Sprintf(" After ShellSort: %s \n", ShellSortedTime.String())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println(" ShellSort:", ShellSortedTime)
	if allowoutput {
		for _, w := range ShellSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	fmt.Println()
	/*  */

	// HeapSort
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before HeapSort:", sliceofwords)
	}
	t4 := time.Now()
	HeapSortedWords := HeapSort(sliceofwords)
	HeapSortedTime := time.Since(t4)
	s = fmt.Sprintf(" After HeapSort: %s, %d ns \n", HeapSortedTime.String(), HeapSortedTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range HeapSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// AnotherHeapSort -- doesn't work
	/*	Does not sort correctly, but does not panic.
		copy(sliceofwords, mastersliceofwords)
		if allowoutput {
			fmt.Println("before AnotherHeapSort:", sliceofwords)
		}
		t5 := time.Now()
		AnotherHeapSortedWords := anotherheapsort(sliceofwords)
		AnotherHeapTime := time.Since(t5)
		fmt.Println(" anotherheapsort:", AnotherHeapTime)
		if allowoutput {
			for _, w := range AnotherHeapSortedWords {
				fmt.Print(w, " ")
			}
			fmt.Println()
		}
		fmt.Println()

	*/

	// QuickSort
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before QuickSort:", sliceofwords)
	}
	t6 := time.Now()
	QuickSortedWords := QuickSort(sliceofwords)
	QuickSortedTime := time.Since(t6)
	s = fmt.Sprintf(" After QuickSort: %s, %d ns \n", QuickSortedTime.String(), QuickSortedTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range QuickSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// MergeSort
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before MergeSort:", sliceofwords)
	}
	t7 := time.Now()
	MergeSortedWords := mergeSort(sliceofwords)
	MergeSortTime := time.Since(t7)
	s = fmt.Sprintf(" After mergeSort: %s, %d ns \n", MergeSortTime.String(), MergeSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range MergeSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// ModifiedMergeSort
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before ModifiedMergeSort:", sliceofwords)
	}
	t7a := time.Now()
	ModifiedMergeSortedWords := ModifiedMergeSort(sliceofwords)
	ModifiedMergeSortTime := time.Since(t7a)
	s = fmt.Sprintf(" After ModifiedMergeSort: %s, %d ns \n", ModifiedMergeSortTime.String(), ModifiedMergeSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range ModifiedMergeSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// NonRecursiveQuickSort -- doesn't work.  Ended with a panic
	/*	I think this paniced
		copy(sliceofwords, mastersliceofwords)
		if allowoutput {
			fmt.Println("before nonrecursiveQuickSort:", sliceofwords)
		}
		t8 := time.Now()
		NonRecursiveQuickSortedWords := NonRecursiveQuickSort(sliceofwords)
		NonRecursiveQuickedTime := time.Since(t8)
		fmt.Println(" NonRecursiveQuickSort:", NonRecursiveQuickedTime)
		if allowoutput {
			for _, w := range NonRecursiveQuickSortedWords {
				fmt.Print(w, " ")
			}
			fmt.Println()
		}
		fmt.Println()
	*/

	// sort.StringSlice
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before 2nd sort.StringSlice:", sliceofwords)
	}
	NativeWords = sort.StringSlice(sliceofwords)
	t9 = time.Now()
	NativeWords.Sort()
	NativeSortTime = time.Since(t9)
	s = fmt.Sprintf(" After NativeSort again: %s, %d ns \n", NativeSortTime.String(), NativeSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range NativeWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// sort.Sort
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before sort.Sort:", sliceofwords)
	}
	NativeWords = sort.StringSlice(sliceofwords)
	t9 = time.Now()
	sort.Sort(NativeWords)
	NativeSortTime = time.Since(t9)
	s = fmt.Sprintf(" After sort.Sort: %s, %d ns \n", NativeSortTime.String(), NativeSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range NativeWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// sort.Stable
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before sort.Stable:", sliceofwords)
	}
	NativeWords = sort.StringSlice(sliceofwords)
	t9 = time.Now()
	sort.Stable(NativeWords)
	NativeSortTime = time.Since(t9)
	s = fmt.Sprintf(" After sort.Stable: %s, %d ns \n", NativeSortTime.String(), NativeSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range NativeWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// sort.Strings
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before sort.Strings:", sliceofwords)
	}
	t10 := time.Now()
	sort.Strings(sliceofwords)
	StringsSortTime := time.Since(t10)
	s = fmt.Sprintf(" After sort.Strings: %s, %d ns \n", StringsSortTime.String(), StringsSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range sliceofwords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// sort.Slice
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before sort.Slice:", sliceofwords)
	}
	lessfunction := func(i, j int) bool {
		return sliceofwords[i] < sliceofwords[j]
	}
	t11 := time.Now()
	sort.Slice(sliceofwords, lessfunction)
	SliceSortTime := time.Since(t11)
	s = fmt.Sprintf(" After sort.Slice: %s, %d ns \n", SliceSortTime.String(), SliceSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range sliceofwords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// sort.SliceStable
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before sort.SliceStable:", sliceofwords)
	}
	t12 := time.Now()
	sort.SliceStable(sliceofwords, lessfunction)
	SliceStableSortTime := time.Since(t12)
	s = fmt.Sprintf(" After sort.SliceStable: %s, %d ns \n", SliceStableSortTime.String(), SliceStableSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range sliceofwords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// Wrap it up by writing number of words, etc.
	s = fmt.Sprintf(" requestedwordcount= %d, numberofwords= %d, len(mastersliceofwords)= %d \n",
		requestedwordcount, numberofwords, len(mastersliceofwords))
	_, err = OutBufioWriter.WriteString(s)
	if len(mastersliceofwords) > 1000 {
		fmt.Println(s)
		//		fmt.Println(" Number of words to be sorted is", len(mastersliceofwords))
	}
	_, err = OutBufioWriter.WriteString("------------------------------------------------------\n")
	check(err)

	// Close the output file and exit
	OutBufioWriter.Flush()
	OutputFile.Close()
}

//===========================================================
func check(e error) {
	if e != nil {
		panic(e)
	}
}

/*
  Timing for first full data file, ScienceOfHappiness.dat, ~67,500 words.  More complete information is now in the .sorted file

 after NativeSort: 47.745145ms
 After StraightSelection: 47.340594191s
 After StraightInsertion: 14.074816209s
 BinaryInsertion has been fixed, but now timings are in the ScienceOfHappiness.sorted file
 After HeapSort: 84.269188ms
 After QuickSort: 55.141166ms
 After mergeSort: 128.025583ms
 After NativeSort again: 60.068087ms

7/10/19, timing for PaulKrugman.dat, 104,603 words, on z76.  Windows has trouble w/ timing below a ms.  Now sorted.
QuickSort: 25.14 ms
NativeSort: 29.51 ms
sort.Strings: 30.24 ms
NativeSort again: 32.29 ms
HeapSort: 46.23 ms
MergeSort: 62.61 ms
BinaryInsertion: 3.66 s
StraightInsertion: 28.86 s
StraightSelection: 45.28 s

 Conclusion:
   Fastest for large files is the NativeSort and QuickSort.
   HeapSort is faster than MergeSort, by a factor of about 1.35, or 35%
   StraightInsertion is faster than StraightSelection, by about a factor of ~1.6
*/
