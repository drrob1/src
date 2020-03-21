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

// The approach taken here was inspired by an example on the gonuts mailing
// list by Roger Peppe.
/*
  REVISION HISTORY
  ----------------
  20 Mar 20 -- Made comparisons case insensitive.  And decided to make this cgrepi.go.
                 And then I figured I could not improve performance by using more packages.
                 But I can change the side effect of displaying altered case.
  21 Mar 20 -- Now called testgogrep, to test the separate module I wrote.  So most of the original code has been moved to gogrep.
                 It works, but is not helpful.  I had to rethink the problem and anack is the result.  And this works.
*/
package main

import (
	"flag"
	"fmt"
	"gogrep"
	"log"
	"strings"
)

const LastAltered = "21 Mar 2020"

func main() {
	//	runtime.GOMAXPROCS(runtime.NumCPU()) // not needed here
	log.SetFlags(0)
	var timeoutOpt *int = flag.Int("timeout", 0, "seconds < 240, where 0 means no timeout.")
	flag.Parse()
	if *timeoutOpt < 0 || *timeoutOpt > 240 {
		log.Fatalln("timeout must be in the range [0,240] seconds")
	}
	args := flag.Args()
	if len(args) < 1 {
		log.Fatalln("a regexp to match must be specified")
	}
	pattern := args[0]
	pattern = strings.ToLower(pattern) // this is the change for the pattern.
	files := args[1:]
	if len(files) < 1 {
		log.Fatalln("must provide at least one filename")
	}

	fmt.Println(" Concurrent grep insensitive case last altered", LastAltered, ".")
	fmt.Println()
	fmt.Println()

	err := gogrep.Gogrep(pattern, files, *timeoutOpt) // this fails vet because it's in the platform specific code files.
	if err != nil {
		log.Fatalf(" Error from gogrep is %v\n", err)
	}
}
