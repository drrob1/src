package main

import (
	"bytes"
	"fmt"
	"github.com/alexbrainman/printer"
	"github.com/disintegration/imaging"
	"github.com/godoes/printers"
	"golang.org/x/image/draw"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	//"github.com/godoes/printers"
)

/*
  23 Feb 25 -- now called winprint, copied from showwinprinters.go.  I'm going to include some code I got from perplexity.
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

	printerNames, err := printer.ReadNames()
	if err != nil {
		log.Fatalf("Error reading printer names: %v", err)
	}

	fmt.Println("Available printers from alex brainman:")
	for _, name := range printerNames {
		fmt.Println(name)
	}

	printNames, err := printers.ReadNames()
	if err != nil {
		log.Fatalf("Error reading printer names: %v", err)
		return
	}
	fmt.Printf("Available printers from godoes/printers:\n")
	for _, name := range printNames {
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

	driverInfo, err := p.DriverInfo()
	if err != nil {

	}
	fmt.Printf(" Driver info: %v\n", driverInfo)

	dataType, err := p.GetDataType()
	if err != nil {
		fmt.Printf("Error getting data type: %v", err)
	}
	fmt.Printf(" Data type: %v\n", dataType)

	if pause() {
		return
	}

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
	fmt.Printf(" Imaging file read and processed: %s\n", fullFilename)

	//var buf bytes.Buffer  not part of most recent example.
	//err = jpeg.Encode(&buf, imgRead, nil)
	//if err != nil {
	//	fmt.Printf(" Error from jpeg.Encode(%s) is %v\n", fi.Name(), err)
	//	return
	//}
	//fmt.Printf(" Jpg Name %s, Fi.Size: %d, buf len: %d\n", fi.Name(), fi.Size(), buf.Len())

	//fmt.Printf(" Next step would be n, err := p.Write(buf) and then to check err and show n\n")
	//err = p.StartRawDocument(fi.Name())
	//if err != nil {
	//	fmt.Printf(" Error from p.StartRawDocument(%s) is %v\n", fi.Name(), err)
	//}
	//n, err := p.Write(buf.Bytes())
	//if err != nil {
	//	fmt.Printf(" Error from p.Write(%s) is %v\n", fi.Name(), err)
	//}
	//fmt.Printf(" n: %d\n", n)

	/*
		// Open the JPG file
		file, err := os.Open("path/to/your/image.jpg")
		if err != nil {
		        log.Fatalf("Failed to open image: %v", err)
		}
		defer file.Close()

		// Decode the JPG
		img, err := jpeg.Decode(file)
		if err != nil {
		        log.Fatalf("Failed to decode image: %v", err)
		}
	*/

	jobs, err := p.Jobs()
	if err != nil {
		fmt.Printf(" Error from p.Jobs() is %v\n", err)
		return
	}
	fmt.Printf(" Jobs opened successfully, and are:\n")
	for _, job := range jobs {
		fmt.Printf(" Job: %v\n", job)
	}
	fmt.Printf(" Jobs processed successfully, and are above.\n\n")

	paperSizes, err := p.Forms()
	if err != nil {
		fmt.Printf(" Error from p.Forms() is %v\n", err)
		return
	}
	fmt.Printf(" Forms opened successfully, and there are %d paper sizes\n", len(paperSizes)) // 269 forms(paper sizes) listed here
	//if pause() {
	//	return
	//}
	//for _, paperSize := range paperSizes {
	//	fmt.Printf(" Form: %s, size: %v, flags: %v\n", paperSize.Name, paperSize.Size, paperSize.Flags)
	//}
	// Start a new print job

	err = p.StartDocument("RAW", "RAW")
	if err != nil {
		fmt.Printf(" Error from p.StartDocument(\"RAW\", RAW): %s\n", err.Error())
		//return
	}
	fmt.Printf(" Document started successfully\n")

	// Start a new page
	err = p.StartPage()
	if err != nil {
		log.Fatalf("Failed to start page: %v", err)
		//return
	}
	fmt.Printf(" Page started successfully\n")

	// Get the page size
	bounds := imgRead.Bounds()
	fmt.Printf(" Image bounds: %v\n", bounds)

	scaledImg := image.NewNRGBA(bounds)
	draw.ApproxBiLinear.Scale(scaledImg, bounds, imgRead, imgRead.Bounds(), draw.Over, nil)

	// Draw the image on the page
	draw.Draw(scaledImg, bounds, imgRead, image.Point{}, draw.Src)

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, scaledImg, nil)

	n, err := p.Write(buf.Bytes())
	if err != nil {
		fmt.Printf(" Error: n=%d,  p.Write(): %v\n", n, err)
	}
	fmt.Printf(" Wrote successfully %d bytes\n", n)

	// End the page
	err = p.EndPage()
	if err != nil {
		log.Fatalf("Failed to end page: %v", err)
	}

	err = p.EndDocument()
	if err != nil {
		log.Fatalf("Failed to end document: %v", err)
	}
	log.Printf("%s printed successfully %d bytes\n", fi.Name(), n)
}

func pause() bool {
	var ans string
	fmt.Printf(" Pausing.  Stop [y/N]: ")
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	return strings.HasPrefix(ans, "y") // suggested by staticcheck.
}
