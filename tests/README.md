# How to run e2e test

## 如何运行集成测试

首先编译 goc，并加入 `PATH`。

```
make install
make e2e
```

## 如何添加 sample

为了不让 case 之间执行时互相干扰，集测设计了 samples 管理系统。

在 `tests/e2e/samples` 目录中，按如下格式在 `meta.yaml` 中添加 sample 的元信息：

```yaml
samples:
  basic:
    dir: basic-project
    description: a basic project only print hello world
  gomod:
    dir: invalidmod-project
    description: a project which contains invalid go.mod
```

其中 basic 是键值，dir 是目录名，description 你可以添加一些说明，让大家一目了然这个 sample 的特点。

然后就按照 `meta.yaml` 填写的信息在 samples 目录内添加相应的工程目录即可。