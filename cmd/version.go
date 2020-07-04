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
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// the version value will be injected when publishing
var version = "Unstable"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the goc version information",
	Example: `
# Print the client and server versions for the current context
goc version
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// if it is "Unstable", means user build local or with go get
		if version == "Unstable" {
			if info, ok := debug.ReadBuildInfo(); ok {
				fmt.Println(info.Main.Version)
			}
		} else {
			// otherwise the value is injected in CI
			fmt.Println(version)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
