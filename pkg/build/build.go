package build

import (
	"os"
	"os/exec"

	"github.com/qiniu/goc/v2/pkg/config"
	"github.com/qiniu/goc/v2/pkg/cover"
	"github.com/qiniu/goc/v2/pkg/flag"
	"github.com/qiniu/goc/v2/pkg/log"
)

// Build struct a build
// most configurations are stored in global variables: config.GocConfig & config.GoConfig
type Build struct {
}

// NewBuild creates a Build struct
//
func NewBuild(args []string) *Build {
	b := &Build{}

	// 2. 解析 go 包位置
	flag.GetPackagesDir(args)
	// 3. 读取工程元信息：go.mod, pkgs list ...
	b.readProjectMetaInfo()
	// 4. 展示元信息
	b.displayProjectMetaInfo()

	return b
}

// Build starts go build
//
// 1. copy project to temp,
// 2. inject cover variables and functions into the project,
// 3. build the project in temp.
func (b *Build) Build() {
	// 1. 拷贝至临时目录
	b.copyProjectToTmp()
	defer b.clean()

	log.Donef("project copied to temporary directory")
	// 2. inject cover vars
	cover.Inject()
	// 3. build in the temp project
	b.doBuildInTemp()
}

func (b *Build) doBuildInTemp() {
	log.StartWait("building the injected project")

	goflags := config.GocConfig.Goflags
	// 检查用户是否自定义了 -o
	oSet := false
	for _, flag := range goflags {
		if flag == "-o" {
			oSet = true
		}
	}

	// 如果没被设置就加一个至原命令执行的目录
	if !oSet {
		goflags = append(goflags, "-o", config.GocConfig.CurWd)
	}

	pacakges := config.GocConfig.Packages

	goflags = append(goflags, pacakges...)

	args := []string{"build"}
	args = append(args, goflags...)
	// go 命令行由 go build [-o output] [build flags] [packages] 组成
	cmd := exec.Command("go", args...)
	cmd.Dir = config.GocConfig.TmpWd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Infof("go build cmd is: %v, in path [%v]", cmd.Args, cmd.Dir)
	if err := cmd.Start(); err != nil {
		log.Fatalf("fail to execute go build: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatalf("fail to execute go build: %v", err)
	}

	// done
	log.StopWait()
	log.Donef("go build done")
}
