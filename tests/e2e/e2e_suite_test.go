package e2e

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGoc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Goc e2e Test Suite")
}
