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
	"strings"
	"testing"

	"github.com/qiniu/goc/pkg/cover"
	"github.com/stretchr/testify/assert"
)

// copy in cpProject of invalid src, dst name
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

	output := captureOutput(b.cpProject)
	assert.Equal(t, strings.Contains(output, "Failed to Copy"), true)
}

// copy goMod project without go.mod
func TestGoModProjectCopyWithUnexistedModFile(t *testing.T) {
	pkgs := make(map[string]*cover.Package)
	pkgs["main"] = &cover.Package{
		Module: &cover.ModulePublic{
			Dir: "not exied, ia mas duser", // not real one, should fail copy
		},
		Dir:  "not exit, iasdfs",
		Name: "main",
	}
	pkgs["another"] = &cover.Package{}
	b := &Build{
		TmpDir: "sdfsfev2234444", // not real one, should fail copy
		Pkgs:   pkgs,
		IsMod:  true,
	}

	output := captureOutput(b.cpProject)
	assert.Equal(t, strings.Contains(output, "Failed to Copy the go mod file"), true)
}

// copy needed files to tmpDir
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

func initSlicesForTest() *cover.Package {
	var pkg cover.Package
	pkg.GoFiles = []string{"a1.go", "b1.go", "c1.go"}
	pkg.CompiledGoFiles = []string{"a2.go", "b2.go", "c2.go"}
	pkg.IgnoredGoFiles = []string{"a3.go", "b3.go", "c3.go"}
	pkg.CFiles = []string{"a4.go", "b4.go", "c4.go"}
	pkg.CXXFiles = []string{"a5.go", "b5.go", "c5.go"}
	pkg.MFiles = []string{"a6.go", "b6.go", "c6.go"}
	pkg.HFiles = []string{"a7.go", "b7.go", "c7.go"}
	pkg.FFiles = []string{"a8.go", "b8.go", "c8.go"}
	pkg.SFiles = []string{"a9.go", "b9.go", "c9.go"}
	pkg.SwigCXXFiles = []string{"a10.go", "b10.go", "c10.go"}
	pkg.SwigFiles = []string{"a11.go", "b11.go", "c11.go"}
	pkg.SysoFiles = []string{"a12.go", "b12.go", "c12.go"}

	return &pkg
}

// benchmark getFileListNeedsCopy
// goos: darwin
// goarch: amd64
// pkg: github.com/qiniu/goc/pkg/build
// BenchmarkGetFileList-4           3884557               298 ns/op             576 B/op          1 allocs/op
func BenchmarkGetFileList(b *testing.B) {
	//var files []string
	pkg := initSlicesForTest()
	for n := 0; n < b.N; n++ {
		getFileListNeedsCopy(pkg)
	}
}
