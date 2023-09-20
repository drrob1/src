package main

import (
	"bufio"
	"bytes"
	"container/heap"
	"fmt"
	"runtime"

	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

/*
  REVISION HISTORY
  ----------------
  July 2017 -- First version
  26 July 17 -- Will try to learn delve (dlv) by using it to debug the routines here that don't work.
   7 Aug  17 -- Thinking about a mergeSort with an insertionsort below, maybe 5 elements.
   8 Nov  17 -- Added comparing to sort.Slice.  I need to remember how I did this, so it will take a day or so.
  10 July 19 -- Added better comments and output strings.
  28 July 19 -- Adding Stable, and SliceStable
  29 July 19 -- Changing some formatting of the output.
  15 May  20 -- Fixed ShellSort after starting to read High Performance Go by Ron Stephen.  I then remembered that ShellSort
                  is a modified BubbleSort, so I coded it as such.  And now it works.
  16 May  20 -- Decided to try to fix the old ShellSort, now called BadShellSort.
  18 May  20 -- Made 12 the break point in Modified merge sort.
  19 May  20 -- Created ModifiedHeapSort which also uses InsertionSort for < 12 items.  And took another crack at fixing AnotherHeapSort.
                Neither of these work.
                However, I also took another crack at NonRecursiveQuickSort.
  21 May  20 -- Removed unneeded commented out code.  I'm not recompiling.
  23 May  20 -- Copied ShellSort by Sedgewick here.  Renamed ShellSort that I based on bubble sort to MyShellSort
  24 May  20 -- All the nonrecursive Quicksort routines I found create their own stack of indices.  I'll try my hand at creating my own
                  stack operations push and pop.  I was not able to write a routine based on code in Sedgewick.
  25 May  20 -- Thoughts on mysorts.go and mysorts2.go.
        Over the last 2 weeks I've been able to get the bad code working.  I now have 3 versions of ShellSort, and 2 versions of nonrecursive quick sort.
        I got ShellSort working by noticing a non-idiomatic for loop, and Rob Pike in his book says the advantages of using idioms is that they help avoid bugs.
        The idiom is a pattern known to work.  Look closely at non-idiomatic code for the source of bugs.  I did and that's where I found the bug in ShellSort.
        The non-recursive quick sort routines depend on creating a stack themselves instead of relying on recursion to use the system's stack.
        When I switched to idiomatic stack code using push and pop, the code started working.  One reason for this is that I made the stack much
        bigger, so one of the issues may have been that the stack was too small in the code published in these books.  Mostly, I used Wirth's code which
        differs between his Modula-2 and Oberon versions of "Programs and Data Structures."  The idea to use explicit push and pop came from Sedgewick's book.

               I will have the non-recursive quick sort routines print out the max size of their respective stacks, so I can gauge if
               a stack too small was my problem all along.

               And I fixed a bug in that MyShellSort was not being tested after all.
               When I correctly tested Sedgewick's approch to the ShellSort interval, I found it to be substantially better than what
               Wirth did.  So I changed all of them to Sedgewick's approach.  The routines became much faster as a result.
  27 May 20 -- Fixing some comments.  I won't recompile.
  21 Jul 20 -- Now called heapsorter.go, and will compare against the container heap, which is essentially a heapsort.
  23 Jul 20 -- Got it to work yesterday when I noticed something in the documentation that I initially missed.  Now I'm playing with
                 the code to Pop to see if my understanding is correct and the modification also works.
  28 Jul 20 -- Changing how the timing is measured for all.  I'm including the copy operation, so that the container/heap measurement is more fair.
  29 Jul 20 -- Removing many of the newlines that are displayed and written to the file.  There are too many.
   3 Aug 20 -- Fixing some comments, and in one place in StraightSelection I made code more idiomatic for Go.
  16 Aug 20 -- Making StraightInsertion more idiomatic Go, based on code shown in High Performance Go.
                 I added BasicInsertion, also from High Performance Go.
                 I changed the variable names in HeapSort and sift, to help me understand the code.  L is lo, and R is hi.
  27 Aug 20 -- Added timing comments from txt file that has just over 1 million words.
   2 Jan 21 -- Removed 3 redundant type conversions, as flagged by GoLand.
   4 Jan 21 -- Adding ModifiedQuickSort, and will sort output times.  And will not run N**2 sorts on datasets > tooBig.
  27 Aug 21 -- Converting to modules by removing getcommandline, and removed the depricated ioutil.
  23 Oct 21 -- Removed call to make slice to receive words file.
  12 Mar 22 -- I'm back in the code to refactor based on what I've learned from Bill Kennedy's course.  I'm now using bytes.Reader and strings.Builder.
  14 Jul 22 -- Added comment in both non-recursive quicksorts about stack being blown in book code, hence the broken routines.
                 In Modula-2 the stack was [1..12] which is too small, as the index went to 12 in this go code, which would blow the stack.
                 In the oberon version, the stack is [12], which is 0..11.  My measurements went to an index of 24, which blew well past the stack limit.
                 I don't know why the code didn't panic w/ array index out of bounds; it merely didn't work.  It would be hard to pull out the non-working code now.
  7 Oct 22 -- Updated output message
 21 Nov 22 -- static linter found a minor issue, now fixed.
 31 Mar 23 -- StaticCheck found an issue where I forgot to do timeSort = append(timeSort, ts)
 19 Sep 23 -- added const stackSize
 20 Sep 23 -- Tweaked output from container/heap to match the others.
*/

const LastAlteredDate = "Sep 20, 2023"
const tooBig = 170_000
const stackSize = 50

var intStack []int // for non-recursive quick sorts

type hiloIndexType struct { // for non-recursive quick sorts
	lo, hi int
}

var hiloStack []hiloIndexType // for non-recursive quick sorts

// {{{
// var maxStackSize int // so I can determine the max stack size.
//                         Using PaulKrugman.dat full file, code from Modula-2 book showed 12, and Oberon book showed 24.
//                         Oberon version would have blown its stack as it defined a stack of 0..12.
//                         The Modula-2 version would have made it w/ one element to spare.
// }}}

type stringHeap []string

type timesortType struct {
	description string
	duration    time.Duration
}

// ----------------------------------------------------------- container/heap -----------------------------------------

func (h stringHeap) Len() int           { return len(h) }
func (h stringHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h stringHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

// Push and Pop use pointer receivers because they modify the slice's length, not just its contents.
// Note that Push and Pop in this interface are for package heap's implementation to call.
// To add and remove things from the heap, use heap.Push and heap.Pop.  I missed this point my first time thru the documentation.
func (h *stringHeap) Push(x interface{}) {
	*h = append(*h, x.(string))
}

func (h *stringHeap) Pop() interface{} {
	n := len(*h)
	x := (*h)[n-1]
	*h = (*h)[:n-1]
	return x
}

/*
func (h *stringHeap) Pop() interface{} {
	H := *h // this immediately dereferences the passed pointer.  I want to see if this works as a way to essentially create a ref param as in C++ and D.
	n := len(H)
	x := (H)[n-1]
	H = (H)[:n-1]  Nope, this did not work because it did not return a shorter length slice so there was an infinite loop waiting for length == 0
	return x
}

func (h *stringHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

*/

// -----------------------------------------------------------
// -----------------------------------------------------------

func StraightInsertion(input []string) []string {
	n := len(input)
	for i := 1; i < n; i++ {
		x := input[i]
		j := i
		for ; (j > 0) && (x < input[j-1]); j-- { // looks like it's moving elements right
			input[j] = input[j-1] //  until it comes to a smaller value entry.  This line does not swap anything.
		}
		input[j] = x // then it inserts the element at the spot at which it stopped moving elements.
	}
	return input
} // END StraightInsertion

func BasicInsertion(input []string) []string {
	n := len(input)
	for i := 1; i < n; i++ {
		for j := i; (j > 0) && (input[j] < input[j-1]); j-- {
			input[j], input[j-1] = input[j-1], input[j]
		}
	}
	return input
}

// -----------------------------------------------------------

func BinaryInsertion(a []string) []string {
	n := len(a)
	for i := 1; i < n; i++ {
		x := a[i]
		L := 0 //
		R := i
		for L < R { // look for smallest item in remaining range, to insert into the correct spot.
			m := (L + R) / 2
			if a[m] <= x {
				L = m + 1
			} else {
				R = m
			}
		} //END while L < R

		for j := i; j >= R+1; j-- { // it's moving elements right from current point to the right, to make room for the insertion.
			a[j] = a[j-1] // then it inserts the
		}
		a[R] = x
	} // END for i :=
	return a
} // END BinaryInsertion

// -----------------------------------------------------------

func StraightSelection(a []string) []string {
	n := len(a)
	for i := 0; i < n-1; i++ { // don't include the last element in this loop.
		k := i
		x := a[i]
		for j := i + 1; j < n; j++ { // include last element in this loop.  And this loop is now more idiomatic for Go.
			if a[j] < x {
				k = j
				x = a[k]
			}
		}
		a[k] = a[i]
		a[i] = x
	}
	return a
} // END StraightSelection

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
			} // END for/while originally (j >= k) & (x < a[j]) DO
			a[j+h] = x
		} // END FOR i := h; ...     originally i := k+1 TO n-1 DO in original code based on Pascal
	} // END FOR h
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
// I based this on bubble sort pseudocode above that I found in "Essential Algorithms", by Rod Stephens.
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
	} // END FOR h
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
// The principal of heapsort is that in phase 1, the array to be sorted is turned into a heap.
// In phase 2, the items are removed from the heap in the order just created.
// Orig code has L and R, that I made lo and hi, respectively, to help me understand it.
func sift(a []string, L, R int) []string {
	lo := L // left is lo
	hi := R // right is hi
	i := lo
	j := 2*i + 1
	x := a[i]
	if (j < hi) && (a[j] < a[j+1]) {
		j++
	} // end if
	for (j <= hi) && (x < a[j]) {
		a[i] = a[j]
		i = j
		j = 2*j + 1
		if (j < hi) && (a[j] < a[j+1]) {
			j++
		} // end if
	} //END for (j <= R) & (x < a[j])
	a[i] = x
	return a
} // END sift;

