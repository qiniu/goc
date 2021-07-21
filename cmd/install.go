package cmd

import (
	"github.com/qiniu/goc/v2/pkg/build"
	"github.com/qiniu/goc/v2/pkg/config"
	"github.com/qiniu/goc/v2/pkg/flag"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use: "install",
	Run: installAction,

	DisableFlagParsing: true, // install 命令需要用原生 go 的方式处理 flags
}

func init() {
	installCmd.Flags().StringVarP(&config.GocConfig.Mode, "gocmode", "", "count", "coverage mode: set, count, atomic, watch")
	installCmd.Flags().StringVarP(&config.GocConfig.Host, "gochost", "", "127.0.0.1:7777", "specify the host of the goc sever")
	rootCmd.AddCommand(installCmd)
}

func installAction(cmd *cobra.Command, args []string) {
	// 1. 解析 goc 命令行和 go 命令行
	remainedArgs := flag.BuildCmdArgsParse(cmd, args, flag.GO_INSTALL)
	b := build.NewInstall(remainedArgs)
	b.Install()
}
