package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicInstallForModProject(t *testing.T) {
	workingDir := filepath.Join(baseDir, "../tests/samples/simple_project")
	gopath := filepath.Join(baseDir, "../tests/samples/simple_project", "testhome")

	os.Setenv("GOPATH", gopath)
	os.Setenv("GO111MODULE", "on")

	buildFlags, packages := "", []string{"."}
	gocBuild, err := NewInstall(buildFlags, packages, workingDir)
	assert.Equal(t, err, nil, "should create temporary directory successfully")

	err = gocBuild.Install()
	assert.Equal(t, err, nil, "temporary directory should build successfully")
}
