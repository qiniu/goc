package build

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/qiniu/goc/v2/pkg/config"
	"github.com/qiniu/goc/v2/pkg/log"
)

// readProjectMetaInfo reads all meta informations of the corresponding project
func (b *Build) readProjectMetaInfo() {
	// get gopath & gobin
	config.GocConfig.GOPATH = b.readGOPATH()
	config.GocConfig.GOBIN = b.readGOBIN()
	// 获取 [packages] 及其依赖的 package list
	pkgs := b.listPackages(config.GocConfig.CurWd)

	// get mod info
	for _, pkg := range pkgs {
		// check if go modules is enabled
		if pkg.Module == nil {
			log.Fatalf("Go module is not enabled, please set GO111MODULE=auto or on")
		}
		// 工程根目录
		config.GocConfig.CurModProjectDir = pkg.Root
		config.GocConfig.ImportPath = pkg.Module.Path

		break
	}

	// 如果当前目录不是工程根目录，那再次 go list 一次，获取整个工程的包信息
	if config.GocConfig.CurWd != config.GocConfig.CurModProjectDir {
		config.GocConfig.Pkgs = b.listPackages(config.GocConfig.CurModProjectDir)
	} else {
		config.GocConfig.Pkgs = pkgs
	}

	// get tmp folder name
	config.GocConfig.TmpModProjectDir = filepath.Join(os.TempDir(), tmpFolderName(config.GocConfig.CurModProjectDir))
	// get working dir in the corresponding tmp dir
	config.GocConfig.TmpWd = filepath.Join(config.GocConfig.TmpModProjectDir, config.GocConfig.CurWd[len(config.GocConfig.CurModProjectDir):])
	// get GlobalCoverVarImportPath
	config.GocConfig.GlobalCoverVarImportPath = path.Join(config.GocConfig.ImportPath, tmpFolderName(config.GocConfig.CurModProjectDir))
	log.Donef("project meta information parsed")
}

// displayProjectMetaInfo prints basic infomation of this project to stdout
func (b *Build) displayProjectMetaInfo() {
	log.Infof("GOPATH: %v", config.GocConfig.GOPATH)
	log.Infof("GOBIN: %v", config.GocConfig.GOBIN)
	log.Infof("Project Directory: %v", config.GocConfig.CurModProjectDir)
	log.Infof("Temporary Project Directory: %v", config.GocConfig.TmpModProjectDir)
	log.Infof("")
}

// readGOPATH reads GOPATH use go env GOPATH command
func (b *Build) readGOPATH() string {
	out, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		log.Fatalf("fail to read GOPATH: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// readGOBIN reads GOBIN use go env GOBIN command
func (b *Build) readGOBIN() string {
	out, err := exec.Command("go", "env", "GOBIN").Output()
	if err != nil {
		log.Fatalf("fail to read GOBIN: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// listPackages list all packages under specific via go list command.
func (b *Build) listPackages(dir string) map[string]*config.Package {
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = dir

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("execute go list -json failed, err: %v, stdout: %v, stderr: %v", err, string(out), errBuf.String())
	}
	// 有些时候 go 命令会打印一些信息到 stderr，但其实命令整体是成功运行了
	if errBuf.String() != "" {
		log.Errorf("%v", errBuf.String())
	}

	dec := json.NewDecoder(bytes.NewBuffer(out))
	pkgs := make(map[string]*config.Package, 0)

	for {
		var pkg config.Package
		if err := dec.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("reading go list output error: %v", err)
		}
		if pkg.Error != nil {
			log.Fatalf("list package %s failed with output: %v", pkg.ImportPath, pkg.Error)
		}

		pkgs[pkg.ImportPath] = &pkg
	}

	return pkgs
}
