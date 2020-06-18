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

package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInstalledBinaryForMod(t *testing.T) {
	startTime := time.Now()

	workingDir := filepath.Join(baseDir, "../tests/samples/simple_project")
	gopath := filepath.Join(baseDir, "../tests/samples/simple_project", "testhome")

	os.Chdir(workingDir)
	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")

	buildFlags, buildOutput = "", ""
	args := []string{"."}
	runInstall(args)

	obj := filepath.Join(gopath, "bin", "simple-project")
	fInfo, err := os.Lstat(obj)
	assert.Equal(t, err, nil, "the binary should be generated.")
	assert.Equal(t, startTime.Before(fInfo.ModTime()), true, obj+"new binary should be generated, not the old one")

	cmd := exec.Command("go", "tool", "objdump", "simple-project")
	cmd.Dir = workingDir
	out, _ := cmd.CombinedOutput()
	cnt := strings.Count(string(out), "main.registerSelf")
	assert.Equal(t, cnt > 0, true, "main.registerSelf function should be in the binary")

	cnt = strings.Count(string(out), "GoCover")
	assert.Equal(t, cnt > 0, true, "GoCover varibale should be in the binary")
}

func TestInstalledBinaryForLegacy(t *testing.T) {
	startTime := time.Now()

	workingDir := filepath.Join(baseDir, "../tests/samples/simple_gopath_project/src/qiniu.com/simple_gopath_project")
	gopath := filepath.Join(baseDir, "../tests/samples/simple_gopath_project")

	os.Chdir(workingDir)
	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "off")

	buildFlags, buildOutput = "", ""
	args := []string{"."}
	runInstall(args)

	obj := filepath.Join(gopath, "bin", "simple_gopath_project")
	fInfo, err := os.Lstat(obj)
	assert.Equal(t, err, nil, "the binary should be generated.")
	assert.Equal(t, startTime.Before(fInfo.ModTime()), true, obj+"new binary should be generated, not the old one")

	cmd := exec.Command("go", "tool", "objdump", obj)
	cmd.Dir = workingDir
	out, _ := cmd.CombinedOutput()
	cnt := strings.Count(string(out), "main.registerSelf")
	assert.Equal(t, cnt > 0, true, "main.registerSelf function should be in the binary")

	cnt = strings.Count(string(out), "GoCover")
	assert.Equal(t, cnt > 0, true, "GoCover varibale should be in the binary")
}
