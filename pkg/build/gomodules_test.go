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
	"os"
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
	}
	pkgs["another"] = &cover.Package{}
	b := &Build{
		TmpDir: "sdfsfev2234444", // not real one, should fail copy
		Pkgs:   pkgs,
	}

	output := captureOutput(b.cpGoModulesProject)
	assert.Equal(t, strings.Contains(output, "Failed to Copy"), true)
}
