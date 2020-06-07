/*
 Copyright 2020 Qiniu Cloud (qiniu.com)

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/qiniu/goc/pkg/build"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Do cover for all go files and execute go build command",
	Long: `
First of all, this build command will copy the project code and its necessary dependencies to a temporary directory, then do cover for the target in this temporary directory, finally go build command will be executed and binaries generated to their original place.

To pass original go build flags to goc command, place them after "--", see examples below for reference.
	`,
	Example: `
# Build the current binary with cover variables injected. The binary will be generated in the current folder.
goc build

# Build the current binary with cover variables injected, and set the registry center to http://127.0.0.1:7777.
goc build --center=http://127.0.0.1:7777 

# Build the current binary with cover variables injected, and redirect output to /to/this/path.
goc build -- -o /to/this/path

# Build the current binary with cover variables injected, and set necessary build flags: -ldflags "-extldflags -static" -tags="embed kodo".
goc build -- -ldflags "-extldflags -static" -tags="embed kodo"
`,
	Run: func(cmd *cobra.Command, args []string) {
		newgopath, newwd, tmpdir, pkgs := build.MvProjectsToTmp(target, args)
		doCover(cmd, args, newgopath, tmpdir)
		newArgs, modified := modifyOutputArg(args)
		doBuild(newArgs, newgopath, newwd)

		// if not modified
		// find the binary in temp build dir
		// and copy them into original dir
		if false == modified {
			build.MvBinaryToOri(pkgs, tmpdir)
		}
	},
}

func init() {
	buildCmd.Flags().StringVarP(&center, "center", "", "http://127.0.0.1:7777", "cover profile host center")

	rootCmd.AddCommand(buildCmd)
}

func doBuild(args []string, newgopath string, newworkingdir string) {
	log.Println("Go building in temp...")
	newArgs := []string{"build"}
	newArgs = append(newArgs, args...)
	cmd := exec.Command("go", newArgs...)
	cmd.Dir = newworkingdir

	if newgopath != "" {
		// Change to temp GOPATH for go install command
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%v", newgopath))
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Fail to execute: go build %v. The error is: %v, the stdout/stderr is: %v", strings.Join(args, " "), err, string(out))
	}
	log.Println("Go build exit successful.")
}

// As we build in the temp build dir, we have to modify the "-o output",
// if output is a relative path, transform it to abspath
func modifyOutputArg(args []string) (newArgs []string, modified bool) {
	var output string
	fs := flag.NewFlagSet("goc-build", flag.PanicOnError)
	fs.StringVar(&output, "o", "", "output dir")

	// parse the go args after "--"
	fs.Parse(args)

	// skip if output is not present
	if output == "" {
		modified = false
		newArgs = args
		return
	}

	abs, err := filepath.Abs(output)
	if err != nil {
		log.Fatalf("Fail to transform the path: %v to absolute path, the error is: %v", output, err)
	}

	// the second -o arg will overwrite the first one
	newArgs = append(args, "-o", abs)
	modified = true
	return
}
