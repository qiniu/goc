package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicInstallForModProject(t *testing.T) {
	workingDir := filepath.Join(baseDir, "../../tests/samples/simple_project")
	gopath := filepath.Join(baseDir, "../../tests/samples/simple_project", "testhome")

	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")

	buildFlags, packages := "", []string{"."}
	gocBuild, err := NewInstall(buildFlags, packages, workingDir)
	if !assert.Equal(t, err, nil) {
		assert.FailNow(t, "should create temporary directory successfully")
	}

	err = gocBuild.Install()
	if !assert.Equal(t, err, nil) {
		assert.FailNow(t, "temporary directory should build successfully")
	}
}

func TestInvalidPackageNameForInstall(t *testing.T) {
	workingDir := filepath.Join(baseDir, "../../tests/samples/simple_project")
	gopath := filepath.Join(baseDir, "../../tests/samples/simple_project", "testhome")

	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")

	buildFlags, packages := "", []string{"main.go"}
	_, err := NewInstall(buildFlags, packages, workingDir)
	if !assert.Equal(t, err, ErrWrongPackageTypeForInstall) {
		assert.FailNow(t, "should not success with non . or ./... package")
	}
}