func HeapSort(a []string) []string { // I think this is based on Wirth's code in either Oberon or Modula-2.
	n := len(a)
	lo := n / 2
	hi := n - 1
	for lo > 0 { // heap creation phase.
		lo--
		a = sift(a, lo, hi)
	}
	for hi > 0 { // heap removal phase.
		a[0], a[hi] = a[hi], a[0]
		hi--
		a = sift(a, lo, hi)
	}
	return a
} // END HeapSort
// -----------------------------------------------------------
// -----------------------------------------------------------
// I did this myself, but it doesn't work.  I'm keeping this here so I don't do this again.  I didn't understand how
// a heap sort works, so this idea was wrong-headed.
/*
func ModifiedHeapSort(a []string) []string {
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

*/
//------------------------------------------------------------------------
//------------------------------------------------------------------------
/*  Don't remember where this came from, but I'm leaving it here in case I find out some day.
func siftup(items []string, n int) []string {
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
*/

func NRsiftdown(items []string, L, R int) []string { // Numerical Recipes 3rd ed (C) 2007, p 428
	i := L
	x := items[L]
	j := 2*L + 1
	for j <= R {
		if j < R && items[j] < items[j+1] { // if next element is better, use it.
			j++
		}
		if x >= items[j] { // found correct level/location for x, so terminate the sift-down.  Else demote x (by swapping) and continue.
			break
		}
		items[i] = items[j]
		i = j
		j = 2*j + 1
	}
	items[i] = x // put x into its correct level/location.
	return items
} // END siftdown;

