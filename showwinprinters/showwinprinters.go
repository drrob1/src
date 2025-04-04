package main

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/godoes/printers"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

/*
  23 Feb 25 -- Yesterday and today I'm playing w/ this package.
*/

func main() {
	fmt.Printf(" Show Windows Printers\n")
	onWin := runtime.GOOS == "windows"
	if onWin {
		fmt.Printf(" Only works on windows.  This is a Windows system so this should work\n")
	} else {
		fmt.Printf(" Only works on Windows.  This is not windows.  Bye-bye.\n")
		return
	}

	printerNames, err := printers.ReadNames()
	if err != nil {
		log.Fatalf("Error reading printer names: %v", err)
	}

	fmt.Println("Available printers:")
	for _, name := range printerNames {
		fmt.Println(name)
	}

	defaultPrinter, err := printers.GetDefault()
	if err != nil {
		fmt.Printf("Error getting default printer: %v", err)
		return
	}
	fmt.Printf(" Default Printer: %s\n", defaultPrinter)

	p, err := printers.Open(defaultPrinter)
	if err != nil {
		fmt.Printf("Error opening printer: %v", err)
	}
	defer func() {
		err = p.Close()
		if err != nil {
			fmt.Printf("Error closing printer: %v", err)
		}
	}()
	fmt.Printf(" Printer opened successfully\n")

	// Now load an image
	if len(os.Args) == 1 {
		fmt.Printf(" No image files on command line.  Bye\n")
		return
	}
	imgName := os.Args[1]
	fullFilename, err := filepath.Abs(imgName)
	if err != nil {
		fmt.Printf(" Error from filepath.Abs(%s) is %v\n", imgName, err)
		return
	}

	fi, err := os.Stat(fullFilename)
	if err != nil {
		fmt.Printf(" Error from os.Stat(%s) is %v\n", fullFilename, err)
		return
	}
	fmt.Printf(" File %s, Size: %d, timestamp: %s\n", fullFilename, fi.Size(), fi.ModTime())

	imgRead, err := imaging.Open(fullFilename)
	if err != nil {
		fmt.Printf(" Error from imaging.Open(%s) is %v\n", fullFilename, err)
		return
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, imgRead, nil)
	if err != nil {
		fmt.Printf(" Error from jpeg.Encode(%s) is %v\n", fi.Name(), err)
		return
	}

	fmt.Printf(" Next step would be n, err := p.Write(buf) and then to check err and show n\n")
	err = p.StartRawDocument(fi.Name())
	if err != nil {
		fmt.Printf(" Error from p.StartRawDocument(%s) is %v\n", fi.Name(), err)
	}
	n, err := p.Write(buf.Bytes())
	if err != nil {
		fmt.Printf(" Error from p.Write(%s) is %v\n", fi.Name(), err)
	}
	fmt.Printf(" n: %d\n", n)
}
