DEFAULT_EXCEPT_PKGS := e2e

all:
	go install ./...

test:
	go test -cover -p 1 `go list ./... | grep -v -E ${DEFAULT_EXCEPT_PKGS}`

fmt:
	go fmt ./...

govet-check:
	go vet ./...