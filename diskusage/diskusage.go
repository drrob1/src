package main

import (
	"fmt"
	"github.com/ricochet2200/go-disk-usage/du"
	"os"
	"runtime"
	"strings"
	"unicode"
)

//var KB = uint64(1024)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(" Usage: diskusage [/dev/sdX] | [drive:]")
		os.Exit(1)
	}
	volumePath := ""
	volume := os.Args[1]
	if runtime.GOOS == "linux" {
		//if !(strings.HasPrefix(volume, "/dev/") && strings.HasSuffix(volume, string(os.PathSeparator))) {
		if !(strings.HasPrefix(volume, "/dev/")) {
			fmt.Println(" On linux, volume must begin /dev/.")
			os.Exit(1)
		}
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
	usage := du.NewDiskUsage(volumePath)
	fmt.Println("Free:", usage.Free())
	fmt.Println("Available:", usage.Available())
	fmt.Println("Size:", usage.Size())
	fmt.Println("Used:", usage.Used())
	//fmt.Println("Usage:", usage.Usage()*100, "%")
	fmt.Println()
}
