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
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/qiniu/goc/v2/pkg/log"
)

// readProjectMetaInfo reads all meta informations of the corresponding project
func (b *Build) readProjectMetaInfo() {
	// get gopath & gobin
	b.GOPATH = b.readGOPATH()
	b.GOBIN = b.readGOBIN()
	// 获取 [packages] 及其依赖的 package list
	pkgs := b.listPackages(b.CurWd)

	// get mod info
	for _, pkg := range pkgs {
		// check if go modules is enabled
		if pkg.Module == nil {
			log.Fatalf("Go module is not enabled, please set GO111MODULE=auto or on")
		}
		// 工程根目录
		b.CurModProjectDir = pkg.Module.Dir
		b.ImportPath = pkg.Module.Path

		break
	}

	// 如果当前目录不是工程根目录，那再次 go list 一次，获取整个工程的包信息
	if b.CurWd != b.CurModProjectDir {
		b.Pkgs = b.listPackages(b.CurModProjectDir)
	} else {
		b.Pkgs = pkgs
	}

	// get tmp folder name
	b.TmpModProjectDir = filepath.Join(os.TempDir(), TmpFolderName(b.CurModProjectDir))
	// get working dir in the corresponding tmp dir
	b.TmpWd = filepath.Join(b.TmpModProjectDir, b.CurWd[len(b.CurModProjectDir):])
	// get GlobalCoverVarImportPath
	b.GlobalCoverVarImportPath = path.Join(b.ImportPath, TmpFolderName(b.CurModProjectDir))
	log.Donef("project meta information parsed")
}

// displayProjectMetaInfo prints basic infomation of this project to stdout
func (b *Build) displayProjectMetaInfo() {
	log.Infof("GOPATH: %v", b.GOPATH)
	log.Infof("GOBIN: %v", b.GOBIN)
	log.Infof("Project Directory: %v", b.CurModProjectDir)
	log.Infof("Temporary Project Directory: %v", b.TmpModProjectDir)
	log.Infof("")
}

// readGOPATH reads GOPATH use go env GOPATH command
func (b *Build) readGOPATH() string {
	out, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		log.Fatalf("fail to read GOPATH: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// readGOBIN reads GOBIN use go env GOBIN command
func (b *Build) readGOBIN() string {
	out, err := exec.Command("go", "env", "GOBIN").Output()
	if err != nil {
		log.Fatalf("fail to read GOBIN: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// listPackages list all packages under specific via go list command.
func (b *Build) listPackages(dir string) map[string]*Package {
	listArgs := []string{"list", "-json"}
	if goflags.BuildTags != "" {
		listArgs = append(listArgs, "-tags", goflags.BuildTags)
	}
	listArgs = append(listArgs, "./...")

	cmd := exec.Command("go", listArgs...)
	cmd.Dir = dir

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("execute go list -json failed, err: %v, stdout: %v, stderr: %v", err, string(out), errBuf.String())
	}
	// 有些时候 go 命令会打印一些信息到 stderr，但其实命令整体是成功运行了
	if errBuf.String() != "" {
		log.Errorf("%v", errBuf.String())
	}

	dec := json.NewDecoder(bytes.NewBuffer(out))
	pkgs := make(map[string]*Package)

	for {
		var pkg Package
		if err := dec.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("reading go list output error: %v", err)
		}
		if pkg.Error != nil {
			log.Fatalf("list package %s failed with output: %v", pkg.ImportPath, pkg.Error)
		}

		pkgs[pkg.ImportPath] = &pkg
	}

	return pkgs
}
