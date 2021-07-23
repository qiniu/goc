package e2e

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	. "github.com/onsi/ginkgo"

	"github.com/gofrs/flock"
	"github.com/tongjingran/copy"
	"gopkg.in/yaml.v3"
)

type Sample struct {
	Dir         string `yaml:"dir"`
	Description string `yaml:"description"`
}

// SamplesMgr create and return sample for test case
//
// ginkgo 的并发执行时的运行模型是多进程模型，会有多个独立的 go test 进程。
// Ginkgo has support for running specs in parallel. It does this by spawning separate go test processes and serving specs to each process off of a shared queue.
//
// 所以这里设计成每个 test case 在各自的临时目录中生成 sample，以便将来测试用例可以并发执行。
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

// GetAvailablePort get next available host port among multiprocess ginkgo test cases
//
// 利用文件锁和端口探活，获取下一个可用的 port
func (m *SamplesMgr) GetAvailablePort() (string, error) {
	fileLockPath := filepath.Join(m.path, "tmp", "port.lock")

	// 文件锁，ginkgo parallel 模式是多进程，必须用跨平台跨进程的同步方式
	lock := flock.New(fileLockPath)
	lock.Lock()
	defer lock.Unlock()

	// 该文件记录 counter，代表下一个可用的 port
	portFilePath := filepath.Join(m.path, "tmp", "port.record")
	data, err := os.ReadFile(portFilePath)
	if err != nil {
		os.Create(portFilePath)
	}
	port, err := strconv.Atoi(string(data))
	if err != nil {
		port = 7777
	} else {
		port += 1
	}

	// 循环检测直到找到可用 port
	for {
		if port == 65534 {
			port = 7777
		}
		conn, err := net.Dial("tcp", ":"+strconv.Itoa(port))
		if err == nil {
			port += 1
			conn.Close()
		} else {
			break
		}
	}

	err = os.WriteFile(portFilePath, []byte(strconv.Itoa(port)), os.ModePerm)

	return "127.0.0.1:" + strconv.Itoa(port), err
}
