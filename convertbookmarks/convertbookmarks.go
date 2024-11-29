package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

/*

29 Nov 24 -- First version, needed because I changed the structure of the bookmarks.  Now the bookmarks will only contain a directory name.  I'm removing the cdd part

*/

const bookmarkFilename = "bookmarkfile.gob"

var LastAltered = "Nov 29, 2024"

var bookmark map[string]string // used in all routines.

func main() {
	var err error
	fmt.Println(" convertbookmarks is intended to be used only once, to do what it's name says.  Written in Go, last altered", LastAltered, "and compiled w/", runtime.Version())

	verboseFlag := flag.Bool("v", false, "verbose message and exit.")
	flag.Parse()

	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")

	if *verboseFlag {
		fmt.Printf(" %s last compiled %s by %s.  Full binary is %s with timestamp of %s.\n", os.Args[0], LastAltered, runtime.Version(), execName, ExecTimeStamp)
	}

	configDir, err := os.UserConfigDir() // this routine became available in Go 1.14
	if err != nil {
		fmt.Println(err, "Exiting")
		os.Exit(1)
	}

	fullBookmarkFilename := filepath.Join(configDir, bookmarkFilename)

	if *verboseFlag { // added 11/29/24
		fmt.Printf(" fullBookmarkFilename: %q\n", fullBookmarkFilename)
	}

	_, err = os.Stat(fullBookmarkFilename)
	if err != nil {
		ctfmt.Printf(ct.Red, true, "No bookmark file found at %s.\n", fullBookmarkFilename)
		return
	}

	bookmarkfile, err := os.Open(fullBookmarkFilename)
	if err != nil {
		log.Fatalln(" cannot open", fullBookmarkFilename, " as input bookmark file, because of", err)
	}
	defer bookmarkfile.Close()
	decoder := gob.NewDecoder(bookmarkfile)
	err = decoder.Decode(&bookmark)
	if err != nil {
		log.Println(" cannot decode", fullBookmarkFilename, ", error is", err, ".  Aborting")
		return
	}
	bookmarkfile.Close()
	fmt.Printf(" %d Bookmarks read in from %s\n", len(bookmark), fullBookmarkFilename)
	fmt.Println()

	// this is the converting bookmark file loop
	for k, v := range bookmark {
		entry := v[4:] // this is to remove the beginning "cdd " part
		bookmark[k] = entry
	}

	// write out bookmarkFile
	bookmarkFile, er := os.Create(fullBookmarkFilename)
	if er != nil {
		log.Println(" could not create bookmarkfile upon exiting, because", er)
		return // better to return, because this allows the deferred functions to run.  os.Exit does not run deferred functions.
	}
	defer bookmarkFile.Close()

	encoder := gob.NewEncoder(bookmarkFile)
	e := encoder.Encode(&bookmark)
	if e != nil {
		log.Println(" could not encode bookmarkfile upon exiting, because", e)
	}
	bookmarkFile.Close()
}
