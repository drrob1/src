// (C) 1990-2019.  Robert W Solomon.  All rights reserved.
// convertreg.go (from rpnterm.go)
package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"src/hpcalc"
	"src/tokenize"
)

const LastAltered = "2 Dec 2018"

type Register struct {
	Value float64
	Name  string
}

var InStorage [36]float64   // 0 .. 9, a .. z
var OutStorage [36]Register // 0 .. 9, a .. z
var Stk hpcalc.StackType    // global for WriteStack function

/*
REVISION HISTORY
----------------
 2 Dec 18 -- Trying GoLand for editing this code now.  Converting old register files to new register format which includes string names for the registers.
*/

func main() {
	var HomeDir string

	const InStorage1File = "RPNStorage.gob" // Old register file format, without string reg names.
	// Remember that this includes the stack.
	const OutStorage1FileName = "RPNStorageName.gob" // New register file format w/ string reg names.

	var err error

	fmt.Println()
	fmt.Println()

	if runtime.GOOS == "linux" {
		HomeDir = os.Getenv("HOME")
	} else if runtime.GOOS == "windows" {
		HomeDir = os.Getenv("userprofile")
	} else { // then HomeDir will be empty.
		fmt.Println(" runtime.GOOS does not say linux or windows.  Is this a Mac?")
	}

	fmt.Println(" Convert old HP RPN reg file to new file format with reg names.  Written in Go.  Last altered", LastAltered)

	execname, _ := os.Executable()
	ExecFI, _ := os.Stat(execname)
	LastLinkedTimeStamp := ExecFI.ModTime().Format("Mon Jan 2 2006 15:04:05 MST")
	fmt.Printf("%s was last linked on %s.  Full executable is %s.", ExecFI.Name(), LastLinkedTimeStamp, execname)

	theFileExists := true

	StorageFullFilenameSlice := make([]string, 5)
	StorageFullFilenameSlice[0] = HomeDir
	StorageFullFilenameSlice[1] = string(os.PathSeparator)
	StorageFullFilenameSlice[2] = InStorage1File
	InStorageFullFilename := strings.Join(StorageFullFilenameSlice, "")

	Storage2FullFilenameSlice := make([]string, 5)
	Storage2FullFilenameSlice[0] = HomeDir
	Storage2FullFilenameSlice[1] = string(os.PathSeparator)
	Storage2FullFilenameSlice[2] = OutStorage1FileName
	OutStorageFullFilename := strings.Join(Storage2FullFilenameSlice, "")

	infile, err := os.Open(InStorageFullFilename) // open for reading
	if os.IsNotExist(err) {
		log.Println(" thefile does not exist for reading. ")
		theFileExists = false
	} else if err != nil {
		log.Printf(" Error from os.Open(Storage1FileName) is %v\n", err)
		theFileExists = false
	}
	if !theFileExists {
		log.Fatalln(" Input storage file", InStorageFullFilename, " does not exist.  Exiting.")
	}
	defer infile.Close()
	decoder := gob.NewDecoder(infile) // decoder reads the file.

	err = decoder.Decode(&Stk)       // decoder reads the file.  Stack and Reg are combined into one file.
	check(err)                       // theFileExists, so panic if this is an error.
	err = decoder.Decode(&InStorage) // decoder reads the file.
	check(err)                       // theFileExists, so panic if this is an error.

	infile.Close()

	for i, r := range InStorage {
		OutStorage[i].Value = r
		OutStorage[i].Name = ""
	}

	fmt.Println("Writing the stack")
	WriteStack()
	fmt.Println()
	fmt.Println(" Writing the Storage registers")
	n := WriteStorage()
	fmt.Println(" There are ", n, "non-zero registers.")
	fmt.Println()

	// Time to write files before exiting.

	outfile, err := os.Create(OutStorageFullFilename) // for writing
	check(err)                                        // This should not fail, so panic if it does fail.
	defer outfile.Close()

	encoder := gob.NewEncoder(outfile) // encoder writes the file
	err = encoder.Encode(Stk)          // encoder writes the file
	check(err)                         // Panic if there is an error
	err = encoder.Encode(OutStorage)   // encoder writes the file
	check(err)

	outfile.Close()
} // main in convertreg.go

/* ------------------------------------------------------------ GetRegIdx --------- */
func GetRegIdx(chr byte) int {
	// Return 0..35 w/ A = 10 and Z = 35
	ch := tokenize.CAP(chr)
	if (ch >= '0') && (ch <= '9') {
		ch = ch - '0'
	} else if (ch >= 'A') && (ch <= 'Z') {
		ch = ch - 'A' + 10
	} else { // added 12/11/2016 to fix bug
		ch = 0
	}
	return int(ch)
} // GetRegIdx

/*-------------------------------------------------------------- GetRegChar ------  */
func GetRegChar(idx int) string {
	/* Return '0'..'Z' with A = 10 and Z = 35 */
	const RegNames = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	if (idx < 0) || (idx > 35) {
		return "0"
	}
	ch := RegNames[idx]
	return string(ch)
} // GetRegChar

// ------------------------------------------------------- check -------------------------------
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// ------------------------------------------------------------ WriteRegToScreen --------------
func WriteStorage() int { // Outputs the number of reg's that are not zero.
	FirstNonZeroStorageFlag := true
	n := 0

	for i, r := range OutStorage {
		if r.Value != 0.0 {
			if FirstNonZeroStorageFlag {
				fmt.Println("The following storage registers are not zero")
				FirstNonZeroStorageFlag = false
			} // if firstnonzerostorageflag
			ch := GetRegChar(i)
			s := strconv.FormatFloat(r.Value, 'g', 4, 64) // sigfig of -1 means max sigfig.
			s = hpcalc.CropNStr(s)
			if r.Value >= 10000 {
				s = hpcalc.AddCommas(s)
			}
			fmt.Printf(" Reg [%s], %s =  %s\n", ch, r.Name, s)
			n++
		} // if storage value is not zero
	} // for range over Storage
	if FirstNonZeroStorageFlag {
		fmt.Println(" All storage registers are zero.")
	}
	return n
} // WriteRegToScreen

// --------------------------------------------------- Cap -----------------------------------------
func Cap(c rune) rune {
	r, _, _, _ := strconv.UnquoteChar(strings.ToUpper(string(c)), 0)
	return r
} // Cap

// ------------------------------------------- WriteStack ------------------------------------------
func WriteStack() {

	for i, r := range Stk {
		fmt.Printf("Stack[%d] is %.2f\n", i, r)
	}
} // end WriteStack

// ---------------------------------------------------- End convertreg.go ------------------------------
