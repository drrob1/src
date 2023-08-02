package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
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
*/

const lastUpdated = "Aug 2, 2023"

func main() {
	execName, _ := os.Executable()
	ExecFI, _ := os.Stat(execName)
	ExecTimeStamp := ExecFI.ModTime().Format("Mon Jan-2-2006_15:04:05 MST")

	fmt.Printf(" %s is last altered %s, and has time stamp of %s\n ", os.Args[0], lastUpdated, ExecTimeStamp)

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
		fmt.Print(" Enter filename choice : ")
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

	// Now have the filename, need to nuke the old Go installation

	err := os.Chdir("/usr/local")
	if err != nil {
		fmt.Printf(" Error from os.Chdir(/usr/local) is %s.  Exiting.\n", err)
		os.Exit(1)
	}

	// Define buf, w1, buf2 and w2 here in case the code jumps over the nukeCmd section

	buf := make([]byte, 0, 1000)
	w1 := bytes.NewBuffer(buf)
	buf2 := make([]byte, 0, 1000)
	w2 := bytes.NewBuffer(buf2)

	var skipNuke bool
	_, err = os.Stat("go/")
	if err != nil {
		fmt.Printf(" There is no old go/ directory in /usr/local, so this step will be skipped.")
		skipNuke = true
	}

	if !skipNuke {
		nukeCmd := exec.Command("doas", "rm", "-rfv", "go/") // this is like the JSON command syntax that is used in docker to not need a shell to interpret commands.
		nukeCmd.Stdin = os.Stdin
		nukeCmd.Stdout = w1
		nukeCmd.Stderr = w2
		nukeCmd.Run()
		str := w1.String()
		s := strings.ReplaceAll(str, "\n", "")
		s = strings.ReplaceAll(s, "\r", "")
		s = strings.ReplaceAll(s, ",", "")
		fmt.Printf(" %q was returned in Stdout from the nuke /usr/local/go command, which was processed to %s\n", str, s)

		str2 := w2.String()
		s2 := strings.ReplaceAll(str2, "\n", "")
		s2 = strings.ReplaceAll(s2, "\r", "")
		s2 = strings.ReplaceAll(s2, ",", "")
		fmt.Printf(" %q was returned in Stderr from the nuke /usr/local/go command, which was processed to %s\n", str2, s2)
	}

	// Now have deleted /usr/local/go if it existed, and have the filename to install.  Time to execute tar

	tarCmd := exec.Command("tar", "-C", "/usr/local", "-xzf", "fn") // this is like the JSON command syntax that is used in docker to not need a shell to interpret commands.

	w1.Reset()
	w2.Reset()

	tarCmd.Stdin = os.Stdin
	tarCmd.Stdout = w1
	tarCmd.Stderr = w2
	tarCmd.Run()
	str := w1.String()
	s := strings.ReplaceAll(str, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, ",", "")
	fmt.Printf(" %q was returned in Stdout from the nuke /usr/local/go command, which was processed to %s\n", str, s)

	str2 := w2.String()
	s2 := strings.ReplaceAll(str2, "\n", "")
	s2 = strings.ReplaceAll(s2, "\r", "")
	s2 = strings.ReplaceAll(s2, ",", "")
	fmt.Printf(" %q was returned in Stderr from the nuke /usr/local/go command, which was processed to %s\n", str2, s2)
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
