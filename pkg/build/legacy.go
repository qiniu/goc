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
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/otiai10/copy"
	"github.com/qiniu/goc/pkg/cover"
)

func (b *Build) cpProject() {
	visited := make(map[string]bool)
	for k, v := range b.Pkgs {
		dst := filepath.Join(b.TmpDir, "src", k)
		src := v.Dir

		if _, ok := visited[src]; ok {
			// Skip if already copied
			continue
		}

		if err := b.copyDir(v); err != nil {
			log.Errorf("Failed to Copy the folder from %v to %v, the error is: %v ", src, dst, err)
		}

		visited[src] = true
	}
	if b.IsMod {
		for _, v := range b.Pkgs {
			if v.Name == "main" {
				dst := filepath.Join(b.TmpDir, "go.mod")
				src := filepath.Join(v.Module.Dir, "go.mod")
				if err := copy.Copy(src, dst); err != nil {
					log.Errorf("Failed to Copy the go mod file from %v to %v, the error is: %v ", src, dst, err)
				}

				dst = filepath.Join(b.TmpDir, "go.sum")
				src = filepath.Join(v.Module.Dir, "go.sum")
				if err := copy.Copy(src, dst); err != nil && !strings.Contains(err.Error(), "no such file or directory") {
					log.Errorf("Failed to Copy the go mod file from %v to %v, the error is: %v ", src, dst, err)
				}
				break
			} else {
				continue
			}
		}
	}
}

func (b *Build) copyDir(pkg *cover.Package) error {
	fileList := []string{}
	dir := pkg.Dir
	fileList = append(fileList, pkg.GoFiles...)
	fileList = append(fileList, pkg.CompiledGoFiles...)
	fileList = append(fileList, pkg.IgnoredGoFiles...)
	fileList = append(fileList, pkg.CFiles...)
	fileList = append(fileList, pkg.CXXFiles...)
	fileList = append(fileList, pkg.MFiles...)
	fileList = append(fileList, pkg.HFiles...)
	fileList = append(fileList, pkg.FFiles...)
	fileList = append(fileList, pkg.SFiles...)
	fileList = append(fileList, pkg.SwigCXXFiles...)
	fileList = append(fileList, pkg.SwigFiles...)
	fileList = append(fileList, pkg.SysoFiles...)
	for _, file := range fileList {
		p := filepath.Join(dir, file)
		var src, root string
		if pkg.Root == "" {
			root = b.WorkingDir // use workingDir instead when the root is empty.
		} else {
			root = pkg.Root
		}
		src = strings.TrimPrefix(pkg.Dir, root)   // get the relative path of the files
		dst := filepath.Join(b.TmpDir, src, file) // it will adapt the case where src is ""
		if err := copy.Copy(p, dst); err != nil {
			log.Errorf("Failed to Copy the folder from %v to %v, the error is: %v ", src, dst, err)
			return err
		}
	}
	return nil
}
