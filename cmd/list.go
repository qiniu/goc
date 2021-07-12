package cmd

import (
	"github.com/qiniu/goc/v2/pkg/client"
	"github.com/qiniu/goc/v2/pkg/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all the registered services",
	Long:  "Lists all the registered services",
	Example: `
goc list [flags]
`,

	Run: list,
}

func init() {
	listCmd.Flags().StringVarP(&config.GocConfig.Host, "host", "", "127.0.0.1:7777", "specify the host of the goc server")
	rootCmd.AddCommand(listCmd)
}

func list(cmd *cobra.Command, args []string) {
	client.NewWorker("http://" + config.GocConfig.Host).ListAgents()
}
