.PHONY: build

build:
	- find . -name '*.go' | xargs -I{} goimports -w {}
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/weaccount cmd/main.go
run: build
	bin/weaccount

