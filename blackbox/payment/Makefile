.PHONY: build  test

PKG_NAME=$(shell basename `pwd`)
PKG_LIST := ./config ./server
export GO111MODULE=on

all: build

build:
	go build

test:
	go test -short -v -race -cover -p=1 $(PKG_LIST)
