package main

import (
	"bufio"
	"fmt"
	"github.com/jdeng/goheif"
	flag "github.com/spf13/pflag"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"src/filepicker"
	"strconv"
	"strings"
)

/*
25 Jul 25 -- First version, based on example code at GitHub.com.  Now that it works on linux (but not on windows, I'll enhance it a bit.
				It's too late now.  I'll do this tomorrow.  I want to add more processing on the output name.
26 Jul 25 -- I will add code from fromfx so that if there are no files on the command line, then it will ask.  And it will assume
				heic for the first file and jpg for the 2nd file, if given on command line without extensions.
29 Jul 25 -- Playing with os.Create instead of the example's use of os.OpenFile.  It works.
*/

const lastModified = "July 26, 2025"
const heicExtension = ".heic"
const jpgExtension = ".jpg"
const defaultCompression = 100

var jpgCompression int

func main() {
	fmt.Printf(" heic converter last modified %s\n\n", lastModified)
	flag.IntVarP(&jpgCompression, "jpgcompression", "j", defaultCompression, "jpg compression level")
	flag.Parse()

	var ans, baseFilename string

	fin, fout := flag.Arg(0), flag.Arg(1)

	// if no command line params, search for them.
	if flag.NArg() < 1 {
		filenames, err := filepicker.GetRegexFilenames("heic$") // $ matches end of line
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from GetRegexFilenames is %v, exiting\n", err)
			os.Exit(1)
		}
		for i := 0; i < min(len(filenames), 26); i++ {
			fmt.Printf("filename[%d, %c] is %s \n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice (stop code=99 . , / ;) : ")
		n, err := fmt.Scanln(&ans)
		if len(ans) == 0 || err != nil || n == 0 { // these are redundant.  I'm playing now.
			ans = "0"
		} else if ans == "99" || ans == "." || ans == "," || ans == "/" || ans == ";" {
			fmt.Println(" Stop code entered.")
			os.Exit(0)
		}

		i, err := strconv.Atoi(ans)
		if err == nil {
			fin = filenames[i]
		} else { // allow entering 'a' .. 'z' for 0 to 25.
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			if i > 25 {
				fmt.Printf(" Index out of bounds.  It is %d.\n", i)
				return
			}
			fin = filenames[i]
		}
		fmt.Println(" Picked filename is", fin)
		baseFilename = fin
	} else {
		baseFilename = filepath.Clean(fin)

		if strings.Contains(baseFilename, ".") { // there is an extension here
			_, err := os.Stat(fin)
			if err != nil {
				fmt.Printf(" Error from os.Stat(%s) is %v, exiting\n", fin, err)
				os.Exit(1)
			}
		} else { // no extension given on command line
			fin = baseFilename + heicExtension
			_, err := os.Stat(fin)
			if err != nil {
				fmt.Printf(" Error from os.Stat(%s) is %v, exiting\n", fin, err)
				os.Exit(1)
			}
		}
	}

	// if fout ext not specified, use jpg extension.  If fout is blank, use same root name as fin w/ jpg extension.
	if fout == "" {
		ext := filepath.Ext(fin)
		fout = strings.TrimSuffix(fin, ext) + jpgExtension
	} else if !strings.Contains(fout, jpgExtension) {
		fout = fout + jpgExtension
	}

	fi, err := os.Open(fin)
	if err != nil {
		log.Fatal(err)
	}
	defer fi.Close()

	img, err := goheif.Decode(fi)
	if err != nil {
		log.Fatalf("Failed to parse %s: %v\n", fin, err)
	}

	//fo, err := os.OpenFile(fout, os.O_RDWR|os.O_CREATE, 0644)  I don't know why the example uses os.OpenFile instead of os.Create
	fo, err := os.Create(fout)
	if err != nil {
		log.Fatalf("Failed to create output file %s: %v\n", fout, err)
	}
	defer fo.Close()

	w := bufio.NewWriter(fo)
	defer w.Flush()
	err = jpeg.Encode(w, img, &jpeg.Options{Quality: jpgCompression})
	if err != nil {
		log.Fatalf("Failed to encode %s: %v\n", fout, err)
	}
}
