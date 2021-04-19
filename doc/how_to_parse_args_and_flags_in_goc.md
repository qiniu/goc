# goc 中的参数处理设计

## 背景

goc build/install/run 有不少人反馈使用起来和 go build/install/run 相比，还是有不少的差异。这种差异导致在日常开发、CI/CD 中替换不便，有些带引号的参数会被改写的面目全非。

## 原则

goc build/install/run 会尽可能的模仿 go 原生的方式去处理参数。

## 主要问题

1. goc 使用 cobra 库来组织各个子命令。cobra 对 flag 处理采用的是 posix 风格（两个个短横线），和 go 的 flag 处理差异很大（一个短横线）。
2. go 命令中 args 和 flags 有着严格先后顺序。而 cobra 库对 flags 和 args 的位置没有要求。
3. 参数中 `[packages]` 有多种组合情况，会影响到插桩的起始位置。
4. goc 还有自己参数，且需要和**非** goc build/install/run 的子命令保持一致（两个短横线）。