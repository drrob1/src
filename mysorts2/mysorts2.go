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

const LastAlteredDate = "9 July 2019"

/*
  REVISION HISTORY
  ----------------
  July 2017 -- First version
  26 Jul 17 -- Will try to learn delve (dlv) by using it to debug the routines here that don't work.
   7 Aug 17 -- Thinking about a mergeSort with an insertionshort below, maybe 5 elements.
   8 Nov 17 -- Added comparing to sort.Slice.  I need to remember how I did this, so it will take a day or so.
  30 Dec 17 -- Added comparing to sort.Strings.  Nevermind.  It's already done.
   9 Jul 19 -- Looking at fixing the routines that don't work.  And now it's called mysorts2.go
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

// From Algorithms, 2nd Ed, by Robert Sedgewick (C) 1989.  Code based on Pascal and 1 origin arrays.  So I subt 1 for each subscript reference.
func ShellSort(a []string) []string {
	var i, j, h int

	n := len(a)
	if n > 9 {
		h = 9
	} else if n > 7 {
		h = 7
	} else if n > 5 {
		h = 5
	} else if n > 3{
		h = 3
	} else {
		h = 1
	}

loop:
	for ; h > 0; h -= 2 {
		for i = h; i <= n; i++ { // FOR i = h+1 TO n DO
			v := a[i-1]
			j = i
			for a[j-h] > v {
				a[j-1] = a[j-h]
				j -= h
				if j < h { // IF j <= h THEN continue
					continue loop
				}
			}
			a[j-1] = v
		}
	} // end for h
	return a
} //END ShellSort

// -----------------------------------------------------------
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

func anotherheapsort(items []string) []string { // doesn't work, but doesn't panic, either.
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

func QuickSort(a []string) []string {
	n := len(a) - 1
	a = qsort(a, 0, n)
	return a
} // END QuickSort

func NonRecursiveQuickSort(a []string) []string { // this paniced
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

//-----------------------------------------------------------------------+
//                               MAIN PROGRAM                            |
//-----------------------------------------------------------------------+

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

	if allowoutput {
		fmt.Println("master before:", mastersliceofwords)
	}
	sliceofwords := make([]string, numberofwords)
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("slice before sort.StringSlice:", sliceofwords)
	}
	NativeWords := sort.StringSlice(sliceofwords)
	t9 := time.Now()
	NativeWords.Sort()
	NativeSortTime := time.Since(t9)
	s = fmt.Sprintf(" after sort.StringSlice: %s \n", NativeSortTime.String())
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
	fmt.Println()

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
	s = fmt.Sprintf(" After StraightSelection: %s \n", StraightSelectionTime.String())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(" after StraightSelection:", StraightSelectionTime)
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

	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before StraightInsertion:", sliceofwords)
	}
	t1 := time.Now()
	sliceofsortedwords := StraightInsertion(sliceofwords)
	StraightInsertionTime := time.Since(t1)
	s = fmt.Sprintf(" After StraightInsertion: %s \n", StraightInsertionTime.String())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(" after StraightInsertion:", StraightInsertionTime)
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

	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before BinaryInsertion:", sliceofwords)
	}
	t2 := time.Now()
	BinaryInsertionSortedWords := BinaryInsertion(sliceofwords)
	BinaryInsertionTime := time.Since(t2)
	s = fmt.Sprintf(" After BinaryInsertion: %s \n", BinaryInsertionTime.String())
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

	/* Does not sort correctly, but doesn't panic.  Now debugging it */
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
	fmt.Println(" ShellSort:", ShellSortedTime)
	if allowoutput {
		for _, w := range ShellSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	fmt.Println()
	/**/

	/*	Does not sort correctly, but does not panic.
		copy(sliceofwords, mastersliceofwords)
		if allowoutput {
			fmt.Println("before:", sliceofwords)
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
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before QuickSort:", sliceofwords)
	}
	t6 := time.Now()
	QuickSortedWords := QuickSort(sliceofwords)
	QuickSortedTime := time.Since(t6)
	s = fmt.Sprintf(" After QuickSort: %s \n", QuickSortedTime.String())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(" QuickSort:", QuickSortedTime)
	if allowoutput {
		for _, w := range QuickSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	/*	I think this paniced
		copy(sliceofwords, mastersliceofwords)
		if allowoutput {
			fmt.Println("before:", sliceofwords)
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
	fmt.Println(" NativeSort:", NativeSortTime)
	if allowoutput {
		for _, w := range NativeWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	s = fmt.Sprintf(" requestedwordcount= %d, numberofwords= %d, len(mastersliceofwords)= %d \n",
		requestedwordcount, numberofwords, len(mastersliceofwords))
	_, err = OutBufioWriter.WriteString(s)
	if len(mastersliceofwords) > 1000 {
		fmt.Println(s)
		//		fmt.Println(" Number of words to be sorted is", len(mastersliceofwords))
	}

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
	fmt.Println(" StringsSortTime:", StringsSortTime)
	if allowoutput {
		for _, w := range NativeWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	s = fmt.Sprintf(" requestedwordcount= %d, numberofwords= %d, len(mastersliceofwords)= %d \n",
		requestedwordcount, numberofwords, len(mastersliceofwords))
	_, err = OutBufioWriter.WriteString(s)
	if len(mastersliceofwords) > 1000 {
		fmt.Println(s)
		//		fmt.Println(" Number of words to be sorted is", len(mastersliceofwords))
	}
	_, err = OutBufioWriter.WriteString("------------------------------------------------------\n")
	check(err)

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