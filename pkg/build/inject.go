/*
 Copyright 2021 Qiniu Cloud (qiniu.com)
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
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/qiniu/goc/v2/pkg/build/internal/tool"
	"github.com/qiniu/goc/v2/pkg/build/internal/websocket"
	"github.com/qiniu/goc/v2/pkg/log"
)

// Inject injects cover variables for all the .go files in the target directory
func (b *Build) Inject() {
	log.StartWait("injecting cover variables")

	var seen = make(map[string]*PackageCover)

	// 所有插桩变量定义声明
	allDecl := ""

	pkgs := b.Pkgs
	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			// 该 main 二进制所关联的所有插桩变量的元信息
			// 每个 main 之间是不相关的，需要重新定义
			allMainCovers := make([]*PackageCover, 0)
			// 注入 main package
			mainCover, mainDecl := b.addCounters(pkg)
			// 收集插桩变量的定义和元信息
			allDecl += mainDecl
			allMainCovers = append(allMainCovers, mainCover)

			// 向 main package 的依赖注入插桩变量
			for _, dep := range pkg.Deps {
				if packageCover, ok := seen[dep]; ok {
					allMainCovers = append(allMainCovers, packageCover)
					continue
				}

				// 依赖需要忽略 Go 标准库和 go.mod 引入的第三方
				if depPkg, ok := pkgs[dep]; ok {
					// 注入依赖的 package
					packageCover, depDecl := b.addCounters(depPkg)
					// 收集插桩变量的定义和元信息
					allDecl += depDecl
					allMainCovers = append(allMainCovers, packageCover)
					// 避免重复访问
					seen[dep] = packageCover
				}
			}
			// 为每个 main 包注入 websocket handler
			b.injectGocAgent(b.getPkgTmpDir(pkg.Dir), allMainCovers)
			if b.Mode == "watch" {
				log.Donef("inject main package [%v] with rpcagent and watchagent", pkg.ImportPath)
			} else {
				log.Donef("inject main package [%v] with rpcagent", pkg.ImportPath)
			}
		}
	}
	// 在工程根目录注入所有插桩变量的声明+定义
	b.injectGlobalCoverVarFile(allDecl)

	// 添加自定义 websocket 依赖
	// 用户代码可能有 gorrila/websocket 的依赖，为避免依赖冲突，以及可能的 replace/vendor，
	// 这里直接注入一份完整的 gorrila/websocket 实现
	websocket.AddCustomWebsocketDep(b.GlobalCoverVarImportPathDir)
	log.Donef("websocket library injected")

	log.StopWait()
	log.Donef("global cover variables injected")
}

// addCounters is different from official go tool cover
//
// 1. only inject covervar++ into source file
//
// 2. no declarartions for these covervars
//
// 3. return the declarations as string
func (b *Build) addCounters(pkg *Package) (*PackageCover, string) {
	mode := b.Mode
	gobalCoverVarImportPath := b.GlobalCoverVarImportPath

	coverVarMap := declareCoverVars(pkg)

	decl := ""
	for file, coverVar := range coverVarMap {
		decl += "\n" + tool.Annotate(filepath.Join(b.getPkgTmpDir(pkg.Dir), file), mode, coverVar.Var, coverVar.File, gobalCoverVarImportPath) + "\n"
	}

	return &PackageCover{
		Package: pkg,
		Vars:    coverVarMap,
	}, decl
}

// getPkgTmpDir gets corresponding pkg dir in temporary project
//
// the reason is that config.GocConfig.Pkgs is get in the original project.
// we need to transfer the direcory.
//
// 在原工程目录已经做了一次 go list -json，在临时目录没有必要再做一遍，直接转换一下就能得到
// 临时目录中的 pkg.Dir。
func (b *Build) getPkgTmpDir(pkgDir string) string {
	relDir, err := filepath.Rel(b.CurModProjectDir, pkgDir)
	if err != nil {
		log.Fatalf("go json -list meta info wrong: %v", err)
	}

	return filepath.Join(b.TmpModProjectDir, relDir)
}

// injectGocAgent inject handlers like following
//
// - xxx.go
// - yyy_package
// - main.go
// - goc-cover-agent-apis-auto-generated-11111-22222-bridge.go
// - goc-cover-agent-apis-auto-generated-11111-22222-package
//  |
//  -- rpcagent.go
//  -- watchagent.go
//
// 11111_22222_bridge.go 仅仅用于引用 11111_22222_package, where package contains ws agent main logic.
// 使用 bridge.go 文件是为了避免插桩逻辑中的变量名污染 main 包
func (b *Build) injectGocAgent(where string, covers []*PackageCover) {
	if len(covers) == 0 {
		return
	}

	if len(covers[0].Vars) == 0 {
		return
	}

	injectPkgName := "goc-cover-agent-apis-auto-generated-11111-22222-package"
	injectBridgeName := "goc-cover-agent-apis-auto-generated-11111-22222-bridge.go"
	wherePkg := filepath.Join(where, injectPkgName)
	err := os.MkdirAll(wherePkg, os.ModePerm)
	if err != nil {
		log.Fatalf("fail to generate %v directory: %v", injectPkgName, err)
	}

	// create bridge file
	whereBridge := filepath.Join(where, injectBridgeName)
	f1, err := os.Create(whereBridge)
	if err != nil {
		log.Fatalf("fail to create cover bridge file in temporary project: %v", err)
	}
	defer f1.Close()

	tmplBridgeData := struct {
		CoverImportPath string
	}{
		// covers[0] is the main package
		CoverImportPath: covers[0].Package.ImportPath + "/" + injectPkgName,
	}

	if err := coverBridgeTmpl.Execute(f1, tmplBridgeData); err != nil {
		log.Fatalf("fail to generate cover bridge in temporary project: %v", err)
	}

	// create ws agent files
	dest := filepath.Join(wherePkg, "rpcagent.go")

	f2, err := os.Create(dest)
	if err != nil {
		log.Fatalf("fail to create cover agent file in temporary project: %v", err)
	}
	defer f2.Close()

	var _coverMode string
	if b.Mode == "watch" {
		_coverMode = "cover"
	} else {
		_coverMode = b.Mode
	}
	tmplData := struct {
		Covers                   []*PackageCover
		GlobalCoverVarImportPath string
		Package                  string
		Host                     string
		Mode                     string
	}{
		Covers:                   covers,
		GlobalCoverVarImportPath: b.GlobalCoverVarImportPath,
		Package:                  injectPkgName,
		Host:                     b.Host,
		Mode:                     _coverMode,
	}

	if err := coverMainTmpl.Execute(f2, tmplData); err != nil {
		log.Fatalf("fail to generate cover agent handlers in temporary project: %v", err)
	}

	// 写入 watch
	if b.Mode != "watch" {
		return
	}
	f, err := os.Create(filepath.Join(wherePkg, "watchagent.go"))
	if err != nil {
		log.Fatalf("fail to create watchagent file: %v", err)
	}

	tmplwatchData := struct {
		Random                   string
		Host                     string
		GlobalCoverVarImportPath string
	}{
		Random:                   filepath.Base(b.TmpModProjectDir),
		Host:                     b.Host,
		GlobalCoverVarImportPath: b.GlobalCoverVarImportPath,
	}

	if err := coverWatchTmpl.Execute(f, tmplwatchData); err != nil {
		log.Fatalf("fail to generate watchagent in temporary project: %v", err)
	}
}

// injectGlobalCoverVarFile 写入所有插桩变量的全局定义至一个单独的文件
func (b *Build) injectGlobalCoverVarFile(decl string) {
	globalCoverVarPackage := path.Base(b.GlobalCoverVarImportPath)
	globalCoverDef := filepath.Join(b.TmpModProjectDir, globalCoverVarPackage)
	b.GlobalCoverVarImportPathDir = globalCoverDef

	err := os.MkdirAll(globalCoverDef, os.ModePerm)
	if err != nil {
		log.Fatalf("fail to create global cover definition package dir: %v", err)
	}
	coverFile, err := os.Create(filepath.Join(globalCoverDef, "cover.go"))
	if err != nil {
		log.Fatalf("fail to create global cover definition file: %v", err)
	}

	defer coverFile.Close()

	packageName := "package coverdef\n\n"

	random := filepath.Base(b.TmpModProjectDir)
	varWatchDef := fmt.Sprintf(`
var WatchChannel_%v = make(chan *blockInfo, 1024)

var WatchEnabled_%v = false

type blockInfo struct {
	Name  string
	Pos   []uint32
	I     int
	Stmts int
}

// UploadCoverChangeEvent_%v is non-blocking
func UploadCoverChangeEvent_%v(name string, pos []uint32, i int, stmts uint16) {

	if WatchEnabled_%v == false {
		return
	}

	// make sure send is non-blocking
	select {
	case WatchChannel_%v <- &blockInfo{
		Name:  name,
		Pos:   pos,
		I:     i,
		Stmts: int(stmts),
	}:
	default:
	}
}

`, random, random, random, random, random, random)

	_, err = coverFile.WriteString(packageName + varWatchDef + decl)
	if err != nil {
		log.Fatalf("fail to write to global cover definition file: %v", err)
	}
}
