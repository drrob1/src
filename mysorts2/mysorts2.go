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

const LastAlteredDate = "22 May 2020"

/*
  REVISION HISTORY
  ----------------
  July 2017 -- First version
  26 Jul 17 -- Will try to learn delve (dlv) by using it to debug the routines here that don't work.
   7 Aug 17 -- Thinking about a mergeSort with an insertionshort below, maybe 5 elements.
   8 Nov 17 -- Added comparing to sort.Slice.  I need to remember how I did this, so it will take a day or so.
  30 Dec 17 -- Added comparing to sort.Strings.  Nevermind.  It's already done.
   9 Jul 19 -- Looking at fixing the routines that don't work.  And now it's called mysorts2.go
  20 May 20 -- Looking at fixing the version of ShellSort that's here.  Then I decided NEVERMIND.
  21 May 20 -- Decided to try again, but remove those that don't work.
                 Now that Sedgewick's ShellSort works, I'm adding BadShellSort to compare more directly.
  22 May 20 -- Adding ModifiedQuickSort to see if it's faster to insertionsort when < 12 items.
*/

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

// -----------------------------------------------------------

func BinaryInsertion(a []string) []string {
	n := len(a)
	//                                                             for i := 0; i < n; i++ {
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
			} // END if a[m] <= x
		} //END while L < R
		//                                                             for i := 0; i < n; i++ {
		for j := i; j >= R+1; j-- {
			a[j] = a[j-1]
		} //END for j := i TO R+1 BY -1 DO
		a[R] = x
	} // END for i := left+1 to right
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
			} // end if
		} // END for j := i+1 TO n-1
		a[k] = a[i]
		a[i] = x
	} // END for i := 0 to n-2
	return a
} // END StraightSelection

// -----------------------------------------------------------

// From Algorithms, 2nd Ed, by Robert Sedgewick (C) 1988 p 108.  Code based on Pascal and 1 origin arrays.
func ShellSort(a []string) []string {
	var h int

	n := len(a)
	if n > 9 {
		h = 9
	} else if n > 7 {
		h = 7
	} else if n > 5 {
		h = 5
	} else if n > 3 {
		h = 3
	} else {
		h = 1
	}

	for ; h > 0; h -= 2 {
		// original code has this line as for i := h+1 to N do.  Here is the conversion to zero origin array
		for i := h; i < n; i++ {
			j := i
			v := a[j]
			for a[j-h] > v {
				a[j] = a[j-h]
				j -= h
				if j < h {
					break
				}
			}
			a[j] = v
		}
	} // end for h
	return a
} //END ShellSort

// -----------------------------------------------------------

func BadShellSort(a []string) []string { // From Wirth's Algorithms and Data Structures, don't remember which edition.
	const T = 4
	var h [T]int

	h[0] = 9
	h[1] = 5
	h[2] = 3
	h[3] = 1
	n := len(a)
	for m := 0; m < T; m++ {
		k := h[m]
		for i := k; i < n; i++ {
			x := a[i]
			j := i - k
			// this works, and now I recognize this is the straight insertion sort pattern.
			for (j+1 >= k) && (x < a[j]) {
				a[j+k] = a[j]
				j = j - k
			} // END for/while (j >= k) & (x < a[j]) DO
			a[j+k] = x
		} // END FOR i := k+1 TO n-1 DO
	} // END FOR m := 0 TO T-1 DO
	return a
} //END BadShellSort

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

func QuickSort(a []string) []string {
	n := len(a) - 1
	a = qsort(a, 0, n)
	return a
} // END QuickSort

//-------------------------------------------------------------

