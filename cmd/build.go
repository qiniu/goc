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
