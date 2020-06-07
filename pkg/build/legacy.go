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
	"log"
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
	"github.com/qiniu/goc/pkg/cover"
)

func cpLegacyProject(tmpBuildDir string, pkgs map[string]*cover.Package) {
	visited := make(map[string]bool)
	for k, v := range pkgs {
		dst := filepath.Join(tmpBuildDir, "src", k)
		src := v.Dir

		if _, ok := visited[src]; ok {
			// Skip if already copied
			continue
		}

		if err := copy.Copy(src, dst); err != nil {
			log.Printf("Failed to Copy the folder from %v to %v, the error is: %v ", src, dst, err)
		}

		visited[src] = true

		cpDepPackages(tmpBuildDir, v, visited)
	}
}

// only cp dependency in root(current gopath),
// skip deps in other GOPATHs
func cpDepPackages(tmpBuildDir string, pkg *cover.Package, visited map[string]bool) {
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

		dst := filepath.Join(tmpBuildDir, "src", dep)

		if err := copy.Copy(src, dst); err != nil {
			log.Printf("Failed to Copy the folder from %v to %v, the error is: %v ", src, dst, err)
		}

		visited[src] = true
	}
}
