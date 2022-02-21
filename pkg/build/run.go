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
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/qiniu/goc/v2/pkg/log"
	"github.com/qiniu/goc/v2/pkg/server"
	"github.com/qiniu/goc/v2/pkg/server/store"
)

func NewRun(opts ...gocOption) *Build {
	b := &Build{}

	for _, opt := range opts {
		opt(b)
	}

	// 1. 解析 goc 命令行和 go 命令行
	b.runCmdArgsParse()
	// 2. 解析 go 包位置
	// b.getPackagesDir()
	// 3. 读取工程元信息：go.mod, pkgs list ...
	b.readProjectMetaInfo()
	// 4. 展示元信息
	b.displayProjectMetaInfo()

	return b
}

// Run starts go run
//
// 1. copy project to temp,
// 2. inject cover variables and functions into the project,
// 3. run the project in temp.
func (b *Build) Run() {
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

	// 4. run in the temp project
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt)
		<-ch
		b.clean()
	}()
	b.doRunInTemp()
}

func (b *Build) doRunInTemp() {
	log.Infof("running the injected project")

	s := store.NewFakeStore()
	go func() {
		gin.SetMode(gin.ReleaseMode)
		err := server.RunGocServerUntilExit(b.Host, s)
		if err != nil {
			log.Fatalf("goc server fail to run: %v", err)
		}
	}()

	args := []string{"run"}
	args = append(args, b.GoArgs...)
	cmd := exec.Command("go", args...)
	cmd.Dir = b.TmpWd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Infof("go run cmd is: %v, in path [%v]", nicePrintArgs(cmd.Args), cmd.Dir)
	if err := cmd.Start(); err != nil {
		log.Errorf("fail to execute go run: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Errorf("fail to execute go run: %v", err)
	}

	// done
	log.Donef("go run done")
}
