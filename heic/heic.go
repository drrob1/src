package main

import (
	"bufio"
	"fmt"
	"github.com/jdeng/goheif"
	flag "github.com/spf13/pflag"
	"image/jpeg"
	"log"
	"os"
)

func main() {
	flag.Parse()
	fin, fout := flag.Arg(0), flag.Arg(1)
	fi, err := os.Open(fin)
	if err != nil {
		log.Fatal(err)
	}
	defer fi.Close()

	exif, err := goheif.ExtractExif(fi)
	if err != nil {
		fmt.Printf("Warning: no EXIF from %s: %v\n", fin, err)
	}

	img, err := goheif.Decode(fi)
	if err != nil {
		log.Fatalf("Failed to parse %s: %v\n", fin, err)
	}

	fo, err := os.OpenFile(fout, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Failed to create output file %s: %v\n", fout, err)
	}
	defer fo.Close()

	w := bufio.NewWriter(fo)
	defer w.Flush()
	if exif != nil {
		_, err = w.Write(exif)
		if err != nil {
			log.Fatalf("Failed to write EXIF to %s: %v\n", fout, err)
		}
		fmt.Printf("Wrote EXIF to %s\n", fout)
	}
	err = jpeg.Encode(w, img, &jpeg.Options{Quality: 100})
	if err != nil {
		log.Fatalf("Failed to encode %s: %v\n", fout, err)
	}
}
