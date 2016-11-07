// Packages and Initialization video code.  Now with weighted shuffle
// Now will add ability to count words by splitting strings on white space, using the strings library
// package

package poetry

import (
	"bufio"
	"os"
	"strings"
//	"io"
//	"fmt"
//	"time"
//	"math/rand"
//	"shuffler"
)

type Line string;
type Stanza []Line;
type Poem []Stanza;

func (s Stanza) Len() int {  // going to make a stanza sortable, just for fun
	return len(s);
}

func (s Stanza) Swap(i,j int) {
	s[i],s[j] = s[j],s[i];
}

func (s Stanza) Less(i,j int) bool {
	return len(s[i]) < len(s[j]);
}

func (p Poem) GetNumStanzas() int {
	return len(p);
}

func (s Stanza) GetNumLines() int {
	return len(s);
}

func (p Poem) GetNumLines() (count int) {
	for _, s := range p {
	  count += s.GetNumLines();
	}
	return;
}

func (p Poem) GetNumWords() int {
	count := 0;
	for _, s := range p {
	  for _, l := range s {
	    sl := string(l);
	    parts := strings.Split(sl," "); // split string line on white space
	    count += len(parts);
	  }
	}
	return count;
}

func (p Poem) GetNumThe() int {
	count := 0;
	for _, s := range p {
	  for _, l := range s {
	    sl := string(l);
	    if strings.Contains(sl,"The") {   // counts lines w/ 'The", not occurrances of this word.
	      count++;
	    }
	  }
	}
	return count;
}

// Unicode package is a better solution here.  It has IsPunct, IsDigit, and other similar functions
// which would be a better way to count
func (p Poem) Stats() (numVowels, numConsonants, numPuncts int) { // numVowels and numConstant are output params
	for _, s := range p {
	  for _, l := range s {
	    for _, r := range l {
		switch r {
		  case 'a','e','i','o','u' : 
		    numVowels++;
		  case ',',' ','!','?' :
		    numPuncts++;
		  default:
		    numConsonants++;
		}
	    }
	  }
	}
	return;

}

func NewPoem() Poem {
	return Poem{};
}


func (p Poem) String() string {   // note that it has the same signature as needed for the interface.
  result := "";
  for _,s := range p { // iterate over stanzas
    for _,l := range s {  // iterate over lines
      result += string(l);        // then he decided to use fmt.Sprintf("%s\n",l); 
      result += "\n";     // not needed here if use Sprintf
    }
    result += "\n";     // after each stanza, will print an extra newline
  }
  return result;
}

func LoadPoem(filename string) (Poem, error) {
	f,err := os.Open(filename);
	if err != nil {
	  return nil, err;
	}
	defer f.Close();
	p := Poem{};
	var s Stanza;

	// read a poem line by line
	scan := bufio.NewScanner(f);
	for scan.Scan() {      // while true loop
	  l := scan.Text();  // grab the line
	  if l == "" {   // read a blank line in
	    p = append(p,s);
	    s = Stanza{};
	    continue;
	  }
	  s = append(s,Line(l));
	}

	if scan.Err() != nil {
	  return nil, scan.Err();
	}
	p = append(p,s);
	return p,nil;
}


/*
func main() {
	t := time.Now();
	h, m, s := t.Clock();
	ns := t.Nanosecond();

	fmt.Println("Time from time.Now() is", t);
	fmt.Println(h, ":", m, ":", s, " nanosec ", ns);






	fmt.Println()
}

*/
