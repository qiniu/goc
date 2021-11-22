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
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/qiniu/goc/pkg/cover"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// MvProjectsToTmp moves the projects into a temporary directory
func (b *Build) MvProjectsToTmp() error {
	listArgs := []string{"-json"}
	if len(b.BuildFlags) != 0 {
		listArgs = append(listArgs, b.BuildFlags)
	}
	listArgs = append(listArgs, "./...")
	var err error
	b.Pkgs, err = cover.ListPackages(b.WorkingDir, strings.Join(listArgs, " "), "")
	if err != nil {
		log.Errorln(err)
		return err
	}

	err = b.mvProjectsToTmp()
	if err != nil {
		log.Errorf("Fail to move the project to temporary directory")
		return err
	}
	b.OriGOPATH = os.Getenv("GOPATH")
	if b.IsMod {
		b.NewGOPATH = ""
	} else if b.OriGOPATH == "" {
		b.NewGOPATH = b.TmpDir
	} else {
		b.NewGOPATH = fmt.Sprintf("%v:%v", b.TmpDir, b.OriGOPATH)
	}
	// fix #14: unable to build project not in GOPATH in legacy mode
	// this kind of project does not have a pkg.Root value
	// go 1.11, 1.12 has no pkg.Root,
	// so add b.IsMod == false as secondary judgement
	if b.NewGOPATH == "" && b.Root == "" && !b.IsMod {
		b.NewGOPATH = b.OriGOPATH
	}
	log.Infof("New GOPATH: %v", b.NewGOPATH)
	return nil
}

func (b *Build) mvProjectsToTmp() error {
	b.TmpDir = filepath.Join(os.TempDir(), tmpFolderName(b.WorkingDir))

	// Delete previous tmp folder and its content
	os.RemoveAll(b.TmpDir)
	// Create a new tmp folder and a new importpath for storing cover variables
	b.GlobalCoverVarImportPath = filepath.Join("src", tmpPackageName(b.WorkingDir))
	err := os.MkdirAll(filepath.Join(b.TmpDir, b.GlobalCoverVarImportPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Fail to create the temporary build directory. The err is: %v", err)
	}
	log.Infof("Tmp project generated in: %v", b.TmpDir)

	// traverse pkg list to get project meta info
	b.IsMod, b.Root, err = b.traversePkgsList()
	log.Infof("mod project? %v", b.IsMod)
	if errors.Is(err, ErrShouldNotReached) {
		return fmt.Errorf("mvProjectsToTmp with a empty project: %w", err)
	}
	// we should get corresponding working directory in temporary directory
	b.TmpWorkingDir, err = b.getTmpwd()
	if err != nil {
		return fmt.Errorf("getTmpwd failed with error: %w", err)
	}
	// issue #14
	// if b.Root == "", then the project is non-standard project
	// known cases:
	// 1. a legacy project, but not in any GOPATH, will cause the b.Root == ""
	if b.IsMod == false && b.Root != "" {
		b.cpLegacyProject()
	} else if b.IsMod == true { // go 1.11, 1.12 has no Build.Root
		b.cpGoModulesProject()
		updated, newGoModContent, err := b.updateGoModFile()
		if err != nil {
			return fmt.Errorf("fail to generate new go.mod: %v", err)
		}
		if updated {
			log.Infoln("go.mod needs rewrite")
			tmpModFile := filepath.Join(b.TmpDir, "go.mod")
			err := ioutil.WriteFile(tmpModFile, newGoModContent, os.ModePerm)
			if err != nil {
				return fmt.Errorf("fail to update go.mod: %v", err)
			}
		}
	} else if b.IsMod == false && b.Root == "" {
		b.TmpWorkingDir = b.TmpDir
		b.cpNonStandardLegacy()
	}

	log.Infof("New workingdir in tmp directory in: %v", b.TmpWorkingDir)
	return nil
}

// tmpFolderName uses the first six characters of the input path's SHA256 checksum
// as the suffix.
func tmpFolderName(path string) string {
	sum := sha256.Sum256([]byte(path))
	h := fmt.Sprintf("%x", sum[:6])

	return "goc-build-" + h
}

// tmpPackageName uses the first six characters of the input path's SHA256 checksum
// as the suffix.
func tmpPackageName(path string) string {
	sum := sha256.Sum256([]byte(path))
	h := fmt.Sprintf("%x", sum[:6])

	return "gocbuild" + h
}

// traversePkgsList travse the Build.Pkgs list
// return Build.IsMod, tell if the project is a mod project
// return Build.Root:
// 1. the project root if it is a mod project,
// 2. current GOPATH if it is a legacy project,
// 3. some non-standard project, which Build.IsMod == false, Build.Root == nil
func (b *Build) traversePkgsList() (isMod bool, root string, err error) {
	for _, v := range b.Pkgs {
		// get root
		root = v.Root
		if v.Module == nil {
			return
		}
		isMod = true
		b.ModRoot = v.Module.Dir
		b.ModRootPath = v.Module.Path
		return
	}
	log.Error(ErrShouldNotReached)
	err = ErrShouldNotReached
	return
}

// getTmpwd get the corresponding working directory in the temporary working directory
// and store it in the Build.tmpWorkdingDir
func (b *Build) getTmpwd() (string, error) {
	for _, pkg := range b.Pkgs {
		var index int
		var parentPath string
		if b.IsMod == false {
			index = strings.Index(b.WorkingDir, pkg.Root)
			parentPath = pkg.Root
		} else {
			index = strings.Index(b.WorkingDir, pkg.Module.Dir)
			parentPath = pkg.Module.Dir
		}

		if index == -1 {
			return "", ErrGocShouldExecInProject
		}
		// b.TmpWorkingDir = filepath.Join(b.TmpDir, path[len(parentPath):])
		return filepath.Join(b.TmpDir, b.WorkingDir[len(parentPath):]), nil
	}

	return "", ErrShouldNotReached
}

func (b *Build) findWhereToInstall() (string, error) {
	if GOBIN := os.Getenv("GOBIN"); GOBIN != "" {
		return GOBIN, nil
	}

	if false == b.IsMod {
		if b.Root == "" {
			return "", ErrNoPlaceToInstall
		}
		return filepath.Join(b.Root, "bin"), nil
	}
	if b.OriGOPATH != "" {
		return filepath.Join(strings.Split(b.OriGOPATH, ":")[0], "bin"), nil
	}
	return filepath.Join(os.Getenv("HOME"), "go", "bin"), nil
}

// Clean clears up the temporary workspace
func (b *Build) Clean() error {
	if !viper.GetBool("debug") {
		return os.RemoveAll(b.TmpDir)
	}
	return nil
}
