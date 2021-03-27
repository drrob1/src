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

func globCommandLineFiles(files []string) []string { // orig code w/ new name
	args := make([]string, 0, len(files))
	for _, name := range files {
		if matches, err := filepath.Glob(name); err != nil {
			args = append(args, name) // Invalid pattern
		} else if matches != nil { // At least one match
			args = append(args, matches...)
		}
	}
	return args
}

func commandLineFiles(files []string) []string {
	workingDirname, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "from commandlinefiles:", err)
		return nil
	}
	dirname, err := os.Open(workingDirname)
	if err != nil {
		return nil
	}
	defer dirname.Close()

	names, err := dirname.Readdirnames(0) // zero means read all names into the returned []string
	if err != nil {
		return nil
	}

	pattern := strings.ToLower(files[0])
	matchingNames := make([]string, 0, len(names))
	for _, s := range names {
		bool, err := filepath.Match(pattern, strings.ToLower(s))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		if bool {
			matchingNames = append(matchingNames, s)
		}
	}
	return matchingNames
}