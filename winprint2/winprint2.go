package main

import (
	"fmt"
	"github.com/disintegration/imaging"
	"os"
	"path/filepath"
	"strings"
)

/*
  23 Feb 25 -- now called winprint, copied from showwinprinters.go.  I'm going to include some code I got from perplexity.
----------------------------------------------------------------------------------------------------
  28 Apr 25 -- Now called winprint2, copied from winprint.go.  I'm going to include code I got from the AI here.  First I'll scale the image, and then I'll print it using
				IP protocol, which is working here at home.  I just need the IP adr.
				For the HP Officejet_Pro 8620, it is 192.168.1.208.  For the HP Officejet_Pro 8028e, it is 192.168.1.197.
				So far, the Officejet_Pro 8620 is working.  The Officejet_Pro 8028e is not tested.
*/

func main() {
	// Load an image
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
	fmt.Printf(" File %s, Size in bytes: %d, timestamp: %s\n", fullFilename, fi.Size(), fi.ModTime())

	imgRead, err := imaging.Open(fullFilename)
	if err != nil {
		fmt.Printf(" Error from imaging.Open(%s) is %v\n", fullFilename, err)
		return
	}
	fmt.Printf(" Imaging file read and processed: %s\n", fullFilename)
	bounds := imgRead.Bounds()
	fmt.Printf(" Image %s bounds: %v\n", fullFilename, bounds)
	fmt.Printf(" Image %s size as an image: %v\n", fullFilename, bounds.Size())
	fmt.Printf(" Image %s width: %d, height: %d\n", fullFilename, bounds.Dx(), bounds.Dy())

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

	//jobs, err := p.Jobs()
	//if err != nil {
	//	fmt.Printf(" Error from p.Jobs() is %v\n", err)
	//	return
	//}
	//fmt.Printf(" Jobs opened successfully, and are:\n")
	//for _, job := range jobs {
	//	fmt.Printf(" Job: %v\n", job)
	//}
	//fmt.Printf(" Jobs processed successfully, and are above.\n\n")
	//
	//paperSizes, err := p.Forms()
	//if err != nil {
	//	fmt.Printf(" Error from p.Forms() is %v\n", err)
	//	return
	//}
	//fmt.Printf(" Forms opened successfully, and there are %d paper sizes\n", len(paperSizes)) // 269 forms(paper sizes) listed here
	//if pause() {
	//	return
	//}
	//for _, paperSize := range paperSizes {
	//	fmt.Printf(" Form: %s, size: %v, flags: %v\n", paperSize.Name, paperSize.Size, paperSize.Flags)
	//}
	// Start a new print job

	//err = p.StartDocument("RAW", "RAW")
	//if err != nil {
	//	fmt.Printf(" Error from p.StartDocument(\"RAW\", RAW): %s\n", err.Error())
	//	//return
	//}
	//fmt.Printf(" Document started successfully\n")
	//
	//// Start a new page
	//err = p.StartPage()
	//if err != nil {
	//	log.Fatalf("Failed to start page: %v", err)
	//	//return
	//}
	//fmt.Printf(" Page started successfully\n")
	//
	//// Get the page size
	//bounds := imgRead.Bounds()
	//fmt.Printf(" Image bounds: %v\n", bounds)
	//
	//scaledImg := image.NewNRGBA(bounds)
	//draw.ApproxBiLinear.Scale(scaledImg, bounds, imgRead, imgRead.Bounds(), draw.Over, nil)
	//
	//// Draw the image on the page
	//draw.Draw(scaledImg, bounds, imgRead, image.Point{}, draw.Src)
	//
	//var buf bytes.Buffer
	//err = jpeg.Encode(&buf, scaledImg, nil)
	//
	//n, err := p.Write(buf.Bytes())
	//if err != nil {
	//	fmt.Printf(" Error: n=%d,  p.Write(): %v\n", n, err)
	//}
	//fmt.Printf(" Wrote successfully %d bytes\n", n)
	//
	//// End the page
	//err = p.EndPage()
	//if err != nil {
	//	log.Fatalf("Failed to end page: %v", err)
	//}
	//
	//err = p.EndDocument()
	//if err != nil {
	//	log.Fatalf("Failed to end document: %v", err)
	//}
	//log.Printf("%s printed successfully %d bytes\n", fi.Name(), n)
}

func pause() bool {
	var ans string
	fmt.Printf(" Pausing.  Stop [y/N]: ")
	fmt.Scanln(&ans)
	ans = strings.ToLower(ans)
	return strings.HasPrefix(ans, "y") // suggested by staticcheck.
}
