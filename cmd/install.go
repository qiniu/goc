package cmd

import (
	"github.com/qiniu/goc/v2/pkg/build"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use: "install",
	Run: installAction,

	DisableFlagParsing: true, // install 命令需要用原生 go 的方式处理 flags
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func installAction(cmd *cobra.Command, args []string) {
	b := build.NewInstall(cmd, args)
	b.Install()
}
