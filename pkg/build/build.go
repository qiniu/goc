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

	"github.com/qiniu/goc/pkg/cover"
	log "github.com/sirupsen/logrus"
)

// Build is to describe the building/installing process of a goc build/install
type Build struct {
	Pkgs          map[string]*cover.Package // Pkg list parsed from "go list -json ./..." command
	NewGOPATH     string                    // the new GOPATH
	OriGOPATH     string                    // the original GOPATH
	WorkingDir    string                    // the working directory
	TmpDir        string                    // the temporary directory to build the project
	TmpWorkingDir string                    // the working directory in the temporary directory, which is corresponding to the current directory in the project directory
	IsMod         bool                      // determine whether it is a Mod project
	Root          string
	// go 1.11, go 1.12 has no Root
	// Project Root:
	// 1. legacy, root == GOPATH
	// 2. mod, root == go.mod Dir
	ModRoot string // path for go.mod
	Target  string // the binary name that go build generate
	// keep compatible with go commands:
	// go run [build flags] [-exec xprog] package [arguments...]
	// go build [-o output] [-i] [build flags] [packages]
	// go install [-i] [build flags] [packages]
	BuildFlags     string // Build flags
	Packages       string // Packages that needs to build
	GoRunExecFlag  string // for the -exec flags in go run command
	GoRunArguments string // for the '[arguments]' parameters in go run command
}

// NewBuild creates a Build struct which can build from goc temporary directory,
// and generate binary in current working directory
func NewBuild(buildflags string, args []string, workingDir string, outputDir string) (*Build, error) {
	if err := checkParameters(args, workingDir); err != nil {
		return nil, err
	}
	// buildflags = buildflags + " -o " + outputDir
	b := &Build{
		BuildFlags: buildflags,
		Packages:   strings.Join(args, " "),
		WorkingDir: workingDir,
	}
	if false == b.validatePackageForBuild() {
		log.Errorln(ErrWrongPackageTypeForBuild)
		return nil, ErrWrongPackageTypeForBuild
	}
	if err := b.MvProjectsToTmp(); err != nil {
		return nil, err
	}
	dir, err := b.determineOutputDir(outputDir)
	b.Target = dir
	if err != nil {
		return nil, err
	}
	return b, nil
}

// Build calls 'go build' tool to do building
func (b *Build) Build() error {
	log.Infoln("Go building in temp...")
	// new -o will overwrite  previous ones
	b.BuildFlags = b.BuildFlags + " -o " + b.Target
	cmd := exec.Command("/bin/bash", "-c", "go build "+b.BuildFlags+" "+b.Packages)
	cmd.Dir = b.TmpWorkingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if b.NewGOPATH != "" {
		// Change to temp GOPATH for go install command
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%v", b.NewGOPATH))
	}

	log.Printf("go build cmd is: %v", cmd.Args)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("fail to execute: %v, err: %w", cmd.Args, err)
	}
	if err = cmd.Wait(); err != nil {
		return fmt.Errorf("fail to execute: %v, err: %w", cmd.Args, err)
	}
	log.Infoln("Go build exit successful.")
	return nil
}

// determineOutputDir, as we only allow . as package name,
// the binary name is always same as the directory name of current directory
func (b *Build) determineOutputDir(outputDir string) (string, error) {
	if b.TmpDir == "" {
		return "", fmt.Errorf("can only be called after Build.MvProjectsToTmp(): %w", ErrEmptyTempWorkingDir)
	}

	// fix #43
	if outputDir != "" {
		abs, err := filepath.Abs(outputDir)
		if err != nil {
			return "", fmt.Errorf("Fail to transform the path: %v to absolute path: %v", outputDir, err)

		}
		return abs, nil
	}
	// fix #43
	// use target name from `go list -json ./...` of the main module
	targetName := ""
	for _, pkg := range b.Pkgs {
		if pkg.Name == "main" {
			_, file := filepath.Split(pkg.Target)
			targetName = file
			break
		}
	}

	return filepath.Join(b.WorkingDir, targetName), nil
}

// validatePackageForBuild only allow . as package name
func (b *Build) validatePackageForBuild() bool {
	if b.Packages == "." || b.Packages == "" {
		return true
	}
	return false
}

func checkParameters(args []string, workingDir string) error {
	if len(args) > 1 {
		log.Errorln(ErrTooManyArgs)
		return ErrTooManyArgs
	}

	if workingDir == "" {
		return ErrInvalidWorkingDir
	}

	log.Infof("Working directory: %v", workingDir)
	return nil
}
