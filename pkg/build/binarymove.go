/*
 Copyright 2020 Qiniu Cloud (七牛云)

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

func MvBinaryToOri(pkgs map[string]*cover.Package, newgopath string) {
	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			_, binaryTarget := filepath.Split(pkg.Target)

			binaryTmpPath := filepath.Join(getTmpwd(newgopath, pkgs), binaryTarget)

			if false == checkIfFileExist(binaryTmpPath) {
				continue
			}

			curwd, err := os.Getwd()
			if err != nil {
				log.Fatalf("Cannot get current working directoy, the error is: %v", err)
			}
			binaryOriPath := filepath.Join(curwd, binaryTarget)

			if checkIfFileExist(binaryOriPath) {
				// if we have file in the original place with same name,
				// but this file is not a binary,
				// then we skip it
				if false == checkIfExecutable(binaryOriPath) {
					log.Printf("Skipping binary: %v, as we find a file in the original place with same name but not executable.", binaryOriPath)
					continue
				}
			}

			log.Printf("Generating binary: %v", binaryOriPath)
			if err = copy.Copy(binaryTmpPath, binaryOriPath); err != nil {
				log.Println(err)
			}
		}
	}
}

func checkIfExecutable(path string) bool {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fileInfo.Mode()&0100 != 0
}

func checkIfFileExist(path string) bool {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !fileInfo.IsDir()
}
