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
	"github.com/qiniu/goc/v2/pkg/client"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Get coverage profile from service registry center",
	Long:  `Get code coverage profile for the services under test at runtime.`,
	Example: `
# Get coverage counter from default register center http://127.0.0.1:7777, the result output to stdout.
goc profile
# Get coverage counter from specified register center, the result output to specified file.
goc profile --host=http://192.168.1.1:8080 --output=./coverage.cov
`,
	Run: profile,
}

var (
	profileHost   string
	profileoutput string // --output flag
)

func init() {
	profileCmd.Flags().StringVar(&profileHost, "host", "127.0.0.1:7777", "specify the host of the goc server")
	profileCmd.Flags().StringVarP(&profileoutput, "output", "o", "", "download cover profile")
	rootCmd.AddCommand(profileCmd)
}

func profile(cmd *cobra.Command, args []string) {
	client.NewWorker("http://" + profileHost).Profile(profileoutput)
}
