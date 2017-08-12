package main

/*
  REVISION HISTORY
  ================
  19 Aug 16 -- First Go version completed to test all parts of tokenize.go package
  21 Sep 16 -- Now need to test my new GetTknStrPreserveCase routine.  And test the change I made to GETCHR.
   7 Oct 16 -- Changed the scanner to scan by words.  I hope.  This is to test the scanner for rpng.  Default is scan by lines.
  11 Aug 17 -- Now named testtokenptr, and will test using pointer receivers and scanning whole lines.
*/

import (
	"bufio"
	"fmt"
	"getcommandline"
	"log"
	"os"
	"strings"
	"tknptr"
)

// var FSAnameType = [...]string{"DELIM","OP","DGT","ALLELSE"};

func main() {
	// testingstate = 0: gettkn, 1: gettknreal, 2: gettknstr, 3: gettkneol, 4: string lower case,
	//                5: token lower case
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

	//  scanner.Split(bufio.ScanWords); // testing scanning by words to see what happens.
	for {
		fmt.Print(" Input test text: ")
		scanner.Scan()
		inputline := scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
		if len(inputline) == 0 {
			os.Exit(0)
		}
		fmt.Println(" After the call to scanner.Text(), before TrimSpace: ", inputline, ".")
		inputline = strings.TrimSpace(inputline)
		fmt.Println(" After call to TrimSpace: ", inputline)
		if strings.ToUpper(inputline) == "QUIT" {
			log.Println(" Test Token finished.")
			os.Exit(0)
		}
		tokenbuffer := tknptr.NewToken(inputline)
		EOL := false
		token := tknptr.TokenType{}
		for !EOL {
			if floatflag || testingstate == 1 {
				token, EOL = tokenbuffer.GETTKNREAL()
			} else if testingstate == 2 {
				token, EOL = tokenbuffer.GETTKNSTR()
			} else if testingstate == 3 {
				token, EOL = tokenbuffer.GETTKNEOL()
			} else if testingstate == 4 {
				token, EOL = tokenbuffer.GetTokenString(false)
			} else if testingstate == 5 {
				token, EOL = tokenbuffer.GetToken(false)
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
			}
		}
		fmt.Println()
		log.Println(" Finished processing the inputline.")
	}
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
