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
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/qiniu/goc/pkg/build"
	"github.com/qiniu/goc/pkg/cover"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Do cover for all go files and execute go build command",
	Long: `
Build command will copy the project code and its necessary dependencies to a temporary directory, then do cover for the target, binaries will be generated to their original place.
`,
	Example: `
# Build the current binary with cover variables injected. The binary will be generated in the current folder.
goc build .

# Build the current binary with cover variables injected, and set the registry center to http://127.0.0.1:7777.
goc build --center=http://127.0.0.1:7777 

# Build the current binary with cover variables injected, and redirect output to /to/this/path.
goc build --output /to/this/path

# Build the current binary with cover variables injected, and set necessary build flags: -ldflags "-extldflags -static" -tags="embed kodo".
goc build --buildflags="-ldflags '-extldflags -static' -tags='embed kodo'"
`,
	Run: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Fail to build: %v", err)
		}
		runBuild(args, wd)
	},
}

var buildOutput string

func init() {
	addBuildFlags(buildCmd.Flags())
	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", "", "it forces build to write the resulting executable to the named output file")
	rootCmd.AddCommand(buildCmd)
}

func runBuild(args []string, wd string) {
	gocBuild, err := build.NewBuild(buildFlags, args, wd, buildOutput)
	if err != nil {
		log.Fatalf("Fail to build: %v", err)
	}
	// remove temporary directory if needed
	defer gocBuild.Clean()
	// doCover with original buildFlags, with new GOPATH( tmp:original )
	// in the tmp directory
	cover.Execute(buildFlags, gocBuild.NewGOPATH, gocBuild.TmpDir, mode, agentPort, center)
	// do install in the temporary directory
	err = gocBuild.Build()
	if err != nil {
		log.Fatalf("Fail to build: %v", err)
	}
	return
}
