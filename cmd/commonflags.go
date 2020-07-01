package cmd

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	target     string
	center     string
	mode       string
	debugGoc   bool
	buildFlags string

	goRunExecFlag  string
	goRunArguments string
)

// addBasicFlags adds a
func addBasicFlags(cmdset *pflag.FlagSet) {
	cmdset.StringVar(&center, "center", "http://127.0.0.1:7777", "cover profile host center")
	// bind to viper
	viper.BindPFlags(cmdset)
}

func addCommonFlags(cmdset *pflag.FlagSet) {
	addBasicFlags(cmdset)
	cmdset.StringVar(&mode, "mode", "count", "coverage mode: set, count, atomic")
	cmdset.StringVar(&buildFlags, "buildflags", "", "specify the build flags")
	// bind to viper
	viper.BindPFlags(cmdset)
}

func addBuildFlags(cmdset *pflag.FlagSet) {
	addCommonFlags(cmdset)
	// bind to viper
	viper.BindPFlags(cmdset)
}

func addRunFlags(cmdset *pflag.FlagSet) {
	addBuildFlags(cmdset)
	cmdset.StringVar(&goRunExecFlag, "exec", "", "same as -exec flag in 'go run' command")
	cmdset.StringVar(&goRunArguments, "arguments", "", "same as 'arguments' in 'go run' command")
	// bind to viper
	viper.BindPFlags(cmdset)
}
