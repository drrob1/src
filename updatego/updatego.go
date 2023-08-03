package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"src/filepicker"
	"strconv"
	"strings"
)

/*
  REVISION HISTORY
  ----------------
   1 Aug 23 -- Conceived of this routine to update go after the latest file is already downloaded and checked w/ a sha routine.
                Since Windows does itself upon installation, I intend this for linux.  It has to nuke the current Go installation
                and unTar the new one correctly.  I'm going to shell out to a command line for both of these operations.
   2 Aug 23 -- Turned out that for the shelling out to the tar cmd, I had to use a fully qualified path name.  When I didn't do that, I got a file not found error.
                I finally figured out why, because I changed dir to /usr/local and forgot to change back.  Anyway, the code's working so I'll leave it alone.
*/

const lastUpdated = "Aug 2, 2023"

func main() {
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")
	workingDir, e := os.Getwd()
	if e != nil {
		fmt.Printf(" os.Getwd() error is %s\n", e)
		os.Exit(1)
	}

	fmt.Printf(" %s is last altered %s, and has time stamp of %s \n", os.Args[0], lastUpdated, ExecTimeStamp)

	// filepicker stuff.

	var ans string
	var fn string
	if len(os.Args) <= 1 {
		filenames, err := filepicker.GetFilenames("go1*.tar.gz")
		if err != nil {
			fmt.Fprintf(os.Stderr, " Error from filepicker is %v.  Exiting \n", err)
			os.Exit(1)
		}
		for i := 0; i < min(len(filenames), 26); i++ {
			fmt.Printf("filename[%d, %c] is %s\n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice (stop code is 999) : ")
		n, err := fmt.Scanln(&ans)
		if n == 0 || err != nil {
			ans = "0"
		} else if ans == "999" {
			fmt.Println(" Stop code entered.  Exiting.")
			os.Exit(0)
		}
		i, err := strconv.Atoi(ans)
		if err == nil {
			fn = filenames[i]
		} else {
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			fn = filenames[i]
		}
	} else { // will use filename entered on commandline
		//            Filename = getcommandline.GetCommandLineString()  removed 3/3/21, as os.Args is fine.
		fn = os.Args[1]
	}

	if fn == "" {
		fmt.Printf(" Filename is empty.  Exiting.\n")
		os.Exit(1)
	}
	fmt.Printf(" Filename is %s\n\n", fn)

	ans = "" // blank out the above answer.  I got confused about this.  Turns out that when there's no input, the var, ans, is not altered.

	// Now have the filename, need to nuke the old Go installation

	err := os.Chdir("/usr/local") // I forgot to change dir back to where the go1 tarball is.  I probably don't need the -C flag to the tar command.  But this code works, so I'll leave it alone.
	if err != nil {
		fmt.Printf(" Error from os.Chdir(/usr/local) is %s.  Exiting.\n", err)
		os.Exit(1)
	}

	buf := make([]byte, 0, 500_000)
	w1 := bytes.NewBuffer(buf)
	buf2 := make([]byte, 0, 1000)
	w2 := bytes.NewBuffer(buf2)

	var skipNuke bool
	_, err = os.Stat("go/")
	if err != nil {
		fmt.Printf(" There is no old go/ directory in /usr/local, so this step will be skipped.\n")
		skipNuke = true
	}

	//fmt.Printf(" Ok to nuke /usr/local/go? (y/N) ")
	//n, err := fmt.Scanln(&ans)
	//ans = strings.ToLower(ans)
	//fmt.Printf(" Answer is %q about stopping nuking old Go tree, n = %d, and err: %s.\n", ans, n, err)
	//if ans != "y" {
	//	fmt.Printf("\n Not nuking /usr/local/go.\n")
	//	os.Exit(1)
	//}

	if !skipNuke {
		nukeCmd := exec.Command("doas", "rm", "-rfv", "go/") // this is like the JSON command syntax that is used in docker to not need a shell to interpret commands.
		nukeCmd.Stdin = os.Stdin
		nukeCmd.Stdout = w1
		nukeCmd.Stderr = w2
		nukeCmd.Run()
		if w1.Len() < 1000 {
			fmt.Printf(" Output from nuke /usr/local/go is %s\n", w1.String())
		} else {
			fmt.Printf(" Output frum nuke /usr/local/go is %d characters long, which is long enough for me to say it was successful.\n", w1.Len())
			fmt.Printf(" Beginning of output is:\n%s\n", w1.String()[:500])
		}

		if w2.Len() == 0 {
			fmt.Printf(" There were no errors from doas nuke go/\n")
		} else {
			fmt.Printf(" nuke go Stderr: %s \n\n", w2.String())
		}
	}

	// Now have deleted /usr/local/go if it existed, and have the filename to install.  Time to execute tar

	//fmt.Printf("\n Untar %s? (y/N) ", fn)
	//n, err = fmt.Scanln(&ans)
	//ans = strings.ToLower(ans)
	//fmt.Printf(" Answer is %q about untar %s, n = %d, err = %s.\n", ans, fn, n, err)
	//if ans == "n" || n == 0 || err != nil {
	//	fmt.Printf("\n Not continuing.\n")
	//	os.Exit(0)
	//}
	//
	//fmt.Printf(" About to call os.Stat(%s)\n", fn)
	//_, err = os.Stat(fn)
	//if err != nil {
	//	fmt.Printf(" Err from os.Stat(%s) is %s\n", fn, err)
	//}

	fullFileName := workingDir + string(filepath.Separator) + fn

	//fmt.Printf(" About to call os.Stat(%s)\n", fullFileName)
	//_, err = os.Stat(fullFileName)
	//if err != nil {
	//	fmt.Printf(" Err from os.Stat(%s) is %s\n", fullFileName, err)
	//} else {
	//	fmt.Printf(" Looks like os.Stat(%s) worked.\n", fullFileName)
	//}

	tarArg := []string{"tar", "-C", "/usr/local/", "-xzf"}
	tarArg = append(tarArg, fullFileName)
	//                                                                             fmt.Printf(" tarArg = %+v\n", tarArg)

	//fmt.Printf(" Should I actually call tar %s: (y/N) ", fn)
	//n, err = fmt.Scanln(&ans)
	//ans = strings.ToLower(ans)
	//if ans == "n" || n == 0 || err != nil {
	//	fmt.Printf(" Bye-Bye.\n")
	//	os.Exit(1)
	//}

	tarCmd := exec.Command("doas", tarArg...) // this is like the JSON command syntax that is used in docker to not need a shell to interpret commands.

	w1.Reset()
	w2.Reset()

	tarCmd.Stdin = os.Stdin
	tarCmd.Stdout = w1
	tarCmd.Stderr = w2
	tarCmd.Run()

	if w1.Len() == 0 {
		fmt.Printf(" There was no output sent to Stdout from the untar %s command\n", fullFileName)
	} else {
		fmt.Printf(" %q\n was returned in Stdout from the untar %s command \n", w1.String(), fullFileName)
	}
	if w2.Len() == 0 {
		fmt.Printf(" There was no output sent to Stderr from the untar %s command\n", fullFileName)
	} else {
		fmt.Printf(" %q\n was returned in Stderr from the untar %s command \n", w2.String(), fullFileName)
	}

	fmt.Printf("\n Bye-Bye.  Hope it worked.\n\n")
}

// ------------------------------------------- min -------------------------------------------------------------------------

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
