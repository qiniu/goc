package e2e

import (
	"time"

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
			_, err = RunShortRunCmd([]string{"goc", "build", "."}, dir, nil)
			Expect(err).To(BeNil(), "goc build 运行错误")

			By("检查代码是否被插入二进制")
			contain, err := SearchSymbolInBinary(dir, "basic", "basic/goc-cover-agent-apis-auto-generated-11111-22222-package.loadFileCover")
			Expect(err).To(BeNil(), "二进制检查失败")
			Expect(contain).To(BeTrue(), "二进制中未找到插桩的符号")
		})
	})

	Describe("2 测试 server 命令", func() {
		It("1.2.1 测试编译/list/profile基础场景", func() {
			dir, err := mgr.GetSampleByKey("basic2")
			Expect(err).To(BeNil(), "找不到 sample")

			By("启动 goc server")
			lc := NewLongRunCmd([]string{"goc", "server", "."}, dir, nil)
			lc.Run()
			defer lc.Stop()

			By("编译一个长时间运行的程序")
			output, err := RunShortRunCmd([]string{"goc", "build", "."}, dir, nil)
			Expect(err).To(BeNil(), "编译失败："+output)

			By("长时间运行 basic2")
			basicC := NewLongRunCmd([]string{"./basic2"}, dir, nil)
			basicC.Run()
			defer basicC.Stop()

			By("使用 goc list 获取服务列表")
			output, err = RunShortRunCmd([]string{"goc", "list"}, dir, nil)
			Expect(err).To(BeNil(), "goc list 运行错误")
			Expect(output).To(ContainSubstring("127.0.0.1   ./basic2"), "goc list 输出应该包含 basic 服务")

			By("使用 goc profile 获取覆盖率")
			profileStr := `mode: count
basic2/main.go:8.13,9.6 1 1
basic2/main.go:9.6,12.3 2 2`
			time.Sleep(time.Second)
			output, err = RunShortRunCmd([]string{"goc", "profile"}, dir, nil)
			Expect(err).To(BeNil(), "goc profile 运行错误")
			Expect(output).To(ContainSubstring(profileStr), "goc profile 获取的覆盖率有误")
		})
	})
})
