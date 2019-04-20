// types, empty interfaces, type assertions and type switches example.  He says this is a dangerous
// type, the empty interface type.  Avoid it.  Or do a type assertion and turn it into a safe type.  If
// the type assertion is wrong, like asserting an int is a string, this will crash, likely as a run-time
// panic error.

package main

import (
	"fmt"
	//	"net/http"
	//	"io/ioutil"
	//	"time"
	//	"math/rand"
	//	"sync/atomic"
)

func whatIsThis(i interface{}) {
	//	fmt.Printf("%T\n",i);

	switch i.(type) {
	case string:
		fmt.Println("It's a string", i.(string)) // this asserts the type
	case uint32:
		fmt.Println("It's a uint32", i.(uint32)) // this asserts the type
	case int:
		I := i.(int)
		fmt.Println(" It's an int", I)
	default:
		fmt.Println("IDK")
	}

	switch v := i.(type) {
	case string:
		fmt.Println("It's a string", v)
	case uint32:
		fmt.Println("It's a uint32", v)
	default:
		fmt.Printf("IDK %v\n", v)
	}
}

func main() {
	whatIsThis(42)

	whatIsThis(uint32(42))

	whatIsThis("42")

	whatIsThis(42.5)

	whatIsThis([...]string{"a", "b", "c"})

	whatIsThis([]string{"A", "B", "C"})

	fmt.Println()
}
