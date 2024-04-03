package stack

import (
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"testing"
)

/*
to be tested with Windows 11 on 4/3/24
go test -bench=.
Results:
         Generic1: ~35 ns/op
         intStack:  ~2.9 ns/op
        HiLoStack:  ~5.3 ns/op


Adding Generic2 changes results
         Generic1: ~35.5 ns/op
         Generic2: ~32 ns/op
         intStack:  ~3.2 ns/op
        HiLoStack:  ~5 ns/op

Adding linked list type of stack
         Generic1: ~35.5 ns/op
         Generic2: ~32 ns/op
      linked list: ~ same as Generic2
         intStack:  ~3.2 ns/op
        HiLoStack:  ~5 ns/op


*/

import (
	ct "github.com/daviddengcn/go-colortext"
)

func BenchmarkGeneric1(b *testing.B) {
	iStack := New[int]()
	for i := range b.N {
		iStack.Push(b.N - i - 1)
	}
	for i := range b.N {
		a, err := iStack.Pop()
		if err != nil {
			ctfmt.Printf(ct.Red, true, " Pop of %d generated error %s\n", i, err)
		}
		if a != i {
			ctfmt.Printf(ct.Red, true, "i = %d, a = %d\n", i, a)
		}
	}
}

func BenchmarkGeneric2(b *testing.B) {
	iStack := New[hiloIndexType]()
	for i := range b.N {
		iStack.Push(hiloIndexType{b.N - i - 1, b.N - i - 1})
	}
	for i := range b.N {
		a, err := iStack.Pop()
		if err != nil {
			ctfmt.Printf(ct.Red, true, " Pop of %d generated error %s\n", i, err)
		}
		if a.lo != i {
			ctfmt.Printf(ct.Red, true, "i = %d, a = %d\n", i, a)
		}
	}
}

func BenchmarkLinkedList(b *testing.B) {
	list := MakeLinkedList()
	for i := range b.N {
		list.Push(b.N - i - 1)
	}
	for i := range b.N {
		a := list.Pop()
		if a != i {
			ctfmt.Printf(ct.Red, false, "i = %d, a = %d\n", i, a)
		}
	}
}

func BenchmarkIntStack(b *testing.B) {
	intStackInit(b.N)
	for i := range b.N {
		intStackPush(b.N - i - 1)
	}
	for i := range b.N {
		a := intStackPop()
		if a != i {
			ctfmt.Printf(ct.Red, false, "i = %d, a = %d \n", i, a)
		}
	}
}

func BenchmarkHiLoStack(b *testing.B) {
	hiloInit(b.N)
	for i := range b.N {
		hiloStackPush(hiloIndexType{b.N - i - 1, b.N - i - 1})
	}
	for i := range b.N {
		a := hiloStackPop()
		if i != a.hi {
			ctfmt.Printf(ct.Red, false, "i = %d, a.hi = %d\n", i, a.hi)
		}
	}
}
