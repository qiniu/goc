package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var version = "unstable"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the goc version information",
	Example: `
# Print the client and server versions for the current context
goc version
	`,
	Run: func(cmd *cobra.Command, args []string) {
		// if it is "Unstable", means user build local or with go get
		if version == "unstable" {
			if info, ok := debug.ReadBuildInfo(); ok {
				fmt.Println(info.Main.Version)
			}
		} else {
			// otherwise the value is injected in CI
			fmt.Println(version)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
