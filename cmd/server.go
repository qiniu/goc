package cmd

import (
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

var (
	serverHost string
)

func init() {
	serverCmd.Flags().StringVarP(&serverHost, "host", "", "127.0.0.1:7777", "specify the host of the goc server")
	rootCmd.AddCommand(serverCmd)
}

func serve(cmd *cobra.Command, args []string) {
	server.RunGocServerUntilExit(serverHost)
}
