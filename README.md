# goc

goc v2 版本开发中

## Quick Start

### 编译要求 

`Go 1.16+`

### 新特性一览

#### 1. 只支持 go module 工程

考虑到 GOPATH 已被官方明确将淘汰，以及支持 GOPATH 工程带来的巨大工作量，v2 不再支持 GOPATH 工程。

#### 2. 命令行 flag 解析优化

在 v1 版本中，`go build -o -ldflags 'foo=bar' ./app/main.go` 与 goc 命令并不等价，首先你需要切换到 `.app/` 目录中，然后执行 `goc build --buildflags="-o -ldflags 'foo=bar' ."`。这个转换给使用者带来不小的负担，特别是 `'"` 混杂在一起时，感觉会更难受。

在 v2 版本中，goc 编译命令和 go 编译命令已经极为相似，例如

```bash
go build -o -ldflags 'foo=bar' ./app/main.go
# 等价于
goc build -o -ldflags 'foo=bar' ./app/main.go
#
go build -o -ldflags 'foo=bar' ./app
# 等价于
goc build -o -ldflags 'foo=bar' ./app
#
go install ./app/...
# 等价于
goc install ./app/...
#
```

由于 go 命令对 flags 和 args 的相对位置有着严格要求：`go build [-o output] [build flags] [packages]`，所以在指定 goc 自己的 flags （所有 goc flags 都是 `--` 开头）必须和 `build flags` 位置保持相同，即：

```bash
goc build --debug -o /home/app . # 合法

goc build -o /home/app . --debug # 非法
```

#### 3. 日志优化

带颜色日志，以及长时间操作时（例如 build, copy）会有转圈动画。

#### 4. 被测服务部署优化

在 v1 版本中，当被测服务在 docker 中，goc server 在外部时，会要求在容器启动时额外开启端口转发，并且编译还需带额外参数。这给部署带来不便。

在 v2 版本中，不再有这一限制，只需要 goc server 能够被被测服务访问即可。

#### 5. watch 模式

当使用 `goc build --mode watch .` 编译后，被测服务任何覆盖率变化都将实时推送到 goc server。

用户可以使用该 websocket 连接 `ws://[goc_server_host]/cover/ws/watch` 观察到被测服务的新触发代码块，推送信息格式如下：

```bash
qiniu.com/kodo/apiserver/server/main.go:42.49,43.13 1 0
#
# importpath/filename.go:Line0.Col1,Line1,Col1 1 0
```

除此之外，原来的全局整体覆盖率可正常获取，不受影响。

#### 6. 跨平台支持

1. 支持 `Linux/Macos/Windows`
2. 支持 go 的交叉编译