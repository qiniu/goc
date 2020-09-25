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
	"github.com/qiniu/goc/pkg/cover"
	"github.com/spf13/cobra"
	"log"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a service registry center",
	Long:  `Start a service registry center.`,
	Example: `
# Start a service registry center, default port :7777.
goc server

# Start a service registry center with port :8080.
goc server --port=:8080

# Start a service registry center with localhost:8080.
goc server --port=localhost:8080
`,
	Run: func(cmd *cobra.Command, args []string) {
		server, err := cover.NewFileBasedServer(localPersistence)
		if err != nil {
			log.Fatalf("New file based server failed, err: %v", err)
		}
		server.Run(port)
	},
}

var port, localPersistence string

func init() {
	serverCmd.Flags().StringVarP(&port, "port", "", ":7777", "listen port to start a coverage host center")
	serverCmd.Flags().StringVarP(&localPersistence, "local-persistence", "", "_svrs_address.txt", "the file to save services address information")
	rootCmd.AddCommand(serverCmd)
}
