package e2e

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("1 [基础测试]", func() {
	var (
		mgr *SamplesMgr
	)

	BeforeEach(func() {
		mgr = NewSamplesMgr()
	})

	Describe("1 测试 build 命令", func() {
		It("1.1.1 简单工程", func() {
			dir, err := mgr.GetSampleByKey("basic")
			Expect(err).To(BeNil(), "找不到 sample")

			By("使用 goc build 命令编译")
			_, err = RunShortRunCmd(dir, []string{"goc", "build", "."})
			Expect(err).To(BeNil(), "goc build 运行错误")
		})
	})

	Describe("2 测试 server 命令", func() {
		It("1.2.1 测试 API 接口", func() {
			dir, err := mgr.GetSampleByKey("basic")
			Expect(err).To(BeNil(), "找不到 sample")

			By("启动 goc server")
			lc := NewLongRunCmd(dir, []string{"goc", "server", "."})
			lc.Run()
			defer lc.Stop()

			By("使用 goc list 获取服务列表")
			output, err := RunShortRunCmd(dir, []string{"goc", "list"})
			Expect(err).To(BeNil(), "goc list 运行错误")
			Expect(output).To(ContainSubstring("REMOTEIP"), "goc list 输出应该包含 REMOTEIP")
		})
	})
})
