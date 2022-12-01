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

	"github.com/qiniu/goc/pkg/qiniu"
	log "github.com/sirupsen/logrus"
)

// Run excutes the main package in addition with the internal goc features
func (b *Build) Run() error {
	cmd := qiniu.ShellCommand("go run " + b.BuildFlags + " " + b.GoRunExecFlag + " " + b.Packages + " " + b.GoRunArguments)
	cmd.Dir = b.TmpWorkingDir

	if b.NewGOPATH != "" {
		// Change to temp GOPATH for go install command
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%v", b.NewGOPATH))
	}

	log.Infof("go build cmd is: %v", cmd.Args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("fail to execute: %v, err: %w", cmd.Args, err)
	}

	if err = cmd.Wait(); err != nil {
		return fmt.Errorf("fail to execute: %v, err: %w", cmd.Args, err)
	}

	return nil
}
