/*
 Copyright 2020 Qiniu Cloud (七牛云)

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

package work

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// An Action holds global state about an action.
type Action struct {
	WorkDir string // the temporary work directory
}

// NewAction construct a new action
func NewAction() *Action {
	tmp, err := ioutil.TempDir(os.Getenv("GOTMPDIR"), "goc-build")
	if err != nil {
		log.Fatalf("goc: create work dir: %v", err)
	}
	if !filepath.IsAbs(tmp) {
		abs, err := filepath.Abs(tmp)
		if err != nil {
			os.RemoveAll(tmp)
			log.Fatalf("goc: create work dir: %v", err)
		}
		tmp = abs
	}

	a := &Action{
		WorkDir: tmp,
	}
	return a
}

func (a *Action) runList() error {
	return nil
}

func (a *Action) run() error {
	return nil
}

// cover runs, in effect,
// go tool cover -mode=a.coverMode -var="varName" -o dst.go src.go
// refer:https://github.com/golang/go/blob/f092be8fd839f5e61745c1b7f3b5990b4b8d6565/src/cmd/go/internal/work/exec.go#L1735
func (a *Action) cover(dst, src string, varName string) error {
	return nil
}
