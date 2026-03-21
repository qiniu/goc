/*
 Copyright 2020 Qiniu Cloud (qiniu.com)

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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/qiniu/goc/pkg/cover"
	"github.com/tongjingran/copy"
)

func (b *Build) cpLegacyProject() {
	visited := make(map[string]bool)
	for k, v := range b.Pkgs {
		dst := filepath.Join(b.TmpDir, "src", k)
		src := v.Dir

		if _, ok := visited[src]; ok {
			// Skip if already copied
			continue
		}

		// 调用cp命令拷贝代码，避免编译时插桩污染源代码，构造一个沙箱环境
		cmd := exec.Command("bash", "-c", fmt.Sprintf("cp -rf %s/* %s/", src, dst))
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("Failed to Copy the folder from %v to %v, the error is: %v, the output is: %v", src, dst, err, string(output))
		}

		if err := copy.Copy(src, dst, copy.Options{Skip: skipCopy}); err != nil {
			log.Errorf("Failed to Copy the folder from %v to %v, the error is: %v ", src, dst, err)
		}

		visited[src] = true

		b.cpDepPackages(v, visited)
	}
}

// only cp dependency in root(current gopath),
// skip deps in other GOPATHs
func (b *Build) cpDepPackages(pkg *cover.Package, visited map[string]bool) {
	gopath := pkg.Root
	for _, dep := range pkg.Deps {
		src := filepath.Join(gopath, "src", dep)
		// Check if copied
		if _, ok := visited[src]; ok {
			// Skip if already copied
			continue
		}
		// Check if we can found in the root gopath
		_, err := os.Stat(src)
		if err != nil {
			continue
		}

		dst := filepath.Join(b.TmpDir, "src", dep)

		cmd := exec.Command("bash", "-c", fmt.Sprintf("cp -rf %s/* %s/", src, dst))
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("Failed to Copy the folder from %v to %v, the error is: %v, the output is: %v", src, dst, err, string(output))
		}

		visited[src] = true
	}
}

func (b *Build) cpNonStandardLegacy() {
	for _, v := range b.Pkgs {
		if v.Name == "main" {
			dst := b.TmpDir
			src := v.Dir

			cmd := exec.Command("bash", "-c", fmt.Sprintf("cp -rf %s/* %s/", src, dst))
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Errorf("Failed to Copy the folder from %v to %v, the error is: %v, the output is: %v", src, dst, err, string(output))
			}
			break
		}
	}
}

// skipCopy skip copy .git dir and irregular files
func skipCopy(src string, info os.FileInfo) (bool, error) {
	irregularModeType := os.ModeNamedPipe | os.ModeSocket | os.ModeDevice | os.ModeCharDevice | os.ModeIrregular
	if strings.HasSuffix(src, "/.git") {
		log.Infof("Skip .git dir [%s]", src)
		return true, nil
	}
	if info.Mode()&irregularModeType != 0 {
		log.Warnf("Skip file [%s], the file mode is [%s]", src, info.Mode().String())
		return true, nil
	}
	return false, nil
}
