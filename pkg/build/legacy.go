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
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/otiai10/copy"
	"github.com/qiniu/goc/pkg/cover"
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

		if err := copy.Copy(src, dst); err != nil {
			log.Errorf("Failed to Copy the folder from %v to %v, the error is: %v ", src, dst, err)
		}

		visited[src] = true

		b.cpDepPackages(v, visited)
	}
}

// only cp dependency in root(current gopath),
// skip deps in other GOPATHs
func (b *Build) cpDepPackages(pkg *cover.Package, visited map[string]bool) {
	/*
		oriGOPATH := os.Getenv("GOPATH")
		if oriGOPATH == "" {
			oriGOPATH = filepath.Join(os.Getenv("HOME"), "go")
		}
		gopaths := strings.Split(oriGOPATH, ":")
	*/
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

		if err := copy.Copy(src, dst); err != nil {
			log.Errorf("Failed to Copy the folder from %v to %v, the error is: %v ", src, dst, err)
		}

		visited[src] = true
	}
}

func (b *Build) cpNonStandardLegacy() {
	for _, v := range b.Pkgs {
		if v.Name == "main" {
			dst := b.TmpDir
			src := v.Dir

			if err := copy.Copy(src, dst); err != nil {
				log.Printf("Failed to Copy the folder from %v to %v, the error is: %v ", src, dst, err)
			}
			break
		} else {
			continue
		}
	}
}
