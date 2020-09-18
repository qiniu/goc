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

package cover

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
)

func testCoverage() (c *Coverage) {
	return &Coverage{FileName: "fake-coverage", NCoveredStmts: 200, NAllStmts: 300}
}

func TestCoverageRatio(t *testing.T) {
	c := testCoverage()
	actualRatio, _ := c.Ratio()
	assert.Equal(t, float32(c.NCoveredStmts)/float32(c.NAllStmts), actualRatio)
}

func TestRatioErr(t *testing.T) {
	c := &Coverage{FileName: "fake-coverage", NCoveredStmts: 200, NAllStmts: 0}
	_, err := c.Ratio()
	assert.NotEqual(t, err, nil)
}

func TestPercentageNA(t *testing.T) {
	c := &Coverage{FileName: "fake-coverage", NCoveredStmts: 200, NAllStmts: 0}
	assert.Equal(t, "N/A", c.Percentage())
}

func TestCovList(t *testing.T) {
	fileName := "qiniu.com/kodo/apiserver/server/main.go"
	fileName1 := "qiniu.com/kodo/apiserver/server/svr.go"

	items := []struct {
		profile   string
		expectPer []string
	}{
		// percentage is 100%
		{
			profile: "mode: atomic\n" +
				fileName + ":32.49,33.13 1 30\n",
			expectPer: []string{"100.0%"},
		},
		// percentage is 50%
		{profile: "mode: atomic\n" +
			fileName + ":32.49,33.13 1 30\n" +
			fileName + ":42.49,43.13 1 0\n",
			expectPer: []string{"50.0%"},
		},
		// two files
		{
			profile: "mode: atomic\n" +
				fileName + ":32.49,33.13 1 30\n" +
				fileName1 + ":42.49,43.13 1 0\n",
			expectPer: []string{"100.0%", "0.0%"},
		},
	}

	for _, tc := range items {
		r := strings.NewReader(tc.profile)
		c, err := CovList(r)
		c.Sort()
		assert.Equal(t, err, nil)
		for k, v := range c {
			assert.Equal(t, tc.expectPer[k], v.Percentage())
		}
	}
}

func TestReadFileToCoverList(t *testing.T) {
	path := "unknown"
	_, err := ReadFileToCoverList(path)
	assert.Equal(t, err.Error(), "open unknown: no such file or directory")
}

func TestTotalPercentage(t *testing.T) {
	items := []struct {
		list      CoverageList
		expectPer string
	}{
		{
			list:      CoverageList{Coverage{FileName: "fake-coverage", NCoveredStmts: 15, NAllStmts: 0}},
			expectPer: "N/A",
		},
		{
			list:      CoverageList{Coverage{FileName: "fake-coverage", NCoveredStmts: 15, NAllStmts: 20}},
			expectPer: "75.0%",
		},
		{
			list: CoverageList{Coverage{FileName: "fake-coverage", NCoveredStmts: 15, NAllStmts: 20},
				Coverage{FileName: "fake-coverage-1", NCoveredStmts: 10, NAllStmts: 30}},
			expectPer: "50.0%",
		},
	}

	for _, tc := range items {
		assert.Equal(t, tc.expectPer, tc.list.TotalPercentage())
	}
}

func TestBuildCoverCmd(t *testing.T) {
	var testCases = []struct {
		name      string
		file      string
		coverVar  *FileVar
		pkg       *Package
		mode      string
		newgopath string
		expectCmd *exec.Cmd
	}{
		{
			name: "normal",
			file: "c.go",
			coverVar: &FileVar{
				File: "example/b/c/c.go",
				Var:  "GoCover_0_643131623532653536333031",
			},
			pkg: &Package{
				Dir: "/go/src/goc/cmd/example-project/b/c",
			},
			mode:      "count",
			newgopath: "",
			expectCmd: &exec.Cmd{
				Path: lookCmdPath("go"),
				Args: []string{"go", "tool", "cover", "-mode", "count", "-var", "GoCover_0_643131623532653536333031", "-o",
					"/go/src/goc/cmd/example-project/b/c/c.go", "/go/src/goc/cmd/example-project/b/c/c.go"},
			},
		},
		{
			name: "normal with gopath",
			file: "c.go",
			coverVar: &FileVar{
				File: "example/b/c/c.go",
				Var:  "GoCover_0_643131623532653536333031",
			},
			pkg: &Package{
				Dir: "/go/src/goc/cmd/example-project/b/c",
			},
			mode:      "set",
			newgopath: "/go/src/goc",
			expectCmd: &exec.Cmd{
				Path: lookCmdPath("go"),
				Args: []string{"go", "tool", "cover", "-mode", "set", "-var", "GoCover_0_643131623532653536333031", "-o",
					"/go/src/goc/cmd/example-project/b/c/c.go", "/go/src/goc/cmd/example-project/b/c/c.go"},
				Env: append(os.Environ(), fmt.Sprintf("GOPATH=%v", "/go/src/goc")),
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cmd := buildCoverCmd(testCase.file, testCase.coverVar, testCase.pkg, testCase.mode, testCase.newgopath)
			if !reflect.DeepEqual(cmd, testCase.expectCmd) {
				t.Errorf("generated incorrect commands:\nGot: %#v\nExpected:%#v", cmd, testCase.expectCmd)
			}
		})
	}

}

