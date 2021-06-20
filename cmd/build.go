package cmd

import (
	"github.com/qiniu/goc/v2/pkg/build"
	"github.com/qiniu/goc/v2/pkg/config"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use: "build",
	Run: buildAction,

	DisableFlagParsing: true, // build 命令需要用原生 go 的方式处理 flags
}

func init() {
	buildCmd.Flags().StringVarP(&config.GocConfig.Mode, "mode", "", "count", "coverage mode: set, count, atomic")
	buildCmd.Flags().StringVarP(&config.GocConfig.Host, "host", "", "127.0.0.1:7777", "specify the host of the goc sever")
	rootCmd.AddCommand(buildCmd)
}

func buildAction(cmd *cobra.Command, args []string) {
	b := build.NewBuild(cmd, args)
	b.Build()
}
