/*
 Copyright 2021 Qiniu Cloud (qiniu.com)
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
	"github.com/qiniu/goc/v2/pkg/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "goc",
	Short: "goc is a comprehensive coverage testing tool for go language",
	Long: `goc is a comprehensive coverage testing tool for go language.

Find more information at:
 https://github.com/qiniu/goc
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.DisplayGoc()
		// init logger
		log.NewLogger(globalDebug)
	},

	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		log.Sync()
	},
}

var globalDebug bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&globalDebug, "gocdebug", false, "run goc in debug mode")
}

// Execute the goc tool
func Execute() {
	if err := rootCmd.Execute(); err != nil {
	}
}
