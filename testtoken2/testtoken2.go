
package main

/*
  REVISION HISTORY
  ================
  19 Aug 16 -- First Go version completed to test all parts of tokenize.go package
  21 Sep 16 -- Now need to test my new GetTknStrPreserveCase routine.  And test the change I made to GETCHR.
   7 Oct 16 -- Changed the scanner to scan by words.  I hope.  This is to test the scanner for rpng.  Default is scan by lines.
*/

import (
        "fmt"
//        "tokenize"
        "strings"
        "log"
        "os"
        "bufio"
        "getcommandline"
       )

// var FSAnameType = [...]string{"DELIM","OP","DGT","ALLELSE"};

func main() {
// testingstate = 0: gettkn, 1: gettknreal, 2: gettknstr, 3: gettkneol, 4: string lower case, 
// 5: token lower case
  testingstate := 0;
  commandline := getcommandline.GetCommandLineString();
  commandline = strings.ToUpper(commandline);
  floatflag := false;
  if (commandline == "REAL") || (commandline == "FLOAT") {
    floatflag = true;
    testingstate = 1;
  }else if commandline == "STRING" {
    testingstate = 2;
  }else if commandline == "EOL" {
    testingstate = 3;
  }else if commandline == "STR" {
    testingstate = 4;
  }else if commandline == "LOWER"{
    testingstate = 5;
  }

  fmt.Print(" Will be testing GetTknReal? ",floatflag,", testingstate is ",testingstate);
  fmt.Println();

  scanner := bufio.NewScanner(os.Stdin);

  scanner.Split(bufio.ScanWords); // testing scanning by words to see what happens.

  for {
    fmt.Print(" Input test text: ");
    scanner.Scan();
    inputline := scanner.Text();
    if err := scanner.Err(); err != nil {
      fmt.Fprintln(os.Stderr, "reading standard input:", err)
      os.Exit(1);
    }
    if len(inputline) == 0 {
      os.Exit(0);
    }
//    fmt.Println(" After the call to scanner.Text(), before TrimSpace: ",inputline," len:",len(inputline));
//    inputline = strings.TrimSpace(inputline);
//    fmt.Println(" After call to TrimSpace: ",inputline," len:",len(inputline)," err:",scanner.Err());
    fmt.Println(" After call to scanner.Text: ",inputline," len:",len(inputline)," err:",scanner.Err());
    if strings.ToUpper(inputline) == "QUIT" {
      log.Println(" Test Token finished.");
      os.Exit(0);
    }

/*  Just testing the scan by word operation here
    tokenize.INITKN(inputline);
    EOL := false;
    token := tokenize.TokenType{};
    for !EOL {
      if floatflag || testingstate == 1 {
        token,EOL = tokenize.GETTKNREAL();
      } else if testingstate == 2 {
        token,EOL = tokenize.GETTKNSTR();
      }else if testingstate == 3 {
        token,EOL = tokenize.GETTKNEOL();
      }else if testingstate == 4 {
        token,EOL = tokenize.GetTokenString(false);
      }else if testingstate == 5 {
        token,EOL = tokenize.GetToken(false);
      }else{
        token,EOL = tokenize.GETTKN();
      }

      fmt.Printf(" Token : %#v \n",token);
      fmt.Println(" EOL : ",EOL);
      if EOL { break }  // I don't want it to ask about ungettkn if there is an EOL condition.
      fmt.Print(" call UnGetTkn? (Y/N) ");
      scanner.Scan();
      ans := scanner.Text();
      if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "reading standard input:", err)
        os.Exit(1);
      }
      ans = strings.TrimSpace(ans);
      ans = strings.ToUpper(ans);
      if strings.HasPrefix(ans,"Y") {
        tokenize.UNGETTKN();
      }
    }
*/
    fmt.Println();
    log.Println(" Finished processing the inputline.");
  }
}
