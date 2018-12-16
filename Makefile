GOBIN ?= $(shell go env GOPATH)/bin

install:	update
	go  install -ldflags  "-s -w"  ./cmd/gvim

update:
	git pull origin master
	git submodule update  --init
