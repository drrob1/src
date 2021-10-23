package main

import (
	"bufio"
	"bytes"
	"fmt"
	"src/getcommandline"
	// "io/ioutil"  Depracated as of Go 1.16
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const LastAlteredDate = "22 Oct 2021"

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
  16 May  20 -- Decided to try to fix the old ShellSort, now called BadShellSort.
  18 May  20 -- Made 12 the break point in Modified merge sort.
  19 May  20 -- Created ModifiedHeapSort which also uses InsertionSort for < 12 items.  And took another crack at fixing AnotherHeapSort.
                Neither of these work.
                However, I also took another crack at NonRecursiveQuickSort.
  21 May  20 -- Removed unneeded commented out code.  I'm not recompiling.
  23 May  20 -- Copied ShellSort by Sedgewick here.  Renamed ShellSort that I based on bubble sort to MyShellSort
  24 May  20 -- All the nonrecursive Quicksort routines I found create their own stack of indices.  I'll try
                  my hand at creating my own stack operations push and pop.  I was not able to write a routine based on code inSedgewick.
  25 May  20 -- Thoughts on mysorts.go and mysorts2.go.
               Over the last 2 weeks I've been able to get the bad code working.  I now have 3 versions of ShellSort, and 2 versions of nonrecursive quick sort.
        I got ShellSort working by noticing a non idiomatic for loop, and Rob Pike in his book says the advantages of using idioms is that they help avoid bugs.
         The idiom is a pattern known to work.  Look closely at non-idiomatic code for the source of bugs.  I did and that's where I found the bug in ShellSort.
               The non-recursive quick sort routines depend on creating a stack themselves instead of relying on recursion to use the system's stack.
               When I switched to idiomatic stack code using push and pop, the code started working.  One reason for this is that I made the stack much
               bigger, so one of the issues may have been that the stack was too small in the code published in these books.  Mostly, I used Wirth's code which
          differs between his Modula-2 and Oberon versions of "Programs and Data Structures."  The idea to use explicit push and pop came from Sedgewick's book.

               I will have the non-recursive quick sort routines print out the max size of their respective stacks, so I can gauge if
               a stack too small was my problem all along.

               And I fixed a bug in that MyShellSort was not being tested after all.
               When I correctly tested Sedgewick's approch to the ShellSort interval, I found it to be substantially better than what
               Wirth did.  So I changed all of them to Sedgewick's approach.  The routines became must faster as a result.
  26 May  20 -- Only updating some comments.
                Now called fastsorts, in which I will only compare the n*log(n) sorts.  I will leave in the binary insertion sort.
                I will use more idiomatic Go for my stack routines in the nonrecursive quick sort routines.
  22 Oct 21 --  Optimizing code that uses bytes.NewBuffer().  And removed the depracated ioutil to use os.ReadFile.
*/

type intStackType []int

func NewIntStack(n int) *intStackType {
	s := make(intStackType, 0, n)
	return &s
}

func (s *intStackType) push(i int) {
	*s = append(*s, i)
}

func (s *intStackType) pop() int {
	i := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return i
}

func (s *intStackType) len() int {
	return len(*s)
}

type hiloIndexType struct {
	lo, hi int
}

type hiloStackType []hiloIndexType

func NewHiloStack(n int) *hiloStackType {
	s := make(hiloStackType, 0, n) // this did not work for me as one line.
	return &s
}

func (s *hiloStackType) push(i hiloIndexType) {
	*s = append(*s, i)
}

func (s *hiloStackType) pop() hiloIndexType {
	i := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return i
}

func (s *hiloStackType) len() int {
	length := len(*s) // this did not work for me as one line.
	return length
}

// var maxStackSize int // so I can determine the max stack size.
//                         Using PaulKrugman.dat full file, Modula-2 version showed 12, and Oberon version showed 24.
//                         Oberon version would have blown its stack as it defined a stack of 0..12.
//                         The Modula-2 version would have made it w/ one element to spare.

// -----------------------------------------------------------

func StraightInsertion(input []string) []string { // keeping this here as it's needed in some rtn's below.
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
/*
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
*/

// -----------------------------------------------------------

func BadShellSort(a []string) []string {
	var h int

	n := len(a)
	for h = 1; h < n; h = h*3 + 1 {
	}

	for ; h > 0; h /= 3 {
		for i := h; i < n; i++ {
			x := a[i]
			j := i - h
			// this works, and now I recognize this is the straight insertion sort pattern.
			for (j+1 >= h) && (x < a[j]) {
				a[j+h] = a[j]
				j = j - h
			} // END for/while (j >= k) & (x < a[j]) DO
			a[j+h] = x
		} // END FOR i := k+1 TO n-1 DO
	} // END FOR k ranges over h
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
// revisiting this as I'm reading "High Performance Go."
// I based this on bubble sort pseudo-code above that I found in "Essential Algorithms", by Rod Stephens.
// I have this as an ebook.
// Now I'm going to add the improvement used by Sedgewick in the determination of h.

func MyShellSort(a []string) []string {
	var h int

	n := len(a)
	for h = 1; h < n; h = h*3 + 1 {
	}
	//	fmt.Println(" in MyShellSort: h=", h)

	// t0 := time.Now()
	for ; h > 0; h /= 3 {
		//		fmt.Println(" in MyShellSort sorting loop: h=",h)
		//		pause()
		for { // loop until sorted
			sorted := true
			for i := h; i < n; i++ {
				if a[i] < a[i-h] {
					a[i], a[i-h] = a[i-h], a[i]
					sorted = false
					//fmt.Println("  ShellSort:  i =", i, ", sorted=", sorted)
				}
			} // END FOR i := h TO last item DO
			if sorted {
				break
			}
			// elapsed := time.Since(t0)
			// if elapsed > 30*time.Second { return a }
		} // end loop until sorted
	} // END FOR range h
	return a
} //END MyShellSort

// -----------------------------------------------------------

// From Algorithms, 2nd Ed, by Robert Sedgewick (C) 1988 p 108.  Code based on Pascal and 1 origin arrays.

func ShellSort(a []string) []string {
	var h int

	n := len(a)
	for h = 1; h > n; h = h*3 + 1 {
	}

	for ; h > 0; h /= 3 {
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
// -----------------------------------------------------------
func sift(a []string, L, R int) []string {
	i := L
	j := 2*i + 1
	x := a[i]
	if (j < R) && (a[j] < a[j+1]) {
		j++
	} // end if
	for (j <= R) && (x < a[j]) {
		a[i] = a[j]
		i = j
		j = 2*j + 1
		if (j < R) && (a[j] < a[j+1]) {
			j++
		} // end if
	} //END for (j <= R) & (x < a[j])
	a[i] = x
	return a
} // END sift;

func HeapSort(a []string) []string { // I think this is based on Wirth's code in either Oberon or Modula-2.
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
// -----------------------------------------------------------
// -----------------------------------------------------------

func ModifiedHeapSort(a []string) []string { // I did this myself, but it doesn't work.
	n := len(a)
	L := n / 2
	R := n - 1
	for L > 0 {
		L--
		if R-L < 12 {
			b := a[L : R+1]
			b = StraightInsertion(b)
			for i, v := range b { // copy the insertion sorted fragment back into a.
				a[L+i] = v
			}
		} else {
			a = sift(a, L, R)
		}
	} // END for-while L>0
	for R > 0 {
		a[0], a[R] = a[R], a[0]
		R--
		if R-L < 12 {
			b := a[L : R+1]
			b = StraightInsertion(b)
			for i, v := range b { // copy the insertion sorted fragment back into a.
				a[L+i] = v
			}
		} else {
			a = sift(a, L, R)
		}
	} // END for-while R > 0
	return a
} // END ModifiedHeapSort
//------------------------------------------------------------------------
//------------------------------------------------------------------------

func NRsiftdown(items []string, L, R int) []string { // Numerical Recipes 3rd ed (C) 2007, p 428
	// execute the sift down on element ra[L] to maintain heap structure.

	i := L
	x := items[L]
	j := 2*L + 1
	for j <= R {
		if j < R && items[j] < items[j+1] { // if the next item is a more suitable comparison, switch to that.
			j++
		}
		if x >= items[j] {
			break
		}
		items[i] = items[j]
		i = j
		j = 2*j + 1
	}
	items[i] = x
	return items
} // END siftdown;

func NRheapsort(items []string) []string { // copied from Numerical Recipes 3rd ed (C) 2007, p 428
	// sort array[0 .. n-1] using HeapSort.
	n := len(items)
	// the index i determines the left range of the siftdown.  Heap creation phase is also call hiring phase.
	for i := n/2 - 1; i >= 0; i-- {
		items = NRsiftdown(items, i, n-1)
	}

	// Right range of the siftdown is decremented from n-2 to 0 during the retirement and promotion phase,
	// also called heap selection
	for i := n - 1; i > 0; i-- {
		// clear a space at the end of the array and retire the top of the heap into it, by swapping
		items[0], items[i] = items[i], items[0]
		NRsiftdown(items, 0, i-1)
	}
	return items
} // END NRheapsort;
//------------------------------------------------------------------------
//------------------------------------------------------------------------
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
// -----------------------------------------------------------
// -----------------------------------------------------------
/*  I'm coding idiomatic Go code for stacks.  Or just playing with functions, whatever.
func intStackInit(n int) {
	intStack = make([]int, 0, n)
}

func intStackPush(i int) {
	intStack = append(intStack, i)
}

func intStackPop() int {
	i := intStack[len(intStack)-1]
	intStack = intStack[:len(intStack)-1]
	return i
}

func intStackLen() int {
	return len(intStack)
}

func hiloInit(n int) {
	hiloStack = make([]hiloIndexType, 0, n)
}

func hiloStackPush(i hiloIndexType) {
	hiloStack = append(hiloStack, i)
}

func hiloStackPop() hiloIndexType {
	i := hiloStack[len(hiloStack)-1]
	hiloStack = hiloStack[:len(hiloStack)-1]
	return i
}

func hiloStackLen() int {
	return len(hiloStack)
}
*/

// -----------------------------------------------------------
// From Wirth p. 94ff in my copy of "Algorithms and Data Structures," (C) 1986.  In Modula-2.
// The code declares S : [0 .. M];  I don't know what happens if try to increment beyond M.
// That request may be ignored, hence the code Wirth wrote.
// Other books also have a nonrecursive quicksort, and makes the algorithm nonrecursive by creating a stack of
// subarrays that are to be sorted.  IE, these create their own recursion.  Sedgewick says that this is what a
// compiler does, since computer architecture is not inherently recursive.

func NonRecursiveQuickSort(a []string) []string {
	var k hiloIndexType
	t0 := time.Now()
	n := len(a) - 1

	s := NewHiloStack(n / 2)

	k.lo = 0
	k.hi = n
	s.push(k)
	//fmt.Println(" initial hi lo stack push.  Stack is", s.hiloStack)

	for s.len() > 0 {
		//		stacksize := hiloStackLen()
		//		if stacksize > maxStackSize {
		//			maxStackSize = stacksize
		//		}

		i0 := s.pop()
		lo := i0.lo
		hi := i0.hi
		//fmt.Println(" outer for loop, for stack not empty.  Stack is", hiloStack)

		if time.Since(t0) > 10*time.Second {
			fmt.Printf(" timeout outer loop.  i0= %v, lo=%d, hi=%d \n", i0, lo, hi)
			return a
		}

		for lo < hi { // REPEAT partition a[L] ...  a[R]
			i := lo
			j := hi
			x := a[(hi+lo)/2] // x is mid point element
			//fmt.Println(" inner for lo < hi.  Stack is", hiloStack)

			if time.Since(t0) > 100*time.Second { // 100 seconds effectively removes it
				fmt.Printf(" timeout inner loop.  lo=%d, hi=%d, i=%d, j=%d, x=%s \n", lo, hi, i, j, x)
			}

			for i <= j { // REPEAT UNTIL i > j
				if time.Since(t0) > 100*time.Second { // 100 seconds effective removes it
					fmt.Printf(" timeout innermost loop.  lo=%d, hi=%d, i=%d, , j=%d,  x=%s \n", lo, hi, i, j, x)
					return a
				}
				//fmt.Println(" innermost for loop for i <= j.  Stack is", hiloStack)
				//pause()

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
				//				if i > j { // UNTIL i > j
				//					break
				//				}
			} // REPEAT ... UNTIL i > j
			if j-lo < hi-i {
				if i < hi { // push request to sort right partition
					k.lo = i
					k.hi = hi
					s.push(k)
				}
				hi = j // now L and R delimit the left partition, and continue sorting the left partition.
			} else {
				if lo < j { // push request for sorting left partition onto the stack
					k.lo = lo
					k.hi = j
					s.push(k)
				}
				lo = i // continue sorting right partition
			}

			//			if L >= R { // UNTIL L >= R
			//				break
			//			}
		} // REPEAT ... UNTIL L >= R

	} // REPEAT ... UNTIL hiloStack is empty
	// fmt.Println(" Modula-2 NonRecursiveQuickSort maxStackSize =", maxStackSize)  This showed 12 on the full PaulKrugman.dat file.
	return a
} // END NonRecursiveQuickSort

func NonRecursiveQuickSortOberon(a []string) []string {
	n := len(a)
	s := NewIntStack(n / 2)
	s.push(0)
	s.push(n - 1)
	for s.len() > 0 { // REPEAT (*take top request from stack*)
		//		stacksize := intStackLen()
		//		if stacksize > maxStackSize {
		//			maxStackSize = stacksize
		//		}

		R := s.pop()
		L := s.pop()
		for L < R { // REPEAT partition a[L] ... a[R]
			i := L
			j := R
			x := a[(L+R)/2]
			for i <= j { //REPEAT
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
				//				if i > j {
				//					break
				//				}
			} // for i <= j, or UNTIL i > j;
			if j-L < R-i {
				if i < R { // THEN push request to sort right partition onto the stack
					s.push(i)
					s.push(R)
				}
				R = j // (*now L and R delimit the left partition*)
			} else {
				if L < j { // push request for sorting left partition onto the atack
					s.push(L)
					s.push(j)
				}
				L = i // continue sorting right partition
			}
			//			if L >= R {
			//				break
			//			}
		} // for L < R

		//		if s == 0 {
		//			break
		//		}
	} // for stack not empty
	// fmt.Println(" NonRecursiveQuickSortOberan maxStackSize=", maxStackSize) This showed 24 on the full PaulKrugman.dat file.
	return a
} // 	END NonRecursiveQuickSortOberon

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
	if len(L) < 12 {
		L = StraightInsertion(L)
		return L
	} else {
		middle := len(L) / 2 // middle needs to be of type int
		left := ModifiedMergeSort(L[:middle])
		right := ModifiedMergeSort(L[middle:])
		return merge(left, right)
	}
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

	//byteslice := make([]byte, 0, filesize+5) // add 5 just in case
	//byteslice, err := ioutil.ReadFile(Filename)
	byteslice, err := os.ReadFile(Filename)
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
	s = fmt.Sprintf(" after first sort.StringSlice: %s, %d ns \n", NativeSortTime.String(), NativeSortTimeNano)
	fmt.Println(s)
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	if allowoutput {
		for _, w := range NativeWords {
			fmt.Print(w, " ")
		}
	}
	fmt.Println()
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	/* not in fastsorts.go
	   {{{
	   	// StraightSelection
	   	copy(sliceofwords, mastersliceofwords)
	   	if allowoutput {
	   		fmt.Println(" sliceofwords before StraightSelection: ", sliceofwords)
	   	}
	   	t0 := time.Now()
	   	sortedsliceofwords := StraightSelection(sliceofwords)
	   	StraightSelectionTime := time.Since(t0)
	   	StraightSelectionTimeNano := StraightSelectionTime.Nanoseconds()
	   	s = fmt.Sprintf(" After StraightSelection: %s, %d ns \n", StraightSelectionTime.String(), StraightSelectionTimeNano)
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
	   	check(err)
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
	   	_, err = OutBufioWriter.WriteRune('\n')
	   	check(err)
	   	fmt.Println()
	   }}}
	*/

	/* removed due to slowness.
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
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range ShellSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	fmt.Println()
	*/

	// BadShellSort -- now a misnomer as it finally works.
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before BadShellSort:", sliceofwords)
	}
	t3a := time.Now()
	BadShellSortedWords := BadShellSort(sliceofwords)
	BadShellSortedTime := time.Since(t3a)
	s = fmt.Sprintf(" After BadShellSort: %s, %d ns \n", BadShellSortedTime.String(), BadShellSortedTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range BadShellSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	fmt.Println()

	// MyShellSort
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before MyShellSort:", sliceofwords)
	}
	t3b := time.Now()
	MyShellSortedWords := MyShellSort(sliceofwords)
	MyShellSortedTime := time.Since(t3b)
	s = fmt.Sprintf(" After MyShellSort: %s, %d ns \n", MyShellSortedTime.String(), MyShellSortedTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println(s)
	if allowoutput {
		for _, w := range MyShellSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	fmt.Println()

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

	// NRHeapSort which is from Numerical Recipies and converted from C++ coce.
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before NRHeapSort:", sliceofwords)
	}
	t5 := time.Now()
	NRHeapSortedWords := NRheapsort(sliceofwords)
	NRHeapTime := time.Since(t5)
	s = fmt.Sprintf(" After NRheapsort: %s, %d ns \n", NRHeapTime.String(), NRHeapTime.Nanoseconds())
	fmt.Println(s)
	if allowoutput {
		for _, w := range NRHeapSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteString(s)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// ModifiedHeapSort -- doesn't work.
	/*
		copy(sliceofwords, mastersliceofwords)
		if allowoutput {
			fmt.Println("before ModifiedHeapSort:", sliceofwords)
		}
		t5a := time.Now()
		ModifiedHeapSortedWords := ModifiedHeapSort(sliceofwords)
		ModifiedHeapTime := time.Since(t5a)
		s = fmt.Sprintf(" After Modifiedheapsort: %s, %d ns \n", ModifiedHeapTime.String(), ModifiedHeapTime.Nanoseconds())
		if allowoutput {
			for _, w := range ModifiedHeapSortedWords {
				fmt.Print(w, " ")
			}
			fmt.Println()
		}
		fmt.Println()
		_, err = OutBufioWriter.WriteString(s)
		_, err = OutBufioWriter.WriteRune('\n')
		check(err)
		fmt.Println(s)
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

	// NonRecursiveQuickSort (from Modula-2)
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println(" before Modula-2 nonrecursiveQuickSort:", sliceofwords)
	}
	t8 := time.Now()
	NonRecursiveQuickSortedWords := NonRecursiveQuickSort(sliceofwords)
	NonRecursiveQuickedTime := time.Since(t8)
	s = fmt.Sprintf(" After Modula-2 NonRecursiveQuickSort: %s, %d ns \n", NonRecursiveQuickedTime.String(), NonRecursiveQuickedTime.Nanoseconds())
	fmt.Println(s)
	if allowoutput {
		for _, w := range NonRecursiveQuickSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// NonRecursiveQuickSortOberon
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println(" before nonrecursiveQuickSortOberon:", sliceofwords)
	}
	t8a := time.Now()
	NonRecursiveQuickSortedOberonWords := NonRecursiveQuickSortOberon(sliceofwords)
	NonRecursiveQuickOberonTime := time.Since(t8a)
	s = fmt.Sprintf(" After NonRecursiveQuickSortOberon: %s, %d ns \n", NonRecursiveQuickOberonTime.String(), NonRecursiveQuickOberonTime.Nanoseconds())
	fmt.Println(s)
	if allowoutput {
		for _, w := range NonRecursiveQuickSortedOberonWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)
	fmt.Println()

	// sort.StringSlice
	copy(sliceofwords, mastersliceofwords)
	if allowoutput {
		fmt.Println("before 2nd sort.StringSlice:", sliceofwords)
	}
	NativeWords = sort.StringSlice(sliceofwords)
	t9 = time.Now()
	NativeWords.Sort()
	NativeSortTime = time.Since(t9)
	s = fmt.Sprintf(" After 2nd sort.StringSlice: %s, %d ns \n", NativeSortTime.String(), NativeSortTime.Nanoseconds())
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
} // end main

//===========================================================
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func pause() {
	fmt.Print(" hit <enter> to continue")
	fmt.Scanln()
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
