# 通信协议设计

## 背景

v1 版本中，被插桩的服务会暴露一个 HTTP 接口，由 goc server 来访问获取覆盖率。

该方式要求被插桩的服务要有一个外界可访问的 ip + port。

如果被测服务躲在 NAT 网络下，该方式就不可行了，典型的就是被测服务由 docker 拉起，而 goc server 部署在另外的网络。

## 新设计选择

新设计只需要暴露 goc server 的地址，由插桩服务发起链接，然后保持长链接，在该长链接上构建 goc 自己的业务逻辑。

### 自行设计 TCP 应用层协议

go 语言做网络编程非常适合，非阻塞地处理“粘包”也不麻烦。但设计出来不管是纯二进制的、还是类似 HTTP 的，都不会是通用协议，后续维护和扩展估计是个大坑。

### websocket + jsonrpc2

websocket + jsonrpc2 有流式调用，消息边界。非常适合

我找到 `github.com/goriila/websocket` 和 `github.com/sourcegraph/jsonrpc2` 库，后者 import 了前者，前者没有 import 任何其它外部库，全是标准库。把两个库的代码合并一下：`github.com/lyyyuna/jsonrpc2` 就是一个无任何外部库引用的 jsonrpc 实现。

这非常适合由插桩代码来使用，因为该库没有再引用其它库，**不会污染原服务的依赖关系**。

### gRPC

老实说 gRPC 在这里更适合作为通信协议来使用，更快更通用，流式调用也有，上一小节的 `github.com/sourcegraph/jsonrpc2` 使用广度就很低。

但 gRPC 的 go 实现有一个很大的缺点，用了一些非标准库，且有版本依赖。我们不清楚原服务是不是有特定 gRPC 要求，或是 goc 插入的 gRPC 库会导致编译依赖冲突，或者是编译后运行冲突。

所以不适合。

### 结论

先使用 websocket + jsonrpc2 来做吧。

## 协议内容

### 注册

### 获取覆盖率

### 清空覆盖率

### watch

### 异常处理