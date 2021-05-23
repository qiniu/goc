package flag

import (
	"flag"

	"github.com/qiniu/goc/v2/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var buildUsage string = `Usage:
  goc build [-o output] [build flags] [packages] [goc flags]

The [goc flags] can be placed in anywhere in the command line.
However, other flags' order are same with the go official command.
`

// BuildCmdArgsParse parse both go flags and goc flags, it returns all non-flag arguments.
//
// 吞下 [packages] 之前所有的 flags.
func BuildCmdArgsParse(cmd *cobra.Command, args []string) []string {
	// 首先解析 cobra 定义的 flag
	allFlagSets := cmd.Flags()
	// 因为 args 里面含有 go 的 flag，所以需要忽略解析 go flag 的错误
	allFlagSets.Init("GOC", pflag.ContinueOnError)
	allFlagSets.Parse(args)

	// 重写 help
	helpFlag := allFlagSets.Lookup("help")

	if helpFlag.Changed {
		printHelp(buildUsage, cmd)
	}
	// 删除 help flag
	args = findAndDelHelpFlag(args)

	// 必须手动调用
	// 由于关闭了 cobra 的 flag parse，root PersistentPreRun 调用时，log.NewLogger 并没有拿到 debug 值
	log.NewLogger()

	// 删除 cobra 定义的 flag
	allFlagSets.Visit(func(f *pflag.Flag) {
		args = findAndDelGocFlag(args, f.Name)
	})

	// 然后解析 go 的 flag
	goFlagSets := flag.NewFlagSet("GO", flag.ContinueOnError)
	addBuildFlags(goFlagSets)
	addOutputFlags(goFlagSets)
	err := goFlagSets.Parse(args)
	if err != nil {
		log.Fatalf("%v", err)
	}

	return goFlagSets.Args()
}

func findAndDelGocFlag(a []string, x string) []string {
	new := make([]string, 0, len(a))
	x = "--" + x
	for _, v := range a {
		if v == x {
			continue
		} else {
			new = append(new, v)
		}
	}

	return new
}

func findAndDelHelpFlag(a []string) []string {
	new := make([]string, 0, len(a))
	for _, v := range a {
		if v == "--help" || v == "-h" {
			continue
		} else {
			new = append(new, v)
		}
	}

	return new
}
