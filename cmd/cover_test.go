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

	err = writeFile(workingDir+"/main.go", mainContent)
	assert.NoError(t, err)
	err = writeFile(workingDir+"/go.mod", goModContent)
	assert.NoError(t, err)
	os.Setenv("GO111MODULE", "on")

	runCover(workingDir)

	_, err = os.Lstat(workingDir + "/http_cover_apis_auto_generated.go")
	assert.Equal(t, err, nil, "the generate file should be generated.")
}

func writeFile(name string, data []byte) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}