func NRheapsort(items []string) []string { // copied from Numerical Recipes 3rd ed (C) 2007, p 428
	n := len(items)
	// the index i determines the left range of the siftdown.  Heap creation phase is also call hiring phase.
	for i := n/2 - 1; i >= 0; i-- {
		items = NRsiftdown(items, i, n-1)
	}

	// Right range of the siftdown is decremented from n-1 to 0 during the retirement and promotion phase,
	// also called heap selection
	for i := n - 1; i > 0; i-- {
		// clear a space at the end of the array and retire the top of the heap into it, by swapping
		items[0], items[i] = items[i], items[0]
		NRsiftdown(items, 0, i-1)
	}
	return items
} // END NRheapsort;
//------------------------------------------------------------------------
/*
The author says that this code is from Go's standard library, so he is just illustrating the concept.  That may be why
he does not define his siftDown.  data.Swap is obvious.
func GoHeapSort(items []string) []string { // High Performance Go, by Bob Strecansky, Packt (C) 2020, p41ff
	first := 0
	lo := 0
	hi := len(items) - 1

	// build heap w/ largest item at top
	for i := (hi - 1) / 2; i >= 0; i-- {
		siftDown(items, i, hi, first)
	}
	// pop elements, largest first, onto the end of the data
	for i := (hi - 1); i >= 0; i-- {
		data.Swap(first, first+1)
		siftDown(items, lo, i, first)
	}
	return items
}
*/

// ------------------------------------------------------------------------
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

func qsortmodified(a []string, lo, hi int) []string {
	if hi-lo < 12 {
		b := StraightInsertion(a[lo : hi+1])
		i := lo
		for _, s := range b {
			a[i] = s
			i++
		}
	} else {
		i := lo
		j := hi
		x := a[(lo+hi)/2]
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
		if lo < j {
			a = qsortmodified(a, lo, j)
		}
		if i < hi {
			a = qsortmodified(a, i, hi)
		}
	}
	return a
}

func ModifiedQuickSort(a []string) []string {
	n := len(a) - 1
	a = qsortmodified(a, 0, n)
	return a
}

// -----------------------------------------------------------
// -----------------------------------------------------------
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
	//hiloInit(n / 2)
	hiloInit(stackSize)

	k.lo = 0
	k.hi = n
	hiloStackPush(k)

	for hiloStackLen() > 0 {
		//		stacksize := hiloStackLen()
		//		if stacksize > maxStackSize {
		//			maxStackSize = stacksize
		//		}

		i0 := hiloStackPop()
		lo := i0.lo
		hi := i0.hi

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
			} // REPEAT ... UNTIL i > j
			if j-lo < hi-i {
				if i < hi { // push request to sort right partition
					k.lo = i
					k.hi = hi
					hiloStackPush(k)
				}
				hi = j // now L and R delimit the left partition, and continue sorting the left partition.
			} else {
				if lo < j { // push request for sorting left partition onto the stack
					k.lo = lo
					k.hi = j
					hiloStackPush(k)
				}
				lo = i // continue sorting right partition
			}
		} // REPEAT ... UNTIL L >= R

	} // REPEAT ... UNTIL hiloStack is empty
	// fmt.Println(" Modula-2 NonRecursiveQuickSort maxStackSize =", maxStackSize)  This showed 12 on the full PaulKrugman.dat file.  Code in the book defined stack as [1..12], so the stack was blown in the book code.
	return a
} // END NonRecursiveQuickSort

func NonRecursiveQuickSortOberon(a []string) []string {
	n := len(a)
	//intStackInit(n / 2)
	intStackInit(stackSize)
	intStackPush(0)
	intStackPush(n - 1)
	for intStackLen() > 0 { // REPEAT (*take top request from stack*)
		R := intStackPop()
		L := intStackPop()
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
					intStackPush(i)
					intStackPush(R)
				}
				R = j // (*now L and R delimit the left partition*)
			} else {
				if L < j { // push request for sorting left partition onto the atack
					intStackPush(L)
					intStackPush(j)
				}
				L = i // continue sorting right partition
			}
		} // for L < R

	} // for stack not empty
	// fmt.Println(" NonRecursiveQuickSortOberon maxStackSize=", maxStackSize) This showed 24 on the full PaulKrugman.dat file.  Code in the book defined stack as [1..12], so the stack was blown out of the water in the book code.
	return a
} // 	END NonRecursiveQuickSortOberon

