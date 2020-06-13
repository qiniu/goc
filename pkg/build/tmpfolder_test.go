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
	"path/filepath"
	"strings"
	"testing"
)

func TestNewDirParseInLegacyProject(t *testing.T) {
	workingDir := "../../tests/samples/simple_gopath_project/src/qiniu.com/simple_gopath_project"
	gopath, _ := filepath.Abs("../../tests/samples/simple_gopath_project")

	os.Chdir(workingDir)
	fmt.Println(gopath)
	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "off")

	b := NewInstall("", ".")
	if -1 == strings.Index(b.TmpWorkingDir, b.TmpDir) {
		t.Fatalf("Directory parse error. newwd: %v, tmpdir: %v", b.TmpWorkingDir, b.TmpDir)
	}

	if -1 == strings.Index(b.NewGOPATH, ":") || -1 == strings.Index(b.NewGOPATH, b.TmpDir) {
		t.Fatalf("The New GOPATH is wrong. newgopath: %v, tmpdir: %v", b.NewGOPATH, b.TmpDir)
	}

	b = NewBuild("", ".", "")
	if -1 == strings.Index(b.TmpWorkingDir, b.TmpDir) {
		t.Fatalf("Directory parse error. newwd: %v, tmpdir: %v", b.TmpWorkingDir, b.TmpDir)
	}

	if -1 == strings.Index(b.NewGOPATH, ":") || -1 == strings.Index(b.NewGOPATH, b.TmpDir) {
		t.Fatalf("The New GOPATH is wrong. newgopath: %v, tmpdir: %v", b.NewGOPATH, b.TmpDir)
	}
}

func TestNewDirParseInModProject(t *testing.T) {
	workingDir := "../../tests/samples/simple_project"
	gopath := ""

	os.Chdir(workingDir)
	fmt.Println(gopath)
	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")

	b := NewInstall("", ".")
	if -1 == strings.Index(b.TmpWorkingDir, b.TmpDir) {
		t.Fatalf("Directory parse error. newwd: %v, tmpdir: %v", b.TmpWorkingDir, b.TmpDir)
	}

	if b.NewGOPATH != "" {
		t.Fatalf("The New GOPATH is wrong. newgopath: %v, tmpdir: %v", b.NewGOPATH, b.TmpDir)
	}

	b = NewBuild("", ".", "")
	if -1 == strings.Index(b.TmpWorkingDir, b.TmpDir) {
		t.Fatalf("Directory parse error. newwd: %v, tmpdir: %v", b.TmpWorkingDir, b.TmpDir)
	}

	if b.NewGOPATH != "" {
		t.Fatalf("The New GOPATH is wrong. newgopath: %v, tmpdir: %v", b.NewGOPATH, b.TmpDir)
	}
}
