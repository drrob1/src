# Context
Write an analyzer in Go for an excel spreadsheet in .xlsx format looking for 3 types of
errors

# Task
- this is a CLI greenfield application
- interview me until the scope is concrete
- scan the current directory with recursion, looking for all .xlsx files that contain
  the string "week", sorted by date with the newest file first
- show me the filenames from the last 30 days to allow me to select the file to
  analyze
- Error 1 to search for is to see if the names in the row labeled "MDs out of office"
  also appear in any other cells in that column
- Error 2 to search within a column to see if any names marked with "(*R)" also
  appear in a row labeled "FLUORO FH" or "FLUORO JH"
- Error 3 to search within a column to see if the names in the cell labeled "Late MD"
  are also in a row labeled "FLUORO FH" or "FLUORO JH"
- A sample file is part of this directory

# Constraints
- produce a plan with the smallest sensible architecture
- show me the main risks and assumptions
- do not write code yet
- do not add features beyond the scope we agree on
- keep the plan minimal and specific to this repository
