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
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/qiniu/goc/pkg/cover"
)

func MvProjectsToTmp(target string, args []string) (newgopath string, newWorkingDir string, tmpBuildDir string, pkgs map[string]*cover.Package) {
	listArgs := []string{"list", "-json"}
	if len(args) != 0 {
		listArgs = append(listArgs, args...)
	}
	listArgs = append(listArgs, "./...")
	pkgs = cover.ListPackages(target, listArgs, "")

	tmpBuildDir, newWorkingDir, isMod := mvProjectsToTmp(pkgs)
	origopath := os.Getenv("GOPATH")
	if isMod == true {
		newgopath = ""
	} else if origopath == "" {
		newgopath = tmpBuildDir
	} else {
		newgopath = fmt.Sprintf("%v:%v", tmpBuildDir, origopath)
	}
	log.Printf("New GOPATH: %v", newgopath)
	return
}

func mvProjectsToTmp(pkgs map[string]*cover.Package) (string, string, bool) {
	tmpBuildDir := os.TempDir() + tmpFolderName()

	// Delete previous tmp folder and its content
	os.RemoveAll(tmpBuildDir)
	// Create a new tmp folder
	err := os.MkdirAll(filepath.Join(tmpBuildDir, "src"), os.ModePerm)
	if err != nil {
		log.Fatalf("Fail to create the temporary build directory. The err is: %v", err)
	}
	log.Printf("Temp project generated in: %v", tmpBuildDir)

	tmpWorkingDir := getTmpwd(tmpBuildDir, pkgs)
	isMod := false
	if checkIfLegacyProject(pkgs) {
		cpLegacyProject(tmpBuildDir, pkgs)
	} else {
		cpGoModulesProject(tmpBuildDir, pkgs)
		isMod = true
	}

	log.Printf("New working/building directory in: %v", tmpWorkingDir)
	return tmpBuildDir, tmpWorkingDir, isMod
}

func tmpFolderName() string {
	path, err := os.Getwd()
	if err != nil {
		log.Fatalf("Cannot get current working directoy, the error is: %v", err)
	}
	sum := sha256.Sum256([]byte(path))
	h := fmt.Sprintf("%x", sum[:6])

	return "goc-" + h
}

// Check if it is go module project
// true legacy
// flase go mod
func checkIfLegacyProject(pkgs map[string]*cover.Package) bool {
	for _, v := range pkgs {

		if v.Module == nil {
			return true
		}
		return false
	}
	log.Fatalln("Should never be reached....")
	return false
}

func getTmpwd(tmpBuildDir string, pkgs map[string]*cover.Package) string {
	for _, pkg := range pkgs {
		path, err := os.Getwd()
		if err != nil {
			log.Fatalf("Cannot get current working directoy, the error is: %v", err)
		}
		index := strings.Index(path, pkg.Root)
		if index == -1 {
			log.Fatalf("goc install not executed in project directory.")
		}
		tmpwd := filepath.Join(tmpBuildDir, path[len(pkg.Root):])
		// log.Printf("New building directory in: %v", tmpwd)
		return tmpwd
	}

	log.Fatalln("Should never be reached....")
	return ""
}

func FindWhereToInstall(pkgs map[string]*cover.Package) string {
	if GOBIN := os.Getenv("GOBIN"); GOBIN != "" {
		return GOBIN
	}

	// old GOPATH dir
	GOPATH := os.Getenv("GOPATH")
	if true == checkIfLegacyProject(pkgs) {
		for _, v := range pkgs {
			return filepath.Join(v.Root, "bin")
		}
	}
	if GOPATH != "" {
		return filepath.Join(strings.Split(GOPATH, ":")[0], "bin")
	}
	return filepath.Join(os.Getenv("HOME"), "go", "bin")
}
