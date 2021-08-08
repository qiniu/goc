package flag

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/qiniu/goc/v2/pkg/config"
	"github.com/qiniu/goc/v2/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var buildUsage string = `Usage:
  goc build [-o output] [build flags] [packages] [goc flags]

[build flags] are same with go official command, you can copy them here directly.

The [goc flags] can be placed in anywhere in the command line.
However, other flags' order are same with the go official command.
`

var installUsage string = `Usage:
goc install [-o output] [build flags] [packages] [goc flags]

[build flags] are same with go official command, you can copy them here directly.

The [goc flags] can be placed in anywhere in the command line.
However, other flags' order are same with the go official command.
`

const (
	GO_BUILD = iota
	GO_INSTALL
)

// BuildCmdArgsParse parse both go flags and goc flags, it rewrite go flags if
// necessary, and returns all non-flag arguments.
//
// 吞下 [packages] 之前所有的 flags.
func BuildCmdArgsParse(cmd *cobra.Command, args []string, cmdType int) []string {
	// 首先解析 cobra 定义的 flag
	allFlagSets := cmd.Flags()
	// 因为 args 里面含有 go 的 flag，所以需要忽略解析 go flag 的错误
	allFlagSets.Init("GOC", pflag.ContinueOnError)
	// 忽略 go flag 在 goc 中的解析错误
	allFlagSets.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{
		UnknownFlags: true,
	}
	allFlagSets.Parse(args)

	// 重写 help
	helpFlag := allFlagSets.Lookup("help")

	if helpFlag.Changed {
		if cmdType == GO_BUILD {
			printHelp(buildUsage, cmd)
		} else if cmdType == GO_INSTALL {
			printHelp(installUsage, cmd)
		}

		os.Exit(0)
	}
	// 删除 help flag
	args = findAndDelHelpFlag(args)

	// 必须手动调用
	// 由于关闭了 cobra 的 flag parse，root PersistentPreRun 调用时，log.NewLogger 并没有拿到 debug 值
	log.NewLogger()

	// 删除 cobra 定义的 flag
	allFlagSets.Visit(func(f *pflag.Flag) {
		args = findAndDelGocFlag(args, f.Name, f.Value.String())
	})

	// 然后解析 go 的 flag
	goFlagSets := flag.NewFlagSet("GO", flag.ContinueOnError)
	addBuildFlags(goFlagSets)
	addOutputFlags(goFlagSets)
	err := goFlagSets.Parse(args)
	if err != nil {
		log.Fatalf("%v", err)
	}

	// 找出设置的 go flag
	curWd, err := os.Getwd()
	if err != nil {
		log.Fatalf("fail to get current working directory: %v", err)
	}
	config.GocConfig.CurWd = curWd
	flags := make([]string, 0)
	goFlagSets.Visit(func(f *flag.Flag) {
		// 将用户指定 -o 改成绝对目录
		if f.Name == "o" {
			outputDir := f.Value.String()
			outputDir, err := filepath.Abs(outputDir)
			if err != nil {
				log.Fatalf("output flag is not valid: %v", err)
			}
			flags = append(flags, "-o", outputDir)
		} else {
			flags = append(flags, "-"+f.Name, f.Value.String())
		}
	})

	config.GocConfig.Goflags = flags

	return goFlagSets.Args()
}

func findAndDelGocFlag(a []string, x string, v string) []string {
	new := make([]string, 0, len(a))
	x = "--" + x
	x_v := x + "=" + v
	for i := 0; i < len(a); i++ {
		if a[i] == "--gocdebug" {
			// debug 是 bool，就一个元素
			continue
		} else if a[i] == x {
			// 有 goc flag 长这样 --mode watch
			i++
			continue
		} else if a[i] == x_v {
			// 有 goc flag 长这样 --mode=watch
			continue
		} else {
			// 剩下的是 go flag
			new = append(new, a[i])
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
