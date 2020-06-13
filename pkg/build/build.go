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
	"github.com/qiniu/goc/pkg/cover"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Build is to describe the building/installing process of a goc build/install
type Build struct {
	Pkgs          map[string]*cover.Package // Pkg list parsed from "go list -json ./..." command
	NewGOPATH     string                    // the new GOPATH
	OriGOPATH     string                    // the original GOPATH
	TmpDir        string                    // the temporary directory to build the project
	TmpWorkingDir string                    // the working directory in the temporary directory, which is corresponding to the current directory in the project directory
	IsMod         bool                      // determine whether it is a Mod project
	BuildFlags    string                    // Build flags
	Packages      string                    // Packages that needs to build
	Root          string                    // Project Root
	Target        string                    // the binary name that go build generate
}

// NewBuild creates a Build struct which can build from goc temporary directory,
// and generate binary in current working directory
func NewBuild(buildflags string, packages string, outputDir string) *Build {
	// buildflags = buildflags + " -o " + outputDir
	b := &Build{
		BuildFlags: buildflags,
		Packages:   packages,
	}
	if false == b.validatePackageForBuild() {
		log.Fatalln("packages only support \".\"")
	}
	b.Target = b.determineOutputDir(outputDir)
	b.MvProjectsToTmp()
	return b
}

func (b *Build) Build() {
	log.Infoln("Go building in temp...")
	// new -o will overwrite  previous ones
	b.BuildFlags = b.BuildFlags + " -o " + b.Target
	cmd := exec.Command("/bin/bash", "-c", "go build "+b.BuildFlags+" "+b.Packages)
	cmd.Dir = b.TmpWorkingDir

	if b.NewGOPATH != "" {
		// Change to temp GOPATH for go install command
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%v", b.NewGOPATH))
	}

	log.Printf("go build cmd is: %v", cmd.Args)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Fail to execute: %v. The error is: %v, the stdout/stderr is: %v", cmd.Args, err, string(out))
	}
	log.Println("Go build exit successful.")
}

// determineOutputDir, as we only allow . as package name,
// the binary name is always same as the directory name of current directory
func (b *Build) determineOutputDir(outputDir string) string {
	curWorkingDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Cannot get current working directory, the err: %v.", err)
	}
	// if
	if outputDir == "" {
		_, last := filepath.Split(curWorkingDir)
		// replace "_" with "-" in the import path
		last = strings.ReplaceAll(last, "_", "-")
		return filepath.Join(curWorkingDir, last)
	}
	abs, err := filepath.Abs(outputDir)
	if err != nil {
		log.Fatalf("Fail to transform the path: %v to absolute path, the error is: %v", outputDir, err)
	}
	return abs
}

// validatePackageForBuild only allow . as package name
func (b *Build) validatePackageForBuild() bool {
	if b.Packages == "." {
		return true
	} else {
		return false
	}
}
