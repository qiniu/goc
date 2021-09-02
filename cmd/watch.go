package cmd

import (
	cli "github.com/qiniu/goc/v2/pkg/watch"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:     "watch",
	Short:   "watch for profile's real time update",
	Long:    "watch for profile's real time update",
	Example: "",

	Run: watch,
}

var (
	watchHost string
)

func init() {
	watchCmd.Flags().StringVarP(&watchHost, "host", "", "127.0.0.1:7777", "specify the host of the goc server")
	rootCmd.AddCommand(watchCmd)
}

func watch(cmd *cobra.Command, args []string) {
	cli.Watch(watchHost)
}
