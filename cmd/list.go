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

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all the registered services",
	Long:  "Lists all the registered services",
	Example: `
goc list [flags]
`,
	Run: func(cmd *cobra.Command, args []string) {
		res, err := cover.NewWorker(center).ListServices()
		if err != nil {
			log.Fatalf("list failed, err: %v", err)
		}
		fmt.Fprint(os.Stdout, string(res))
	},
}

func init() {
	listCmd.Flags().StringVarP(&center, "center", "", "http://127.0.0.1:7777", "cover profile host center")
	rootCmd.AddCommand(listCmd)
}
