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

package build

import (
	"encoding/json"
	"testing"

	"github.com/qiniu/goc/pkg/cover"
)

const TEST_GO_LIST_LEGACY = `{
	"Dir": "/Users/lyyyuna/gitup/linking/src/qiniu.com/linking/api/linking.v1",
	"ImportPath": "qiniu.com/linking/api/linking.v1",
	"Name": "linking",
	"Target": "/Users/lyyyuna/gitup/linking/pkg/darwin_amd64/qiniu.com/linking/api/linking.v1.a",
	"Root": "/Users/lyyyuna/gitup/linking",
	"Match": [
		"./..."
	],
	"Stale": true,
	"StaleReason": "stale dependency: vendor/github.com/modern-go/concurrent",
	"GoFiles": [
		"client.go"
	],
	"Imports": [
		"vendor/github.com/json-iterator/go",
		"github.com/qiniu/rpc.v2",
		"vendor/github.com/qiniu/xlog.v1",
		"vendor/qiniu.com/auth/qiniumac.v1"
	],
	"ImportMap": {
		"github.com/json-iterator/go": "vendor/github.com/json-iterator/go",
		"github.com/qiniu/xlog.v1": "vendor/github.com/qiniu/xlog.v1",
		"qiniu.com/auth/qiniumac.v1": "vendor/qiniu.com/auth/qiniumac.v1"
	},
	"Deps": [
		"bufio"
	]
}`

const TEST_GO_LIST_MOD = `{
	"Dir": "/Users/lyyyuna/gitup/tonghu-chat",
	"ImportPath": "github.com/lyyyuna/tonghu-chat",
	"Name": "main",
	"Target": "/Users/lyyyuna/go/bin/tonghu-chat",
	"Root": "/Users/lyyyuna/gitup/tonghu-chat",
	"Module": {
		"Path": "github.com/lyyyuna/tonghu-chat",
		"Main": true,
		"Dir": "/Users/lyyyuna/gitup/tonghu-chat",
		"GoMod": "/Users/lyyyuna/gitup/tonghu-chat/go.mod",
		"GoVersion": "1.14"
	},
	"Match": [
		"./..."
	],
	"Stale": true,
	"StaleReason": "not installed but available in build cache",
	"GoFiles": [
		"main.go"
	],
	"Imports": [
		"github.com/gin-gonic/gin",
		"github.com/gorilla/websocket"
	],
	"Deps": [
		"bufio"
	]
}`

func constructPkg(raw string) *cover.Package {
	var pkg cover.Package
	if err := json.Unmarshal([]byte(raw), &pkg); err != nil {
		panic(err)
	}
	return &pkg
}

func TestLegacyProjectJudgement(t *testing.T) {
	pkgs := make(map[string]*cover.Package)
	pkg := constructPkg(TEST_GO_LIST_LEGACY)
	pkgs[pkg.ImportPath] = pkg
	if expect, got := true, checkIfLegacyProject(pkgs); expect != got {
		t.Fatalf("Expected %v, but got %v.", expect, got)
	}
}

func TestModProjectJudgement(t *testing.T) {
	pkgs := make(map[string]*cover.Package)
	pkg := constructPkg(TEST_GO_LIST_MOD)
	pkgs[pkg.ImportPath] = pkg
	if expect, got := false, checkIfLegacyProject(pkgs); expect != got {
		t.Fatalf("Expected %v, but got %v.", expect, got)
	}
}
