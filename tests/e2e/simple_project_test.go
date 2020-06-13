/*
 Copyright 2020 Qiniu Cloud (qiniu.com)

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/qiniu/goc/pkg/build"
)

var TESTS_ROOT string

var _ = BeforeSuite(func() {
	TESTS_ROOT, _ = os.Getwd()
	By("Current working directory: " + TESTS_ROOT)
	TESTS_ROOT = filepath.Join(TESTS_ROOT, "..")
})

var _ = Describe("E2E", func() {
	var GOPATH string

	BeforeEach(func() {
		GOPATH = os.Getenv("GOPATH")
		// in GitHub Action, this value is empty
		if GOPATH == "" {
			GOPATH = filepath.Join(os.Getenv("HOME"), "go")
		}
	})

	Context("Go module", func() {
		It("Simple project", func() {
			startTime := time.Now()

			By("goc build")
			testProjDir := filepath.Join(TESTS_ROOT, "samples/simple_project")
			cmd := exec.Command("goc", "build")
			cmd.Dir = testProjDir

			out, err := cmd.CombinedOutput()
			Expect(err).To(BeNil(), "goc build on this project should be successful", string(out))

			By("goc install")
			testProjDir = filepath.Join(TESTS_ROOT, "samples/simple_project")
			cmd = exec.Command("goc", "install", "--debuggoc")
			cmd.Dir = testProjDir

			out, err = cmd.CombinedOutput()
			Expect(err).To(BeNil(), "goc install on this project should be successful", string(out))

			By("check files in generated temporary directory")
			tempDir := filepath.Join(os.TempDir(), build.TmpFolderName(testProjDir))
			_, err = os.Lstat(tempDir)
			Expect(err).To(BeNil(), "projects should be copied to temporary directory")

			By("check if cover variables are injected")
			_, err = os.Lstat(filepath.Join(tempDir, "http_cover_apis_auto_generated.go"))
			Expect(err).To(BeNil(), "a http server file should be generated")

			By("check generated binary")
			objects := []string{GOPATH + "/bin", testProjDir}
			for _, dir := range objects {
				obj := filepath.Join(dir, "simple-project")
				fInfo, err := os.Lstat(obj)
				Expect(err).To(BeNil())
				Expect(startTime.Before(fInfo.ModTime())).To(Equal(true), "new binary should be generated, not the old one")

				cmd := exec.Command("go", "tool", "objdump", "simple-project")
				cmd.Dir = dir
				out, err = cmd.CombinedOutput()
				Expect(err).To(BeNil(), "the binary cannot be disassembled")

				cnt := strings.Count(string(out), "GoCover")
				Expect(cnt).To(BeNumerically(">", 0), "GoCover varibale should be in the binary")

				cnt = strings.Count(string(out), "main.registerSelf")
				Expect(cnt).To(BeNumerically(">", 0), "main.registerSelf function should be in the binary")
			}
		})
	})

	Context("GOPATH", func() {
		var GOPATH string

		BeforeEach(func() {
			GOPATH = os.Getenv("GOPATH")
		})

		It("Simple GOPATH project", func() {
			startTime := time.Now()
			testProjDir := filepath.Join(TESTS_ROOT, "samples/simple_gopath_project")
			oriWorkingDir := filepath.Join(testProjDir, "src/qiniu.com/simple_gopath_project")
			GOPATH = testProjDir

			By("goc build")
			cmd := exec.Command("goc", "build")
			cmd.Dir = oriWorkingDir
			// use GOPATH mode to compile project
			cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%v", GOPATH), "GO111MODULE=off")

			out, err := cmd.CombinedOutput()
			Expect(err).To(BeNil(), "goc build on this project should be successful", string(out), cmd.Dir)

			By("goc install")
			testProjDir = filepath.Join(TESTS_ROOT, "samples/simple_gopath_project")
			cmd = exec.Command("goc", "install", "--debuggoc")
			cmd.Dir = filepath.Join(testProjDir, "src/qiniu.com/simple_gopath_project")
			// use GOPATH mode to compile project
			cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%v", testProjDir), "GO111MODULE=off")

			out, err = cmd.CombinedOutput()
			Expect(err).To(BeNil(), "goc install on this project should be successful", string(out))

			By("check files in generated temporary directory")
			tempDir := filepath.Join(os.TempDir(), build.TmpFolderName(oriWorkingDir))
			_, err = os.Lstat(tempDir)
			Expect(err).To(BeNil(), "projects should be copied to temporary directory")

			By("check if cover variables are injected")
			newWorkingDir := filepath.Join(tempDir, "src/qiniu.com/simple_gopath_project")
			_, err = os.Lstat(filepath.Join(newWorkingDir, "http_cover_apis_auto_generated.go"))
			Expect(err).To(BeNil(), "a http server file should be generated")

			By("check generated binary")
			objects := []string{GOPATH + "/bin", oriWorkingDir}
			for _, dir := range objects {
				obj := filepath.Join(dir, "simple-gopath-project")
				fInfo, err := os.Lstat(obj)
				Expect(err).To(BeNil())
				Expect(startTime.Before(fInfo.ModTime())).To(Equal(true), "new binary should be generated, not the old one")

				cmd := exec.Command("go", "tool", "objdump", "simple_gopath_project")
				cmd.Dir = dir
				out, err = cmd.CombinedOutput()
				Expect(err).To(BeNil(), "the binary cannot be disassembled")

				cnt := strings.Count(string(out), "GoCover")
				Expect(cnt).To(BeNumerically(">", 0), "GoCover varibale should be in the binary")

				cnt = strings.Count(string(out), "main.registerSelf")
				Expect(cnt).To(BeNumerically(">", 0), "main.registerSelf function should be in the binary")
			}

		})
	})
})
