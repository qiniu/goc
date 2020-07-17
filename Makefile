DEFAULT_EXCEPT_PKGS := e2e

all:
	go install ./...

test:
	go test -cover -p 1 `go list ./... | grep -v -E ${DEFAULT_EXCEPT_PKGS}`

fmt:
	go fmt ./...

govet-check:
	go vet ./...

clean:
	find tests/ -type f -name '*.bak' -delete 
	find tests/ -type f -name '*.cov' -delete 
	find tests/ -type f -name 'simple-project' -delete 
	find tests/ -type f -name '*_profile_listen_addr' -delete 
	find tests/ -type f -name 'simple_gopath_project' -delete 
	