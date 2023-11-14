package unzipanddel

import (
	"fmt"
	"github.com/evilsocket/islazy/zip"
	"os"
)

/*
REVISION HISTORY
-------- -------
13 Nov 23 -- Started working on the first version of this pgm.
*/

const lastModified = "17 Nov 2023" // I doubt this will be finished quickly.

func unzipAndShow(src, dest string) error {
	filenames, err := zip.Unzip(src, dest)
	if err != nil {
		return err
	}
	fmt.Printf(" filenames: %+v\n", filenames)
	return nil
}

func unzipAndDel(src, dest string) error {
	_, err := zip.Unzip(src, dest)
	if err != nil {
		return err
	}
	err = os.Remove(src)
	return err
}

/*
github.com/evilsocket/islazy/zip

func Unzip(src string, dest string) ([]string, error)
Unzip will decompress a zip archive, moving all files and folders within the zip file (parameter 1) to an output directory (parameter 2).

package main

import (
	"fmt"
	"github.com/evilsocket/islazy/zip"
)

func main() {
	if err := zip.Files("archive.zip", []string{"README.md", "release.sh"}); err != nil {
		panic(err)
	}

	if files, err := zip.Unzip("archive.zip", "./dest"); err != nil {
		panic(err)
	} else {
		fmt.Println(files)
	}
}
*/
