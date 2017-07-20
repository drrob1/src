I did Modula-2 versions of these long ago.  I found my results in a file called SortPointerStuff.mod.  
Here are the results from back then (Sept 2011 or so).


                 20,916 words | 75,131 words
heapsort         0.015625 sec |  0.078125 sec
BinInsSort       0.44141 sec  |  6.53906 sec
StraightInsSort  1.53516 sec  | 70.83984 sec
qsort           12.82812 sec  |  0.062500 sec
qsort2                        |  0.062500 sec
QUICKSORT                     |  0.062500 sec
TREESORT                      |  0.078125 sec

As I remembered, heapsort and treesort are essentially the same.  I don't have a merge sort here, and Wirth's book and this MIT course on Python
both favor mergesort.  Maybe I'll do something about this in Go.  Looks like I never debugged the mergesort that's there in the file.

"Numerical Recipies" favors heapsort.  I don't know how close heapsort is to mergesort.  That chapter says that quicksort is, on average,
the fastest algorithm.  But it's worst case is O(n**2).  In fact, its worst case is an already ordered list.


I have a python version of mergesort.  I'm going to ignore treesort, as it's superfluous.


const MaxDim = 10000000;  // 10,000,000 for testing
const floatequaltolerance = 1.e-5
var inputstring []string
inputstring = make([]string,whatever was read in)


In Oberon
PROCEDURE StraightInsertion;
VAR
   i, j: INTEGER; 
   x: Item;
