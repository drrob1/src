package main // for copycv

import (
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"src/list"
)

/*
  10 Apr 23 -- CopyAFile is now separate, and will delete a file if there's an error from the io.copy or Sync()
                I haven't yet seen any errors from Close(), so I'll wait to see what those errors may be to determine what I will then do in the future.
  24 Apr 23 -- I found a bug in copyC that I also have to fix here.  This routine can only return 1 message, because that 1 message decrements the wait group and the main pgm exits.
*/

// CopyAFile                    ------------------------------------ Copy ----------------------------------------------
// CopyAFile(srcFile, destDir string) where src is a regular file.  destDir is a directory
func CopyAFile(srcFile, destDir string) {
	if list.VerboseFlag {
		fmt.Printf(" In CopyAFile.  srcFile is %s, destDir %s.\n", srcFile, destDir)
	}

	in, err := os.Open(srcFile)
	if err != nil {
		msg := msgType{
			s:       "",
			e:       fmt.Errorf("%s", err),
			color:   ct.Red,
			success: false,
		}
		msgChan <- msg
		return
	}
	defer in.Close()

	destD, err := os.Open(destDir)
	if err != nil {
		msg := msgType{
			s:       "",
			e:       err,
			color:   ct.Red,
			success: false,
		}
		msgChan <- msg
		return
	}

	destFI, err := destD.Stat()
	if err != nil {
		msg := msgType{
			s:       "",
			e:       err,
			color:   ct.Red,
			success: false,
		}
		msgChan <- msg
		return
	}
	if !destFI.IsDir() {
		msg := msgType{
			s:       "",
			e:       fmt.Errorf("os.Stat(%s) must be a directory, but it's not c/w a directory", destDir),
			color:   ct.Red,
			success: false,
		}
		msgChan <- msg
		return
	}

	baseFile := filepath.Base(srcFile)
	outName := filepath.Join(destDir, baseFile)
	inFI, _ := in.Stat()
	outFI, err := os.Stat(outName)
	if err == nil { // this means that the file exists.  I have to handle a possible collision now.  I'm ignoring err != nil because that means that file's not already there.
		if !outFI.ModTime().Before(inFI.ModTime()) { // this condition is true if the current file in the destDir is newer than the file to be copied here.
			ErrNotNew = fmt.Errorf(" Skipping %s as it's same or older than destination %s", baseFile, destDir)
			msg := msgType{
				s:       "",
				e:       ErrNotNew,
				color:   ct.Red,
				success: false,
			}
			msgChan <- msg
			return
		}
	}

	out, err := os.Create(outName)
	if err != nil {
		msg := msgType{
			s:       "",
			e:       err,
			color:   ct.Red,
			success: false,
		}
		msgChan <- msg
		return
	}

	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		//msg := msgType{
		//	s:       "",
		//	e:       err,
		//	color:   ct.Red,
		//	success: false,
		//}
		//msgChan <- msg  too soon.  Don't return a message yet.
		var msg msgType

		er := os.Remove(outName)
		if er == nil {
			msg = msgType{
				s:        "",
				e:        fmt.Errorf("ERROR from io.Copy was %s, so os.Remove(%s) was called.  There was no error", err, outName),
				color:    ct.Yellow, // to make sure I see the message.
				success:  false,
				verified: false,
			}
			msgChan <- msg
		} else {
			msg = msgType{
				s:        "",
				e:        fmt.Errorf("ERROR from io.Copy was %s, so os.Remove(%s) was called.  The error from os.Remove was %s", err, outName, er),
				color:    ct.Yellow, // to make sure I see the message
				success:  false,
				verified: false,
			}
			msgChan <- msg
		}
		return
	}

	err = out.Sync()
	if err != nil {
		//msg := msgType{
		//	s:       "",
		//	e:       err,
		//	color:   ct.Magenta,
		//	success: false,
		//}
		//msgChan <- msg  Too soon to return a message.

		var msg msgType

		er := os.Remove(outName)
		if er == nil {
			msg = msgType{
				s:        "",
				e:        fmt.Errorf("ERROR from Sync() was %s, so os.Remove(%s) was called.  There was no error", err, outName),
				color:    ct.Yellow, // to make sure I see this message.
				success:  false,
				verified: false,
			}
			msgChan <- msg
		} else {
			msg = msgType{
				s:        "",
				e:        fmt.Errorf("ERROR from Sync() was %s, so os.Remove(%s) was called.  The error from os.Remove was %s", err, outName, er),
				color:    ct.Yellow, // to make sure I see this message.
				success:  false,
				verified: false,
			}
			msgChan <- msg
		}
		return
	}

	err = out.Close()
	if err != nil {
		msg := msgType{
			s:       "",
			e:       err,
			color:   ct.Red,
			success: false,
		}
		msgChan <- msg
		return
	}
	t := inFI.ModTime()
	if runtime.GOOS == "linux" {
		t = t.Add(timeFudgeFactor)
	}
	err = os.Chtimes(outName, t, t)
	if err != nil {
		msg := msgType{
			s:       "",
			e:       err,
			color:   ct.Red,
			success: false,
		}
		msgChan <- msg
		return
	}

	if verifyFlag {
		vmsg := verifyType{
			srcFile:  srcFile,
			destFile: outName,
			destDir:  destDir, // this is here so the messages can be shorter.
		}
		verifyChan <- vmsg
		return
	}

	msg := msgType{
		s:        fmt.Sprintf("%s copied to %s", srcFile, destDir),
		e:        nil,
		color:    ct.Green,
		success:  true,
		verified: verifyFlag, // this flag must be false by now.
	}
	msgChan <- msg
	//return  this is implied.
} // end CopyAFile
