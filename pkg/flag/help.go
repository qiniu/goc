package flag

import (
	"fmt"

	"github.com/spf13/cobra"
)

func printHelp(usage string, cmd *cobra.Command) {
	fmt.Println(usage)

	flags := cmd.LocalFlags()
	globalFlags := cmd.Parent().PersistentFlags()

	fmt.Println("Flags:")
	fmt.Println(flags.FlagUsages())

	fmt.Println("Global Flags:")
	fmt.Println(globalFlags.FlagUsages())
}