func lookCmdPath(name string) string {
	if filepath.Base(name) == name {
		if lp, err := exec.LookPath(name); err != nil {
			log.Fatalf("find exec %s err: %v", name, err)
		} else {
			return lp
		}
	}
	return ""
}

func TestDeclareCoverVars(t *testing.T) {
	var testCases = []struct {
		name           string
		pkg            *Package
		expectCoverVar map[string]*FileVar
	}{
		{
			name: "normal",
			pkg: &Package{
				Dir:        "/go/src/goc/cmd/example-project/b/c",
				GoFiles:    []string{"c.go"},
				ImportPath: "example/b/c",
			},
			expectCoverVar: map[string]*FileVar{
				"c.go": {File: "example/b/c/c.go", Var: "GoCover_0_643131623532653536333031"},
			},
		},
		{
			name: "more go files",
			pkg: &Package{
				Dir:        "/go/src/goc/cmd/example-project/a/b",
				GoFiles:    []string{"printf.go", "printf1.go"},
				ImportPath: "example/a/b",
			},
			expectCoverVar: map[string]*FileVar{
				"printf.go":  {File: "example/a/b/printf.go", Var: "GoCover_0_326535623364613565313464"},
				"printf1.go": {File: "example/a/b/printf1.go", Var: "GoCover_1_326535623364613565313464"},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			coverVar := declareCoverVars(testCase.pkg)
			if !reflect.DeepEqual(coverVar, testCase.expectCoverVar) {
				t.Errorf("generated incorrect cover vars:\nGot: %#v\nExpected:%#v", coverVar, testCase.expectCoverVar)
			}
		})
	}

}

func TestGetInternalParent(t *testing.T) {
	var tcs = []struct {
		ImportPath     string
		expectedParent string
	}{
		{
			ImportPath:     "a/internal/b",
			expectedParent: "a",
		},
		{
			ImportPath:     "internal/b",
			expectedParent: "",
		},
		{
			ImportPath:     "a/b/internal/b",
			expectedParent: "a/b",
		},
		{
			ImportPath:     "a/b/internal",
			expectedParent: "a/b",
		},
		{
			ImportPath:     "a/b/internal/c",
			expectedParent: "a/b",
		},
		{
			ImportPath:     "a/b/c",
			expectedParent: "",
		},
		{
			ImportPath:     "",
			expectedParent: "",
		},
		{
			ImportPath:     "a/b/internal/c/d/internal/e",
			expectedParent: "a/b",
		},
	}

	for _, tc := range tcs {
		actual := getInternalParent(tc.ImportPath)
		if actual != tc.expectedParent {
			t.Errorf("getInternalParent failed for importPath %s, expected %s, got %s", tc.ImportPath, tc.expectedParent, actual)
		}
	}
}

