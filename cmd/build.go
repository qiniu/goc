package cmd

import (
	"github.com/qiniu/goc/v2/pkg/flag"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use: "build",
	Run: build,

	DisableFlagParsing: true, // build 命令需要用原生 go 的方式处理 flags
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

func build(cmd *cobra.Command, args []string) {
	remainedArgs := flag.BuildCmdArgsParse(cmd, args)
	where, buildName := flag.GetPackagesDir(remainedArgs)

}
