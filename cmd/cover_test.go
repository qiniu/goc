package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mainContent = []byte(`package main

import (
	"fmt"
)

func main() {
	fmt.Println("hello world")
}`)

var goModContent = []byte(`module example.com/simple-project

go 1.11
`)

func TestCoverSuccess(t *testing.T) {
	workingDir := filepath.Join(baseDir, "../tests/samples/tmp/project")
	err := os.MkdirAll(workingDir, os.ModePerm)
	assert.NoError(t, err)
	defer os.RemoveAll(workingDir)

	os.WriteFile(workingDir+"/main.go", mainContent, 0644)
	os.WriteFile(workingDir+"/go.mod", goModContent, 0644)
	os.Setenv("GO111MODULE", "on")

	runCover(workingDir)

	_, err = os.Lstat(workingDir + "/http_cover_apis_auto_generated.go")
	assert.Equal(t, err, nil, "the generate file should be generated.")
}
