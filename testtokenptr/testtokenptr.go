package main

import (
	"bufio"
	"flag"
	"fmt"

	"log"
	"os"
	"strings"

	"src/tknptr" // converted to their module system June 6, 2021.
)

const LastAltered = "19 July 2023"

/*
REVISION HISTORY
================
19 Aug 16 -- First Go version completed to test all parts of tokenize.go package
21 Sep 16 -- Now need to test my new GetTknStrPreserveCase routine.  And test the change I made to GETCHR.
 7 Oct 16 -- Changed the scanner to scan by words.  I hope.  This is to test the scanner for rpng.  Default is scan by lines.
11 Aug 17 -- Now named testtokenptr, and will test using pointer receivers and scanning whole lines.
13 Oct 17 -- Testing the inclusion of horizontal tab as a delim, needed for comparehashes.
30 Jan 18 -- Will use flags to set the mode now.
28 Sep 20 -- Testing new use code of tknptr, in which the StateMap is part of the pointer structure that is passed around.
 6 Jun 21 -- Testing GetTokenSlice
12 Jun 21 -- Testing TokenSlice and TokenRealSlice
 7 Jul 23 -- Testing TokenReal, and making sure I didn't break anything.  But I may not finish this today.
19 Jul 23 -- Didn't do it until today.  I debugged the TokenReal() routine using unit table driven testing.  It's not working here correctly w/ neg numbers. To be continued.
               The unit table based tests worked.  And the code is working in rpn and rpn2.  So something is wrong here.
20 Jul 23 -- All seems to be working now.  I changed how negative exponents are entered.  I almost never do this so it doesn't really matter much.
               Of note, entering negative exponents is different btwn TokenReal() and GETTKNREAL().  The old GETTKNREAL() doesn't treat '_' as a negation operator.
*/

// var FSAnameType = [...]string{"DELIM","OP","DGT","ALLELSE"};

