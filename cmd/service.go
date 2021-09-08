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

var listCmd = &cobra.Command{
	Use:   "service",
	Short: "Lists all the registered services",
	Long:  "Lists all the registered services",
}

var (
	listHost string
	listWide bool
	listIds  []string
)

func init() {
	listCmd.PersistentFlags().StringVar(&listHost, "host", "127.0.0.1:7777", "specify the host of the goc server")
	listCmd.PersistentFlags().BoolVar(&listWide, "wide", false, "list all services with more information (such as pid)")
	listCmd.PersistentFlags().StringSliceVar(&listIds, "id", nil, "specify the ids of the services")

	listCmd.AddCommand(getServiceCmd)
	listCmd.AddCommand(deleteServiceCmd)
	rootCmd.AddCommand(listCmd)
}

func list(cmd *cobra.Command, args []string) {
	client.ListAgents(listHost, listIds, listWide)
}

var getServiceCmd = &cobra.Command{
	Use: "get",
	Run: getAgents,
}

func getAgents(cmd *cobra.Command, args []string) {
	client.ListAgents(listHost, listIds, listWide)
}

var deleteServiceCmd = &cobra.Command{
	Use: "delete",
	Run: deleteAgents,
}

func deleteAgents(cmd *cobra.Command, args []string) {
	client.DeleteAgents(listHost, listIds)
}
