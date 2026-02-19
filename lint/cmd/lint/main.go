package main // lint.go, from lint2.go from lint.go

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"src/lint"
	"src/whichexec"
	"strconv"
	"strings"

	ct "github.com/daviddengcn/go-colortext"
	ctfmt "github.com/daviddengcn/go-colortext/fmt"
	flag "github.com/spf13/pflag"
	//"flag"
)

/*
  26 Sep 24 -- Started first version.  Intended as a linter for the weekly work schedule.  It will need a .conf or .ini file to establish the suitable params.
               It will need lists to specify who can be covering a specific area, and to make sure that if someone is on vacation, their name does not appear anywhere else
               for that day.  So I'll need categories in the .conf or .ini file, such as:
				weekdayOncall row 3
				neuro row 4
				body row 5
				ER row 6
				Xrays row 6
				IR row 7
				Nuclear row 8
				US row 9
				Pediatrics row 10
				FLUORO JH row 11
				FLUORO FH row 12
				MSK row 13
				MAMMO row 14
				BONE (DENSITY) row 15
				LATE row 16
				Moonlighters row 17
				Weekend JH row 18
				Weekend FH row 19
				Weekend IR row 20
				MD's Off (vacation) row 21
				Below row 21 are the MD phone #'s.

				if the line begins with any of [# ; /] then it's a comment.  If a line doesn't begin w/ a keyword, then it's an error and the pgm exits.
				I think I'll just check the vacation rule first.  Then expand it to the other rules.

                I have to read the weekly schedule into an appropriate data structure, as also the .conf/.ini file.

 xlsx (github.com/tealeg/xlsx/v3)
----------------------------------------------------------------------------------------------------
  28 Jan 25 -- Now called lint2, and added detection of having the late doc also be on fluoro.  That happened today at Flushing, and it was a mistake.
				I think I can do it without much difficulty.
  29 Jan 25 -- I'm going to make the week a 2D array and use a map to get the names from the row #.  Just to see if it will work.  It does.
				Now I'm going to try to see if a remote doc is on fluoro, like there was today.
  30 Jan 25 -- And I shortened the main loop looking for vacation docs assigned to clinical work.
----------------------------------------------------------------------------------------------------
  31 Jan 25 -- Renamed back to lint.go
   2 Feb 25 -- Made dayNames an array instead of a slice.  It's fixed, so it doesn't need to be a slice.
  14 Mar 25 -- Today is Pi Day.  But that's not important now.
				I want to refactor this so it works in the environment it's needed.  It needs to get the files from o: drive and then homeDir/Documents, both.
				So I want to write the routine here as taking a param of a full filename and scanning that file.
				First I want to see if the xlsx.OpenFile takes a full file name as its param.  If so, that'll be easier for me to code.  It does.
  16 Mar 25 -- Changing colors that are displayed.
  18 Mar 25 -- Still doesn't work for Nikki, as it doesn't find the files on O-drive.  I'll broaden the expression to include all Excel files.
  20 Mar 25 -- 1.  I need a switch to only search c:, for my use.  I'll call this conly, apprev as c.
               2.  Nikki uses a much more complex directory structure on o-drive than I expected.  I think I'm going to need a walk function to search for all files
					timestamped this month or next month, and add their full name to the slice, sort the slice by date, newest first.
				And changed to use pflag.
  22 Mar 25 -- It now works as intended.  So now I want to add a flag to set the time interval that's valid, and a config file setting for the directory searched on o:
				And I'll add a veryverboseFlag, using vv and w.  The veryverbose setting will be for the Excel tests I don't need anymore.
  27 Mar 25 -- I ran this using the -race flag; it's clean.  No data races.  Just checking.
   9 Apr 25 -- Made observation that since the walk function sorts its list, the first file that doesn't meet the threshold date can stop the search since the rest are older.
  10 Apr 25 -- Made the change by adding an else clause to an if statement in the walk function.
   8 May 25 -- Fixed the help message.
  26 May 25 -- When I installed this on Caity's computer, I got the idea that I should filter out the files that Excel controls, i.e., those that begin w/ ~, tilda.
  31 May 25 -- Changed an error message.
   2 Aug 25 -- Completed getdocnames yesterday which extracts the names from the schedule itself.  This way I don't need to specify them in the config file, making this routine
				more robust.  I'm going to add it.  I have to ignore the line of the config file that begins w/ "off".  I have to check what the code does if there is no
				startdirectory line in the config file, or the line has invalid syntax (like not beginning w/ the correctly spelled keyword).
				Processing the config file used to do so by using global var's.  I'm changing that to use params.  This way, I can ignore a return param if I want to.
				I'm tagging this as lint v2.0
   3 Aug 25 -- walk function will skip .git
   4 Aug 25 -- Our 40th Anniversary.  But that's not important now.  I'm using soundex codes to report likely spelling errors so they can be fixed.
   6 Aug 25 -- I found out today that the hospital will retire the o: drive, in favor of OneDrive.  I'll need to change the code to use OneDrive.
               There's an environment varible called OneDrive that is set to the path of OneDrive.  And another one called OneDriveConsumer.
               At work, there's OneDriveCommercial, which is set to the same value as OneDrive.  This is also true at home in that OneDrive and OneDriveConsumer have the same value.
				I first coded this to use a filepicker function, but that doesn't exclude old files.  The walk function will skip files that are older than the threshold.
				I need to modify the walk function to take a param that is the start directory, and then combine the results of all the walk function calls.
                And O: drive is going away at work as of Aug 8, 2025.  I'll need to change the code to use OneDrive.  I'll remove the conly flag as it's not needed now.
                I tagged this version that knows about OneDrive, and auto-updating, as lint v2.1.
  11 Aug 25 -- Time to add the code to autoupdate.
  12 Aug 25 -- If this is run w/ the verboseFlag, I'll pass that to upgradelint.
  16 Aug 25 -- Fixed an error in a param message.  And will use workingDir to run upgradelint.  And add flags to use the other websites as backup, which have to get passed to
				upgradelint.
  17 Aug 25 -- Clarified a comment to the walk function, saying that it skips files that begin w/ a tilda, ~.
				And change behavior of walk function so that veryverbose is needed for it to display the walk function's output.
  22 Aug 25 -- Added exclusion of "ra" as it seems that the schedule now includes Radiology Assistant initials for Murina and Payal.
  24 Aug 25 -- Replaced excludeMe using new code I tested in getDocNames.  It uses slices of strings to define the strings to exclude.  And it makes it much easier to add new strings.
  26 Aug 25 -- Added "dr" to excludeMe string, which occurs when the period is forgotten.  And added the -1 and -2 shortcuts that are passed to upgradelint.
				This code is now saved in lintprior1sep25.go
------------------------------------------------------------------------------------------------------------------------------------------------------
    Lint v 3.0 coming up soon.
   1 Sep 25 -- The department changed the format for the schedule, highlighting on call and weekend docs.  I'll need to change the code to use the new format.
   8 Sep 25 -- I got them to add back an indication of who's off, but it's all in 1 box in the Friday column.  I have to think about this, and wonder about it changing again.
				And the numbering changed.  I have to move weekdayOncall, which used to be at the top, and is now near the bottom.  Basically, I have to completely change
				whosOnVacationToday.  When I do that, then the rest of the code should be ok.  I already changed some of the const names for the sections to be covered.
				Previously, scanXLSfile processes the file and displays the error messages.  Main sets up the data structures leading into scanXLSfile.  scanXLSfile was able to
				read each day separately, which worked before.  Now, the off data is in Friday's column.  So would have to read the whole file, and then process the off data.
   9 Sep 25 -- I'm coming back to do these changes in lint.go itself.  I'll stop using extractoff.go.  My plan is to create the vacation string slice in the format that it used to be in,
				and then pass that to whosOnVacationToday.  And then I'll have to change the code to use the new format.
				Currently, whosOnVacationToday is returning a slice of strings just for that day of current interest.  I'll need to return this, but do it differently.  I don't know how, yet.
  11 Sep 25 -- My plan is to populate a vacationStructSlice with the data from docsOffStringForWeek.  I need year from index 2 from each dayType.
                 The populateVacStructSlice function is working.  And now the vacation scanning, late doc on fluoro, and remote doc on fluoro are all working.  Hooray!
  12 Sep 25 -- I got the format Greg and Carol made working last night.  And, as I suspected would happen, it was changed this morning.  Anyway, I'm glad I got it working as it was a challenge for me.
                 Since the new format is very similar to the original format, I think it will be easy to implement that.  It was.
                 I'll tag this lint v3.0 in git when I'm comfortable that I won't need v3.0.1, etc.
------------------------------------------------------------------------------------------------------------------------------------------------------
   1 Feb 26 -- I'm starting to think about using fyne to make this a GUI pgm.  I may have to convert much of this code to functions that are merely connected together in main.go,
				which will be in ./cmd/main.go.  The code works as is.  Time to refactor to package lint and have package main be in main.go separately.
------------------------------------------------------------------------------------------------------------------------------------------------------
   2 Feb 26 -- Now only contains main() and calls package lint.  A sticking point in the conversion to separate pacckages was the globals.  When I copied those correctly, the code started working.
   3 Feb 26 -- In the process of writing lintGUI, I had to change the scanXLSfile function to return a slice of messages.  So now I'm refactoring here to test them.
   7 Feb 26 -- Changed GetFilenames to GetScheduleFilenames.  Also changed when the schedule filenames are sorted.  Now, the entire returns slice of names is sorted by date stamp.
  13 Feb 26 -- Added rowOffset to allow for a date row in the schedule to the lint library.  Here, I added a message for the lint library last modified date.
  18 Feb 26 -- Debugging problem w/ not finding upgradelint.exe
*/

