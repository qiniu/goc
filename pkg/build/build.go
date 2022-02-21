/*
 Copyright 2021 Qiniu Cloud (qiniu.com)
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
     http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package build

import (
	"os"
	"os/exec"
	"strings"

	"github.com/qiniu/goc/v2/pkg/log"
	"github.com/spf13/pflag"
)

// Build struct a build
type Build struct {
	Args      []string // all goc + go command line args + flags
	FlagSets  *pflag.FlagSet
	BuildType int

	Debug bool
	Host  string
	Mode  string // cover mode

	GOPATH           string
	GOBIN            string
	CurWd            string
	TmpWd            string
	CurModProjectDir string
	TmpModProjectDir string

	Goflags  []string // go command line flags
	GoArgs   []string // go command line args
	Packages []string // go command line [packages]

	IsVendorMod bool // vendor, readonly, or mod?
	IsModEdit   bool // is mod file edited?

	ImportPath                  string // the whole import path of the project
	Pkgs                        map[string]*Package
	GlobalCoverVarImportPath    string
	GlobalCoverVarImportPathDir string
}

// NewBuild creates a Build struct
//
func NewBuild(opts ...gocOption) *Build {
	b := &Build{}

	for _, opt := range opts {
		opt(b)
	}

	// 1. 解析 goc 命令行和 go 命令行
	b.buildCmdArgsParse()
	// 2. 解析 go 包位置
	b.getPackagesDir()
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

	// 2. update go.mod file if needed
	b.updateGoModFile()
	// 3. inject cover vars
	b.Inject()

	if b.IsVendorMod && b.IsModEdit {
		b.reVendor()
	}

	// 4. build in the temp project
	b.doBuildInTemp()
}

func (b *Build) doBuildInTemp() {
	log.StartWait("building the injected project")

	goflags := b.Goflags
	// 检查用户是否自定义了 -o
	oSet := false
	for _, flag := range goflags {
		if flag == "-o" {
			oSet = true
		}
	}

	// 如果没被设置就加一个至原命令执行的目录
	if !oSet {
		goflags = append(goflags, "-o", b.CurWd)
	}

	if b.IsVendorMod && b.IsModEdit {
		b.reVendor()
	}

	pacakges := b.Packages

	goflags = append(goflags, pacakges...)

	args := []string{"build"}
	args = append(args, goflags...)
	// go 命令行由 go build [-o output] [build flags] [packages] 组成
	cmd := exec.Command("go", args...)
	cmd.Dir = b.TmpWd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Infof("go build cmd is: %v, in path [%v]", nicePrintArgs(cmd.Args), cmd.Dir)
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

// nicePrintArgs 优化 args 打印内容
//
// 假如：go build -ldflags "-X my/package/config.Version=1.0.0" -o /home/lyy/gitdown/gin-test/cmd .
//
// 实际输出会变为：go build -ldflags -X my/package/config.Version=1.0.0 -o /home/lyy/gitdown/gin-test/cmd .
func nicePrintArgs(args []string) []string {
	output := make([]string, 0)
	for _, arg := range args {
		if strings.Contains(arg, " ") {
			output = append(output, "\""+arg+"\"")
		} else {
			output = append(output, arg)
		}
	}

	return output
}

func (b *Build) reVendor() {
	log.StartWait("re-vendoring the project")
	cmd := exec.Command("go", "mod", "vendor")
	cmd.Dir = b.TmpModProjectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatalf("fail to execute go vendor: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatalf("fail to execute go vendor: %v", err)
	}

	log.StopWait()
	log.Donef("re-vendor the project done")
}
