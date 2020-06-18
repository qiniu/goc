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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidPackage(t *testing.T) {

	workingDir := filepath.Join(baseDir, "../../tests/samples/simple_project")
	gopath := ""

	os.Chdir(workingDir)
	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")

	_, err := NewBuild("", []string{"example.com/simple-project"}, "")
	assert.Equal(t, err, ErrWrongPackageTypeForBuild, "the package name should be invalid")
}

func TestBasicBuildForModProject(t *testing.T) {
	workingDir := filepath.Join(baseDir, "../tests/samples/simple_project")
	gopath := ""

	os.Chdir(workingDir)
	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")

	buildFlags, args, buildOutput := "", []string{"."}, ""
	gocBuild, err := NewBuild(buildFlags, args, buildOutput)
	assert.Equal(t, err, nil, "should create temporary directory successfully")

	err = gocBuild.Build()
	assert.Equal(t, err, nil, "temporary directory should build successfully")
}
