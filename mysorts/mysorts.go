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
	//                                                             for i := 0; i < n; i++ {
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
		//                                                             for i := 0; i < n; i++ {
		for j := i; j >= R+1; j-- {
			a[j] = a[j-1]
		} //END for j := i TO R+1 BY -1 DO
		a[R] = x
	} // END for i := 1 to n-1
	return a
} // END BinaryInsertion

func StraightSelection(a []string) []string {
	n := len(a)
	for i := 0; i < n-1; i++ {
		//                                                  for i := 0; i <= n-2; i++ {
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
			for (j >= k) && (x < a[j]) {
				a[j+k] = a[j]
				j = j - k
			} // END for/while (j >= k) & (x < a[j]) DO
			a[j+k] = x
		} // END FOR i := k+1 TO n-1 DO
	} // END FOR m := 0 TO T-1 DO
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

func QuickSort(a []string) []string {
	n := len(a) - 1
	a = qsort(a, 0, n)
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
	answer := scanner.Text()
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}
	requestedwordcount, err := strconv.Atoi(answer)
	if err != nil {
		fmt.Println(" No valid answer entered")
		os.Exit(1)
	}

	if requestedwordcount == 0 {
		requestedwordcount = int(filesize / 5)
	}

	mastersliceofwords := make([]string, 0, requestedwordcount)

	for totalwords := 0; totalwords < requestedwordcount; totalwords++ { // Main processing loop
		word, err := bytesbuffer.ReadString('\n')
		if err != nil {
			break
		}
		//                                                     word = strings.TrimSpace(word)
		word = strings.ToLower(strings.TrimSpace(word))
		if len(word) < 4 {
			continue
		}
		mastersliceofwords = append(mastersliceofwords, word)
	}

	s := ""
	fmt.Println("master before:", mastersliceofwords)
	sliceofwords := make([]string, requestedwordcount)
	copy(sliceofwords, mastersliceofwords)
	fmt.Println("slice before:", sliceofwords)
	NativeWords := sort.StringSlice(sliceofwords[:])
	t9 := time.Now()
	NativeWords.Sort()
	NativeSortTime := time.Since(t9)
	s = fmt.Sprintf(" after NativeSort: %s \n", NativeSortTime.String())
	fmt.Println(s)
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	//	s = fmt.Sprintf("%v\n", NativeWords)
	//	_, err = OutBufioWriter.WriteString(s)
	//	check(err)
	for _, w := range NativeWords {
		fmt.Print(w, " ")
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()
	fmt.Println()

	copy(sliceofwords, mastersliceofwords)
	fmt.Print(" sliceofwords before: ")
	for _, w := range sliceofwords {
		fmt.Print(w, " ")
	}
	fmt.Println()

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
	for _, w := range sortedsliceofwords {
		fmt.Print(w, " ")
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()
	fmt.Println()

	copy(sliceofwords, mastersliceofwords)
	fmt.Println("before:", sliceofwords)
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
	for _, w := range sliceofsortedwords {
		fmt.Print(w, " ")
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	fmt.Println()
	fmt.Println()
	/* Does not sort correctly, but doesn't panic anymore
	copy(sliceofwords, mastersliceofwords)
	fmt.Println("before:", sliceofwords)
	t2 := time.Now()
	BinaryInsertionSortedWords := BinaryInsertion(sliceofwords)
	BinaryInsertionTime := time.Since(t2)
	fmt.Println(" after BinaryInsertion:", BinaryInsertionTime)
	for _, w := range BinaryInsertionSortedWords {
		fmt.Print(w, " ")
	}
	fmt.Println()
	fmt.Println()
	*/
	/* Does not sort correctly, but doesn't panic
	copy(sliceofwords, mastersliceofwords)
	fmt.Println("before:", sliceofwords)
	t3 := time.Now()
	ShellSortedWords := ShellSort(sliceofwords)
	ShellSortedTime := time.Since(t3)
	fmt.Println(" ShellSort:", ShellSortedTime)
	for _, w := range ShellSortedWords {
		fmt.Print(w, " ")
	}
	fmt.Println()
	fmt.Println()
	*/
	copy(sliceofwords, mastersliceofwords)
	fmt.Println("before:", sliceofwords)
	t4 := time.Now()
	HeapSortedWords := HeapSort(sliceofwords)
	HeapSortedTime := time.Since(t4)
	s = fmt.Sprintf(" After HeapSort: %s \n", HeapSortedTime.String())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(" HeapSort:", HeapSortedTime)
	for _, w := range HeapSortedWords {
		fmt.Print(w, " ")
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()
	fmt.Println()
	/*  Does not sort correctly, but does not panic.
	copy(sliceofwords, mastersliceofwords)
	fmt.Println("before:", sliceofwords)
	t5 := time.Now()
	AnotherHeapSortedWords := anotherheapsort(sliceofwords)
	AnotherHeapTime := time.Since(t5)
	fmt.Println(" anotherheapsort:", AnotherHeapTime)
	for _, w := range AnotherHeapSortedWords {
		fmt.Print(w, " ")
	}
	fmt.Println()
	fmt.Println()
	*/
	copy(sliceofwords, mastersliceofwords)
	fmt.Println("before:", sliceofwords)
	t6 := time.Now()
	QuickSortedWords := QuickSort(sliceofwords)
	QuickSortedTime := time.Since(t6)
	s = fmt.Sprintf(" After QuickSort: %s \n", QuickSortedTime.String())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(" QuickSort:", QuickSortedTime)
	for _, w := range QuickSortedWords {
		fmt.Print(w, " ")
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()
	fmt.Println()

	copy(sliceofwords, mastersliceofwords)
	fmt.Println("before:", sliceofwords)
	t7 := time.Now()
	MergeSortedWords := mergeSort(sliceofwords)
	MergeSortTime := time.Since(t7)
	s = fmt.Sprintf(" After mergeSort: %s \n", MergeSortTime.String())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(" mergeSort:", MergeSortTime)
	for _, w := range MergeSortedWords {
		fmt.Print(w, " ")
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()
	fmt.Println()
	/*  I think this paniced
	copy(sliceofwords, mastersliceofwords)
	fmt.Println("before:", sliceofwords)
	t8 := time.Now()
	NonRecursiveQuickSortedWords := NonRecursiveQuickSort(sliceofwords)
	NonRecursiveQuickedTime := time.Since(t8)
	fmt.Println(" NonRecursiveQuickSort:", NonRecursiveQuickedTime)
	for _, w := range NonRecursiveQuickSortedWords {
		fmt.Print(w, " ")
	}
	fmt.Println()
	fmt.Println()
	*/
	copy(sliceofwords, mastersliceofwords)
	fmt.Println("before:", sliceofwords)
	NativeWords = sort.StringSlice(sliceofwords)
	t9 = time.Now()
	NativeWords.Sort()
	NativeSortTime = time.Since(t9)
	s = fmt.Sprintf(" After NativeSort again: %s \n", NativeSortTime.String())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Println(" NativeSort:", NativeSortTime)
	for _, w := range NativeWords {
		fmt.Print(w, " ")
	}
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()
	fmt.Println()
}

//===========================================================
func check(e error) {
	if e != nil {
		panic(e)
	}
}

/*
  Timing for full data file, ScienceOfHappiness.dat

 after NativeSort: 47.745145ms

 After StraightSelection: 47.340594191s

 After StraightInsertion: 14.074816209s

 After HeapSort: 84.269188ms

 After QuickSort: 55.141166ms

 After mergeSort: 128.025583ms

 After NativeSort again: 60.068087ms


 Conclusion:
   Fastest for large files is the NativeSort, then QuickSort.
   HeapSort is faster than mergeSort, by a factor of about 1.5.
   StraightInsertion is faster than StraightSelection, by about a factor of 3.
*/
