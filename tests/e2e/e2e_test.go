/*
 Copyright 2021 Qiniu Cloud (qiniu.com)
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
		It("1.2.1 测试编译/service/profile基础场景", func() {
			dir, err := mgr.GetSampleByKey("basic2")
			Expect(err).To(BeNil(), "找不到 sample")

			By("启动 goc server")
			lc := NewLongRunCmd([]string{"goc", "server"}, dir, nil)
			lc.Run()
			defer lc.Stop()

			By("编译 basic2")
			output, err := RunShortRunCmd([]string{"goc", "build", "."}, dir, nil)
			Expect(err).To(BeNil(), "编译失败："+output)

			By("运行 basic2")
			basicC := NewLongRunCmd([]string{"./basic2"}, dir, nil)
			basicC.Run()
			defer basicC.Stop()

			By("使用 goc service get 获取服务列表")
			Eventually(func() {
				output, err = RunShortRunCmd([]string{"goc", "service", "get"}, dir, nil)
				Expect(err).To(BeNil(), "goc servive get 运行错误")
				Expect(output).To(ContainSubstring("127.0.0.1   ./basic2"), "goc service get 输出应该包含 basic 服务")
			}, 3*time.Second, 1*time.Second).Should(Succeed())

			By("使用 goc profile get 获取覆盖率")
			profileStr := `mode: count
basic2/main.go:8.13,9.6 1 1`
			time.Sleep(time.Second)
			output, err = RunShortRunCmd([]string{"goc", "profile", "get"}, dir, nil)
			Expect(err).To(BeNil(), "goc profile get运行错误")
			Expect(output).To(ContainSubstring(profileStr), "goc profile get 获取的覆盖率有误")
		})

		It("1.2.2 测试 server 重启", func() {
			dir, err := mgr.GetSampleByKey("basic2")
			Expect(err).To(BeNil(), "找不到 sample")

			By("启动 goc server")
			lc := NewLongRunCmd([]string{"goc", "server"}, dir, nil)
			lc.Run()
			defer lc.Stop()

			By("编译 basic2")
			output, err := RunShortRunCmd([]string{"goc", "build", "."}, dir, nil)
			Expect(err).To(BeNil(), "编译失败："+output)

			By("运行 basic2")
			basicC := NewLongRunCmd([]string{"./basic2"}, dir, nil)
			basicC.Run()
			defer basicC.Stop()

			By("获取 service 列表")
			Eventually(func() {
				output, err = RunShortRunCmd([]string{"goc", "service", "get"}, dir, nil)
				Expect(err).To(BeNil(), "goc servive get 运行错误")
				Expect(output).To(ContainSubstring("1    CONNECT   127.0.0.1   ./basic2"), "goc service get 输出应该包含 basic 服务")
			}, 3*time.Second, 1*time.Second).Should(Succeed())

			By("重启 goc server")
			lc.Stop()
			lc2 := NewLongRunCmd([]string{"goc", "server"}, dir, nil)
			lc2.Run()
			defer lc2.Stop()

			By("再次获取 service 列表")
			Eventually(func() {
				output, err = RunShortRunCmd([]string{"goc", "service", "get"}, dir, nil)
				Expect(err).To(BeNil(), "goc servive get 运行错误")
				Expect(output).To(ContainSubstring("1    CONNECT   127.0.0.1   ./basic2"), "goc service get 输出应该包含 basic 服务")
			}, 15*time.Second, 1*time.Second).Should(Succeed(), "10s 内 basic2 服务重新注册成功，且 id 不变")
		})
	})

	Describe("3 测试 goc service 相关命令", func() {
		It("1.3.1 get/delete 组合", func() {
			dir, err := mgr.GetSampleByKey("basic2")
			Expect(err).To(BeNil(), "找不到 sample")

			By("启动 goc server")
			lc := NewLongRunCmd([]string{"goc", "server"}, dir, nil)
			lc.Run()
			defer lc.Stop()

			By("编译 basic2")
			output, err := RunShortRunCmd([]string{"goc", "build", "."}, dir, nil)
			Expect(err).To(BeNil(), "编译失败："+output)

			By("长时间运行 basic2")
			basicC := NewLongRunCmd([]string{"./basic2"}, dir, nil)
			basicC.Run()
			defer basicC.Stop()

			By("再长时间运行 basic2 -f hello")
			time.Sleep(time.Second * 1)
			basicC2 := NewLongRunCmd([]string{"./basic2", "-f", "hello"}, dir,
				[]string{
					"GOC_REGISTER_EXTRA=fantastic", // 额外的注册信息
				})
			basicC2.Run()

			time.Sleep(time.Second)

			By("测试获取 service 信息")
			output, err = RunShortRunCmd([]string{"goc", "service", "get"}, dir, nil)
			Expect(err).To(BeNil(), "goc servive get 运行错误")
			Expect(output).To(ContainSubstring("2    CONNECT   127.0.0.1   ./basic2 -f"), "goc service get 应该显示第二个服务")

			By("测试获取指定 id 的 service 信息")
			output, err = RunShortRunCmd([]string{"goc", "service", "get", "--id", "1"}, dir, nil)
			Expect(err).To(BeNil(), "goc servive get 运行错误")
			Expect(output).To(ContainSubstring("1    CONNECT   127.0.0.1   ./basic2"), "id=1 的服务能返回")
			Expect(output).NotTo(ContainSubstring("2    CONNECT   127.0.0.1   ./basic2 -f"), "id=2 的服务没有返回")

			By("测试能否获取 extra")
			output, err = RunShortRunCmd([]string{"goc", "service", "get", "--wide"}, dir, nil)
			Expect(err).To(BeNil(), "goc servive get --wide 运行错误")
			Expect(output).To(ContainSubstring("fantastic"), "注入的 extra 信息没有获取到")

			By("basic2 -f hello 退出")
			basicC2.Stop()

			By("测试 get 能否获取 DISCONNECT 状态")
			Eventually(func() {
				output, err = RunShortRunCmd([]string{"goc", "service", "get"}, dir, nil)
				Expect(err).To(BeNil(), "goc servive get 运行错误")
				Expect(output).To(ContainSubstring("2    DISCONNECT   127.0.0.1   ./basic2 -f"), "应该在 10s 内感知到 agent 断连")
			}, 10*time.Second, 1*time.Second).Should(Succeed())

			By("测试删除 id=2(已经退出) 的服务列表")
			output, err = RunShortRunCmd([]string{"goc", "service", "delete", "--id", "2"}, dir, nil)
			Expect(err).To(BeNil(), "删除服务列表失败")

			output, err = RunShortRunCmd([]string{"goc", "service", "get"}, dir, nil)
			Expect(err).To(BeNil(), "goc servive get 运行错误")
			Expect(output).NotTo(ContainSubstring("./basic2 -f"), "DISCONNECT 的服务未被删除")

			By("测试删除 id=1(还在运行) 的服务列表")
			output, err = RunShortRunCmd([]string{"goc", "service", "delete", "--id", "1"}, dir, nil)
			Expect(err).To(BeNil(), "删除服务列表失败")

			Eventually(func() {
				output, err = RunShortRunCmd([]string{"goc", "service", "get"}, dir, nil)
				Expect(err).To(BeNil(), "goc servive get 运行错误")
				Expect(output).To(ContainSubstring("3    CONNECT"), "20s 内 basic2 重新注册成功")
				Expect(output).NotTo(ContainSubstring("1    "), "id=1 的服务不在")

			}, 20*time.Second, 1*time.Second).Should(Succeed())
		})
	})

	Describe("4 测试 goc profile 相关命令", func() {
		It("1.4.1 测试 get/clear", func() {
			dir, err := mgr.GetSampleByKey("basic3")
			Expect(err).To(BeNil(), "找不到 sample")

			By("启动 goc server")
			lc := NewLongRunCmd([]string{"goc", "server"}, dir, nil)
			lc.Run()
			defer lc.Stop()

			By("编译 basic3")
			output, err := RunShortRunCmd([]string{"goc", "build", "."}, dir, nil)
			Expect(err).To(BeNil(), "编译失败："+output)

			By("运行 basic3")
			basicC := NewLongRunCmd([]string{"./basic3"}, dir, nil)
			basicC.Run()
			defer basicC.Stop()

			By("运行 basic3 -f")
			basicC2 := NewLongRunCmd([]string{"./basic3", "-f"}, dir, nil)
			basicC2.Run()
			defer basicC2.Stop()

			time.Sleep(time.Second)

			By("获取覆盖率")
			Eventually(func() {
				output, err = RunShortRunCmd([]string{"goc", "profile", "get"}, dir, nil)
				Expect(err).To(BeNil(), "获取覆盖率失败")
				Expect(output).To(ContainSubstring("basic3/main.go:8.13,11.2 2 2"))
			}, 3*time.Second, 1*time.Second).Should(Succeed())

			By("只获取 id=1 的覆盖率")
			output, err = RunShortRunCmd([]string{"goc", "profile", "get", "--id", "1"}, dir, nil)
			Expect(err).To(BeNil(), "获取覆盖率失败")
			Expect(output).To(ContainSubstring("basic3/main.go:8.13,11.2 2 1"), "id=1 只有 1 个覆盖率")

			By("运行 basic3 -g")
			basicC3 := NewLongRunCmd([]string{"./basic3", "-g"}, dir, []string{
				"GOC_REGISTER_EXTRA=fantastic", // 额外的注册信息
			})
			basicC3.Run()
			defer basicC3.Stop()

			By("只获取 extra=fantastic 的覆盖率")
			Eventually(func() {
				output, err = RunShortRunCmd([]string{"goc", "profile", "get", "--extra", "fantastic"}, dir, nil)
				Expect(err).To(BeNil(), "获取覆盖率失败")
				Expect(output).To(ContainSubstring("basic3/main.go:8.13,11.2 2 1"), "extra=fantastic 的覆盖率只有 1")
			}, 3*time.Second, 1*time.Second).Should(Succeed())

			By("获取 id=10 的覆盖率")
			output, err = RunShortRunCmd([]string{"goc", "profile", "get", "--id", "10"}, dir, nil)
			Expect(err).NotTo(BeNil(), "获取覆盖率不应该成功")
			Expect(output).To(ContainSubstring("can't merge zero profiles"), "错误信息不对")

			By("清空 id=2 的覆盖率")
			output, err = RunShortRunCmd([]string{"goc", "profile", "get"}, dir, nil)
			Expect(err).To(BeNil(), "获取覆盖率失败")
			Expect(output).To(ContainSubstring("basic3/main.go:8.13,11.2 2 3"))

			output, err = RunShortRunCmd([]string{"goc", "profile", "clear", "--id", "2"}, dir, nil)
			Expect(err).To(BeNil(), "清空覆盖率失败")

			Eventually(func() {
				output, err = RunShortRunCmd([]string{"goc", "profile", "get"}, dir, nil)
				Expect(err).To(BeNil(), "获取覆盖率失败")
				Expect(output).To(ContainSubstring("basic3/main.go:8.13,11.2 2 2"))
			}, 3*time.Second, 1*time.Second).Should(Succeed())

			By("清空 extra=fantastic 的覆盖率")
			output, err = RunShortRunCmd([]string{"goc", "profile", "clear", "--extra", "fantastic"}, dir, nil)
			Expect(err).To(BeNil(), "清空覆盖率失败")

			Eventually(func() {
				output, err = RunShortRunCmd([]string{"goc", "profile", "get"}, dir, nil)
				Expect(err).To(BeNil(), "获取覆盖率失败")
				Expect(output).To(ContainSubstring("basic3/main.go:8.13,11.2 2 1"))
			}, 3*time.Second, 1*time.Second).Should(Succeed())

			By("运行 basic3 -h")
			basicC4 := NewLongRunCmd([]string{"./basic3", "-h"}, dir, []string{
				"GOC_REGISTER_EXTRA=fantastic", // 额外的注册信息
			})
			basicC4.Run()
			defer basicC4.Stop()

			Eventually(func() {
				output, err = RunShortRunCmd([]string{"goc", "profile", "get"}, dir, nil)
				Expect(err).To(BeNil(), "获取覆盖率失败")
				Expect(output).To(ContainSubstring("basic3/main.go:8.13,11.2 2 2"))
			}, 3*time.Second, 1*time.Second).Should(Succeed())

			By("清空所有的覆盖率")
			output, err = RunShortRunCmd([]string{"goc", "profile", "clear"}, dir, nil)
			Expect(err).To(BeNil(), "清空覆盖率失败")

			Eventually(func() {
				output, err = RunShortRunCmd([]string{"goc", "profile", "get"}, dir, nil)
				Expect(err).To(BeNil(), "获取覆盖率失败")
				Expect(output).To(ContainSubstring("basic3/main.go:8.13,11.2 2 0"))
			}, 3*time.Second, 1*time.Second).Should(Succeed())
		})
	})

	Describe("5 测试 install 命令", func() {
		It("1.5.1 简单工程", func() {
			dir, err := mgr.GetSampleByKey("basic")
			Expect(err).To(BeNil(), "找不到 sample")

			By("使用 goc install 命令编译")
			gobinEnv := "GOBIN=" + dir + "/bin"
			_, err = RunShortRunCmd([]string{"goc", "install", "."}, dir, []string{
				gobinEnv,
			})
			Expect(err).To(BeNil(), "goc install 运行错误")

			By("检查代码是否被插入二进制")
			contain, err := SearchSymbolInBinary(dir+"/bin", "basic", "basic/goc-cover-agent-apis-auto-generated-11111-22222-package.loadFileCover")
			Expect(err).To(BeNil(), "二进制检查失败")
			Expect(contain).To(BeTrue(), "二进制中未找到插桩的符号")
		})
	})
})
