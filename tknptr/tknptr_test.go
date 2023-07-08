package tknptr

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

/*
   7 July 23 -- I'm going to try and code another set of table based testing functions
*/

/*
const (DELIM, OP, DGT, ALLELSE)

	type TokenType struct {
		Str        string
		FullString string // includes minus sign character, if present.
		State      int
		DelimCH    byte
		DelimState int
		Isum       int
		Rsum       float64
		RealFlag   bool // flag so integer processing stops when it sees a dot, E or e.
	}
*/
var testRealStrings = []struct {
	inputString string
	outputToken TokenType
}{
	{"fix", TokenType{"FIX", "FIX", ALLELSE, '0', DELIM, 0, 0, false}},
	{"+", TokenType{
		Str:        "+",
		FullString: "+",
		State:      OP,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       8,
		Rsum:       0,
		RealFlag:   false,
	}},
	{"567", TokenType{
		Str:        "567",
		FullString: "567",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       567,
		Rsum:       0,
		RealFlag:   false,
	}},
	{"-782", TokenType{
		Str:        "782",
		FullString: "-782",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       -782,
		Rsum:       0,
		RealFlag:   false,
	}},
	{"7", TokenType{
		Str:        "7",
		FullString: "7",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       7,
		Rsum:       0,
		RealFlag:   false,
	}},
	{"-8", TokenType{
		Str:        "8",
		FullString: "-8",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       -8,
		Rsum:       0,
		RealFlag:   false,
	}},
	{"7.6", TokenType{"7.6", "7.6", DGT, 0, DELIM, 7, 7.6, true}},
	{"-3.14159", TokenType{
		Str:        "3.14159",
		FullString: "-3.14159",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       -3,
		Rsum:       -3.14159,
		RealFlag:   true,
	}},
	{"8e5", TokenType{
		Str:        "8E5",
		FullString: "8E5",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       8,
		Rsum:       800000,
		RealFlag:   true,
	}},
	{"7.68e4", TokenType{
		Str:        "7.68E4",
		FullString: "7.68E4",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       7,
		Rsum:       76000,
		RealFlag:   true,
	}},
	{"-3.14e2", TokenType{
		Str:        "3.14E2",
		FullString: "-3.14E2",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       -3,
		Rsum:       -314,
		RealFlag:   true,
	}},
	{"8.623e-1", TokenType{ // neg exponent is an issue.  Fixed by allowing '-' to follow 'E' or 'e', and also '_' substituted for '-' before conversion to float64.
		Str:        "8.623e-1",
		FullString: "",
		State:      0,
		DelimCH:    0,
		DelimState: 0,
		Isum:       0,
		Rsum:       0,
		RealFlag:   false,
	}},
	{"-23.456e-2", TokenType{
		Str:        "",
		FullString: "",
		State:      0,
		DelimCH:    0,
		DelimState: 0,
		Isum:       0,
		Rsum:       0,
		RealFlag:   false,
	}},
}

func TestMain(m *testing.M) { // this example is in the docs of the testing package, that I was referred to by the golang nuts google group.
	flag.Parse()
	os.Exit(m.Run())
}

func TestTokenReal(t *testing.T) { // test ALLELSE FIX keyword, OP of +, and then test the numbers.  DelimCH not important, DelimState not important
	for _, tkn := range testRealStrings {
		bs := New(tkn.inputString)
		token, EOL := bs.TokenReal()
		if EOL {
			t.Errorf(" EOL should have been false, but it's true.  TestString=%q, token=%+v\n", tkn.inputString, token)
		}

		fmt.Printf(" InputString = %q, token is %+v\n", tkn.inputString, token)
		if tkn.outputToken.State == ALLELSE && token.Str != tkn.outputToken.Str {
			t.Errorf(" inputString is %q, token.Str is %q which was not expected.  token.FullString = %q, Isum = %d, Rsum=%g\n", tkn.inputString, token.Str, token.FullString,
				token.Isum, token.Rsum)
		}
		if tkn.outputToken.State == OP && token.Str != tkn.outputToken.Str {
			t.Errorf(" inputstring=%q, token.Str=%q which was not expected.  Token=%+v\n", tkn.inputString, token.Str, token)
		}

		if tkn.outputToken.State == DGT && token.FullString != tkn.outputToken.FullString {
			t.Errorf(" inputstring=%q, token.FullString=%q which was not expected.  Token=%+v\n", tkn.inputString, token.FullString, token)
		}
	}
}
