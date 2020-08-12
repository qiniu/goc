module example.com/gg/a

replace (
    github.com/qiniu/bar => ../home/foo/bar
    github.com/qiniu/bar2 => github.com/baniu/bar3 v1.2.3
)

require (
	github.com/qiniu/bar v1.0.0
    github.com/qiniu/bar2 v1.2.0
)
