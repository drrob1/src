## gocode
All of my go code as of Feb 5, 2023.

1. sha -- easier way for me to validate d/l file hashes that works the same on linux and windows.  Multisha and consha do this w/ conconcurrency.
2. rpn, rpn2, rpng -- RPN style calculator like the HP-25 that I used while in college.  Has CLI.
3. rpnterm -- Same RPN style calculator but written to use termbox-go.  Not up to date as I switched to rpntcell.
4. rpntcell -- Same RPN style calculator but written to use tcell, but doesn't play nice w/ tcmd by JPSoft
5. rpnf -- uses fyne to give more of a GUI interface
6. vlc -- Intended to shuffle the individual files in a vlc .xspf file.  Has CLI.
7. cal -- Simple CLI program that creates a file for paper output, a year on a page, and a 2nd file to import into Excel to be used to make a schedule.
8. calg -- Displays 6 months of a calendar just using colortext.
9. calgo -- calendar pgm written in Go using termbox-go.  Not up to date.
10. caltcell -- calendar pgm written in Go using tcell.
11. dsrt -- directory sort program.  Sorted by timestamp or size.
12. ds -- truncated directory sort intended for narrow terminal windows or columns
13. rex -- uses regular expressions to match the file names.
14. ofx2csv -- takes open financial exchange datafile and writes in csv format.
15. fromfx, fromfx2, queuefx -- takes qfx, ofx or qbo bank file and writes xls and csv formats.
16. eols -- counts end of line characters.
17. nocr -- removes CR characters.
18. gastricgo, gastric2, gastric3 -- computes gastric emptying T-1/2 given a text inputfile of the times and counts.
19. solve -- Linear algebra equation solver.  Rather primitive but works for me.
20. showutf8, toascii, utf8toascii, trimtoascii -- show or convert utf8 to straight ascii codepoints.
                                                   utf8toascii can also convert line endings.
21. primes, primes2, makeprimesslice -- does prime factoring.
22. Several date convert programs -- needed for sqlite3 formatted datafiles.\
23. todo -- text based todo list manager.
24. bj, bj2 -- blackjack simulators.  bj2 uses same deck for all runs, and needs cardshuffler to write the deck
25. cgrepi -- my stab at a fast concurrent grep insensitive-case
26. multack -- my stab at a concurrent case insensitive ack
27. detox -- my stab at the linux detox utility intended for windows
28. dirb -- by stab at porting the bashDirB script to windows in tcmd by jpsoft
29. feq, few -- file equal pgms mostly comparing different hash functions
30. img, imga, img2 -- GUI pgm to show an image, and then switch easily
31. freq -- letter counting and sorting, for use in wordle
32. goclick, gofshowtime -- for work to keep a window active
33. hideme -- for work, intended for the Aidoc demon.
34. launchv, lvr -- used to start vlc with a random list of files on its command line.
35. copylist, dellist -- create a list on which copy or del is executed.
36. copyc, copyc2 -- use same list concept to use concurrency in the copying of selected files.
37. copying -- same basic concept but implemented differently.  Command line params are list of output destinations.  Needs flags to specify inputs.


All the other files were either used in testing or are support files for the above programs.
