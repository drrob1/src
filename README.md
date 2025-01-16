## gocode
All of my go code as of Nov 9, 2024.

1.  sha -- easier way for me to validate d/l file hashes that works the same on linux and windows.  Multisha, consha, csha, fsha and fsha3 do this w/ conconcurrency.
2.  rpn, rpn2, rpng -- RPN style calculator like the HP-25 that I used while in college.  Has CLI.
3.  rpnterm -- Same RPN style calculator but written to use termbox-go.  Not up to date as I switched to rpntcell.
4.  rpnt -- Same RPN style calculator but written to use tcell, but doesn't play nice w/ tcmd by JPSoft
5.  rpnf -- uses fyne to give more of a GUI interface
6.  shufv, shufv2 -- Intended to shuffle the individual files in a vlc .xspf file.  Has CLI.
7.  launchv -- takes a regexp, shuffles the matches and starts vlc w/ the first n files on the command line
8.  lv2 -- takes a regexp, shuffles the matches, writes an .xspf file and starts vlc to use that file.
9.  cal -- Simple CLI program that creates a file for paper output, a year on a page, and a 2nd file to import into Excel to be used to make a schedule.
10. calg -- Displays 6 months of a calendar just using colortext.  Writes schedule template using hardcoded names.
11. calgo -- calendar pgm written in Go based on calg code.  Writes schedule template using a config file for the names.
12. oldcalgo -- calendar pgm written in Go using tcell.  Name implies it's not up to date.
12. caltcell -- calendar pgm written in Go using tcell.  Not up to date.
13. dsrt -- directory sort program.  Sorted by mod timestamp or size.
14. fdsrt -- directory sort pgm sorted by modified timestamp or size, using concurrency to collect filenames
14. ds -- truncated directory sort intended for narrow terminal windows or columns
15. dvfirst -- directory sort program based on dsrt, but using viper for its options
16. dv -- directory sort program based on fdsrt, using viper for its options
15. rex -- uses regular expressions to match the file names.
16. ofx2csv -- takes open financial exchange datafile and writes in csv format.
17. fromfx, fromfx2, queuefx -- takes qfx, ofx or qbo bank file and writes xls, csv, xlsx and directly to sqlite3 .db formats.
18. taxproc -- takes the taxesyy file I create every year and processes the file for my taxes.mdb and taxes.db files.
18. eols -- counts end of line characters.
19. nocr, nocr2 -- removes CR characters.
20. gastricgo, gastric2, gastric3 -- computes gastric emptying T-1/2 given a text inputfile of the times and counts.
21. solve, solve2 -- Linear algebra equation solver.  Rather primitive but works for me.
22. mattest, mattest2, mattest3 -- generate test data for solve and solve2 to use for debugging.
22. showutf8, toascii, utf8toascii, trimtoascii -- show or convert utf8 to straight ascii codepoints.  utf8toascii can also convert line endings.
23. primes, primes2, makeprimesslice -- does prime factoring.
24. Several date convert programs -- needed for sqlite3 formatted datafiles.
25. todo -- text based todo list manager.
26. bj, bj2 -- blackjack simulators.  bj2 uses same deck for all runs, and needs cardshuffler to write the deck
27. cgrepi -- a fast concurrent grep insensitive-case
28. multack -- a concurrent case insensitive ack
29. detox -- a linux-like detox utility for windows
30. dirb -- porting the bashDirB script to windows in tcmd by jpsoft.  Uses makedirbkmk to create its map.
31. feq, few -- file equal pgms mostly comparing different hash functions
32. img, imga, img2 -- GUI pgm to show an image, and then switch easily
33. freq -- letter counting and sorting, for use in wordle
34. goclick, gofshowtime -- to keep a window active at work
35. hideme -- for work, intended for the Aidoc demon.
36. runlst, run, runx -- attempt to emulate executable extension, but starting w/ the data and then launching the correct program.
37. copylist, dellist -- create a list on which copy or del is executed.
38. copyc, copyc1, copyc2, copycv, cf, cf2 -- use same list concept to use concurrency in the copying of selected files.  I primarily use cf2 now.
39. copying, copyingc -- same basic concept but implemented differently.  Command line params are list of output destinations.  Needs flags to specify inputs.


All the other files were either used in testing or are support files for the above programs.
