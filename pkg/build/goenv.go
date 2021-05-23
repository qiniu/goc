package build

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/qiniu/goc/v2/pkg/config"
	"github.com/qiniu/goc/v2/pkg/cover"
	"github.com/qiniu/goc/v2/pkg/log"
)

// readProjectMetaInfo reads all meta informations of the corresponding project
func (b *Build) readProjectMetaInfo() {
	// get gopath & gobin
	config.GocConfig.GOPATH = b.readGOPATH()
	config.GocConfig.GOBIN = b.readGOBIN()
	// 获取当前目录及其依赖的 package list
	config.GocConfig.Pkgs = cover.ListPackages(config.GocConfig.CurPkgDir)

	// get mod info
	pkgs := config.GocConfig.Pkgs
	for _, pkg := range pkgs {
		// check if go modules is enabled
		if pkg.Module == nil {
			log.Fatalf("Go module is not enabled, please set GO111MODULE=auto or on")
		}
		// 工程根目录
		config.GocConfig.CurModProjectDir = pkg.Root

		break
	}

	// get tmp folder name
	config.GocConfig.TmpModProjectDir = filepath.Join(os.TempDir(), tmpFolderName(config.GocConfig.CurModProjectDir))
	// get cur pkg dir in the corresponding tmp dir
	config.GocConfig.TmpPkgDir = filepath.Join(config.GocConfig.TmpModProjectDir, config.GocConfig.CurPkgDir[len(config.GocConfig.CurModProjectDir):])
	log.Donef("project meta information parsed")
}

// displayProjectMetaInfo prints basic infomation of this project to stdout
func (b *Build) displayProjectMetaInfo() {
	log.Infof("Project Infomation")
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