func TestGetInternalCount(t *testing.T) {
	var tcs = []struct {
		ImportPath    string
		expectedCount int
	}{
		{
			ImportPath:    "a/internal/b",
			expectedCount: 1,
		},
		{
			ImportPath:    "internal/b",
			expectedCount: 1,
		},
		{
			ImportPath:    "a/b/internal/b",
			expectedCount: 1,
		},
		{
			ImportPath:    "a/b/internal",
			expectedCount: 1,
		},
		{
			ImportPath:    "a/b/internal/c",
			expectedCount: 1,
		},
		{
			ImportPath:    "a/b/c",
			expectedCount: 0,
		},
		{
			ImportPath:    "",
			expectedCount: 0,
		},
		{
			ImportPath:    "a/b/internal/c/d/internal/e",
			expectedCount: 2,
		},
		{
			ImportPath:    "a/b/internal/c/d/internal",
			expectedCount: 2,
		},
		{
			ImportPath:    "a/b/internal/c/d/internal/e/f",
			expectedCount: 2,
		},
		{
			ImportPath:    "a/b/internal/internal/e/f",
			expectedCount: 2,
		},
		{
			ImportPath:    "a/b/internal/internal",
			expectedCount: 2,
		},
		{
			ImportPath:    "a/b/internal/internal/d/f",
			expectedCount: 2,
		},
		{
			ImportPath:    "internal/internal/d/f",
			expectedCount: 2,
		},
		{
			ImportPath:    "internal/internal",
			expectedCount: 2,
		},
		{
			ImportPath:    "/internal/internal",
			expectedCount: 2,
		},
		{
			ImportPath:    "/a/b/internal/c/d/internal/e",
			expectedCount: 2,
		},
	}
	for _, tc := range tcs {
		actual := getInternalCount(tc.ImportPath)
		assert.Equal(t, actual, tc.expectedCount, fmt.Sprintf("getInternalCount failed for importPath %s, expected %d, got %d", tc.ImportPath, tc.expectedCount, actual))
	}
}

func TestFindInternal(t *testing.T) {
	var tcs = []struct {
		ImportPath     string
		expectedParent bool
	}{
		{
			ImportPath:     "a/internal/b",
			expectedParent: true,
		},
		{
			ImportPath:     "internal/b",
			expectedParent: true,
		},
		{
			ImportPath:     "a/b/internal",
			expectedParent: true,
		},
		{
			ImportPath:     "a/b/c",
			expectedParent: false,
		},
		{
			ImportPath:     "internal",
			expectedParent: true,
		},
	}

	for _, tc := range tcs {
		actual := hasInternalPath(tc.ImportPath)
		if actual != tc.expectedParent {
			t.Errorf("hasInternalPath check failed for importPath %s", tc.ImportPath)
		}
	}
}

func TestExecuteForSimpleModProject(t *testing.T) {
	workingDir := "../../tests/samples/simple_project"
	gopath := ""

	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")

	testDir := filepath.Join(os.TempDir(), "goc-build-test")
	copy.Copy(workingDir, testDir)

	Execute("", gopath, testDir, "count", "", "http://127.0.0.1:7777")

	_, err := os.Lstat(filepath.Join(testDir, "http_cover_apis_auto_generated.go"))
	if !assert.Equal(t, err, nil) {
		assert.FailNow(t, "should generate http_cover_apis_auto_generated.go")
	}
}

func TestListPackagesForSimpleModProject(t *testing.T) {
	workingDir := "../../tests/samples/simple_project"
	gopath := ""

	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")

	pkgs, _ := ListPackages(workingDir, "-json ./...", "")
	if !assert.Equal(t, len(pkgs), 1) {
		assert.FailNow(t, "should only have one pkg")
	}
	if pkg, ok := pkgs["example.com/simple-project"]; ok {
		assert.Equal(t, pkg.Module.Path, "example.com/simple-project")
	} else {
		assert.FailNow(t, "cannot get the pkg: example.com/simple-project")
	}

}

// test if goc can get variables in internal package
func TestCoverResultForInternalPackage(t *testing.T) {

	workingDir := "../../tests/samples/simple_project_with_internal"
	gopath := ""

	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")

	testDir := filepath.Join(os.TempDir(), "goc-build-test")
	copy.Copy(workingDir, testDir)

	Execute("", gopath, testDir, "count", "", "http://127.0.0.1:7777")

	_, err := os.Lstat(filepath.Join(testDir, "http_cover_apis_auto_generated.go"))
	if !assert.Equal(t, err, nil) {
		assert.FailNow(t, "should generate http_cover_apis_auto_generated.go")
	}

	out, err := ioutil.ReadFile(filepath.Join(testDir, "http_cover_apis_auto_generated.go"))
	if err != nil {
		assert.FailNow(t, "failed to read http_cover_apis_auto_generated.go file")
	}
	cnt := strings.Count(string(out), "GoCacheCover")
	assert.Equal(t, cnt > 0, true, "GoCacheCover variable should be in http_cover_apis_auto_generated.go")

	cnt = strings.Count(string(out), "example.com/simple-project/internal/foo.go")
	assert.Equal(t, cnt > 0, true, "`example.com/simple-project/internal/foo.go` should be in http_cover_apis_auto_generated.go")
}
