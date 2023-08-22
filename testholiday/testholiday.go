package main

/*
  REVISION HISTORY
  ================
  19 Aug 16 -- First Go version completed to test all parts of tokenize.go package
  21 Sep 16 -- Now need to test my new GetTknStrPreserveCase routine.  And test the change I made to GETCHR.
   7 Apr 17 -- Used TestToken as code base for this, which is now TestHoliday.  Holiday is pannicing from calgo.
*/

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"src/tokenize"
	//        "getcommandline"
	"src/holidaycalc"
)

// var FSAnameType = [...]string{"DELIM","OP","DGT","ALLELSE"};

func main() {

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(" Input holiday year: ")
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
		tokenize.INITKN(inputline)
		token, EOL := tokenize.GetToken(false)
		year := token.Isum

		fmt.Printf(" Token : %#v \n", token)
		fmt.Println(" EOL : ", EOL)
		fmt.Println()

		Holiday := holidaycalc.GetHolidays(year)
		fmt.Printf(" Holiday: %#v \n", Holiday)

		holiday := holidaycalc.GetHolidays(2017)
		fmt.Printf(" holiday 2017: %#v \n", holiday)

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
