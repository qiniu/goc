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

	"github.com/qiniu/goc/v2/pkg/log"
)

func NewInstall(opts ...gocOption) *Build {
	return NewBuild(opts...)
}

// Install starts go install
//
// 1. copy project to temp,
// 2. inject cover variables and functions into the project,
// 3. install the project in temp.
func (b *Build) Install() {
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

	// 4. install in the temp project
	b.doInstallInTemp()
}

func (b *Build) doInstallInTemp() {
	log.StartWait("installing the injected project")

	goflags := b.Goflags

	pacakges := b.Packages

	goflags = append(goflags, pacakges...)

	args := []string{"install"}
	args = append(args, goflags...)
	// go 命令行由 go install [build flags] [packages] 组成
	cmd := exec.Command("go", args...)
	cmd.Dir = b.TmpWd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Infof("go install cmd is: %v, in path [%v]", cmd.Args, cmd.Dir)
	if err := cmd.Start(); err != nil {
		log.Fatalf("fail to execute go install: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatalf("fail to execute go install: %v", err)
	}

	// done
	log.StopWait()
	log.Donef("go install done")
}
