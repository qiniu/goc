package cmd

import (
	"github.com/qiniu/goc/v2/pkg/client"
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

var (
	listHost string
	listWide bool
)

func init() {
	listCmd.Flags().StringVar(&listHost, "host", "127.0.0.1:7777", "specify the host of the goc server")
	listCmd.Flags().BoolVar(&listWide, "wide", false, "list all services with more information (such as pid)")
	rootCmd.AddCommand(listCmd)
}

func list(cmd *cobra.Command, args []string) {
	client.NewWorker("http://" + listHost).ListAgents(listWide)
}