const lastModified = "18 Feb 2026"

//const conf = "lint.conf"
//const ini = "lint.ini"
//const numOfDocs = 40 // used to dimension a string slice.
//const maxDimensions = 200
//var numLines = 15 // I don't expect to need more than these, as I display only the first 26 elements (a-z) so far.

var verboseFlag bool
var veryVerboseFlag bool
var monthsThreshold int
var startDirFromConfigFile string // this needs to be a global, esp for the walk function.

func main() {
	var err error
	var noUpgradeLint bool
	var whichURL int
	flag.BoolVarP(&verboseFlag, "verbose", "v", false, "verbose debugging output")
	flag.BoolVarP(&veryVerboseFlag, "vv", "w", false, "very verbose debugging output")
	flag.IntVarP(&monthsThreshold, "months", "m", 1, "months threshold for schedule files")
	flag.BoolVarP(&noUpgradeLint, "noupgrade", "n", false, "do not upgrade lint.exe")
	flag.IntVarP(&whichURL, "url", "u", 0, "which URL to use for the auto updating of lint.exe")
	u1 := flag.BoolP("u1", "1", false, "Shortcut for -u 1")
	u2 := flag.BoolP("u2", "2", false, "Shortcut for -u 2")

	flag.Usage = func() {
		fmt.Printf(" %s last modified main.go %s and lint.go %s, compiled with %s, using pflag.\n", os.Args[0],
			lastModified, lint.LastModified, runtime.Version())
		fmt.Printf(" Usage: %s <weekly xlsx file> \n", os.Args[0])
		fmt.Printf(" Looks for lint.conf or lint.ini in current, home and config directories.\n")
		fmt.Printf(" First line must begin with off, and 2nd line, if present, must begin with startdirectory.\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if veryVerboseFlag {
		verboseFlag = true
	}

	if *u1 {
		whichURL = 1
	} else if *u2 {
		whichURL = 2
	}

	var filename, ans string

	fmt.Printf(" lint V 3.1 for the weekly schedule, last modified %s, last modified lint library %s\n", lastModified, lint.LastModified)

	_, startDirFromConfigFile, err = lint.FindAndReadConfIni() // ignore the doc names list from the config file, as that's now extracted from the schedule itself.
	if err != nil {
		if verboseFlag { // only show this message if verbose flag is set.  Otherwise, it's too much.
			fmt.Printf(" Warning from findAndReadConfIni: %s.  Ignoring. \n", err)
			ctfmt.Printf(ct.Red, true, " Warning message from findAndReadConfINI: %s. \n", err)
			//   return  No longer need the names from the file.  And don't absolutely need startDirectory.
		}
	}
	if verboseFlag {
		fmt.Printf(" After findAndReadConfIni, Start Directory: %s, NArg(): %d\n", startDirFromConfigFile, flag.NArg())
	}

	lint.VeryVerboseFlag = veryVerboseFlag
	lint.VerboseFlag = verboseFlag
	lint.StartDirFromConfigFile = startDirFromConfigFile
	lint.MonthsThreshold = monthsThreshold
	whichexec.VerboseFlag = verboseFlag

	if flag.NArg() == 0 {
		//if includeODrive {  O: drive is gone as of 8/8/25.
		//	filenames, err = walkRegexFullFilenames() // function is below.  "o:\\week.*xls.?$"
		//	if err != nil {
		//		ctfmt.Printf(ct.Red, false, " Error from walkRegexFullFilenames is %s.  Exiting \n", err)
		//		return
		//	}
		//	if *verboseFlag {
		//		fmt.Printf(" Filenames length from o drive: %d\n", len(filenames))
		//	}
		//}

		//if startDirFromConfigFile != "" {
		//	filenamesStartDir, err := walkRegexFullFilenames(startDirFromConfigFile)
		//	if err != nil {
		//		ctfmt.Printf(ct.Red, false, " Error from walkRegexFullFilenames(%s) is %s.  Ignoring. \n",
		//			startDirFromConfigFile, err)
		//	}
		//	if len(filenamesStartDir) > 0 {
		//		filenames = append(filenames, filenamesStartDir...)
		//		if *verboseFlag {
		//			fmt.Printf(" Filenames length after append %s: %d\n", filenamesStartDir, len(filenames))
		//		}
		//	}
		//	if *verboseFlag {
		//		fmt.Printf(" Filenames length after append %s: %d\n", filenamesStartDir, len(filenames))
		//	}
		//}
		//
		//homeDir, err = os.UserHomeDir()
		//if err != nil {
		//	fmt.Printf(" Error from os.UserHomeDir: %s\n", err)
		//	return
		//}
		////                               docs = filepath.Join(filepath.Join(homeDir, "Documents"), "week.*xls.?$")  Don't want the regex as part of this expression.
		//docs = filepath.Join(homeDir, "Documents") // this walks this directory below to collect filenames
		//if *verboseFlag {
		//	fmt.Printf(" homedir=%q, Joined Documents: %q\n", homeDir, docs)
		//}
		//oneDriveString := os.Getenv("OneDrive")
		//if *verboseFlag {
		//	fmt.Printf(" oneDriveString = %s  \n", oneDriveString)
		//}
		//filenamesOneDrive, err := walkRegexFullFilenames(oneDriveString)
		//if err != nil {
		//	fmt.Printf(" Error from walkRegexFullFilenames(%s) is %s.  Ignoring \n", oneDriveString, err)
		//}
		//if *verboseFlag {
		//	fmt.Printf(" FilenamesDocs length: %d\n", len(filenamesOneDrive))
		//}
		//
		//filenames = append(filenames, filenamesOneDrive...)
		//if *verboseFlag {
		//	fmt.Printf(" Filenames length after append %s: %d\n", filenamesOneDrive, len(filenames))
		//}
		//
		//filenamesDocs, err := walkRegexFullFilenames(docs)
		//if err != nil {
		//	fmt.Printf(" Error from walkRegesFullFilenames(%s) is %s.  Ignored \n", docs, err)
		//} else {
		//	filenames = append(filenames, filenamesDocs...)
		//	if *verboseFlag {
		//		fmt.Printf(" Filenames length after append %s: %d\n", docs, len(filenames))
		//	}
		//}
		if verboseFlag {
			fmt.Printf("main ln ~257: VerboseFlag is true, NArg() is zero, and startdirectoryFromConfigFile = %s, about to call GetFilenames() \n",
				startDirFromConfigFile)
		}

		filenames, err := lint.GetScheduleFilenames()
		if err != nil {
			fmt.Printf(" Error from GetFilenames is %s.  Exiting \n", err)
			return
		}

		for i := 0; i < min(len(filenames), 26); i++ {
			fmt.Printf("filename[%d, %c] is %s\n", i, i+'a', filenames[i])
		}
		fmt.Print(" Enter filename choice : ")
		n, err := fmt.Scanln(&ans)
		if n == 0 || err != nil {
			ans = "0"
		} else if ans == "999" || ans == "." || ans == "," || ans == ";" {
			fmt.Println(" No files entered.  Exiting.")
			return
		}
		i, e := strconv.Atoi(ans)
		if e == nil {
			filename = filenames[i]
		} else {
			s := strings.ToUpper(ans)
			s = strings.TrimSpace(s)
			s0 := s[0]
			i = int(s0 - 'A')
			filename = filenames[i]
		}
		fmt.Println(" Picked spreadsheet is", filename)
	} else { // will use filename entered on commandline
		filename = flag.Arg(0)
	}

	if verboseFlag {
		fmt.Printf(" spreadsheet picked is %s\n", filename)
	}
	fmt.Println()

	names, err := lint.GetDocNames(filename)
	if err != nil {
		ctfmt.Printf(ct.Red, true, " Error from getDocNames: %s.  Exiting \n", err)
		return
	}
	if verboseFlag {
		fmt.Printf(" doc names extracted from %s length: %d\n", filename, len(names))
		fmt.Printf(" names: %#v\n\n", names)
	}
	lint.Names = names

	// detecting and reporting likely spelling errors based on the soundex algorithm

	soundx := lint.GetSoundex(names)
	spellingErrors := lint.ShowSpellingErrors(soundx)
	if len(spellingErrors) > 0 {
		ctfmt.Printf(ct.Cyan, true, "\n\n %d spelling error(s) detected in %s: ", len(spellingErrors)/2, filename)
		for _, spell := range spellingErrors {
			ctfmt.Printf(ct.Red, true, " %s  ", spell)
		}
		fmt.Printf("\n\n\n")
	}

	// scan the xlsx schedule file

	messages, err := lint.ScanXLSfile(filename)
	if len(messages) > 0 {
		ctfmt.Printf(ct.Cyan, true, "\n\n %d message(s) generated from %s: \n", len(messages), filename)
		for _, msg := range messages {
			ctfmt.Printf(ct.Yellow, true, " %s \n", msg)
		}
	}

	if err == nil {
		ctfmt.Printf(ct.Green, true, "\n\n Finished scanning %s\n\n", filename)
	} else {
		ctfmt.Printf(ct.Red, true, "\n\n Error scanning %s is %s\n\n", filename, err)
		return
	}

	if noUpgradeLint {
		return
	} // this flag is a param above.

	// Time to run the updatelist cmd.

	workingDir, err := os.Getwd()
	if err != nil {
		ctfmt.Printf(ct.Red, true, "\n\n Error getting working directory: %s.  Contact Rob Solomon\n\n", err)
		return
	}
	// fullUpgradeLintPath := filepath.Join(workingDir, "upgradelint.exe") // not needed now that I'm using whichexec to find it.
	upgradeExecPath := whichexec.Find("upgradelint.exe", workingDir)
	if verboseFlag {
		fmt.Printf(" workingDir=%s, upgradeExecPath=%s\n", workingDir, upgradeExecPath)
	}

	variadicArgs := make([]string, 0, 2)
	if verboseFlag {
		variadicArgs = append(variadicArgs, "-v")
	}
	if whichURL > 0 {
		variadicArgs = append(variadicArgs, "-u", strconv.Itoa(whichURL))
	}

	// execcmd := exec.Command(fullUpgradeLintPath, variadicArgs...)
	execcmd := exec.Command(upgradeExecPath, variadicArgs...)
	execcmd.Stdin = os.Stdin
	execcmd.Stdout = os.Stdout
	execcmd.Stderr = os.Stderr
	err = execcmd.Start()
	if err != nil {
		ctfmt.Printf(ct.Red, true, "\n\n Error starting upgradelint: %s.  Contact Rob Solomon\n\n", err)
	}
}
