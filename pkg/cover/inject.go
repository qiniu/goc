package cover

import (
	"os"
	"path"
	"path/filepath"

	"github.com/qiniu/goc/v2/pkg/config"
	"github.com/qiniu/goc/v2/pkg/cover/internal/tool"
	"github.com/qiniu/goc/v2/pkg/log"
)

// Inject injects cover variables for all the .go files in the target directory
func Inject() {
	log.StartWait("injecting cover variables")

	var seen = make(map[string]*config.PackageCover)

	// 所有插桩变量定义声明
	allDecl := ""

	pkgs := config.GocConfig.Pkgs
	for _, pkg := range pkgs {
		if pkg.Name == "main" {
			log.Infof("handle main package: %v", pkg.ImportPath)
			// 该 main 二进制所关联的所有插桩变量的元信息
			// 每个 main 之间是不相关的，需要重新定义
			allMainCovers := make([]*config.PackageCover, 0)
			// 注入 main package
			mainCover, mainDecl := addCounters(pkg)
			// 收集插桩变量的定义和元信息
			allDecl += mainDecl
			allMainCovers = append(allMainCovers, mainCover)

			// 向 main package 的依赖注入插桩变量
			for _, dep := range pkg.Deps {
				if _, ok := seen[dep]; ok {
					continue
				}

				// 依赖需要忽略 Go 标准库和 go.mod 引入的第三方
				if depPkg, ok := pkgs[dep]; ok {
					// 注入依赖的 package
					packageCover, depDecl := addCounters(depPkg)
					// 收集插桩变量的定义和元信息
					allDecl += depDecl
					allMainCovers = append(allMainCovers, packageCover)
					// 避免重复访问
					seen[dep] = packageCover
				}
			}
			// 为每个 main 包注入 websocket handler
			injectCoverHandler(getPkgTmpDir(pkg.Dir), allMainCovers)
		}
	}
	// 在工程根目录注入所有插桩变量的声明+定义
	injectGlobalCoverVarFile(allDecl)

	log.StopWait()
	log.Donef("cover variables injected")
}

// addCounters is different from official go tool cover
//
// 1. only inject covervar++ into source file
//
// 2. no declarartions for these covervars
//
// 3. return the declarations as string
func addCounters(pkg *config.Package) (*config.PackageCover, string) {
	mode := config.GocConfig.Mode
	gobalCoverVarImportPath := config.GocConfig.GlobalCoverVarImportPath

	coverVarMap := declareCoverVars(pkg)

	decl := ""
	for file, coverVar := range coverVarMap {
		decl += "\n" + tool.Annotate(filepath.Join(getPkgTmpDir(pkg.Dir), file), mode, coverVar.Var, gobalCoverVarImportPath) + "\n"
	}

	return &config.PackageCover{
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
func getPkgTmpDir(pkgDir string) string {
	relDir, err := filepath.Rel(config.GocConfig.CurModProjectDir, pkgDir)
	if err != nil {
		log.Fatalf("go json -list meta info wrong: %v", err)
	}

	return filepath.Join(config.GocConfig.TmpModProjectDir, relDir)
}

// injectCoverHandler inject handlers like following
//
// - xxx.go
// - yyy_package
// - main.go
// - goc-http-cover-apis-auto-generated-11111-22222-bridge.go
// - goc-http-cover-apis-auto-generated-11111-22222-package
//  |
//  -- init.go
//
// 11111_22222_bridge.go just import 11111_22222_package, where package contains ws handlers.
// 使用 bridge.go 文件是为了避免插桩逻辑中的变量名污染 main 包
func injectCoverHandler(where string, covers []*config.PackageCover) {
	injectPkgName := "goc-http-cover-apis-auto-generated-11111-22222-package"
	wherePkg := filepath.Join(where, injectPkgName)
	err := os.MkdirAll(wherePkg, os.ModePerm)
	if err != nil {
		log.Fatalf("fail to generate %v directory: %v", injectPkgName, err)
	}

	// create bridge file
	whereBridge := filepath.Join(where, "goc-http-cover-apis-auto-generated-11111-22222-bridge.go")
	f, err := os.Create(whereBridge)
	if err != nil {
		log.Fatalf("fail to create cover bridge file in temporary project: %v", err)
	}

	tmplBridgeData := struct {
		CoverImportPath string
	}{
		// covers[0] is the main package
		CoverImportPath: covers[0].Package.ImportPath + "/" + injectPkgName,
	}

	if err := coverBridgeTmpl.Execute(f, tmplBridgeData); err != nil {
		log.Fatalf("fail to generate cover bridge in temporary project: %v", err)
	}

	// create ws handler files
	dest := filepath.Join(wherePkg, "init.go")

	f, err = os.Create(dest)
	if err != nil {
		log.Fatalf("fail to create cover handlers in temporary project: %v", err)
	}
	defer f.Close()

	tmplData := struct {
		Covers                   []*config.PackageCover
		GlobalCoverVarImportPath string
		Package                  string
	}{
		Covers:                   covers,
		GlobalCoverVarImportPath: config.GocConfig.GlobalCoverVarImportPath,
		Package:                  injectPkgName,
	}

	if err := coverMainTmpl.Execute(f, tmplData); err != nil {
		log.Fatalf("fail to generate cover handlers in temporary project: %v", err)
	}
}

// injectGlobalCoverVarFile 写入所有插桩变量的全局定义至单独一个文件
func injectGlobalCoverVarFile(decl string) {
	globalCoverVarPackage := path.Base(config.GocConfig.GlobalCoverVarImportPath)
	globalCoverDef := filepath.Join(config.GocConfig.TmpModProjectDir, globalCoverVarPackage)
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

	_, err = coverFile.WriteString(packageName + decl)
	if err != nil {
		log.Fatalf("fail to write to global cover definition file: %v", err)
	}
}
