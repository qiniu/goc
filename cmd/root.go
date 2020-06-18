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
	"path/filepath"
	"runtime"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "goc",
	Short: "goc is a comprehensive coverage testing tool for go language",
	Long: `goc is a comprehensive coverage testing tool for go language.

Find more information at:
 https://github.com/qiniu/goc
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetReportCaller(true)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				dirname, filename := filepath.Split(f.File)
				lastelem := filepath.Base(dirname)
				filename = filepath.Join(lastelem, filename)
				line := strconv.Itoa(f.Line)
				return "", "[" + filename + ":" + line + "]"
			},
		})
		if debugGoc == false {
			// we only need log in debug mode
			log.SetLevel(log.FatalLevel)
			log.SetFormatter(&log.TextFormatter{
				DisableTimestamp: true,
				CallerPrettyfier: func(f *runtime.Frame) (string, string) {
					return "", ""
				},
			})
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&debugGoc, "debug", false, "run goc in debug mode")
	viper.BindPFlags(rootCmd.PersistentFlags())
}

// Execute the goc tool
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