// -----------------------------------------------------------
// mergesort.go
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
} // ModifiedMergeSort

// ----------------------------------------------------------
// readLine

func readLine(r *bytes.Reader) (string, error) {
	var sb strings.Builder
	for {
		byte, err := r.ReadByte()
		if err != nil {
			return strings.TrimSpace(sb.String()), err
		}
		if byte == '\n' {
			return strings.TrimSpace(sb.String()), nil
		}
		err = sb.WriteByte(byte)
		if err != nil {
			return strings.TrimSpace(sb.String()), err
		}
	}
} // readLine

//-----------------------------------------------------------------------+
//                               MAIN PROGRAM                            |
//-----------------------------------------------------------------------+

func main() {
	var filesize int64
	fmt.Println(" Sort a slice of strings, using the different algorithms.  Last altered", LastAlteredDate, ", compiled by", runtime.Version())
	fmt.Println()

	// File I/O.  Construct filenames
	if len(os.Args) <= 1 {
		fmt.Println(" Usage: heapsorter <filename>")
		os.Exit(0)
	}

	Ext1Default := ".dat"
	Ext2Default := ".txt"
	OutDefault := ".sorted"

	date := time.Now()
	datestring := date.Format("Mon Jan 2 2006 15:04:05 MST") // written to output file below.

	//commandline := getcommandline.GetCommandLineString()
	commandline := os.Args[1]
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

	//byteslice := make([]byte, 0, filesize+5) // add 5 just in case.  Now I don't think this is needed, anyway.
	byteslice, err := os.ReadFile(Filename)
	if err != nil {
		fmt.Println(" Error from os.ReadFile when reading ", Filename, ".  Exiting.")
		os.Exit(1)
	}

	//bytesbuffer := bytes.NewBuffer(byteslice)
	bytesReader := bytes.NewReader(byteslice)

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
	OutBufioWriter.WriteString("------------------------------------------------------\n")
	OutBufioWriter.WriteString(datestring)
	_, err = OutBufioWriter.WriteRune('\n')
	check(err)

	// Read in the words to sort
	scanner := bufio.NewScanner(os.Stdin) // this reads from stdin.  I would not do this today; I'd use fmt.Scanln.
	fmt.Print(" Enter number of words for this run.  0 means full file: ")
	scanner.Scan()
	answer := scanner.Text()
	if err = scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		os.Exit(1)
	}
	requestedwordcount, er := strconv.Atoi(answer)
	if er != nil {
		//fmt.Println(" No valid answer entered.  Will assume 0.")  This message is silly.
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
		//word, err := bytesbuffer.ReadString('\n')
		//word = strings.TrimSpace(word)
		word, err := readLine(bytesReader)
		if err != nil {
			break
		}
		//	word = strings.ToLower(strings.TrimSpace(word))
		//if len(word) < 4 {  This is already in makewordfile, so I don't need it here, too.
		//	continue          makewordfile also removes all non-alphanumeric characters, also.
		//}
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

	fmt.Printf(" Requested number of words is %d, actual number of words read in is %d.\n\n", requestedwordcount, numberofwords)

	// make the timesort slice to be sorted at the end
	timeSort := make([]timesortType, 0, 50)

	// sort.StringSlice method
	t9 := time.Now()
	copy(sliceofwords, mastersliceofwords)
	NativeWords := sort.StringSlice(sliceofwords)
	NativeWords.Sort()
	NativeSortTime := time.Since(t9)
	NativeSortTimeNano := NativeSortTime.Nanoseconds()
	s = fmt.Sprintf(" after NativeSort: %s, %d ns \n", NativeSortTime.String(), NativeSortTimeNano)
	fmt.Print(s)
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	if allowoutput {
		for _, w := range NativeWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts := timesortType{s, NativeSortTime} // structured constant
	timeSort = append(timeSort, ts)

	// StraightSelection
	if numberofwords < tooBig { // I don't want to wait 2 hrs for this to finish.
		t0 := time.Now()
		copy(sliceofwords, mastersliceofwords)
		sortedsliceofwords := StraightSelection(sliceofwords)
		StraightSelectionTime := time.Since(t0)
		StraightSelectionTimeNano := StraightSelectionTime.Nanoseconds()
		s = fmt.Sprintf(" After StraightSelection: %s, %d ns \n", StraightSelectionTime.String(), StraightSelectionTimeNano)
		_, err = OutBufioWriter.WriteString(s)
		check(err)
		fmt.Print(s)
		if allowoutput {
			for _, w := range sortedsliceofwords {
				fmt.Print(w, " ")
			}
			fmt.Println()
		}
		ts = timesortType{s, StraightSelectionTime} // structured constant
		timeSort = append(timeSort, ts)
	}

	// StraightInsertion
	if numberofwords < tooBig {
		t1 := time.Now()
		copy(sliceofwords, mastersliceofwords)
		sliceofsortedwords := StraightInsertion(sliceofwords)
		StraightInsertionTime := time.Since(t1)
		s = fmt.Sprintf(" After StraightInsertion: %s, %d ns \n", StraightInsertionTime.String(), StraightInsertionTime.Nanoseconds())
		_, err = OutBufioWriter.WriteString(s)
		check(err)
		fmt.Print(s)
		if allowoutput {
			for _, w := range sliceofsortedwords {
				fmt.Print(w, " ")
			}
			fmt.Println()
		}
		ts = timesortType{s, StraightInsertionTime}
		timeSort = append(timeSort, ts)

		t1a := time.Now()
		copy(sliceofwords, mastersliceofwords)
		sliceofsortedwords = BasicInsertion(sliceofwords)
		BasicInsertionTime := time.Since(t1a)
		s = fmt.Sprintf(" After BasicInsertion: %s, %d ns \n", BasicInsertionTime.String(), BasicInsertionTime.Nanoseconds())
		_, err = OutBufioWriter.WriteString(s)
		check(err)
		fmt.Print(s)
		if allowoutput {
			for _, w := range sliceofsortedwords {
				fmt.Print(w, " ")
			}
			fmt.Println()
		}
		ts = timesortType{s, BasicInsertionTime}
		timeSort = append(timeSort, ts)
	}

	// BinaryInsertion
	t2 := time.Now()
	copy(sliceofwords, mastersliceofwords)
	BinaryInsertionSortedWords := BinaryInsertion(sliceofwords)
	BinaryInsertionTime := time.Since(t2)
	s = fmt.Sprintf(" After BinaryInsertion: %s, %d ns \n", BinaryInsertionTime.String(), BinaryInsertionTime.Nanoseconds())
	fmt.Print(s)
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	if allowoutput {
		for _, w := range BinaryInsertionSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, BinaryInsertionTime}
	timeSort = append(timeSort, ts)

	// ShellSort
	if numberofwords < tooBig {
		t3 := time.Now()
		copy(sliceofwords, mastersliceofwords)
		ShellSortedWords := ShellSort(sliceofwords)
		ShellSortedTime := time.Since(t3)
		s = fmt.Sprintf(" After ShellSort: %s, %d ns \n", ShellSortedTime.String(), ShellSortedTime.Nanoseconds())
		_, err = OutBufioWriter.WriteString(s)
		check(err)
		fmt.Print(s)
		if allowoutput {
			for _, w := range ShellSortedWords {
				fmt.Print(w, " ")
			}
			fmt.Println()
		}
		ts = timesortType{s, ShellSortedTime}
		timeSort = append(timeSort, ts)
	}

	// BadShellSort -- now a misnomer as it finally works.
	t3a := time.Now()
	copy(sliceofwords, mastersliceofwords)
	BadShellSortedWords := BadShellSort(sliceofwords)
	BadShellSortedTime := time.Since(t3a)
	s = fmt.Sprintf(" After BadShellSort: %s, %d ns \n", BadShellSortedTime.String(), BadShellSortedTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Print(s)
	if allowoutput {
		for _, w := range BadShellSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, BadShellSortedTime}
	timeSort = append(timeSort, ts)

	// MyShellSort
	t3b := time.Now()
	copy(sliceofwords, mastersliceofwords)
	MyShellSortedWords := MyShellSort(sliceofwords)
	MyShellSortedTime := time.Since(t3b)
	s = fmt.Sprintf(" After MyShellSort: %s, %d ns \n", MyShellSortedTime.String(), MyShellSortedTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Print(s)
	if allowoutput {
		for _, w := range MyShellSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, MyShellSortedTime}
	timeSort = append(timeSort, ts)

	// HeapSort
	t4 := time.Now()
	copy(sliceofwords, mastersliceofwords)
	HeapSortedWords := HeapSort(sliceofwords)
	HeapSortedTime := time.Since(t4)
	s = fmt.Sprintf(" After HeapSort: %s, %d ns \n", HeapSortedTime.String(), HeapSortedTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Print(s)
	if allowoutput {
		for _, w := range HeapSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, HeapSortedTime}
	timeSort = append(timeSort, ts)

	// NRHeapSort which is from Numerical Recipies and converted from C++ code.
	t5 := time.Now()
	copy(sliceofwords, mastersliceofwords)
	NRHeapSortedWords := NRheapsort(sliceofwords)
	NRHeapTime := time.Since(t5)
	s = fmt.Sprintf(" After NRheapsort: %s, %d ns \n", NRHeapTime.String(), NRHeapTime.Nanoseconds())
	fmt.Print(s)
	if allowoutput {
		for _, w := range NRHeapSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, NRHeapTime}
	timeSort = append(timeSort, ts)
	_, err = OutBufioWriter.WriteString(s)
	//	_, err = OutBufioWriter.WriteRune('\n')
	check(err)

	// QuickSort
	t6 := time.Now()
	copy(sliceofwords, mastersliceofwords)
	QuickSortedWords := QuickSort(sliceofwords)
	QuickSortedTime := time.Since(t6)
	s = fmt.Sprintf(" After QuickSort: %s, %d ns \n", QuickSortedTime.String(), QuickSortedTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Print(s)
	if allowoutput {
		for _, w := range QuickSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, QuickSortedTime}
	timeSort = append(timeSort, ts)

	// ModifiedQuickSort
	t6a := time.Now()
	copy(sliceofwords, mastersliceofwords)
	ModifiedQuickSortedWords := ModifiedQuickSort(sliceofwords)
	ModifiedQuickSortedTime := time.Since(t6a)
	s = fmt.Sprintf(" After ModifiedQuickSort: %s, %d ns\n", ModifiedQuickSortedTime.String(), ModifiedQuickSortedTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Print(s)
	if allowoutput {
		for _, w := range ModifiedQuickSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, ModifiedQuickSortedTime}
	timeSort = append(timeSort, ts)

	// MergeSort
	t7 := time.Now()
	copy(sliceofwords, mastersliceofwords)
	MergeSortedWords := mergeSort(sliceofwords)
	MergeSortTime := time.Since(t7)
	s = fmt.Sprintf(" After mergeSort: %s, %d ns \n", MergeSortTime.String(), MergeSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Print(s)
	if allowoutput {
		for _, w := range MergeSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, MergeSortTime}
	timeSort = append(timeSort, ts)

	// ModifiedMergeSort
	t7a := time.Now()
	copy(sliceofwords, mastersliceofwords)
	ModifiedMergeSortedWords := ModifiedMergeSort(sliceofwords)
	ModifiedMergeSortTime := time.Since(t7a)
	s = fmt.Sprintf(" After ModifiedMergeSort: %s, %d ns \n", ModifiedMergeSortTime.String(), ModifiedMergeSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Print(s)
	if allowoutput {
		for _, w := range ModifiedMergeSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, ModifiedMergeSortTime}
	timeSort = append(timeSort, ts)

	// NonRecursiveQuickSort (from Modula-2)
	t8 := time.Now()
	copy(sliceofwords, mastersliceofwords)
	NonRecursiveQuickSortedWords := NonRecursiveQuickSort(sliceofwords)
	NonRecursiveQuickedTime := time.Since(t8)
	s = fmt.Sprintf(" After Modula-2 NonRecursiveQuickSort: %s, %d ns \n", NonRecursiveQuickedTime.String(), NonRecursiveQuickedTime.Nanoseconds())
	fmt.Print(s)
	if allowoutput {
		for _, w := range NonRecursiveQuickSortedWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, NonRecursiveQuickedTime}
	timeSort = append(timeSort, ts)
	_, err = OutBufioWriter.WriteString(s)
	check(err)

	// NonRecursiveQuickSortOberon
	t8a := time.Now()
	copy(sliceofwords, mastersliceofwords)
	NonRecursiveQuickSortedOberonWords := NonRecursiveQuickSortOberon(sliceofwords)
	NonRecursiveQuickOberonTime := time.Since(t8a)
	s = fmt.Sprintf(" After NonRecursiveQuickSortOberon: %s, %d ns \n", NonRecursiveQuickOberonTime.String(), NonRecursiveQuickOberonTime.Nanoseconds())
	fmt.Print(s)
	if allowoutput {
		for _, w := range NonRecursiveQuickSortedOberonWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, NonRecursiveQuickOberonTime}
	timeSort = append(timeSort, ts)
	_, err = OutBufioWriter.WriteString(s)
	check(err)

	// sort.StringSlice
	t9 = time.Now()
	copy(sliceofwords, mastersliceofwords)
	//NativeWords = sort.StringSlice(sliceofwords)  GoLand flagged this as a redundant type conversion.
	NativeWords = sliceofwords
	NativeWords.Sort()
	NativeSortTime = time.Since(t9)
	s = fmt.Sprintf(" After 2nd sort.StringSlice: %s, %d ns \n", NativeSortTime.String(), NativeSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Print(s)
	if allowoutput {
		for _, w := range NativeWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, NativeSortTime}
	timeSort = append(timeSort, ts)

	// sort.Sort
	t9 = time.Now()
	copy(sliceofwords, mastersliceofwords)
	//NativeWords = sort.StringSlice(sliceofwords)  GoLand flagged this as a redundant type conversion
	NativeWords = sliceofwords
	sort.Sort(NativeWords)
	NativeSortTime = time.Since(t9)
	s = fmt.Sprintf(" After sort.Sort: %s, %d ns \n", NativeSortTime.String(), NativeSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Print(s)
	if allowoutput {
		for _, w := range NativeWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, NativeSortTime}
	timeSort = append(timeSort, ts)

	// sort.Stable
	t9 = time.Now()
	copy(sliceofwords, mastersliceofwords)
	//NativeWords = sort.StringSlice(sliceofwords)  GoLand flagged this as a redundant type conversion
	NativeWords = sliceofwords
	sort.Stable(NativeWords)
	NativeSortTime = time.Since(t9)
	s = fmt.Sprintf(" After sort.Stable: %s, %d ns \n", NativeSortTime.String(), NativeSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Print(s)
	if allowoutput {
		for _, w := range NativeWords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, NativeSortTime}
	timeSort = append(timeSort, ts)

	// sort.Strings
	t10 := time.Now()
	copy(sliceofwords, mastersliceofwords)
	sort.Strings(sliceofwords)
	StringsSortTime := time.Since(t10)
	s = fmt.Sprintf(" After sort.Strings: %s, %d ns \n", StringsSortTime.String(), StringsSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Print(s)
	if allowoutput {
		for _, w := range sliceofwords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, StringsSortTime}
	timeSort = append(timeSort, ts)

	// sort.Slice
	t11 := time.Now()
	copy(sliceofwords, mastersliceofwords)
	lessfunction := func(i, j int) bool {
		return sliceofwords[i] < sliceofwords[j]
	}
	sort.Slice(sliceofwords, lessfunction)
	SliceSortTime := time.Since(t11)
	s = fmt.Sprintf(" After sort.Slice: %s, %d ns \n", SliceSortTime.String(), SliceSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Print(s)
	if allowoutput {
		for _, w := range sliceofwords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, SliceSortTime}
	timeSort = append(timeSort, ts)

	// sort.SliceStable
	t12 := time.Now()
	copy(sliceofwords, mastersliceofwords)
	sort.SliceStable(sliceofwords, lessfunction)
	SliceStableSortTime := time.Since(t12)
	s = fmt.Sprintf(" After sort.SliceStable: %s, %d ns \n", SliceStableSortTime.String(), SliceStableSortTime.Nanoseconds())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Print(s)
	if allowoutput {
		for _, w := range sliceofwords {
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, SliceStableSortTime}
	timeSort = append(timeSort, ts)

	// container/heap
	heapofwords := make(stringHeap, 0, numberofwords) // doesn't need to be a separate type, but it makes my intent clear.
	sortedheapofwords := make(stringHeap, 0, numberofwords)
	heap.Init(&heapofwords)
	t13 := time.Now()
	for _, wrd := range mastersliceofwords {
		heap.Push(&heapofwords, wrd)
	}

	var str string
	for heapofwords.Len() > 0 { // Note: as items are popped off of the heap, it's length gets smaller.  So using i < heapofwords.Len() didn't work.
		str = heap.Pop(&heapofwords).(string) // this works to force the interface type treated as the string that it is.
		sortedheapofwords = append(sortedheapofwords, str)
	}
	sortedheapofwordsTime := time.Since(t13)
	s = fmt.Sprintf(" after container/heap: %s \n", sortedheapofwordsTime.String())
	_, err = OutBufioWriter.WriteString(s)
	check(err)
	fmt.Print(s)
	if allowoutput {
		for i := 0; i < len(sortedheapofwords); i++ {
			w := sortedheapofwords[i]
			fmt.Print(w, " ")
		}
		fmt.Println()
	}
	ts = timesortType{s, sortedheapofwordsTime}
	timeSort = append(timeSort, ts)
	//	_, err = OutBufioWriter.WriteRune('\n')
	//	check(err)
	fmt.Println()

	// Sort times and output sorted list
	sortlessfcn := func(i, j int) bool {
		return timeSort[i].duration < timeSort[j].duration
	}
	sort.Slice(timeSort, sortlessfcn)
	fmt.Println(" \n --- Sorted list of times is: ------")
	_, err = OutBufioWriter.WriteString(" \n ---- Sorted List of Times ----\n")
	if err != nil {
		fmt.Fprintf(os.Stderr, " err from OutBufioWriter is: %s\n", err)
	}
	for _, t := range timeSort {
		fmt.Print(t.description)
		_, err = OutBufioWriter.WriteString(t.description)
		check(err)
	}
	fmt.Println("  \n  ") // should print 2 blank lines

	// Wrap it up by writing number of words, etc.
	s = fmt.Sprintf(" requestedwordcount= %d, numberofwords= %d, len(mastersliceofwords)= %d \n",
		requestedwordcount, numberofwords, len(mastersliceofwords))
	_, err = OutBufioWriter.WriteString(s)
	check(err)
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

// ===========================================================
func check(e error) {
	if e != nil {
		panic(e)
	}
}

/*
func pause() {
	fmt.Print(" hit <enter> to continue")
	fmt.Scanln()
}

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

After adding and debugging container/heap, it is the approx the same timing as sort.Stable and sort.SliceStable.
But it does work once I understood how it is expected to be used.
This sort is not in place, so the timing includes the for loops that are not included in the measurement in the other methods.

Thu Jul 23 2020 17:16:06 EDT, and then Tue Jul 28 2020 17:38:53 EDT on z76
 filesize = 834660, requestedwordcount = 119237, numberofwords= 104603, len(mastersliceofwords)= 104603
 After QuickSort: 25.702455 ms -> 26.126064 ms
 After NonRecursiveQuickSortOberon: 26.821662 ms -> 26.96439 ms
 After sort.Slice: 27.491206 ms -> 27.702159 ms
 After 2nd sort.StringSlice: 29.588204 ms -> 31.060037 ms
 After sort.Strings: 29.732935 ms -> 30.637129 ms
 After sort.Sort: 29.759328 ms -> 29.813508 ms
 After sort.SliceStable: 74.335231 ms -> 75.846885 ms
 after NativeSort: 30.757105 ms -> 33.118014 ms
 After NRheapsort: 45.628029 ms -> 46.167062 ms
 After HeapSort: 47.248887 ms -> 48.055007ms
 After Modula-2 NonRecursiveQuickSort: 50.680713 ms ->  48.316853 ms
 After ModifiedMergeSort: 56.346407 ms -> 67.006893 ms
 After BadShellSort: 61.245545 ms -> 61.051531 ms
 After sort.Stable: 79.934308 ms -> 80.047117 ms
 after container/heap based sort time: 87.553214 ms -> 91.973921 ms
 After mergeSort: 91.337021 ms -> 80.618479 ms
 After MyShellSort: 386.692135 ms -> 386.779997 ms
 After BinaryInsertion: 3.547058203 s -> 3.54911457 s
 After StraightInsertion: 28.062582641 s -> 28.038693805 s
 After ShellSort: 28.50378449 s -> 28.594864915 s
 After StraightSelection: 47.00024184 s -> 47.048454729 s
----------------------------------------------------------------------------------------------------
Tue Aug 25 2020 20:56:28 EDT
 filesize = 8,729,361, requestedwordcount= 1,247,051, numberofwords= 1,080,126
ms
 After sort.Strings: 300.6921ms
 After ModifiedQuickSort: 307.6901ms  (added 1/4/21)
 After 2nd sort.StringSlice: 308.6832ms
 After sort.Sort: 313.6777ms
 After QuickSort: 315.6763ms
 After sort.Slice: 315.7076ms
 after NativeSort: 325.6655ms
 After NonRecursiveQuickSortOberon: 325.6664ms
 After Modula-2 NonRecursiveQuickSort: 433.5419ms
 After ModifiedMergeSort: 441.5628ms
 After mergeSort: 563.4229ms
 After NRheapsort: 847.1315ms
 After sort.Stable: 891.0876ms
 After sort.SliceStable: 922.282ms
 After HeapSort: 947.1116ms
s
 after container/heap, sortedheapofwords: time=1.3915749s
 After BadShellSort: 1.2157665s
 After MyShellSort: 19.4820301s
m
 After BinaryInsertion: 8m59.7050462s
h
 After ShellSort: 1h48m15.8470534s
 After StraightInsertion: 1h48m46.1262238s
 After BasicInsertion: 1h56m38.5238831s
 After StraightSelection: 1h57m51.6689395s

------------------------------------------------------
Mon Jan 4 2021 on same  requestedwordcount= 1247051, numberofwords= 1,080,126
 ---- Sorted List of Times ----
 after NativeSort: 283.7105ms, 283710500 ns
 After sort.Sort: 291.7665ms, 291766500 ns
 After sort.Strings: 293.7003ms, 293700300 ns
 After 2nd sort.StringSlice: 301.7085ms, 301708500 ns
 After sort.Slice: 302.2547ms, 302254700 ns
 After ModifiedQuickSort: 307.6901ms, 307690100 ns
 After QuickSort: 315.9826ms, 315982600 ns
 After NonRecursiveQuickSortOberon: 332.6597ms, 332659700 ns
 After Modula-2 NonRecursiveQuickSort: 393.5936ms, 393593600 ns
 After ModifiedMergeSort: 420.5925ms, 420592500 ns
 After mergeSort: 558.1632ms, 558163200 ns
 After NRheapsort: 780.7471ms, 780747100 ns
 After sort.SliceStable: 781.6792ms, 781679200 ns
 After HeapSort: 791.7811ms, 791781100 ns
 After BadShellSort: 1.3078114s, 1307811400 ns
 after container/heap: time=1.3266436s for 0 entries by heapofwords.Len(), and 1080126 by Len(sortedheapofworda)
 After MyShellSort: 18.364497s, 18364497000 ns
 After BinaryInsertion: 8m25.5667533s, 505566753300 ns
 After ShellSort: 1h43m54.8550407s, 6234855040700 ns

------------------------------------------------------




{{{
From the documentation at golang.org for container/heap

// This example demonstrates an integer heap built using the heap interface.
package main

import (
	"container/heap"
	"fmt"
)

// An IntHeap is a min-heap of ints.
type IntHeap []int

func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(int))
}

func (h *IntHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// This example inserts several ints into an IntHeap, checks the minimum,
// and removes them in order of priority.
func main() {
	h := &IntHeap{2, 1, 5}
	heap.Init(h)
	heap.Push(h, 3)
	fmt.Printf("minimum: %d\n", (*h)[0])
	for h.Len() > 0 {
		fmt.Printf("%d ", heap.Pop(h))
	}
}



// This example demonstrates a priority queue built using the heap interface.
package main

import (
	"container/heap"
	"fmt"
)

// An Item is something we manage in a priority queue.
type Item struct {
	value    string // The value of the item; arbitrary.
	priority int    // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, value string, priority int) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}

// This example creates a PriorityQueue with some items, adds and manipulates an item,
// and then removes the items in priority order.
func main() {
	// Some items and their priorities.
	items := map[string]int{
		"banana": 3, "apple": 2, "pear": 4,
	}

	// Create a priority queue, put the items in it, and
	// establish the priority queue (heap) invariants.
	pq := make(PriorityQueue, len(items))
	i := 0
	for value, priority := range items {
		pq[i] = &Item{
			value:    value,
			priority: priority,
			index:    i,
		}
		i++
	}
	heap.Init(&pq)

	// Insert a new item and then modify its priority.
	item := &Item{
		value:    "orange",
		priority: 1,
	}
	heap.Push(&pq, item)
	pq.update(item, item.value, 5)

	// Take the items out; they arrive in decreasing priority order.
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		fmt.Printf("%.2d:%s ", item.priority, item.value)
	}
}
}}}
*/
