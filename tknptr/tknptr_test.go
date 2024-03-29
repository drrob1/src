package tknptr

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

/*
   7 July 23 -- I'm going to try and code another set of table based testing functions
  20 July 23 -- Adjusted the test cases in the table because I changed the logic.  I want to keep '-' as an operator.  To enter neg exponent need '_'.
  24 July 23 -- Added test for hex input.
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
	        HexFlag    bool  added 7/9/23
		}
*/
var fsaNameType = [...]string{"DELIM", "OP", "DGT", "ALLELSE"}
var testRealStrings = []struct {
	inputString string
	outputToken TokenType
}{
	{"fix", TokenType{"FIX", "FIX", ALLELSE, '0', DELIM, 0, 0, false, false}},
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
	{"-", TokenType{
		Str:        "-",
		FullString: "-",
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
	{"7.6", TokenType{"7.6", "7.6", DGT, 0, DELIM, 7, 7.6, true, false}},
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
	{"8.623e_1", TokenType{ // neg exponent is an issue.  Fixed by allowing '-' to follow 'E' or 'e', and also '_' substituted for '-' before conversion to float64.
		Str:        "8.623E-1",
		FullString: "8.623E-1",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       0,
		Rsum:       .8623,
		RealFlag:   true,
	}},
	{"-23.456e_2", TokenType{
		Str:        "23.456E-2",
		FullString: "-23.456E-2",
		State:      DGT,
		DelimCH:    0,
		DelimState: ALLELSE,
		Isum:       0,
		Rsum:       -.23456,
		RealFlag:   true,
	}},
	{"-23.456e_2", TokenType{
		Str:        "23.456E-2",
		FullString: "-23.456E-2",
		State:      DGT,
		DelimCH:    0,
		DelimState: ALLELSE,
		Isum:       0,
		Rsum:       -.23456,
		RealFlag:   true,
	}},
	{"8.623e_10", TokenType{ // neg exponent is an issue.  Fixed by allowing '-' to follow 'E' or 'e', and also '_' substituted for '-' before conversion to float64.
		Str:        "8.623E-10",
		FullString: "8.623E-10",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       0,
		Rsum:       8.623e-10,
		RealFlag:   true,
	}},
	{"8.623e_01", TokenType{ // neg exponent is an issue.  Fixed by allowing '-' to follow 'E' or 'e', and also '_' substituted for '-' before conversion to float64.
		Str:        "8.623E-01",
		FullString: "8.623E-01",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       0,
		Rsum:       .8623,
		RealFlag:   true,
	}},
	{"0x10", TokenType{
		Str:        "010",
		FullString: "010",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       16,
		Rsum:       16,
		RealFlag:   false,
	}},
	{".5", TokenType{
		Str:        ".5",
		FullString: ".5",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       0,
		Rsum:       0.5,
		RealFlag:   true,
	}},
	{"-.5", TokenType{
		Str:        ".5",
		FullString: "-.5",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       0,
		Rsum:       -0.5,
		RealFlag:   true,
	}},
	{"0.5", TokenType{
		Str:        "0.5",
		FullString: "0.5",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       0,
		Rsum:       0.5,
		RealFlag:   true,
	}},
	{"-0.5", TokenType{
		Str:        "0.5",
		FullString: "-0.5",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       0,
		Rsum:       -0.5,
		RealFlag:   true,
	}},
	{"0xa", TokenType{
		Str:        "0A",
		FullString: "0A",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       10,
		Rsum:       10,
		RealFlag:   true, // this flag affects how Isum is calculated, which is fine.
		HexFlag:    true,
	}},
	{"0xA", TokenType{
		Str:        "0A",
		FullString: "0A",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       10,
		Rsum:       10,
		RealFlag:   true, // this flag affects how Isum is calculated, which is fine.
		HexFlag:    true,
	}},
	{"0XA", TokenType{
		Str:        "0A",
		FullString: "0A",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       10,
		Rsum:       10,
		RealFlag:   true, // this flag affects how Isum is calculated, which is fine.
		HexFlag:    true,
	}},
	{"0xf", TokenType{
		Str:        "0F",
		FullString: "0F",
		State:      DGT,
		DelimCH:    0,
		DelimState: DELIM,
		Isum:       15,
		Rsum:       15,
		RealFlag:   true, // this flag affects how Isum is calculated, which is fine.
		HexFlag:    true,
	}},
}

func FSAname(i int) string { // no param checking is done.  This code will panic if the int param is out of bounds, ie, not in the correct range of 0..3
	//var fsaNameType = [...]string{"DELIM", "OP", "DGT", "ALLELSE"}  now made global
	return fsaNameType[i]
}

func TestMain(m *testing.M) { // this example is in the docs of the testing package, that I was referred to by the golang nuts google group.
	flag.Parse()
	os.Exit(m.Run())
}

func TestTokenReal(t *testing.T) { // test ALLELSE FIX keyword, OP of +, and then test the numbers.  DelimCH not important, DelimState not important

	fmt.Printf(" FSA Names are %+v\n", fsaNameType)

	for _, tkn := range testRealStrings {
		bs := New(tkn.inputString)
		token, EOL := bs.TokenReal()
		if EOL {
			t.Errorf(" EOL should have been false, but it's true.  TestString=%q, token=%+v\n", tkn.inputString, token)
		}

		fmt.Printf(" InputString = %q, token is %+v\n", tkn.inputString, token)
		if tkn.outputToken.State != token.State || token.Str != tkn.outputToken.Str || tkn.outputToken.FullString != token.FullString || tkn.outputToken.RealFlag != token.RealFlag {
			t.Errorf(" Error: inputString is %q, token.Str is %q, token.FullString = %q, token.State = %s, Rsum=%g\n", tkn.inputString, token.Str, token.FullString,
				FSAname(token.State), token.Rsum)
		}

		token, EOL = bs.TokenReal()
		if !EOL {
			t.Errorf(" EOL should be true.  It's not.  Look into this.  TestString=%q\n", tkn.inputString)
		}
	}
}
