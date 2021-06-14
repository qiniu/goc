package cmd

import (
	"github.com/qiniu/goc/v2/pkg/build"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "start a service registry center",
	Long:    "start a service registry center",
	Example: "",

	Run: serve,
}

func init() {
	// serverCmd.Flags().IntVarP(&config.GocConfig.Port, "port", "", 7777, "listen port to start a coverage host center")
	// serverCmd.Flags().StringVarP(&config.GocConfig.StorePath, "storepath", "", "goc.store", "the file to save all goc server information")

	rootCmd.AddCommand(serverCmd)
}

func serve(cmd *cobra.Command, args []string) {
	b := build.NewBuild(cmd, args)
	b.Build()
}
