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
	"log"
	"os"
	"os/exec"
)

func (b *Build) Install() {
	log.Println("Go building in temp...")
	cmd := exec.Command("/bin/bash", "-c", "go install " + b.BuildFlags + " " + b.Packages)
	cmd.Dir = b.TmpWorkingDir

	// Change the temp GOBIN, to force binary install to original place
	cmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%v", b.findWhereToInstall()))
	if b.NewGOPATH != "" {
		// Change to temp GOPATH for go install command
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%v", b.NewGOPATH))
	}

	log.Printf("go install cmd is: %v", cmd.Args)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Fail to execute: %v. The error is: %v, the stdout/stderr is: %v", cmd.Args, err, string(out))
	}
	log.Printf("Go install successful. Binary installed in: %v", b.findWhereToInstall())
}
