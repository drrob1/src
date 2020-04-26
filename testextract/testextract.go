package main

import (

	"os"
	"fmt"
	"path/filepath"
	"sort"
)

const LastAltered = "23 Apr 2020"


func main() {
	extensions := make([]string, 0, 100)
	if len(os.Args) < 1 {
		fmt.Println(" need more than one param on the line.  Exiting.")
		os.Exit(1)
	}
	args := os.Args[1:]
	// args := os.Args  I forgot that Args[0] is the name of the binary
	//for i := range args {
	//	fmt.Printf("arg[%d] = %s \n", i, args[i])
	//}

	//fmt.Println(" args:", args)
	if len(args) < 1 {
		extensions = append(extensions, ".txt")
	} else {
		extensions = extractExtensions(args)
	}
    fmt.Println(" len Args", len(args), ", len(ext)=", len(extensions))
	fmt.Println(" args:", args)
	fmt.Println()
	fmt.Println("extension:", extensions)
	fmt.Println()
} // end main

func extractExtensions(files []string) []string {

	var extensions sort.StringSlice
	extensions = make([]string, 0, 100)
	//fmt.Print(" ext:")
	for _, file := range files {
		ext := filepath.Ext(file)
		extensions = append(extensions, ext)
	//	fmt.Print(ext, " ")
	}
	fmt.Println()
	fmt.Println()
	fmt.Println(" in extractext b4 sort. len files=", len(files), ", len(ext)=", len(extensions), ", extensions:", extensions)
	if len(extensions) > 1 {
		extensions.Sort()
		for i, ext := range extensions {
			if i == 0 {
				continue
			}
			if extensions[i-1] == ext {  // recall that ext here = extensions[i]
				fmt.Printf(" %q EQ %q for i=%d \n", extensions[i-1], extensions[i], i)
				extensions[i-1] = ""
			} else {
				//fmt.Printf(" %q NE %q \n", extensions[0], extensions[i])
			}
		}
		fmt.Println(" in extractExtensions after sort and assigning to null string:", extensions)
		sort.Sort(sort.Reverse(extensions))
		// sort.Sort(sort.Reverse(sort.IntSlice(s)))   I don't remember why this line is here.
		trimmedExtensions := make([]string, 0, len(extensions))
		for _, ext := range extensions {
			if ext != "" {
				trimmedExtensions = append(trimmedExtensions, ext)
			}
		}
		fmt.Println(" in extractExtensions after sort trimmedExtensions:", trimmedExtensions)
		fmt.Println()
		return trimmedExtensions
	}
	//fmt.Println(" in extractExtensions without a sort:", extensions)
	//fmt.Println()
	return extensions

} // end extractExtensions