BEGIN
   FOR i := 1 TO n-1 DO
      x := a[i]; 
      j := i;
      WHILE (j > 0) & (x < a[j-1] DO 
        a[j] := a[j-1]; 
        DEC(j) 
      END ;
      a[j] := x
   END
END StraightInsertion 


In go
func StraightInsertion(input []string) []string{
   n := len(input)
   for i := 1; i < n; i++ {
      x := input[i]; 
      j := i;
      for (j > 0) && (x < input[j-1] {
        input[j] = input[j-1]
        j--
      } // while j > 0  and .LT. operator used
      input[j] := x
   } // for i := 1 TO n-1 
   return input
} // END StraightInsertion 


In Oberon
PROCEDURE BinaryInsertion(VAR a: ARRAY OF Item; n: INTEGER);
VAR
   i, j, m, L, R: INTEGER; 
   x: Item; 
BEGIN
   FOR i := 1 TO n-1 DO 
      x := a[i]; 
      L := 1; 
      R := i;
      WHILE L < R DO
         m := (L+R) DIV 2;
         IF a[m] <= x THEN
           L := m+1
         ELSE
           R := m 
         END
      END ;
      FOR j := i TO R+1 BY -1 DO
        a[j] := a[j-1]
      END ;
      a[R] := x
   END
END BinaryInsertion

in Go
func BinaryInsertion(a []string) []string {
   n := len(a)
   for i := 1; i < n; i++ {
      x := a[i]; 
      L := 1; 
      R := i;
      for L < R {
         m := (L+R) / 2
         if a[m] <= x {
           L = m+1
         }else{
           R = m 
         } // END if a[m] <= x
      }//END for L < R
      for j := i; j <= R+1; j-- {
        a[j] := a[j-1]
      }//END for j := i TO R+1 BY -1 DO
      a[R] := x
   } // END for i := 1 to n-1
   return a
} // END BinaryInsertion



In Oberon
PROCEDURE StraightSelection;
VAR
   i, j, k: INTEGER;
   x: Item;
BEGIN
   FOR i := 0 TO n-2 DO 
      k := i;
      x := a[i];
      FOR j := i+1 TO n-1 DO 
         IF a[j] < x THEN
           k := j; 
           x := a[k]
         END
      END ;
      a[k] := a[i]; 
      a[i] := x
   END
END StraightSelection 

In Go
func StraightSelection(a []string) []string {
   for i := 0; i <= n-2; i++ { 
      k := i;
      x := a[i];
      for j := i+1; j <= n-1; j++ {
         if a[j] < x {
           k = j; 
           x = a[k]
         } // end if
      } // END for j := i+1 TO n-1
      a[k] = a[i]; 
      a[i] = x
   } // END for i := 0 to n-2
   return a
} // END StraightSelection 


In Oberon
2.3.1  Insertion Sort by Diminishing Increment, DL Shell, 1959.

PROCEDURE ShellSort;  
CONST T = 4; 
VAR
   i, j, k, m, s: INTEGER;
   x: Item; 
   h: ARRAY T OF INTEGER;
BEGIN
   h[0] := 9;
   h[1] := 5; 
   h[2] := 3;
   h[3] := 1;
   FOR m := 0 TO T-1 DO
      k := h[m]; 
      FOR i := k+1 TO n-1 DO
         x := a[i];
         j := i-k;
         WHILE (j >= k) & (x < a[j]) DO
           a[j+k] := a[j];
           j := j-k
         END ; 
         a[j+k] := x
      END 
   END
END ShellSort 


in Go

func ShellSort(a []string) []string {
   const T = 4
   var h [T]int

   h[0] = 9;
   h[1] = 5; 
   h[2] = 3;
   h[3] = 1;
   n := len(a)
   for m := 0; m < T; m++ {
      k := h[m]; 
      for i := k+1; i < n; i++ {
         x := a[i];
         j := i-k;
         for (j >= k) & (x < a[j]) {
           a[j+k] = a[j];
           j = j-k
         } // END for/while (j >= k) & (x < a[j]) DO 
         a[j+k] = x
      } // END FOR i := k+1 TO n-1 DO
   } // END FOR m := 0 TO T-1 DO
   return a
} //END ShellSort 




HeapSort

Table 2.7  Example of a Heapsort Process.
The example of Table 2.7 shows that the resulting order is actually inverted.  This, however, can easily be 
remedied by changing the direction of the ordering relations in the sift procedure.  This results in the
following procedure Heapsort.  Note that sift should actually be declared local to Heapsort. 


In Oberon
PROCEDURE sift(L, R: INTEGER); 
VAR
   i, j: INTEGER;
   x: Item;
BEGIN 
   i := L; 
   j := 2*i+1; 
   x := a[i];
   IF (j < R) & (a[j] < a[j+1]) THEN
     j := j+1
   END ;
   WHILE (j <= R) & (x < a[j]) DO
      a[i] := a[j];
      i := j;
      j := 2*j+1;
      IF (j < R) & (a[j] < a[j+1]) THEN
        j := j+1 
      END
   END ;
   a[i] := x
END sift; 

PROCEDURE HeapSort;
VAR
   L, R: INTEGER; 
   x: Item; 
BEGIN
   L := n DIV 2; 
   R := n-1;
   WHILE L > 0 DO
     DEC(L); 
     sift(L, R) 
   END ;
   WHILE R > 0 DO
      x := a[0];
      a[0] := a[R];
      a[R] := x;
      DEC(R);
      sift(L, R)
   END
END HeapSort 



In Go
func sift(L, R int) {
   i := L; 
   j := 2*i+1; 
   x := a[i];
   if (j < R) && (a[j] < a[j+1]) {
     j += 1
   } // end if
   for (j <= R) && (x < a[j]) { // was while in original code
      a[i] = a[j]
      i = j;
      j = 2*j+1;
      if (j < R) && (a[j] < a[j+1]) {
        j += 1 
      } // end if
   } //END for/while (j <= R) & (x < a[j]) DO
   a[i] = x
} // END sift; 

func HeapSort(a []string) []string {
VAR
   L, R: INTEGER; 
   x: Item; 

   L := n / 2; 
   R := n-1;
   for L > 0 {
     L--
     func (L,R) {
       sift(L,R)
     } (L, R)
   } // END for-while L>0
   for R > 0 {
      a[0], a[R] = a[R], a[0]
      R--
      func (L,R) {
        sift(L,R)
      } (L, R)
   } // END for-while R > 0
} // END HeapSort 

------------------------------------------------------------------------
Original Modula-2 version of heapsort.  I don't remember where I got it.

func siftup(n int) { // items is global to this function which is called as an anonymous closure.
  i := n
  done := false;
  for ( i > 0 ) &&  ! done { // Originally a while statement
    p := (i - 1) / 2
    if items[i] <= items[p] {
      done = true
    }else{
      items[i], items[p] = items[p], items[i]
      i = p
    } // END (* end if *)
  } // END (* end for-while *)
} // END siftup;

func siftdown(n int) { // items is global to this function which is called as an anonymous closure.
  i := 0
  c := 0
  done := false
  for ( c < n ) && ! done { // originally a while statement
    c = 2 * i + 1
    if c <= n {
      if ( c < n ) && ( items[ c ] <= items[ c+1 ] ) {
        c++
      } // END if
      if items[c] <= items[i] {
        done := true
      }else{
        items[c], items[i] = items[i], items[c]
        i = c
      } // END if items[c] <= items[i]
    } // END if c <= n
  } // END (* for-while *);
  return items
} // END siftdown;

func anotherheapsort(items []string) []string {
  size := len(items)
  number := size - 1;
  for index := 1; index <= number; index++ {
    func (idx int) {
      siftup(idx)
    }(index)
  } // END for

  for index := number, index >= 1; index-- {
    items[0], items[index] = items[index], items[0]
    func (idx int) {
      siftdown(idx)
    }(index-1)
  } // END for
  return items
} // END anotherheapsort;

-------------------------------------------------------------


In Oberon
QuickSort

Note that sort should actually be declared local to Quicksort. 
PROCEDURE sort(L, R: INTEGER);
VAR
   i, j: INTEGER;
   w, x: Item;
BEGIN
   i := L; 
   j := R;
   x := a[(L+R) DIV 2];
   REPEAT 
      WHILE a[i] < x DO 
        INC(i) 
      END ;
      WHILE x < a[j] DO  
        DEC(j) 
      END ;
      IF i <= j THEN
        w := a[i]; 
        a[i] := a[j]; 
        a[j] := w;
        i := i+1; 
        j := j-1
      END
   UNTIL i > j;
   IF L < j THEN
     sort(L, j) 
   END ;
   IF i < R THEN
     sort(i, R) 
   END
END sort; 

PROCEDURE QuickSort;
BEGIN
  sort(0, n-1) 
END QuickSort

In Go
func qsort(L, R int){
   i := L; 
   j := R;
   x := a[(L+R) / 2];
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
   }            // UNTIL i > j;
   if L < j {
     qsort(L, j) 
   }
   if i < R {
     qsort(i, R) 
   }
} // END qsort; 

func QuickSort(a []string) []string {
  n := len(a) - 1
  func (a,b) {
    qsort(a, b) 
  }(0,n)
  return a
} // END QuickSort



In Oberon
PROCEDURE NonRecursiveQuickSort;
CONST
   M = 12;
VAR
   i, j, L, R, s: INTEGER;
   x, w: Item;
   low, high: ARRAY M OF INTEGER;  (*index stack*)
BEGIN
   s := 0; 
   low[0] := 0;
   high[0] := n-1;
   REPEAT (*take top request from stack*) 
      L := low[s];
      R := high[s]; 
      DEC(s);
      REPEAT (*partition a[L] ...  a[R]*) 
         i := L;
         j := R;
         x := a[(L+R) DIV 2];
         REPEAT
            WHILE a[i] < x DO
              INC(i) 
            END ;
            WHILE x < a[j] DO
              DEC(j) 
            END ;
            IF i <= j THEN
               w := a[i];
               a[i] := a[j];
               a[j] := w;
               i := i+1;
               j := j-1
            END
         UNTIL i > j;
         IF i < R THEN  (*stack request to sort right partition*)
            INC(s);
            low[s] := i;
            high[s] := R
         END ;
         R := j  (*now L and R delimit the left partition*)
      UNTIL L >= R
   UNTIL s = 0
END NonRecursiveQuickSort 


In Go 
func NonRecursiveQuickSort(a []string) []string {
const M = 12;
var   low, high: [M]int  // index stack

   n := len(a)
   s := 0; 
   low[0] := 0;
   high[0] := n-1;
  
   for {      // REPEAT take top request from stack
      L := low[s];
      R := high[s]; 
      DEC(s);
      for {     // REPEAT (*partition a[L] ...  a[R]*) 
         i := L;
         j := R;
         x := a[(L+R) / 2];
         for {    // REPEAT
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
            if i > j {       // UNTIL i > j;
              break
            }
         }
         if i < R {  // stack request to sort right partition
            s++
            low[s] = i;
            high[s] = R
         }
         R := j  // now L and R delimit the left partition
         if L >= R {  // UNTIL L >= R
           break
         }
      }
      if s = 0 { // UNTIL s = 0
        break
      }
   }
}  // END NonRecursiveQuickSort 


MergeSort

We may now proceed to describe the entire algorithm in terms of a procedure operating on the global array a with 2n elements. 
In Oberon
PROCEDURE StraightMerge;
VAR
   i, j, k, L, t: INTEGER;   (*index range of a is 0 ..  2*n-1 *)
   h, m, p, q, r: INTEGER; 
   up: BOOLEAN;
BEGIN
   up := TRUE;
   p := 1;
   REPEAT
      h := 1;
      m := n;
      IF up THEN
        i := 0;
        j := n-1;
        k := n; 
        L := 2*n-1 
      ELSE
        k := 0;
        L := n-1;
        i := n;
        j := 2*n-1
      END ; 
      REPEAT (*merge a run from i- and j-sources to k-destination*)
         IF m >= p THEN
           q := p
         ELSE 
           q := m 
         END ;
         m := m-q;
         IF m >= p THEN
           r := p
         ELSE r := m 
         END ;
         m := m-r;
         WHILE (q > 0) & (r > 0) DO
            IF a[i] < a[j] THEN 
               a[k] := a[i];
               k := k+h;
               i := i+1; 
               q := q-1
            ELSE
               a[k] := a[j]; 
               k := k+h; 
               j := j-1;
               r := r-1
            END
         END ;
         WHILE r > 0 DO
            a[k] := a[j]; 
            k := k+h; 
            j := j-1; 
            r := r-1
         END ;
         WHILE q > 0 DO
            a[k] := a[i];
            k := k+h; 
            i := i+1; 
            q := q-1
         END ;
         h := -h; 
         t := k; 
         k := L; 
         L := t
      UNTIL m = 0; 
      up := ~up; 
      p := 2*p
   UNTIL p >= n;
   IF ~up THEN
      FOR i := 1 TO n DO 
        a[i] := a[i+n]
      END
   END
END StraightMerge 


In Go -- I don't get why len is 2n-1.  I'm skipping this.  It looks like it's a mergesort in place algorithm anyway.

func StraightMerge(a []string] []string {
VAR
   i, j, k, L, t: INTEGER;   (*index range of a is 0 ..  2*n-1 *)
   h, m, p, q, r: INTEGER; 
BEGIN
   up := true;
   p := 1;
   REPEAT
      h := 1;
      m := n;
      IF up THEN
        i := 0;
        j := n-1;
        k := n; 
        L := 2*n-1 
      ELSE
        k := 0;
        L := n-1;
        i := n;
        j := 2*n-1
      END ; 
      REPEAT (*merge a run from i- and j-sources to k-destination*)
         IF m >= p THEN
           q := p
         ELSE 
           q := m 
         END ;
         m := m-q;
         IF m >= p THEN
           r := p
         ELSE r := m 
         END ;
         m := m-r;
         WHILE (q > 0) & (r > 0) DO
            IF a[i] < a[j] THEN 
               a[k] := a[i];
               k := k+h;
               i := i+1; 
               q := q-1
            ELSE
               a[k] := a[j]; 
               k := k+h; 
               j := j-1;
               r := r-1
            END
         END ;
         WHILE r > 0 DO
            a[k] := a[j]; 
            k := k+h; 
            j := j-1; 
            r := r-1
         END ;
         WHILE q > 0 DO
            a[k] := a[i];
            k := k+h; 
            i := i+1; 
            q := q-1
         END ;
         h := -h; 
         t := k; 
         k := L; 
         L := t
      UNTIL m = 0; 
      up := ~up; 
      p := 2*p
   UNTIL p >= n;
   IF ~up THEN
      FOR i := 1 TO n DO 
        a[i] := a[i+n]
      END
   END
} // END StraightMerge 



MergeSort.py
# -*- coding: utf-8 -*-
"""
Created on Wed May 18 20:34:31 2016

@author: ericgrimson
"""
import operator

def mergeSort(L, compare = operator.lt):
    if len(L) < 2:
        return L[:]
    else:
        middle = int(len(L)/2)
        left = mergeSort(L[:middle], compare)
        right = mergeSort(L[middle:], compare)
        return merge(left, right, compare)

def merge(left, right, compare):
    result = []
    i,j = 0, 0
    while i < len(left) and j < len(right):
        if compare(left[i], right[j]):
            result.append(left[i])
            i += 1
        else:
            result.append(right[j])
            j += 1
    while (i < len(left)):
        result.append(left[i])
        i += 1
    while (j < len(right)):
        result.append(right[j])
        j += 1
    return result



In Go

mergesort.go

func mergeSort(L []string) []string {
    if len(L) < 2{
        return L
    }else{
        middle := len(L)/2  // middle needs to be of type int
        left := mergeSort(L[:middle])
        right := mergeSort(L[middle:])
        return merge(left, right)
    } // end if else clause
}


func merge(left, right []string) []string {
    result := make([]string,0,len(left)+len(right))
    i,j := 0, 0
    while i < len(left) && j < len(right){
        if left[i] < right[j]{
            result.append(left[i])
            i += 1
        }else{
            result.append(right[j])
            j += 1
        }
    } // end while

    while (i < len(left)) {
        result.append(left[i])
        i += 1
    }

    while (j < len(right)){
        result.append(right[j])
        j += 1
    }

    return result
}

From Essential Algorithms last updated 5/1/15.  It uses a C-like pseudo-code

MakeHeap(Data: values[])
    // Add each item to the heap one at a time.
    For i = 0 To <length of values> - 1
        // Start at the new item, and work up to the root.
        Integer: index = i
        While (index != 0)
            // Find the parent's index.
            Integer: parent = (index - 1) / 2
            // If child <= parent, we're done, so
            // break out of the While loop.
            If (values[index] <= values[parent]) Then Break
            // Swap the parent and child.
            Data: temp = values[index]
            values[index] = values[parent]
            values[parent] = temp
            // Move to the parent.
            index = parent
        End While
    Next i
End MakeHeap

Data: RemoveTopItem (Data: values[], Integer: count)
    // Save the top item to return later.
    Data: result = values[0]
    // Move the last item to the root.
    values[0] = values[count - 1]
    // Restore the heap property.
    Integer: index = 0
    While (True)
        // Find the child indices.
        Integer: child1 = 2 * index + 1
        Integer: child2 = 2 * index + 2
        // If a child index is off the end of the tree,
        // use the parent's index.
        If (child1 >= count) Then child1 = index
        If (child2 >= count) Then child2 = index
        // If the heap property is satisfied,
        // we're done, so break out of the While loop.
        If ((values[index] >= values[child1]) And
            (values[index] >= values[child2])) Then Break
        // Get the index of the child with the larger value.
        Integer: swap_child
        If (values[child1] > values[child2]) Then
            swap_child = child1
        Else
            swap_child = child2
        // Swap with the larger child.
        Data: temp = values[index]
        values[index] = values[swap_child]
        values[swap_child] = temp
        // Move to the child node.
        index = swap_child
    End While
    // Return the value we removed from the root.
    return result
End RemoveTopItem


Heapsort(Data: values)
    <Turn the array into a heap.>
    For i = <length of values> - 1 To 0 Step -1
        // Swap the root item and the last item.
        Data: temp = values[0]
        values[0] = values[i]
        values[i] = temp
        <Consider the item in position i to be removed from the heap, so
         the heap now holds i - 1 items.  Push the new root value down
         into the heap to restore the heap property.>
    Next i
End Heapsort



The following pseudocode shows the entire quicksort algorithm at a low level:

// Sort the indicated part of the array.
Quicksort(Data: values[], Integer: start, Integer: end)
    // If the list has no more than one element, it's sorted.
    If (start >= end) Then Return
    // Use the first item as the dividing item.
    Integer: divider = values[start]
    // Move items < divider to the front of the array and
    // items >= divider to the end of the array.
    Integer: lo = start
    Integer: hi = end
    While (True)
        // Search the array from back to front starting at "hi"
        // to find the last item where value < "divider."
        // Move that item into the hole.  The hole is now where
        // that item was.
        While (values[hi] >= divider)
            hi = hi - 1
            If (hi <= lo) Then <Break out of the inner While loop.>
        End While
        If (hi <= lo) Then
            // The left and right pieces have met in the middle
            // so we're done.  Put the divider here, and
            // break out of the outer While loop.
            values[lo] = divider
            <Break out of the outer While loop.>
        End If
        // Move the value we found to the lower half.
        values[lo] = values[hi]
        // Search the array from front to back starting at "lo"
        // to find the first item where value >= "divider."
        // Move that item into the hole.  The hole is now where
        // that item was.
        lo = lo + 1
        While (values[lo] < divider)
            lo = lo + 1
            If (lo >= hi) Then <Break out of the inner While loop.>
        End While
        If (lo >= hi) Then
            // The left and right pieces have met in the middle
            // so we're done.  Put the divider here, and
            // break out of the outer While loop.
            lo = hi
            values[hi] = divider
            <Break out of the outer While loop.>
        End If
        // Move the value we found to the upper half.
        values[hi] = values[lo]
    End While
    // Recursively sort the two halves.
    Quicksort(values, start, lo - 1)
    Quicksort(values, lo + 1, end)
End Quicksort


Mergesort(Data: values[], Data: scratch[], Integer: start, Integer: end)
    // If the array contains only one item, it is already sorted.
    If (start == end) Then Return
    // Break the array into left and right halves.
    Integer: midpoint = (start + end) / 2
    // Call Mergesort to sort the two halves.
    Mergesort(values, scratch, start, midpoint)
    Mergesort(values, scratch, midpoint + 1, end)
    // Merge the two sorted halves.
    Integer: left_index = start
    Integer: right_index = midpoint + 1
    Integer: scratch_index = left_index
    While ((left_index <= midpoint) And (right_index <= end))
        If (values[left_index] <= values[right_index]) Then
            scratch[scratch_index] = values[left_index]
            left_index = left_index + 1
        Else
            scratch[scratch_index] = values[right_index]
            right_index = right_index + 1
        End If
        scratch_index = scratch_index + 1    End While
    // Finish copying whichever half is not empty.
    For i = left_index To midpoint
        scratch[scratch_index] = values[i]
        scratch_index = scratch_index + 1
    Next i
    For i = right_index To end
        scratch[scratch_index] = values[i]
        scratch_index = scratch_index + 1
    Next i
    // Copy the values back into the original values array.
    For i = start To end
154 Chapter 6 ¦ Sorting
        values[i] = scratch[i]
    Next i
End Mergesort

NOTE It is possible to merge the sorted halves without using a scratch array, but it's more complicated and slower, so most programmers use a scratch array.


STABLE SORTING

A stable sorting algorithm is one that maintains the original relative positioning
of equivalent values.  For example, suppose a program is sorting Car objects
by their Cost properties and Car objects A and B have the same Cost values.  If
object A initially comes before object B in the array, then in a stable sorting algorithm,
object A still comes before object B in the sorted array.

If the items you are sorting are value types such as integers, dates, or strings,
then two entries with the same values are equivalent, so it doesn't matter if the
sort is stable.  For example, if the array contains two entries that have value 47, it 
doesn't matter which 47 comes ?  rst in the sorted array.

In contrast, you might care if Car objects are rearranged unnecessarily.  A
stable sort lets you sort the array multiple times to get a result that is sorted on
multiple keys (such as Maker and Cost for the Car example).
Mergesort is easy to implement as a stable sort (the algorithm described earlier
is stable), so it is used by Java's Arrays.sort library method.

Mergesort is also easy to parallelize, so it may be useful on computers that
have more than one CPU.  See Chapter 18 for information on implementing mergesort
on multiple CPUs.

Quicksort may often be faster, but mergesort still has some advantages.




Algorithms in a nutshell updated 5/2/13.  It uses C.

Example 4-2.  Insertion Sort using value-based information
void sortValues (void *base, int n, int s, int(*cmp)(const void *, const void *)) {
  int j;
  void *saved = malloc (s);
  for (j = 1; j < n; j++) {
    /* start at end, work backward until smaller element or i < 0. */
    int i = j-1;
    void *value = base + j*s;
    while (i >= 0 && cmp(base + i*s, value) > 0) { i--; }
    /* If already in place, no movement needed.  Otherwise save value to be
     * inserted and move as a LARGE block intervening values.  Then insert
     * into proper position. */
    if (++i == j) continue;
    memmove (saved, value, s);
    memmove (base+(i+1)*s, base+i*s, s*(j-i));
    memmove (base+i*s, saved, s);
  }
  free (saved);
}





Example 4-7.  Quicksort implementation in C
/**
 * Sort array ar[left,right] using Quicksort method.
 * The comparison function, cmp, is needed to properly compare elements.
 */
void do_qsort (void **ar, int(*cmp)(const void *,const void *),
               int left, int right) {
  int pivotIndex;
  if (right <= left) { return; }
  /* partition */
  pivotIndex = selectPivotIndex (ar, left, right);
  pivotIndex = partition (ar, cmp, left, right, pivotIndex);
  if (pivotIndex-1-left <= minSize) {
    insertion (ar, cmp, left, pivotIndex-1);
  } else {
    do_qsort (ar, cmp, left, pivotIndex-1);
  }
  if (right-pivotIndex-1 <= minSize) {
    insertion (ar, cmp, pivotIndex+1, right);
  } else {
    do_qsort (ar, cmp, pivotIndex+1, right);
  }
}
/**  Qsort straight */
void sortPointers (void **vals, int total_elems,
                   int(*cmp)(const void *,const void *)) {
  do_qsort (vals, cmp, 0, total_elems-1);
}


The choice of pivot is made by the external method selectPivotIndex(ar, left, right) , which provides the array element for which to partition.


Example 4-9.  Heap Sort implementation in C
static void heapify (void **ar, int(*cmp)(const void *,const void *), int idx, int max) {
  int left = 2*idx + 1;
  int right = 2*idx + 2;
  int largest;
  /* Find largest element of A[idx], A[left], and A[right]. *
  if (left < max && cmp (ar[left], ar[idx]) > 0) {
    largest = left;
  } else {
    largest = idx;
  }
  if (right < max && cmp(ar[right], ar[largest]) > 0) {
    largest = right;
  }
  /* If largest is not already the parent then swap and propagate. */
  if (largest != idx) {
     void *tmp;
     tmp = ar[idx];
     ar[idx] = ar[largest];
     ar[largest] = tmp;
     heapify(ar, cmp, largest, max);
   }
}
static void buildHeap (void **ar, int(*cmp)(const void *,const void *), int n) {
  int i;
  for (i = n/2-1; i>=0; i--) {
    heapify (ar, cmp, i, n);
  }
}
void sortPointers (void **ar, int n, int(*cmp)(const void *,const void *)) {
  int i;
  buildHeap (ar, cmp, n);
  for (i = n-1; i >= 1; i--) {
   void *tmp;
   tmp = ar[0];
   ar[0] = ar[i];
   ar[i] = tmp;
   heapify (ar, cmp, 0, i);
  }
}
HEAP SORT succeeds because of the heapify function.  It is used in two distinct
places, although it serves the same purpose each time.



Criteria for Choosing a Sorting Algorithm

To choose a sorting algorithm, consider the qualitative criteria in Table 4-6.  These
may help your initial decision, but you likely will need more quantitative measures to guide you.

Table 4-6.  Criteria for choosing a sorting algorithm

Criteria                        Sorting algorithm
--------                        -----------------
Only a few items                INSERTION SORT
Items are mostly sorted already INSERTION SORT
Concerned about worst-case
scenarios                       HEAP SORT
Interested in a good 
average-case result             QUICKSORT
Items are drawn from a dense 
universe                        BUCKET SORT
Desire to write as little code 
as possible                     INSERTION SORT


log-linear is the best sort for sorts that depend on comparisons.  Some variants of a hash or radix sort do not depend on comparisons and can perform better if dealing with numbers.


Extracted from GoLang.txt, possibly from "A Way To Go: A thorough Intro"
start := time.Now()
longCalculation()
end := time.Now()
delta := end.Sub(start)
fmt.Printf("longCalculation took this amount of time: %s\n", delta) 


t0 := time.Now()
  longCalculation()
t1 := time.Now()
delta := t1.Sub(t0)
fmt.Printf("longCalculation took this amount of time: %s\n", delta) 


From vlc.go

  timeToShuffle := time.Since(t0);  // timeToShuffle is a Duration type, which is an int64 but has methods.
  timeToShuffleString := timeToShuffle.String();
  fmt.Println(" It took ",timeToShuffleString," to shuffle this file.");


 func Since

func Since(t Time) Duration

Since returns the time elapsed since t. It is shorthand for time.Now().Sub(t). 
--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
:%s/\%x92/'/g
:%s/\%x93/"/g
:%s/\%x94/"/g
:%s/\%x97/ -- /g
:%s/\%x96/--/g
:%s/\%x95/--/g
--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