func main() {
	//	                                                            commandline := getcommandline.GetCommandLineString()
	// 																commandline = strings.ToUpper(commandline)
	var floatflag = flag.Bool("f", false, "call GetTknReal")  // pointer syntax
	var realFlag = flag.Bool("r", false, "call TokenReal")    // pointer syntax
	var noopflag = flag.Bool("noop", false, "Set No OpCodes") // pointer syntax
	var strflag = flag.Bool("s", false, "Call GetTknStr")     // pointer syntax
	var Strflag bool
	flag.BoolVar(&Strflag, "S", false, "Call GetTknStr")                // value syntax
	var eolflag = flag.Bool("e", false, "Call GetTknEOL")               // pointer syntax
	var lowerflag = flag.Bool("l", false, "Call using lower case form") // pointer syntax
	var helpflag = flag.Bool("h", false, "help")                        // pointer syntax
	var mapflag = flag.Bool("m", false, "test setmapdelim routine")

	flag.Parse()

	if *helpflag {
		flag.PrintDefaults()
		os.Exit(0)
	}

	// testingstate = 0: gettkn, 1: gettknreal, 2: gettknstr, 3: gettkneol, 4: string lower case, 5: token lower case, 6: tokenreal
	testingstate := 0
	if *floatflag {
		testingstate = 1
	} else if *realFlag {
		testingstate = 6
	} else if *lowerflag && (*strflag || Strflag) {
		testingstate = 4
	} else if *strflag || Strflag {
		testingstate = 2
	} else if *eolflag {
		testingstate = 3
	} else if *lowerflag {
		testingstate = 5
	}

	fmt.Print(" Test Token Ptr last altered ", LastAltered)
	fmt.Print(",  floatflag is ", *floatflag, ", testingstate is ", testingstate, ", mapflag is ", *mapflag)
	fmt.Println()
	// var FSAnameType = [...]string{"DELIM","OP","DGT","ALLELSE"};
	fmt.Printf(" FSAnameType = [...]string{DELIM, OP, DGT, ALLELSE}\n")

	scanner := bufio.NewScanner(os.Stdin)
	inputline := ""
	//  scanner.Split(bufio.ScanWords); // testing scanning by words to see what happens.
	for {
		fmt.Print(" Input test text: ")
		scanner.Scan()
		inputline = scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
		if len(inputline) == 0 {
			os.Exit(0)
		}
		//		fmt.Println(" After the call to scanner.Text(), before TrimSpace: ", inputline, ".")
		//		inputline = strings.TrimSpace(inputline)
		fmt.Println(" After call to TrimSpace: ", inputline)
		if strings.ToUpper(inputline) == "QUIT" {
			log.Println(" Test Token finished.")
			os.Exit(0)
		}
		tokenbuffer := tknptr.New(inputline)
		if *mapflag || *noopflag {
			tokenbuffer.SetMapDelim('#')
			tokenbuffer.SetMapDelim('*')
			tokenbuffer.SetMapDelim('+')
			tokenbuffer.SetMapDelim('-')
			tokenbuffer.SetMapDelim('=')
			tokenbuffer.SetMapDelim('/')
			tokenbuffer.SetMapDelim('<')
			tokenbuffer.SetMapDelim('>')
			tokenbuffer.SetMapDelim('%')
			tokenbuffer.SetMapDelim('^')
			*mapflag = false
		}
		//fmt.Printf(" After tokenbuffer is initialized.")

		tknslice := tknptr.TokenSlice(inputline)
		fmt.Printf(" Length of token slice is %d, tknslice=%+v\n", len(tknslice), tknslice)
		fmt.Println()

		realtknslice := tknptr.RealTokenSlice(inputline)
		fmt.Printf(" old real token slice length=%d, slice itself is: %+v\n", len(realtknslice), realtknslice)
		for _, t := range realtknslice {
			fmt.Printf(" tknreal str=%q, fullstr=%q, state=%s, delimCH=%q, delimState=%q, iSum=%d, rSum=%g, realFlag=%t, hexflag=%t\n",
				t.Str, t.FullString, FSAname(t.State), t.DelimCH, FSAname(t.DelimState), t.Isum, t.Rsum, t.RealFlag, t.HexFlag)
		}
		fmt.Println()
		tknrealslice := tknptr.TokenRealSlice(inputline)
		fmt.Printf(" Now testing the new TokenReal, len=%d, and the slice is: %+v\n", len(tknrealslice), tknrealslice)
		for _, t := range tknrealslice {
			fmt.Printf(" tknreal str=%q, fullstr=%q, state=%s, delimCH=%q, delimState=%q, iSum=%d, rSum=%g, realFlag=%t, hexflag=%t\n",
				t.Str, t.FullString, FSAname(t.State), t.DelimCH, FSAname(t.DelimState), t.Isum, t.Rsum, t.RealFlag, t.HexFlag)
		}
		fmt.Println()
		fmt.Println()
		// There's a bug here when multiple tokens are on the same line.  This is not the case so far w/ rpn tests.  The letter 'e' is not changed back to ALLELSE as it needs to be.gg

		var EOL bool
		var token tknptr.TokenType
		for !EOL {
			if *floatflag || testingstate == 1 {
				token, EOL = tokenbuffer.GETTKNREAL()
			} else if testingstate == 2 {
				token, EOL = tokenbuffer.GETTKNSTR()
			} else if testingstate == 3 {
				token, EOL = tokenbuffer.GETTKNEOL()
			} else if testingstate == 4 {
				token, EOL = tokenbuffer.GetTokenString(false)
			} else if testingstate == 5 {
				token, EOL = tokenbuffer.GetToken(false)
			} else if testingstate == 6 {
				token, EOL = tokenbuffer.TokenReal()
			} else { // testingstate .EQ. 0
				token, EOL = tokenbuffer.GETTKN()
			}

			fmt.Printf(" Token : %#v \n", token)
			fmt.Println(" EOL : ", EOL)
			if EOL {
				break
			} // I don't want it to ask about ungettkn if there is an EOL condition.
			fmt.Print(" call UnGetTkn? (Y/N) ")
			scanner.Scan()
			ans := scanner.Text()
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "reading standard input:", err)
				os.Exit(1)
			}
			ans = strings.TrimSpace(ans)
			ans = strings.ToUpper(ans)
			if strings.HasPrefix(ans, "Y") {
				tokenbuffer.UNGETTKN()
			} else if ans == "X" {
				break
			}
		}
		fmt.Println()
		log.Println(" Finished processing the inputline.")
		fmt.Println()
	}
}

func FSAname(i int) string {

	var fsaNameType = [...]string{"DELIM", "OP", "DGT", "ALLELSE"}
	return fsaNameType[i]
}

/*  from the web documentation at golang.org
        scanner := bufio.NewScanner(os.Stdin)
        for scanner.Scan() {
          fmt.Println(scanner.Text()) // Println will add back the final '\n'
	}
        if err := scanner.Err(); err != nil {
          fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
*/
