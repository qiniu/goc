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
	log "github.com/sirupsen/logrus"

	"github.com/qiniu/goc/pkg/cover"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Clear the register information in order to start a new round of tests",
	Run: func(cmd *cobra.Command, args []string) {
		if res, err := cover.NewWorker().InitSystem(center); err != nil {
			log.Fatalf("call host %v failed, err: %v, response: %v", center, err, string(res))
		}
	},
}

func init() {
	addBasicFlags(initCmd.Flags())
	rootCmd.AddCommand(initCmd)
}
