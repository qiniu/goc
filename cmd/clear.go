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
	"log"
	"os"

	"github.com/qiniu/goc/pkg/cover"
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear code coverage counters of all the registered services",
	Long:  `Clear code coverage counters for the services under test at runtime.`,
	Example: `
# Clear coverage counter from default register center http://127.0.0.1:7777.
goc clear

# Clear coverage counter from specified register center.
goc clear --center=http://192.168.1.1:8080
`,
	Run: func(cmd *cobra.Command, args []string) {
		res, err := cover.NewWorker().Clear(center)
		if err != nil {
			log.Fatalf("call host %v failed, err: %v, response: %v", center, err, string(res))
		}
		fmt.Fprint(os.Stdout, string(res))
	},
}

func init() {
	addBasicFlags(clearCmd.Flags())
	rootCmd.AddCommand(clearCmd)
}