func modified12Qsort(a []string, L, R int) []string {
	if R-L < 12 {
		b := StraightInsertion(a[L : R+1])
		return b
	}

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

func ModifiedQuickSort(a []string) []string {
	n := len(a) - 1
	a = modified12Qsort(a, 0, n)
	return a
} // END QuickSort

//-----------------------------------------------------------------------+
//                               MAIN PROGRAM                            |
//-----------------------------------------------------------------------+

func main() {
	var filesize int64
	fmt.Println(" Mysorts2: Sort a slice of strings, using the different algorithms.  Last altered", LastAlteredDate)
	fmt.Println()

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

	// Main processing loop where the mastersliceofwords is constructed.
	for totalwords := 0; totalwords < requestedwordcount; totalwords++ {
		word, err := bytesbuffer.ReadString('\n')
		if err != nil {
			break
		}
		word = strings.TrimSpace(word)
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

	if allowoutput {
		fmt.Println("master before:", mastersliceofwords)
	}
	sliceofwords := make([]string, numberofwords)

	// sort.StringSlice
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("slice before sort.StringSlice:", sliceofwords)
	}
	NativeWords := sort.StringSlice(sliceofwords)
	t9 := time.Now()
	NativeWords.Sort()
	NativeSortTime := time.Since(t9)
	s = fmt.Sprintf(" after sort.StringSlice: %s \n", NativeSortTime.String())
	fmt.Println(s) // notice that s has a newline, and Println also prints a newline.  That's why I see 2 newlines.
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	if allowoutput {
		for _, w := range NativeWords {
			fmt.Print(w, " ")
		}
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()
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
	s = fmt.Sprintf(" After StraightSelection: %s, %d ns \n", StraightSelectionTime.String(), StraightSelectionTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
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
	fmt.Println(s)
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
	fmt.Println()

	// ShellSort
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before ShellSort:", sliceofwords)
	}
	t3 := time.Now()
	ShellSortedWords := ShellSort(sliceofwords)
	ShellSortedTime := time.Since(t3)
	s = fmt.Sprintf(" After ShellSort: %s, %d ns \n", ShellSortedTime.String(), ShellSortedTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range ShellSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	fmt.Println()

	// BadShellSort
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before BadShellSort:", sliceofwords)
	}
	t3a := time.Now()
	BadShellSortedWords := ShellSort(sliceofwords)
	BadShellSortedTime := time.Since(t3a)
	s = fmt.Sprintf(" After BadShellSort: %s, %d ns \n", BadShellSortedTime.String(), BadShellSortedTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range BadShellSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	fmt.Println()

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

	// ModifiedQuickSort
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before ModifiedQuickSort:", sliceofwords)
	}
	t6a := time.Now()
	ModifiedQuickSortedWords := ModifiedQuickSort(sliceofwords)
	ModifiedQuickSortedTime := time.Since(t6a)
	s = fmt.Sprintf(" After ModifiedQuickSort: %s, %d ns \n", ModifiedQuickSortedTime.String(), ModifiedQuickSortedTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	fmt.Println(s)
	if allowoutput {
		for _, w := range ModifiedQuickSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// sort.StringSlice again
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before sort.StringSlice again:", sliceofwords)
	}
	NativeWords = sort.StringSlice(sliceofwords)
	t9 = time.Now()
	NativeWords.Sort()
	NativeSortTime = time.Since(t9)
	s = fmt.Sprintf(" After sort.StringSlice again: %s \n", NativeSortTime.String())
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
	s = fmt.Sprintf(" After sort.Strings: %s \n", StringsSortTime.String())
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

	// Wrap up
	s = fmt.Sprintf(" requestedwordcount= %d, numberofwords= %d, len(mastersliceofwords)= %d \n",
		requestedwordcount, numberofwords, len(mastersliceofwords))
	_, err = OutBufioWriter.WriteString(s)
	if len(mastersliceofwords) > 1000 {
		fmt.Println(s)
	}

	/*  I think this is duplicated code.
	{{{
		s = fmt.Sprintf(" requestedwordcount= %d, numberofwords= %d, len(mastersliceofwords)= %d \n",
			requestedwordcount, numberofwords, len(mastersliceofwords))
		_, err = OutBufioWriter.WriteString(s)
		if len(mastersliceofwords) > 1000 {
			fmt.Println(s)
			//		fmt.Println(" Number of words to be sorted is", len(mastersliceofwords))
		}
		_, err = OutBufioWriter.WriteString("------------------------------------------------------\n")
		check(err)
	}}}
	*/

} // end main

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


 Conclusion:
   Fastest for large files is the NativeSort, then QuickSort.
   HeapSort is faster than mergeSort, by a factor of about 1.5.
   StraightInsertion is faster than StraightSelection, by about a factor of 3.
*/
