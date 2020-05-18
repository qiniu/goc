/*
 Copyright 2020 Qiniu Cloud (七牛云)

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

package app

import (
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "cover and run Go program",
	Long: `usage: goc run [flags] . [arguments...]

Run uses 'go tool cover' to cover the .go source files for the package, then compiles and runs the named main Go package like pure 'go run .' command.
if the arguments are given, 'goc' invokes the covered binary like 'go run . [arguments...]'`,
	Run: func(cmd *cobra.Command, args []string) {
		// init workspace

		// list packages

		// cover code

		// inject apis

		// use go run . to run main package
	},
}

var arguments []string

func init() {
	runCmd.Flags().BoolVarP(&BuildWork, "work", "", false, "do not delete the temporary work directory when exiting")
	rootCmd.AddCommand(runCmd)
}
