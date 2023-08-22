package main

import (
	"bufio"
	"fmt"
	"src/getcommandline"
	"log"
	"os"
	"strconv"
	"strings"
)

//-------------------------------------------------------------------- InsertByteSlice
func InsertIntoByteSlice(slice, insertion []byte, index int) []byte {
	return append(slice[:index], append(insertion, slice[index:]...)...)
}

//---------------------------------------------------------------------- AddCommas
func AddCommas(instr string) string {
	var i, decptposn int
	var Comma []byte = []byte{','}

	BS := make([]byte, 0, 100)
	//  outBS := make([]byte,0,100);
	decptposn = strings.LastIndex(instr, ".")
	BS = append(BS, instr...)

	if decptposn < 0 { // decimal point not found
		i = len(BS)
		BS = append(BS, '.')
	} else {
		i = decptposn
	}

	fmt.Println(" In AddCommas and string is: ", instr, ".  ByteSlice is: ", BS, ".  decptposn is: ", decptposn, ".  i= ", i)

	//  copy(outBS, BS);

	for NumberOfCommas := i / 3; (NumberOfCommas > 0) && (i > 3); NumberOfCommas-- {
		i -= 3
		BS = InsertIntoByteSlice(BS, Comma, i)
	}
	return string(BS)
} // AddCommas
//-----------------------------------------------------------------------------------------------------------------------------

//---------------------------------------------------------------------- AddCommasInt
func AddCommasInt(in int) string {
	var Comma []byte = []byte{','}

	BS := make([]byte, 0, 15)
	s := strconv.Itoa(in)
	BS = append(BS, s...)

	i := len(BS)

	fmt.Println(" In AddCommasInt and int is: ", in, ".  ByteSlice is: ", BS, ".  i= ", i)

	for NumberOfCommas := i / 3; (NumberOfCommas > 0) && (i > 3); NumberOfCommas-- {
		i -= 3
		BS = InsertIntoByteSlice(BS, Comma, i)
	}
	return string(BS)
} // AddCommas
//-----------------------------------------------------------------------------------------------------------------------------

func main() {

	commandline := getcommandline.GetCommandLineString()
	//  commandline = strings.ToUpper(commandline);

	number, _ := strconv.Atoi(commandline)
	fmt.Println(" number on command line is ", number)

	//  R,_ := strconv.ParseFloat(commandline,64);  // If err not nil, R becomes 0 by this routine.  So I won't check err.

	//  s := AddCommas(commandline);
	s := AddCommasInt(number)

	//  fmt.Printf(" Commandline: %s.  R= %g.  S after AddCommas %s\n",commandline,R,s);
	fmt.Println(" Number is ", number, ", After added commas it is ", s)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(" Input test number: ")
		scanner.Scan()
		inputline := scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			fmt.Errorf(" Error from reading std input: %v\n", err) // I'm adding this line for fun.
			os.Exit(1)
		}
		inputline = strings.TrimSpace(inputline)
		fmt.Println(" inputline call to TrimSpace: ", inputline)
		if strings.ToUpper(inputline) == "QUIT" || len(inputline) == 0 {
			log.Println(" Test addcommas finished.")
			os.Exit(0)
		}

		WithCommas := AddCommas(inputline)

		fmt.Println(" After AddCommas: ", WithCommas)
	}
}
