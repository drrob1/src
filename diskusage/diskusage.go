package main

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	"github.com/ricochet2200/go-disk-usage/du"
	"os"
	"runtime"
	"unicode"
)

// Results are still off by 3 orders of magnitude, even if I use it as bin/diskusage "/"

func main() {
	if len(os.Args) < 2 {
		fmt.Println(" Usage: diskusage [/dev/sdX] | [drive:]")
		os.Exit(1)
	}
	volumePath := ""
	volume := os.Args[1]

	if runtime.GOOS == "linux" {
		/*
		if !(strings.HasPrefix(volume, "/dev/")) {
			fmt.Println(" On linux, volume must begin /dev/.")
			os.Exit(1)
		}

		 */
		volumePath = volume
	} else if runtime.GOOS == "windows" {
		volByteSlice := []byte(volume)
		driveletter := rune(volByteSlice[0])
		colon := rune(volByteSlice[1])
		if !(unicode.IsLetter(driveletter) && colon == ':') {
			fmt.Println(" On Windows, volume must begin w/ a 2 character drive designator")
			os.Exit(1)
		}
		volumePath = string(driveletter + colon + os.PathSeparator)
	}

	fmt.Println(" volumepath is", volumePath)
	usage := du.NewDiskUsage(volumePath)
	//fmt.Println("Free:", usage.Free())
	free := usage.Free()
	freestr, color := getMagnitudeString(int64(free))

	//fmt.Println("Available:", usage.Available())
	ctfmt.Println(color, false, " Free:", free, ";", freestr)
	avail := usage.Available()
	availstr, availcolor := getMagnitudeString(int64(avail))
	ctfmt.Println(availcolor, false, "Available:", avail, ";", availstr)

	//fmt.Println("Size:", usage.Size())
	size := usage.Size()
	sizestr, sizecolor := getMagnitudeString(int64(size))
	ctfmt.Println(sizecolor, false, " Size:", size, ";", sizestr)

	//fmt.Println("Used:", usage.Used())
	used := usage.Used()
	usedstr, usedcolor := getMagnitudeString(int64(used))
	ctfmt.Println(usedcolor, false, " Used:", used, ";", usedstr)
	fmt.Println()
	fmt.Println()
}
// ----------------------------- getMagnitudeString -------------------------------
func getMagnitudeString(j int64) (string, ct.Color) {

	var s1 string
	var f float64
	var color ct.Color
	switch {
	case j > 1_000_000_000_000: // 1 trillion, or TB
		f = float64(j) / 1000000000000
		s1 = fmt.Sprintf("%.4g TB", f)
		color = ct.Red
	case j > 100_000_000_000: // 100 billion
		f = float64(j) / 1_000_000_000
		s1 = fmt.Sprintf(" %.4g GB", f)
		color = ct.White
	case j > 10_000_000_000: // 10 billion
		f = float64(j) / 1_000_000_000
		s1 = fmt.Sprintf("  %.4g GB", f)
		color = ct.White
	case j > 1_000_000_000: // 1 billion, or GB
		f = float64(j) / 1000000000
		s1 = fmt.Sprintf("   %.4g GB", f)
		color = ct.White
	case j > 100_000_000: // 100 million
		f = float64(j) / 1_000_000
		s1 = fmt.Sprintf("    %.4g mb", f)
		color = ct.Yellow
	case j > 10_000_000: // 10 million
		f = float64(j) / 1_000_000
		s1 = fmt.Sprintf("     %.4g mb", f)
		color = ct.Yellow
	case j > 1_000_000: // 1 million, or MB
		f = float64(j) / 1000000
		s1 = fmt.Sprintf("      %.4g mb", f)
		color = ct.Yellow
	case j > 100_000: // 100 thousand
		f = float64(j) / 1000
		s1 = fmt.Sprintf("       %.4g kb", f)
		color = ct.Cyan
	case j > 10_000: // 10 thousand
		f = float64(j) / 1000
		s1 = fmt.Sprintf("        %.4g kb", f)
		color = ct.Cyan
	case j > 1000: // KB
		f = float64(j) / 1000
		s1 = fmt.Sprintf("         %.3g kb", f)
		color = ct.Cyan
	default:
		s1 = fmt.Sprintf("%3d bytes", j)
		color = ct.Green
	}
	return s1, color
}
