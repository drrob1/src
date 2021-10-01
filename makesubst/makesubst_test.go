package makesubst

/*
I'm going to take notes about this testing stuff.  The Println and Log output was only seen when I used
  go test makesubst -v

from the GOPATH ~/gocode dir.  Without the -v switch, I saw no output other than the OK or pass output.  But if it fails, I would not need the -v flag to see the output.

 1 Oct 21 -- I changed the use of fmt.Println to fmt.Printf and t.log to t.logf to have more control over the output.
               If any of the t.Fail() are executed, then all of the output statements become visible without using the -v flag.
*/
import (
	"fmt"
	"testing"
)

func TestMakeSubst(t *testing.T) {
	s1 := MakeSubst("=")
	fmt.Printf("fmt.Printf -- s1 should be +, and it is %q \n", s1)
	t.Logf(" s1 should be + and it is %q \n", s1)
	if s1 != "+" {
		t.Fail()
	}
	s2 := MakeSubst(";")
	t.Logf(" s2 should be *, and it is %q \n", s2)
	if s2 != "*" {
		t.Fail()
	}
	s1 = MakeSubst("")
	if s1 != "" {
		t.Fail()
	}
	s3 := MakeSubst("this is a =1234 string; that I am testing.")
	fmt.Printf("fmt.Printf -- s3 is %q \n", s3)
	t.Logf("t.Log -- s3 is %q \n", s3)
	if s3 != "this is a +1234 string* that I am testing." {
		t.Fail()
	}

	// Added 1 Oct 21
	s4 := MakeReplaced("=")
	fmt.Printf("fmt.Printf -- s4 replaced should be +, and it is %q \n", s4)
	t.Logf(" s4 replaced should be + and it is %q \n", s4)
	if s4 != "+" {
		t.Fail()
	}
	s5 := MakeReplaced(";")
	t.Logf(" s5 replaced should be *, and it is %q \n", s5)
	if s5 != "*" {
		t.Fail()
	}
	s6 := MakeReplaced("")
	if s6 != "" {
		t.Fail()
	}
	s7 := MakeReplaced("this is a =1234 string; that I am testing.")
	fmt.Printf("fmt.Printf -- s7 replaced is %q \n", s7)
	t.Logf("t.Log -- s7 replaced is %q \n", s7)
	if s7 != "this is a +1234 string* that I am testing." {
		t.Fail()
	}

}
