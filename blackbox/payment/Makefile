DARWIN_TARGET_ENV=GOOS=darwin GOARCH=amd64
LINUX_TARGET_ENV=GOOS=linux GOARCH=amd64

.PHONY: build  test

PKG_NAME=$(shell basename `pwd`)
PKG_LIST := ./config ./server
export GO111MODULE=on

all: build

build:
	$(DARWIN_TARGET_ENV) go build -o ./bin/payment_darwin
	$(LINUX_TARGET_ENV) go build -o ./bin/payment_linux

test:
	go test -short -v -race -cover -p=1 $(PKG_LIST)
	go clean -testcache $(PKG_LIST)
