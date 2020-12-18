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
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qiniu/goc/pkg/cover"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return buf.String()
}

// copy in cpGoModulesProject of invalid src, dst name
func TestModProjectCopyWithUnexistedDir(t *testing.T) {
	pkgs := make(map[string]*cover.Package)
	pkgs["main"] = &cover.Package{
		Name: "main",
		Module: &cover.ModulePublic{
			Dir: "not exied, ia mas duser", // not real one, should fail copy
		},
		GoFiles: []string{"empty.go"},
	}
	pkgs["another"] = &cover.Package{}
	b := &Build{
		TmpDir: "sdfsfev2234444", // not real one, should fail copy
		Pkgs:   pkgs,
	}

	output := captureOutput(b.cpProject)
	assert.Equal(t, strings.Contains(output, "Failed to Copy"), true)
}

// test go mod file udpate
func TestUpdateModFileIfContainsReplace(t *testing.T) {
	workingDir := filepath.Join(baseDir, "../../tests/samples/gomod_samples/a")
	b := &Build{
		TmpDir:  workingDir,
		ModRoot: "/aa/bb/cc",
	}

	// replace with relative local file path should be rewrite
	updated, newmod, err := b.updateGoModFile()
	assert.Equal(t, err, nil)
	assert.Equal(t, updated, true)
	assert.Contains(t, string(newmod), "replace github.com/qiniu/bar => /aa/bb/home/foo/bar")

	// old replace should be removed
	assert.NotContains(t, string(newmod), "github.com/qiniu/bar => ../home/foo/bar")

	// normal replace should not be rewrite
	assert.Contains(t, string(newmod), "github.com/qiniu/bar2 => github.com/baniu/bar3 v1.2.3")
}

// test wrong go mod file
func TestWithWrongGoModFile(t *testing.T) {
	// go.mod not exist
	workingDir := filepath.Join(baseDir, "../../tests/samples/xxxxxxxxxxxx/a")
	b := &Build{
		TmpDir:  workingDir,
		ModRoot: "/aa/bb/cc",
	}

	updated, _, err := b.updateGoModFile()
	assert.Equal(t, errors.Is(err, os.ErrNotExist), true)
	assert.Equal(t, updated, false)

	// a wrong format go mod
	workingDir = filepath.Join(baseDir, "../../tests/samples/gomod_samples/b")
	b = &Build{
		TmpDir:  workingDir,
		ModRoot: "/aa/bb/cc",
	}

	updated, _, err = b.updateGoModFile()
	assert.NotEqual(t, err, nil)
	assert.Equal(t, updated, false)
}
