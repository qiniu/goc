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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidPackage(t *testing.T) {

	workingDir := filepath.Join(baseDir, "../../tests/samples/simple_project")
	gopath := ""

	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")

	_, err := NewBuild("", []string{"example.com/simple-project"}, workingDir, "")
	if !assert.Equal(t, err, ErrWrongPackageTypeForBuild) {
		assert.FailNow(t, "the package name should be invalid")
	}
}

func TestBasicBuildForModProject(t *testing.T) {
	workingDir := filepath.Join(baseDir, "../../tests/samples/simple_project")
	gopath := ""

	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")
	fmt.Println(workingDir)
	buildFlags, args, buildOutput := "", []string{"."}, ""
	gocBuild, err := NewBuild(buildFlags, args, workingDir, buildOutput)
	if !assert.Equal(t, err, nil) {
		assert.FailNow(t, "should create temporary directory successfully")
	}

	err = gocBuild.Build()
	if !assert.Equal(t, err, nil) {
		assert.FailNow(t, "temporary directory should build successfully")
	}
}

func TestCheckParameters(t *testing.T) {
	err := checkParameters([]string{"aa", "bb"}, "aa")
	assert.Equal(t, err, ErrTooManyArgs, "too many arguments should failed")

	err = checkParameters([]string{"aa"}, "")
	assert.Equal(t, err, ErrInvalidWorkingDir, "empty working directory should failed")
}

func TestDetermineOutputDir(t *testing.T) {
	b := &Build{}
	_, err := b.determineOutputDir("")
	assert.Equal(t, errors.Is(err, ErrWrongCallSequence), true, "called before Build.MvProjectsToTmp() should fail")

	b.TmpDir = "fake"
	_, err = b.determineOutputDir("xx")
	assert.Equal(t, err, nil, "should return a directory")
}

func TestInvalidPackageNameForBuild(t *testing.T) {
	workingDir := filepath.Join(baseDir, "../../tests/samples/simple_project")
	gopath := filepath.Join(baseDir, "../../tests/samples/simple_project", "testhome")

	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")

	buildFlags, packages := "", []string{"main.go"}
	_, err := NewBuild(buildFlags, packages, workingDir, "")
	if !assert.Equal(t, err, ErrWrongPackageTypeForBuild) {
		assert.FailNow(t, "should not success with non . or ./... package")
	}
}

// test NewBuild with wrong parameters
func TestNewBuildWithWrongParameters(t *testing.T) {
	_, err := NewBuild("", []string{"a.go", "b.go"}, "cur", "cur")
	assert.Equal(t, err, ErrTooManyArgs)

	_, err = NewBuild("", []string{"a.go"}, "", "cur")
	assert.Equal(t, err, ErrInvalidWorkingDir)
}
