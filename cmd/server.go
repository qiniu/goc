package cmd

import (
	"github.com/qiniu/goc/v2/pkg/config"
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

func init() {
	// serverCmd.Flags().IntVarP(&config.GocConfig.Port, "port", "", 7777, "listen port to start a coverage host center")
	// serverCmd.Flags().StringVarP(&config.GocConfig.StorePath, "storepath", "", "goc.store", "the file to save all goc server information")
	serverCmd.Flags().StringVarP(&config.GocConfig.Host, "host", "", "0.0.0.0:7777", "specify the host of the goc server")
	rootCmd.AddCommand(serverCmd)
}

func serve(cmd *cobra.Command, args []string) {
	server.RunGocServerUntilExit(config.GocConfig.Host)
}
