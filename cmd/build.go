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
	"github.com/qiniu/goc/pkg/build"
	"github.com/qiniu/goc/pkg/cover"
	"github.com/spf13/cobra"
	"log"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Do cover for all go files and execute go build command",
	Long: `
First of all, this build command will copy the project code and its necessary dependencies to a temporary directory, then do cover for the target in this temporary directory, finally go build command will be executed and binaries generated to their original place.
`,
	Example: `
# Build the current binary with cover variables injected. The binary will be generated in the current folder.
goc build

# Build the current binary with cover variables injected, and set the registry center to http://127.0.0.1:7777.
goc build --center=http://127.0.0.1:7777 

# Build the current binary with cover variables injected, and redirect output to /to/this/path.
goc build --output /to/this/path

# Build the current binary with cover variables injected, and set necessary build flags: -ldflags "-extldflags -static" -tags="embed kodo".
goc build --buildflags="-ldflags '-extldflags -static' -tags='embed kodo'"
`,
	Run: func(cmd *cobra.Command, args []string) {
		gocBuild, err := build.NewBuild(buildFlags, packages, buildOutput)
		if err != nil {
			log.Fatalf("Fail to NewBuild: %v", err)
		}
		// remove temporary directory if needed
		defer gocBuild.Clean()
		// doCover with original buildFlags, with new GOPATH( tmp:original )
		// in the tmp directory
		cover.Execute(buildFlags, gocBuild.NewGOPATH, gocBuild.TmpDir, mode, center)
		// do install in the temporary directory
		err = gocBuild.Build()
		if err != nil {
			log.Fatalf("Fail to build: %v", err)
		}
		return
	},
}

var buildOutput string

func init() {
	addBuildFlags(buildCmd.Flags())
	buildCmd.Flags().StringVar(&buildOutput, "output", "", "it forces build to write the resulting executable or object to the named output file or directory")
	rootCmd.AddCommand(buildCmd)
}
