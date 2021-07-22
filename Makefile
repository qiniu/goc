DEFAULT_EXCEPT_PKGS := e2e

build:
	go build .

install:
	go install ./...

test:
	go test -p 1 -coverprofile=coverage.txt ./pkg/...

fmt:
	go fmt ./...

govet-check:
	go vet ./...

e2e:
	ginkgo tests/e2e/...

clean:
	find ./ -type f -name 'coverage.txt' -delete
	find ./ -type f -name 'goc' -delete
	find ./ -type f -name 'gocc' -delete
	rm -rf ./tests/e2e/tmp