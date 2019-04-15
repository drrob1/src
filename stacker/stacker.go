package main

import (
	"errors"
	"fmt"
)

func main() {
	var haystack Stack
	haystack.Push("hey") // this func needs a pointer, but I do not have to call it with the AdrOf op
	haystack.Push(-15)   // here is where the compiler is more forgiving.
	haystack.Push([]string{"one", "two", "three"})
	haystack.Push(83.21)

	for {
		item, err := haystack.Pop()
		if err != nil {
			break
		}
		fmt.Println(item)
	}
}

type Stack []interface{}

//                                 Len()
func (stack Stack) Len() int {
	return len(stack)
}

//                                 Cap()
func (stack Stack) Cap() int {
	return cap(stack)
}

//                                 IsEmpty()
func (stack Stack) IsEmpty() bool {
	if len(stack) == 0 {
		return true
	} else {
		return false
	}
}

//                                 Push
func (stack *Stack) Push(x interface{}) {
	*stack = append(*stack, x) // the book has these dereferenced.  I found that deref must be explicit.
}

func (stack Stack) Top() (interface{}, error) {
	if len(stack) == 0 {
		return nil, errors.New("Cannot Top() an empty stack.")
	}
	return stack[len(stack)-1], nil
}

func (stack *Stack) Pop() (interface{}, error) {
	thestack := *stack // the code did not work until I added this from the book.  Deref op is important here.
	if len(thestack) == 0 {
		return nil, errors.New("Cannot Pop() an empty stack.")
	}
	x := thestack[len(thestack)-1]
	*stack = thestack[:len(thestack)-1] // Deref op is important.  thestack value slice created so don't need to deref everywhere
	return x, nil
}
