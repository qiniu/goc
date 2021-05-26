package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoverSuccess(t *testing.T) {
	workingDir := filepath.Join(baseDir, "../tests/samples/simple_project")
	os.Setenv("GO111MODULE", "on")

	runCover(workingDir)

	_, err := os.Lstat(workingDir + "/http_cover_apis_auto_generated.go")
	assert.Equal(t, err, nil, "the generate file should be generated.")
}
