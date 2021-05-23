package cover

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os/exec"

	"github.com/qiniu/goc/v2/pkg/config"
	"github.com/qiniu/goc/v2/pkg/log"
)

var (
	// ErrCoverPkgFailed represents the error that fails to inject the package
	ErrCoverPkgFailed = errors.New("fail to inject code to project")
	// ErrCoverListFailed represents the error that fails to list package dependencies
	ErrCoverListFailed = errors.New("fail to list package dependencies")
)

// ListPackages list all packages under specific via go list command.
func ListPackages(dir string) map[string]*config.Package {
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
