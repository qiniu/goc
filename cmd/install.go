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

	"github.com/qiniu/goc/pkg/build"
	"github.com/qiniu/goc/pkg/cover"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Do cover for all go files and execute go install command",
	Long: `
Install command will copy the project code and its necessary dependencies to a temporary directory, then do cover for the target, binaries will be generated to their original place.
`,
	Example: `
# Install all binaries with cover variables injected. The binary will be installed in $GOPATH/bin or $HOME/go/bin if directory existed.
goc install ./...

# Install the current binary with cover variables injected, and set the registry center to http://127.0.0.1:7777.
goc install --center=http://127.0.0.1:7777 

# Install the current binary with cover variables injected, and set necessary build flags: -ldflags "-extldflags -static" -tags="embed kodo".
goc build --buildflags="-ldflags '-extldflags -static' -tags='embed kodo'"
`,
	Run: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Fail to build: %v", err)
		}
		runInstall(args, wd)
	},
}

func init() {
	addBuildFlags(installCmd.Flags())
	rootCmd.AddCommand(installCmd)
}

func runInstall(args []string, wd string) {
	gocBuild, err := build.NewInstall(buildFlags, args, wd)
	if err != nil {
		log.Fatalf("Fail to install: %v", err)
	}
	// remove temporary directory if needed
	defer gocBuild.Clean()
	// doCover with original buildFlags, with new GOPATH( tmp:original )
	// in the tmp directory
	cover.Execute(buildFlags, gocBuild.NewGOPATH, gocBuild.TmpDir, mode, center)
	// do install in the temporary directory
	err = gocBuild.Install()
	if err != nil {
		log.Fatalf("Fail to install: %v", err)
	}
	return
}
