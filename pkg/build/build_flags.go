package build

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// CustomParseCmdAndArgs 因为关闭了 cobra 的解析功能，需要手动构造并解析 goc flags
func CustomParseCmdAndArgs(cmd *cobra.Command, args []string) *pflag.FlagSet {
	// 首先解析 cobra 定义的 flag
	allFlagSets := cmd.Flags()
	// 因为 args 里面含有 go 的 flag，所以需要忽略解析 go flag 的错误
	allFlagSets.Init("GOC", pflag.ContinueOnError)
	// 忽略 go flag 在 goc 中的解析错误
	allFlagSets.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{
		UnknownFlags: true,
	}
	allFlagSets.Parse(args)

	return allFlagSets
}

// buildCmdArgsParse parse both go flags and goc flags, it rewrite go flags if
// necessary, and returns all non-flag arguments.
//
// 吞下 [packages] 之前所有的 flags.
func (b *Build) buildCmdArgsParse() {
	args := b.Args
	cmdType := b.BuildType
	allFlagSets := b.FlagSets

	// 重写 help
	helpFlag := allFlagSets.Lookup("help")

	if helpFlag.Changed {
		if cmdType == GO_BUILD {
			printGoHelp(buildUsage)
		} else if cmdType == GO_INSTALL {
			printGoHelp(installUsage)
		}

		os.Exit(0)
	}
	// 删除 help flag
	args = findAndDelHelpFlag(args)

	// 必须手动调用
	// 由于关闭了 cobra 的 flag parse，root PersistentPreRun 调用时，log.NewLogger 并没有拿到 debug 值
	log.NewLogger(b.Debug)

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

	b.Goflags = flags
	b.CurWd = curWd
	b.GoArgs = goFlagSets.Args()
	return
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

type goConfig struct {
	BuildA                 bool
	BuildBuildmode         string // -buildmode flag
	BuildMod               string // -mod flag
	BuildModReason         string // reason -mod flag is set, if set by default
	BuildI                 bool   // -i flag
	BuildLinkshared        bool   // -linkshared flag
	BuildMSan              bool   // -msan flag
	BuildN                 bool   // -n flag
	BuildO                 string // -o flag
	BuildP                 int    // -p flag
	BuildPkgdir            string // -pkgdir flag
	BuildRace              bool   // -race flag
	BuildToolexec          string // -toolexec flag
	BuildToolchainName     string
	BuildToolchainCompiler func() string
	BuildToolchainLinker   func() string
	BuildTrimpath          bool // -trimpath flag
	BuildV                 bool // -v flag
	BuildWork              bool // -work flag
	BuildX                 bool // -x flag
	// from buildcontext
	Installsuffix string // -installSuffix
	BuildTags     string // -tags
	// from load
	BuildAsmflags   string
	BuildCompiler   string
	BuildGcflags    string
	BuildGccgoflags string
	BuildLdflags    string

	// mod related
	ModCacheRW bool
	ModFile    string
}

var goflags goConfig

func addBuildFlags(cmdSet *flag.FlagSet) {
	cmdSet.BoolVar(&goflags.BuildA, "a", false, "")
	cmdSet.BoolVar(&goflags.BuildN, "n", false, "")
	cmdSet.IntVar(&goflags.BuildP, "p", 4, "")
	cmdSet.BoolVar(&goflags.BuildV, "v", false, "")
	cmdSet.BoolVar(&goflags.BuildX, "x", false, "")
	cmdSet.StringVar(&goflags.BuildBuildmode, "buildmode", "default", "")
	cmdSet.StringVar(&goflags.BuildMod, "mod", "", "")
	cmdSet.StringVar(&goflags.Installsuffix, "installsuffix", "", "")

	// 类型和 go 原生的不一样，这里纯粹是为了 parse 并传递给 go
	cmdSet.StringVar(&goflags.BuildAsmflags, "asmflags", "", "")
	cmdSet.StringVar(&goflags.BuildCompiler, "compiler", "", "")
	cmdSet.StringVar(&goflags.BuildGcflags, "gcflags", "", "")
	cmdSet.StringVar(&goflags.BuildGccgoflags, "gccgoflags", "", "")
	// mod related
	cmdSet.BoolVar(&goflags.ModCacheRW, "modcacherw", false, "")
	cmdSet.StringVar(&goflags.ModFile, "modfile", "", "")
	cmdSet.StringVar(&goflags.BuildLdflags, "ldflags", "", "")
	cmdSet.BoolVar(&goflags.BuildLinkshared, "linkshared", false, "")
	cmdSet.StringVar(&goflags.BuildPkgdir, "pkgdir", "", "")
	cmdSet.BoolVar(&goflags.BuildRace, "race", false, "")
	cmdSet.BoolVar(&goflags.BuildMSan, "msan", false, "")
	cmdSet.StringVar(&goflags.BuildTags, "tags", "", "")
	cmdSet.StringVar(&goflags.BuildToolexec, "toolexec", "", "")
	cmdSet.BoolVar(&goflags.BuildTrimpath, "trimpath", false, "")
	cmdSet.BoolVar(&goflags.BuildWork, "work", false, "")
}

func addOutputFlags(cmdSet *flag.FlagSet) {
	cmdSet.StringVar(&goflags.BuildO, "o", "", "")
}

func printGoHelp(usage string) {
	fmt.Println(usage)
}

func printGocHelp(cmd *cobra.Command) {
	flags := cmd.LocalFlags()
	globalFlags := cmd.Parent().PersistentFlags()

	fmt.Println("Flags:")
	fmt.Println(flags.FlagUsages())

	fmt.Println("Global Flags:")
	fmt.Println(globalFlags.FlagUsages())
}

// GetPackagesDir parse [pacakges] part of args, it will fatal if error encountered
//
// 函数获取 1： [packages] 所在的目录位置，供后续插桩使用。
//
// 函数获取 2： 如果参数是 *.go，第一个 .go 文件的文件名。go build 中，二进制名字既可能是目录名也可能是文件名，和参数类型有关。
//
// 如果 [packages] 非法（即不符合 go 原生的定义），则返回对应错误
// 这里只考虑 go mod 的方式
func (b *Build) getPackagesDir() {
	patterns := b.GoArgs
	packages := make([]string, 0)
	for _, p := range patterns {
		// patterns 只支持两种格式
		// 1. 要么是直接指向某些 .go 文件的相对/绝对路径
		if strings.HasSuffix(p, ".go") {
			if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
				// check if valid
				if err := goFilesPackage(patterns); err != nil {
					log.Fatalf("%v", err)
				}

				// 获取相对于 current working directory 对路径
				for _, p := range patterns {
					if filepath.IsAbs(p) {
						relPath, err := filepath.Rel(b.CurWd, p)
						if err != nil {
							log.Fatalf("fail to get [packages] relative path from current working directory: %v", err)
						}
						packages = append(packages, relPath)
					} else {
						packages = append(packages, p)
					}
				}
				// fix: go build ./xx/main.go 需要转换为
				// go build ./xx/main.go ./xx/goc-cover-agent-apis-auto-generated-11111-22222-bridge.go
				dir := filepath.Dir(packages[0])
				packages = append(packages, filepath.Join(dir, "goc-cover-agent-apis-auto-generated-11111-22222-bridge.go"))
				b.Packages = packages

				return
			}
		}
	}

	// 2. 要么是 import path
	b.Packages = patterns
}

