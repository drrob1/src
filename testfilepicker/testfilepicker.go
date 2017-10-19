package main

/*
  REVISION HISTORY
  ================
  19 Aug 16 -- First Go version completed to test all parts of tokenize.go package
  21 Sep 16 -- Now need to test my new GetTknStrPreserveCase routine.  And test the change I made to GETCHR.
  13 Oct 17 -- Testing the inclusion of horizontal tab as a delim, needed for comparehashes.
  18 Oct 17 -- Now called testfilepicker, derived from testtoken.go
*/

import (
	"bufio"
	"filepicker"
	"fmt"
	"getcommandline"
	"log"
	"os"
	"strconv"
	"strings"
	"tknptr"
)

// var FSAnameType = [...]string{"DELIM","OP","DGT","ALLELSE"};
const maxchoices = 10

func main() {
	var ans string
	// testingstate = 0: gettkn, 1: gettknreal, 2: gettknstr, 3: gettkneol, 4: string lower case,
	// 5: token lower case
	testingstate := 0
	commandline := getcommandline.GetCommandLineString()
	filenames := filepicker.GetFilenames(commandline)
	fmt.Println(" Number of filenames in the string slice are", len(filenames))
	for i := 0; i < maxchoices; i++ {
		fmt.Println("filename[", i, "] is", filenames[i])
	}
	fmt.Print(" Enter filename choice : ")
	fmt.Scanln(&ans)
	fmt.Printf(" ans is string %s, hex %x \n", ans, ans)
	i, err := strconv.Atoi(ans)
	if err == nil {
		fmt.Println(" ans as int is", i)
	} else {
		s := strings.ToUpper(ans)
		s = strings.TrimSpace(s)
		s0 := s[0]
		i = int(s0 - 'A') // may need byte(s0) - 'A' or byte(s0-'A') or some other permutation
		fmt.Println(" string ans as int is", i, " referenced filename is", filenames[i])
	}
	fmt.Println(" Picked filename is", filenames[i])
	fmt.Println()

	a := 'a'
	b := a ^ 32
	c := a &^ 32
	fmt.Println(" playing with bit clearing.  a,b,c =", a, b, c)

	fmt.Println()

	//	commandline = strings.ToUpper(commandline)
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
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}
		if len(inputline) == 0 {
			os.Exit(0)
		}
		fmt.Println(" After the call to Scan(), before TrimSpace: ", inputline, ".")
		inputline = strings.TrimSpace(inputline)
		fmt.Println(" After call to TrimSpace: ", inputline)
		if strings.ToUpper(inputline) == "QUIT" {
			log.Println(" Test Token finished.")
			os.Exit(0)
		}
		tkn := tknptr.NewToken(inputline)
		EOL := false
		token := tknptr.TokenType{}
		for !EOL {
			if floatflag || testingstate == 1 {
				token, EOL = tkn.GETTKNREAL()
			} else if testingstate == 2 {
				token, EOL = tkn.GETTKNSTR()
			} else if testingstate == 3 {
				token, EOL = tkn.GETTKNEOL()
			} else if testingstate == 4 {
				token, EOL = tkn.GetTokenString(false)
			} else if testingstate == 5 {
				token, EOL = tkn.GetToken(false)
			} else {
				token, EOL = tkn.GETTKN()
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
			//			scanner.Scan()
			//			ans := scanner.Text()
			//			if err := scanner.Err(); err != nil {
			//				fmt.Fprintln(os.Stderr, "reading standard input:", err)
			//				os.Exit(1)
			//			}
			fmt.Scan(&ans)
			ans = strings.TrimSpace(ans)
			ans = strings.ToUpper(ans)
			if strings.HasPrefix(ans, "Y") {
				tkn.UNGETTKN()
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
