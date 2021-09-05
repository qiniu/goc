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
	"github.com/qiniu/goc/v2/pkg/server"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "start a service registry center",
	Long:    "start a service registry center",
	Example: "",

	Run: serve,
}

var (
	serverHost  string
	serverStore string
)

func init() {
	serverCmd.Flags().StringVarP(&serverHost, "host", "", "127.0.0.1:7777", "specify the host of the goc server")
	serverCmd.Flags().StringVarP(&serverStore, "store", "", ".goc.kvstore", "specify the host of the goc server")

	rootCmd.AddCommand(serverCmd)
}

func serve(cmd *cobra.Command, args []string) {
	server.RunGocServerUntilExit(serverHost, serverStore)
}
