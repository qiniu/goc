# goc

[![Go Report Card](https://goreportcard.com/badge/github.com/qiniu/goc)](https://goreportcard.com/report/github.com/qiniu/goc)
![](https://github.com/qiniu/goc/workflows/ut-check/badge.svg)
![](https://github.com/qiniu/goc/workflows/style-check/badge.svg)
![](https://github.com/qiniu/goc/workflows/e2e%20test/badge.svg)
![Build Release](https://github.com/qiniu/goc/workflows/Build%20Release/badge.svg)
[![codecov](https://codecov.io/gh/qiniu/goc/branch/master/graph/badge.svg)](https://codecov.io/gh/qiniu/goc)
[![GoDoc](https://godoc.org/github.com/qiniu/goc?status.svg)](https://godoc.org/github.com/qiniu/goc)

goc 是专为 Go 语言打造的一个综合覆盖率收集系统，尤其适合复杂的测试场景，比如系统测试时的代码覆盖率收集以及精准测试。

希望你们喜欢～

![Demo](docs/images/intro.gif)

## 安装

最新版本在该页面下载 [Github Releases](https://github.com/qiniu/goc/releases)。

goc 同时支持 `GOPATH` 工程和 `Go Modules` 工程，且 Go 版本要求 **Go 1.11+**。如果想参与 goc 的开发，你必须使用 **Go 1.13+**。

## 例子

goc 有多种使用场景。

### 在系统测试中收集代码覆盖率

goc 可以实时收集长时运行的 golang 服务覆盖率。收集步骤只需要下面三步：

1. 运行 `goc server` 命令启动一个服务注册中心：
    ```
    ➜  simple-go-server git:(master) ✗ goc server
    ```
2. 运行 `goc build` 命令编译目标服务，然后启动插过桩的二进制。下面以 [simple-go-server](https://github.com/CarlJi/simple-go-server) 工程为例：
    ```
    ➜  simple-go-server git:(master) ✗ goc build .
    ... // omit logs
    ➜  simple-go-server git:(master) ✗ ./simple-go-server  
    ```
3. 运行 `goc profile` 命令收集刚启动的 simple server 的代码覆盖率：
    ```
    ➜  simple-go-server git:(master) ✗ goc profile
    mode: atomic
    enricofoltran/simple-go-server/main.go:30.13,48.33 13 1
    enricofoltran/simple-go-server/main.go:48.33,50.3 1 0
    enricofoltran/simple-go-server/main.go:52.2,65.12 5 1
    enricofoltran/simple-go-server/main.go:65.12,74.46 7 1
    enricofoltran/simple-go-server/main.go:74.46,76.4 1 0
    ...   
    ```

### Vscode 中实时展示覆盖率动态变化

我们提供了一个 vscode 插件 - [Goc Coverage](https://marketplace.visualstudio.com/items?itemName=lyyyuna.goc)。该插件可以在运行时高亮覆盖过的代码。

![Extension](docs/images/goc-vscode.gif)

## Tips

1. goc 命令加上 `--debug` 会打印详细的日志。我们建议在提交 bug 时附上详细日志。

2. 默认情况下，插桩过的服务会监听在一个随机的端口，注册中心会通过这个端口与服务通信。然而，对于 [docker](https://docs.docker.com/engine/reference/commandline/run/#publish-or-expose-port--p---expose) 和 [kubernetes](https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service) 容器化运行环境，对外暴露端口需在容器启动前指定。针对这种场景，你可以在 `goc build` 或 `goc install` 时使用 `--agentport` 来指定插桩过的服务监听在固定的端口。

3. 如果注册中心不在本机，你可以在 `goc build` 或 `goc install` 编译目标服务时使用 `--center` 指定远端注册中心地址。

4. 目前覆盖率数据存储在插过桩的服务测，如果某个服务中途需要重启，那么其覆盖率数据在重启后会丢失。针对这个场景，你可以通过以下步骤解决：

    1. 在重启前，通过 `goc profile -o a.cov` 命令收集一次覆盖率
    2. 测试结束后，通过 `goc profile -o b.cov` 命令再收集一次覆盖率
    3. 通过 `goc merge a.cov b.cov -o merge.cov` 命令合并两次的覆盖率

## Blogs

- [Go语言系统测试覆盖率收集利器 goc](https://mp.weixin.qq.com/s/DzXEXwepaouSuD2dPVloOg)
- [聊聊Go代码覆盖率技术与最佳实践](https://mp.weixin.qq.com/s/SQHzsfV5T_B8fmt9NzGA7Q)

## RoadMap

- [x] 支持系统测试中收集代码覆盖率
- [x] 支持运行时对被测服务代码覆盖率计数器清零
- [x] 支持精准测试
- [x] 支持基于 Pull Request 的增量代码覆盖率报告
- [ ] 优化插桩计数器带来的性能损耗

## Contributing

我们欢迎各种形式的贡献，包括提交 bug、提新需求、优化文档和改进 UI 等等。

感谢所有的[贡献者](https://github.com/qiniu/goc/graphs/contributors)!!

## License

Goc is released under the Apache 2.0 license. See [LICENSE.txt](https://github.com/qiniu/goc/blob/master/LICENSE)

## 加入微信群聊
![WeChat](docs/images/wechat.png)