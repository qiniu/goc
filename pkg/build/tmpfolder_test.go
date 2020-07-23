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

	"github.com/stretchr/testify/assert"
)

var baseDir string

func init() {
	baseDir, _ = os.Getwd()
}

func TestNewDirParseInLegacyProject(t *testing.T) {
	workingDir := filepath.Join(baseDir, "../../tests/samples/simple_gopath_project/src/qiniu.com/simple_gopath_project")
	gopath := filepath.Join(baseDir, "../../tests/samples/simple_gopath_project")

	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "off")

	b, _ := NewInstall("", []string{"."}, workingDir)
	if -1 == strings.Index(b.TmpWorkingDir, b.TmpDir) {
		t.Fatalf("Directory parse error. newwd: %v, tmpdir: %v", b.TmpWorkingDir, b.TmpDir)
	}

	if -1 == strings.Index(b.NewGOPATH, ":") || -1 == strings.Index(b.NewGOPATH, b.TmpDir) {
		t.Fatalf("The New GOPATH is wrong. newgopath: %v, tmpdir: %v", b.NewGOPATH, b.TmpDir)
	}

	b, _ = NewBuild("", []string{"."}, workingDir, "")
	if -1 == strings.Index(b.TmpWorkingDir, b.TmpDir) {
		t.Fatalf("Directory parse error. newwd: %v, tmpdir: %v", b.TmpWorkingDir, b.TmpDir)
	}

	if -1 == strings.Index(b.NewGOPATH, ":") || -1 == strings.Index(b.NewGOPATH, b.TmpDir) {
		t.Fatalf("The New GOPATH is wrong. newgopath: %v, tmpdir: %v", b.NewGOPATH, b.TmpDir)
	}
}

func TestNewDirParseInModProject(t *testing.T) {
	workingDir := filepath.Join(baseDir, "../../tests/samples/simple_project")
	gopath := ""

	fmt.Println(gopath)
	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")

	b, _ := NewInstall("", []string{"."}, workingDir)
	if -1 == strings.Index(b.TmpWorkingDir, b.TmpDir) {
		t.Fatalf("Directory parse error. newwd: %v, tmpdir: %v", b.TmpWorkingDir, b.TmpDir)
	}

	if b.NewGOPATH != "" {
		t.Fatalf("The New GOPATH is wrong. newgopath: %v, tmpdir: %v", b.NewGOPATH, b.TmpDir)
	}

	b, _ = NewBuild("", []string{"."}, workingDir, "")
	if -1 == strings.Index(b.TmpWorkingDir, b.TmpDir) {
		t.Fatalf("Directory parse error. newwd: %v, tmpdir: %v", b.TmpWorkingDir, b.TmpDir)
	}

	if b.NewGOPATH != "" {
		t.Fatalf("The New GOPATH is wrong. newgopath: %v, tmpdir: %v", b.NewGOPATH, b.TmpDir)
	}
}

// Test #14
func TestLegacyProjectNotInGoPATH(t *testing.T) {
	workingDir := filepath.Join(baseDir, "../../tests/samples/simple_gopath_project/src/qiniu.com/simple_gopath_project")
	gopath := ""

	fmt.Println(gopath)
	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "off")

	b, _ := NewBuild("", []string{"."}, workingDir, "")
	if b.OriGOPATH != b.NewGOPATH {
		t.Fatalf("New GOPATH should be same with old GOPATH, for this kind of project. New: %v, old: %v", b.NewGOPATH, b.OriGOPATH)
	}

	_, err := os.Stat(filepath.Join(b.TmpDir, "main.go"))
	if err != nil {
		t.Fatalf("There should be a main.go in temporary directory directly, the error: %v", err)
	}
}

// test traversePkgsList error case
func TestTraversePkgsList(t *testing.T) {
	b := &Build{
		Pkgs: nil,
	}
	_, _, err := b.traversePkgsList()
	assert.EqualError(t, err, ErrShouldNotReached.Error())
}

// test getTmpwd error case
func TestGetTmpwd(t *testing.T) {
	b := &Build{
		Pkgs: nil,
	}
	_, err := b.getTmpwd()
	assert.EqualError(t, err, ErrShouldNotReached.Error())
}

// test findWhereToInstall
func TestFindWhereToInstall(t *testing.T) {
	// if a legacy project without project root find
	// should find no plcae to install
	b := &Build{
		Pkgs:  nil,
		IsMod: false,
		Root:  "",
	}
	_, err := b.findWhereToInstall()
	assert.EqualError(t, err, ErrNoplaceToInstall.Error())

	// if $GOBIN not found
	// and if $GOPATH not found
	// should install to $HOME/go/bin
	b = &Build{
		Pkgs:      nil,
		IsMod:     true,
		OriGOPATH: "",
	}
	placeToInstall, err := b.findWhereToInstall()
	expectedPlace := filepath.Join(os.Getenv("HOME"), "go", "bin")
	assert.Equal(t, placeToInstall, expectedPlace)
}
