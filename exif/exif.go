/*
  EXIF data decoding so I can understand it
  20 Feb 25 -- Started playing w/ this.  This is working well enough.  Only jpg's have EXIF data, AFAICT.
*/

package main

import (
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	flag "github.com/spf13/pflag"
	"os"
	"src/filepicker"
	"strconv"
	"strings"
)

//"github.com/rwcarlsen/goexif/exif"

func main() {
	var verboseFlag bool
	var infileName string
	var ans string

	fmt.Printf(" EXIF data viewer. ")
	flag.BoolVarP(&verboseFlag, "verbose", "v", false, "verbose output")
	flag.Parse()

	if flag.NArg() < 1 {
		filenames, err := filepicker.GetRegexFilenames("(png$)|(jpg$)|(webm$)|(tif$}") // $ matches end of line
		if err != nil {
			fmt.Printf(" Error from GetRegexFilenames is %v, exiting\n", err)
			os.Exit(1)
		}
		for i := 0; i < min(len(filenames), 26); i++ {
			fmt.Printf("filename[%d, %c] is %s \n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice (stop code=999 . , / ;) : ")
		n, err := fmt.Scanln(&ans)
		if len(ans) == 0 || err != nil || n == 0 { // these are redundant.  I'm playing now.
			ans = "0"
		} else if ans == "999" || ans == "." || ans == "," || ans == "/" || ans == ";" {
			fmt.Println(" Stop code entered.")
			return
		}

		i, err := strconv.Atoi(ans)
		if err == nil {
			infileName = filenames[i]
		} else { // allow entering 'a' .. 'z' for 0 to 25.
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			if i > 25 {
				fmt.Printf(" Index out of bounds.  It is %d.\n", i)
				return
			}
			infileName = filenames[i]
		}
		fmt.Println(" Picked filename is", infileName)
	} else {
		infileName = flag.Arg(0)
	}

	// open the image file
	f, err := os.Open(infileName)
	if err != nil {
		fmt.Printf(" Error Opening %s is %v, exiting\n", infileName, err)
		return
	}
	defer f.Close()

	// Optionally register camera makenote data parsing - currently Nikon and
	// Canon are supported.
	exif.RegisterParsers(mknote.All...)

	// get EXIF data
	EXIF, err := exif.Decode(f)
	if err != nil {
		fmt.Printf(" Error Decoding %s is %v, exiting\n", infileName, err)
		return
	}

	camModel, err := EXIF.Get(exif.Model)
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Println(camModel.StringVal())
	}

	focal, err := EXIF.Get(exif.FocalLength)
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		numer, denom, _ := focal.Rat2(0) // retrieve first (only) rat. value
		fmt.Printf("Focal length: %v/%v\n", numer, denom)
	}

	// Two convenience functions exist for date/time taken and GPS coords:
	tm, err := EXIF.DateTime()
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Println("Taken: ", tm)
	}

	lat, long, err := EXIF.LatLong()
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Println("lat, long: ", lat, ", ", long)
	}

	orientation, err := EXIF.Get(exif.Orientation)
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Printf(" Orientation is %s\n", orientation)
	}

	software, err := EXIF.Get(exif.Software)
	if err != nil {
		fmt.Printf(" Error %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Printf(" Software is %s\n", software)
	}

	colorspace, err := EXIF.Get(exif.ColorSpace)
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Printf(" Colorspace is %s\n", colorspace)
	}

	exposureTime, err := EXIF.Get(exif.ExposureTime)
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Printf(" ExposureTime is %s\n", exposureTime)
	}

	fnumber, err := EXIF.Get(exif.FNumber)
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Printf(" FNumber is %s\n", fnumber)
	}

	focalLength, err := EXIF.Get(exif.FocalLength)
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Printf(" FocalLength is %s\n", focalLength)
	}

	whiteBalance, err := EXIF.Get(exif.WhiteBalance)
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Printf(" WhiteBalance is %s\n", whiteBalance)
	}

	digitalZoomRatio, err := EXIF.Get(exif.DigitalZoomRatio)
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Printf(" DigitalZoomRatio is %s\n", digitalZoomRatio)
	}

	gainControl, err := EXIF.Get(exif.GainControl)
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Printf(" GainControl is %s\n", gainControl)
	}

	contrast, err := EXIF.Get(exif.Contrast)
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Printf(" Contrast is %s\n", contrast)
	}

	saturation, err := EXIF.Get(exif.Saturation)
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Printf(" Saturation is %s\n", saturation)
	}

	sharpness, err := EXIF.Get(exif.Sharpness)
	if err != nil {
		fmt.Printf(" Error for %s is %v, skipping\n", infileName, err)
	} else {
		fmt.Printf(" Sharpness is %s\n", sharpness)
	}
}
