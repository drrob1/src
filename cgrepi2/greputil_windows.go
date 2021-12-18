// Copyright © 2011-12 Qtrac Ltd.
//
// This program or package and any associated files are licensed under the
// Apache License, Version 2.0 (the "License"); you may not use these files
// except in compliance with the License. You can get a copy of the License
// at: http://www.apache.org/licenses/LICENSE-2.0.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// REVISION HISTORY
// -------- -------
// 27 Mar 21 -- The routine is not working as hoped.  I think the issue is this routine not finding all needed filenames
//               So I'm re-writing it.  I'll not use Glob as that is case sensitive.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
) // Glob is case sensitive.  I want case insensitive.

const estimatedNumberOfFiles = 100

func globCommandLineFiles(patterns []string) []string {
	matchingNames := make([]string, 0, estimatedNumberOfFiles)
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern) // Glob returns names of all files matching the case-sensitive pattern.
		if err != nil {
			fmt.Fprintln(os.Stderr, " Error from filepath.Glob is", err)
			os.Exit(1)
		} else if matches != nil { // At least one match
			matchingNames = append(matchingNames, matches...)
		}
	}
	return matchingNames
}

func commandLineFiles(patterns []string) []string {
	workingDirname, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "from commandlinefiles:", err)
		return nil
	}
	direntries, err := os.ReadDir(workingDirname) // became available as of Go 1.16
	if err != nil {
		return nil
	}

	matchingNames := make([]string, 0, len(direntries))

	for _, pattern := range patterns { // outer loop to test against multiple patterns.
		for _, d := range direntries { // inner loop to test each pattern against the filenames.
			if d.IsDir() {
				continue // skip a subdirectory name
			}
			bool, err := filepath.Match(pattern, strings.ToLower(d.Name()))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if bool {
				matchingNames = append(matchingNames, d.Name())
			}
		}
	}
	return matchingNames
}
