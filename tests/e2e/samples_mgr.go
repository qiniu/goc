package e2e

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	. "github.com/onsi/ginkgo"

	"github.com/tongjingran/copy"
	"gopkg.in/yaml.v3"
)

type Sample struct {
	Dir         string `yaml:"dir"`
	Description string `yaml:"description"`
}

// SamplesMgr create and return sample for test case
//
// ginkgo 的运行模型是多进程模型，每一个 test case 是一个独立的进程
// Ginkgo has support for running specs in parallel. It does this by spawning separate go test processes and serving specs to each process off of a shared queue.
// 所以这里会单独在一个临时目录中生成 sample，以便将来测试用例可以并发执行
type SamplesMgr struct {
	Samples map[string]Sample `yaml:"samples"`
	path    string            `yaml:"-"`
}

func NewSamplesMgr() *SamplesMgr {
	path, _ := os.Getwd()
	metaData, err := os.ReadFile(filepath.Join(path, "samples", "meta.yaml"))
	if err != nil {
		log.Fatalf("fail to read sample meta")
	}

	mgr := SamplesMgr{}
	err = yaml.Unmarshal(metaData, &mgr)
	if err != nil {
		log.Fatalf("fail to parse the meta yaml")
	}
	mgr.path = path
	return &mgr
}

// GetSampleByKey return the sample folder location for test use
func (m *SamplesMgr) GetSampleByKey(key string) (string, error) {
	sample, ok := m.Samples[key]
	if !ok {
		return "", fmt.Errorf("no sample found")
	}

	desc := CurrentGinkgoTestDescription()
	caseTitle := desc.FullTestText + " " + key

	m1 := regexp.MustCompile(`[/\\?%*:|"<>]`)
	caseTitle = m1.ReplaceAllString(caseTitle, "-")

	dst := filepath.Join(m.path, "tmp", caseTitle)
	err := os.RemoveAll(dst)
	if err != nil {
		log.Fatalf("fail to clean the temp sample: %v", err)
	}
	err = os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		log.Fatalf("fail to create sample dir: %v", err)
	}

	src := filepath.Join(m.path, "samples", sample.Dir)

	err = copy.Copy(src, dst)
	if err != nil {
		log.Fatalf("fail to copy the sample project: %v", err)
	}

	return dst, nil
}
