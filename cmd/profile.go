package cmd

import (
	"github.com/qiniu/goc/v2/pkg/client"
	"github.com/qiniu/goc/v2/pkg/config"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Get coverage profile from service registry center",
	Long:  `Get code coverage profile for the services under test at runtime.`,
	Example: `
# Get coverage counter from default register center http://127.0.0.1:7777, the result output to stdout.
goc profile
# Get coverage counter from specified register center, the result output to specified file.
goc profile --host=http://192.168.1.1:8080 --output=./coverage.cov
`,
	Run: profile,
}

var output string // --output flag

func init() {
	profileCmd.Flags().StringVar(&config.GocConfig.Host, "host", "127.0.0.1:7777", "specify the host of the goc server")
	profileCmd.Flags().StringVarP(&output, "output", "o", "", "download cover profile")
	rootCmd.AddCommand(profileCmd)
}

func profile(cmd *cobra.Command, args []string) {
	client.NewWorker("http://" + config.GocConfig.Host).Profile(output)
}
