package main

/*
  REVISION HISTORY
  ================
  19 Aug 16 -- First Go version completed to test all parts of tokenize.go package, and is called testtoken.go.
  21 Sep 16 -- Now need to test my new GetTknStrPreserveCase routine.  And test the change I made to GETCHR.
  31 Aug 17 -- Now this is testinput.go, to test using fmt.Scan and fmt.Scanln to do unbuffered input
               from Stdin.  fmt.Scanln() didn't do what I want.  It only read 1 word in.
   8 May 24 -- Fixed import list so it could compile under modules.
*/

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"src/getcommandline"
	"src/tokenize"
	"strings"
)

// var FSAnameType = [...]string{"DELIM","OP","DGT","ALLELSE"};

func main() {
	// testingstate = 0: gettkn, 1: gettknreal, 2: gettknstr, 3: gettkneol, 4: string lower case,
	// 5: token lower case
	testingstate := 0
	commandline := getcommandline.GetCommandLineString()
	commandline = strings.ToUpper(commandline)
	floatflag := false
	if (commandline == "REAL") || (commandline == "FLOAT") {
		floatflag = true
		testingstate = 1
	} else if commandline == "STRING" {
		testingstate = 2
	} else if commandline == "EOL" {
		testingstate = 3
	} else if commandline == "STR" {
		testingstate = 4
	} else if commandline == "LOWER" {
		testingstate = 5
	}

	fmt.Print(" Will be testing GetTknReal? ", floatflag, ", testingstate is ", testingstate)
	fmt.Println()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(" Input test text: ")
		scanner.Scan()
		inputline := scanner.Text()
		//		n, err := fmt.Scanf("%s \n", &inputline)
		//		fmt.Println(n, "bytes read from Stdin")
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
		if len(inputline) == 0 {
			break
		}
		fmt.Println(" After the call to scanner.Text(), before TrimSpace: ", inputline, ".")
		inputline = strings.TrimSpace(inputline)
		fmt.Println(" After call to TrimSpace: ", inputline)
		if strings.ToUpper(inputline) == "QUIT" {
			log.Println(" Test Token finished.")
			os.Exit(0)
		}
		tokenize.INITKN(inputline)
		EOL := false
		token := tokenize.TokenType{}
		for !EOL {
			if floatflag || testingstate == 1 {
				token, EOL = tokenize.GETTKNREAL()
			} else if testingstate == 2 {
				token, EOL = tokenize.GETTKNSTR()
			} else if testingstate == 3 {
				token, EOL = tokenize.GETTKNEOL()
			} else if testingstate == 4 {
				token, EOL = tokenize.GetTokenString(false)
			} else if testingstate == 5 {
				token, EOL = tokenize.GetToken(false)
			} else {
				token, EOL = tokenize.GETTKN()
			}

			fmt.Printf(" Token : %#v \n", token)
			//      if floatflag {     I think this is just an error.  I missed something in testing
			//        fmt.Println(" R = ",token.Rsum);
			//      }
			fmt.Println(" EOL : ", EOL)
			if EOL {
				break
			} // I don't want it to ask about ungettkn if there is an EOL condition.
			fmt.Print(" call UnGetTkn? (Y/N) ")
			ans := ""
			n, err := fmt.Scan(&ans)
			fmt.Println()
			fmt.Println(n, "bytes read from Stdin.  ans =", ans)
			if err != nil {
				fmt.Fprintln(os.Stderr, "reading standard input:", err)
				os.Exit(1)
			}
			ans = strings.TrimSpace(ans)
			ans = strings.ToUpper(ans)
			if strings.HasPrefix(ans, "Y") {
				tokenize.UNGETTKN()
			}
		}
		fmt.Println()
		log.Println(" Finished processing the inputline.")
	}

	stuff := ""
	for {
		fmt.Print("Input for Scan: ")
		n, err := fmt.Scan(&stuff)
		if err != nil {
			fmt.Println(" Error from fmt.Scan", err)
			os.Exit(1)
		}
		fmt.Println("n =", n, ".  stuff =", stuff)
		if n == 0 || stuff == "Q" {
			break
		}
	}
}
