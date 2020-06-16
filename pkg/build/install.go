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

	log "github.com/sirupsen/logrus"
)

// NewInstall creates a Build struct which can install from goc temporary directory
func NewInstall(buildflags string, packages string) (*Build, error) {
	b := &Build{
		BuildFlags: buildflags,
		Packages:   packages,
	}
	if false == b.validatePackageForInstall() {
		log.Errorln(ErrWrongPackageTypeForInstall)
		return nil, ErrWrongPackageTypeForInstall
	}
	b.MvProjectsToTmp()
	return b, nil
}

func (b *Build) Install() error {
	log.Println("Go building in temp...")
	cmd := exec.Command("/bin/bash", "-c", "go install "+b.BuildFlags+" "+b.Packages)
	cmd.Dir = b.TmpWorkingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	whereToInstall, err := b.findWhereToInstall()
	if err != nil {
		// ignore the err
		log.Errorf("No place to install: %v", err)
	}
	// Change the temp GOBIN, to force binary install to original place
	cmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%v", whereToInstall))
	if b.NewGOPATH != "" {
		// Change to temp GOPATH for go install command
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%v", b.NewGOPATH))
	}

	log.Infof("go install cmd is: %v", cmd.Args)
	err = cmd.Start()
	if err != nil {
		log.Errorf("Fail to execute: %v. The error is: %v", cmd.Args, err)
		return fmt.Errorf("fail to execute: %v: %w", cmd.Args, err)
	}
	if err = cmd.Wait(); err != nil {
		log.Errorf("go install failed. The error is: %v", err)
		return fmt.Errorf("go install failed: %w", err)
	}
	log.Infof("Go install successful. Binary installed in: %v", whereToInstall)
	return nil
}

func (b *Build) validatePackageForInstall() bool {
	if b.Packages == "." || b.Packages == "./..." {
		return true
	}
	return false
}
