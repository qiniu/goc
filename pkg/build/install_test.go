package build

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestBasicInstallForModProject(t *testing.T) {
	workingDir := filepath.Join(baseDir, "../tests/samples/simple_project")
	gopath := filepath.Join(baseDir, "../tests/samples/simple_project", "testhome")

	os.Chdir(workingDir)
	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")

	buildFlags, packages := "", "."
	gocBuild, err := NewInstall(buildFlags, packages)
	assert.Equal(t, err, nil, "should create temporary directory successfully")

	err = gocBuild.Install()
	assert.Equal(t, err, nil, "temporary directory should build successfully")
}
