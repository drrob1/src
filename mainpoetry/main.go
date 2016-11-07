// Packages and Initialization video code.  Now with weighted shuffle

package main;

import (
	"fmt"
	"time"
	"poetry"
//	"math/rand"
//	"shuffler"
)

func main() {
	t := time.Now()
	h, m, s := t.Clock()
	ns := t.Nanosecond()

	fmt.Println("Time from time.Now() is", t)
	fmt.Println(h, ":", m, ":", s, " nanosec ", ns)


//	p0 := poetry.NewPoem();
	p1 := poetry.Poem{{"the quick","brown fox","jumps over","the lazy dog"}};  // create poem directly without use of an empty exported function.

	v,c := p1.Stats();
	fmt.Println(" vowels:",v,", consonants:",c);

	fmt.Println(" stanzas:",p1.GetNumStanzas(),", lines:",p1.GetNumLines());
	fmt.Println();
}
