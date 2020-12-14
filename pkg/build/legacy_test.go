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
	"strings"
	"testing"

	"github.com/qiniu/goc/pkg/cover"
	"github.com/stretchr/testify/assert"
	"os"
)

// copy in cpLegacyProject/cpNonStandardLegacy of invalid src, dst name
func TestLegacyProjectCopyWithUnexistedDir(t *testing.T) {
	pkgs := make(map[string]*cover.Package)
	pkgs["main"] = &cover.Package{
		Module: &cover.ModulePublic{
			Dir: "not exied, ia mas duser", // not real one, should fail copy
		},
		Dir:     "not exit, iasdfs",
		Name:    "main",
		GoFiles: []string{"not_exist.go"},
	}
	pkgs["another"] = &cover.Package{}
	b := &Build{
		TmpDir: "sdfsfev2234444", // not real one, should fail copy
		Pkgs:   pkgs,
	}

	output := captureOutput(b.cpLegacyProject)
	assert.Equal(t, strings.Contains(output, "Failed to Copy"), true)
}

func TestCopyDir(t *testing.T) {
	wd, _ := os.Getwd()
	pkg := &cover.Package{Dir: wd, Root: wd, GoFiles: []string{"build.go", "legacy.go"}, CgoFiles: []string{"run.go"}}
	tmpDir := "/tmp/test/"
	b := &Build{
		WorkingDir: "empty",
		TmpDir:     tmpDir,
	}
	assert.NoError(t, os.MkdirAll(tmpDir, os.ModePerm))
	defer os.RemoveAll(tmpDir)
	assert.NoError(t, b.copyDir(pkg))
}