// goFilesPackage 对一组 go 文件解析，判断是否合法
// go 本身还判断语法上是否是同一个 package，goc 这里不做解析
// 1. 都是 *.go 文件？
// 2. *.go 文件都在同一个目录？
// 3. *.go 文件存在？
func goFilesPackage(gofiles []string) error {
	// 1. 必须都是 *.go 结尾
	for _, f := range gofiles {
		if !strings.HasSuffix(f, ".go") {
			return fmt.Errorf("named files must be .go files: %s", f)
		}
	}

	var dir string
	for _, file := range gofiles {
		// 3. 文件都存在？
		fi, err := os.Stat(file)
		if err != nil {
			return err
		}

		// 2.1 有可能以 *.go 结尾的目录
		if fi.IsDir() {
			return fmt.Errorf("%s is a directory, should be a Go file", file)
		}

		// 2.2 所有 *.go 必须在同一个目录内
		dir1, _ := filepath.Split(file)
		if dir1 == "" {
			dir1 = "./"
		}

		if dir == "" {
			dir = dir1
		} else if dir != dir1 {
			return fmt.Errorf("named files must all be in one directory: have %s and %s", dir, dir1)
		}
	}

	return nil
}

// getDirFromImportPaths return the import path's real abs directory
//
// 该函数接收到的只有 dir 或 import path，file 在上一步已被排除
// 只考虑 go modules 的情况
func getDirFromImportPaths(patterns []string) (string, error) {
	// no import path, pattern = current wd
	if len(patterns) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("fail to parse import path: %w", err)
		}
		return wd, nil
	}

	// 为了简化插桩的逻辑，goc 对 import path 要求必须都在同一个目录
	// 所以干脆只允许一个 pattern 得了 -_-
	// 对于 goc build/run 来说本身就是只能在一个目录内
	// 对于 goc install 来讲，这个行为就和 go install 不同，不过多 import path 较少见 >_<，先忽略
	if len(patterns) > 1 {
		return "", fmt.Errorf("goc only support one import path now")
	}

	pattern := patterns[0]
	switch {
	// case isLocalImport(pattern) || filepath.IsAbs(pattern):
	// 	dir1, err := filepath.Abs(pattern)
	// 	if err != nil {
	// 		return "", fmt.Errorf("error (%w) get directory from the import path: %v", err, pattern)
	// 	}
	// 	if _, err := os.Stat(dir1); err != nil {
	// 		return "", fmt.Errorf("error (%w) get directory from the import path: %v", err, pattern)
	// 	}
	// 	return dir1, nil

	case strings.Contains(pattern, "..."):
		i := strings.Index(pattern, "...")
		dir, _ := filepath.Split(pattern[:i])
		dir, _ = filepath.Abs(dir)
		if _, err := os.Stat(dir); err != nil {
			return "", fmt.Errorf("error (%w) get directory from the import path: %v", err, pattern)
		}
		return dir, nil

	case strings.IndexByte(pattern, '@') > 0:
		return "", fmt.Errorf("import path with @ version query is not supported in goc")

	case isMetaPackage(pattern):
		return "", fmt.Errorf("`std`, `cmd`, `all` import path is not supported by goc")

	default: // 到这一步认为 pattern 是相对路径或者绝对路径
		dir1, err := filepath.Abs(pattern)
		if err != nil {
			return "", fmt.Errorf("error (%w) get directory from the import path: %v", err, pattern)
		}
		if _, err := os.Stat(dir1); err != nil {
			return "", fmt.Errorf("error (%w) get directory from the import path: %v", err, pattern)
		}

		return dir1, nil
	}
}

// isLocalImport reports whether the import path is
// a local import path, like ".", "..", "./foo", or "../foo"
func isLocalImport(path string) bool {
	return path == "." || path == ".." ||
		strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../")
}

// isMetaPackage checks if the name is a reserved package name
func isMetaPackage(name string) bool {
	return name == "std" || name == "cmd" || name == "all"
}

// find direct path of current project which contains go.mod
func findModuleRoot(dir string) string {
	dir = filepath.Clean(dir)

	// look for enclosing go.mod
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}

	return ""
}
