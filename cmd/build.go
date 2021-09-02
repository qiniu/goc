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
	"github.com/qiniu/goc/v2/pkg/build"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use: "build",
	Run: buildAction,

	DisableFlagParsing: true, // build 命令需要用原生 go 的方式处理 flags
}

var (
	gocmode string
	gochost string
)

func init() {
	buildCmd.Flags().StringVarP(&gocmode, "gocmode", "", "count", "coverage mode: set, count, atomic, watch")
	buildCmd.Flags().StringVarP(&gochost, "gochost", "", "127.0.0.1:7777", "specify the host of the goc sever")
	rootCmd.AddCommand(buildCmd)
}

func buildAction(cmd *cobra.Command, args []string) {

	sets := build.CustomParseCmdAndArgs(cmd, args)

	b := build.NewBuild(
		build.WithHost(gochost),
		build.WithMode(gocmode),
		build.WithFlagSets(sets),
		build.WithArgs(args),
		build.WithBuild(),
		build.WithDebug(globalDebug),
	)
	b.Build()

}
