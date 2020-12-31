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
	"strings"
	"testing"
	"time"

	"github.com/qiniu/goc/pkg/cover"
	"github.com/stretchr/testify/assert"
)

// copy in cpLegacyProject/cpNonStandardLegacy of invalid src, dst name
func TestLegacyProjectCopyWithUnexistedDir(t *testing.T) {
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
	}

	output := captureOutput(b.cpLegacyProject)
	assert.Equal(t, strings.Contains(output, "Failed to Copy"), true)

	output = captureOutput(b.cpNonStandardLegacy)
	assert.Equal(t, strings.Contains(output, "Failed to Copy"), true)
}

// copy in cpDepPackages of invalid dst name
func TestDepPackagesCopyWithInvalidDir(t *testing.T) {
	gopath := filepath.Join(baseDir, "../../tests/samples/simple_gopath_project")
	pkg := &cover.Package{
		Module: &cover.ModulePublic{
			Dir: "not exied, ia mas duser",
		},
		Root: gopath,
		Deps: []string{"qiniu.com", "ddfee 2344234"},
	}
	b := &Build{
		TmpDir: "/", // "/" is invalid dst in Linux, it should fail
	}

	output := captureOutput(func() {
		visited := make(map[string]bool)

		b.cpDepPackages(pkg, visited)
	})
	assert.Equal(t, strings.Contains(output, "Failed to Copy"), true)
}

type MockFile struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m MockFile) Name() string {
	return m.name
}

func (m MockFile) Size() int64 {
	return m.size
}

func (m MockFile) Mode() os.FileMode {
	return m.mode
}

func (m MockFile) ModTime() time.Time {
	return m.modTime
}

func (m MockFile) IsDir() bool {
	return m.isDir
}

func (m MockFile) Sys() interface{} {
	return nil
}

// skipCopy verify
func TestSkipCopy(t *testing.T) {
	testCases := map[string]struct {
		inputSrc  string
		inputInfo MockFile
		expected  bool
	}{
		"src with /.git suffix":    {inputSrc: "/test/.git", inputInfo: MockFile{mode: 0}, expected: true},
		"src with ./git suffix":    {inputSrc: "/test.git", inputInfo: MockFile{mode: 0}, expected: false},
		"src with /.gita suffix":   {inputSrc: "/test/.gita", inputInfo: MockFile{mode: 0}, expected: false},
		"src with /.git in middle": {inputSrc: "/test/.git/test", inputInfo: MockFile{mode: 0}, expected: false},
		"irregular file":           {inputSrc: "/test", inputInfo: MockFile{mode: os.ModeIrregular}, expected: true},
		"dir file":                 {inputSrc: "/test", inputInfo: MockFile{isDir: true, mode: os.ModeDir}, expected: false},
		"temporary file":           {inputSrc: "/test", inputInfo: MockFile{mode: os.ModeTemporary}, expected: false},
		"symlink file":             {inputSrc: "/test", inputInfo: MockFile{mode: os.ModeSymlink}, expected: true},
		"device file":              {inputSrc: "/test", inputInfo: MockFile{mode: os.ModeDevice}, expected: true},
		"named pipe file":          {inputSrc: "/test", inputInfo: MockFile{mode: os.ModeNamedPipe}, expected: true},
		"socket file":              {inputSrc: "/test", inputInfo: MockFile{mode: os.ModeSocket}, expected: true},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			output, err := skipCopy(tc.inputSrc, tc.inputInfo)
			assert.NoError(t, err)
			assert.Equal(t, output, tc.expected)
		})
	}
}
