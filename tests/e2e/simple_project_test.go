package e2e_test

import (
	"os"
	"os/exec"
	"strings"

	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/qiniu/goc/pkg/build"
)

var TESTS_ROOT string

var _ = BeforeSuite(func() {
	TESTS_ROOT = os.Getenv("TESTS_ROOT")
	if TESTS_ROOT == "" {
		panic("Please set TESTS_ROOT env, or run in GitHub Actions!")
	}
})

var _ = Describe("E2E", func() {

	Context("Go module", func() {
		It("Simple project", func() {
			By("goc build")
			testProjDir := filepath.Join(TESTS_ROOT, "samples/simple_project")
			cmd := exec.Command("goc", "build")
			cmd.Dir = testProjDir

			out, err := cmd.CombinedOutput()
			Expect(err).To(BeNil(), "goc build on this project should be successful", string(out))

			By("check files in generated temporary directory")
			tempDir := filepath.Join(os.TempDir(), build.TmpFolderName(testProjDir))
			_, err = os.Lstat(tempDir)
			Expect(err).To(BeNil(), "projects should be copied to temporary directory")

			By("check if cover variables are injected")
			_, err = os.Lstat(filepath.Join(tempDir, "http_cover_apis_auto_generated.go"))
			Expect(err).To(BeNil(), "a http server file should be generated")

			By("check generated binary")
			cmd = exec.Command("go", "tool", "objdump", "simple-project")
			cmd.Dir = testProjDir
			out, err = cmd.CombinedOutput()
			Expect(err).To(BeNil(), "the binary cannot be disassembled")

			cnt := strings.Count(string(out), "GoCover")
			Expect(cnt).To(BeNumerically(">", 0), "GoCover varibale should be in the binary")

			cnt = strings.Count(string(out), "main.registerSelf")
			Expect(cnt).To(BeNumerically(">", 0), "main.registerSelf function should be in the binary")

		})
	})
})
