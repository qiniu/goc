github.com/qiniu/api.v7 (Qiniu Go SDK v7.x)
===============

[![Build Status](https://travis-ci.org/qiniu/api.v7.svg?branch=master)](https://travis-ci.org/qiniu/api.v7) [![GoDoc](https://godoc.org/github.com/qiniu/api.v7?status.svg)](https://godoc.org/github.com/qiniu/api.v7)

[![Qiniu Logo](http://open.qiniudn.com/logo.png)](http://qiniu.com/)

# 下载

## 使用 Go mod【推荐】

在您的项目中的 `go.mod` 文件内添加这行代码

```
require github.com/qiniu/api.v7/v7 v7.4.1
```

并且在项目中使用 `"github.com/qiniu/api.v7/v7"` 引用 Qiniu Go SDK。

例如

```go
import (
    "github.com/qiniu/api.v7/v7/auth"
    "github.com/qiniu/api.v7/v7/storage"
)
```

## 不使用 Go mod【不推荐，且只能获取 v7.2.5 及其以下版本】

```bash
go get -u github.com/qiniu/api.v7
```

# go版本需求

需要 go1.10 或者 1.10 以上

#  文档

[七牛SDK文档站](https://developer.qiniu.com/kodo/sdk/1238/go) 或者 [项目WIKI](https://github.com/qiniu/api.v7/wiki)

# 示例

[参考代码](https://github.com/qiniu/api.v7/tree/master/examples)
