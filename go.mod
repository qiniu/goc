module github.com/qiniu/goc/v2

go 1.16

require (
	github.com/gin-gonic/gin v1.7.2
	github.com/gorilla/websocket v1.4.2
	github.com/k0kubun/go-ansi v0.0.0-20180517002512-3bf9e2903213
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.13 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d
	github.com/olekukonko/tablewriter v0.0.5
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/tongjingran/copy v1.4.2
	go.uber.org/zap v1.17.0
	golang.org/x/mod v0.4.2
	golang.org/x/sys v0.0.0-20210608053332-aa57babbf139 // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d
	golang.org/x/tools v0.1.3
	k8s.io/kubectl v0.21.2
	k8s.io/test-infra v0.0.0-20210618100605-34aa2f2aa75b
)

replace k8s.io/client-go => k8s.io/client-go v0.21.1
