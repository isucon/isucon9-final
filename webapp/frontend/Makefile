all: build

build: Dockerfile
	docker build -t isucon9-final/frontend .
	docker run -v $(shell pwd):/opt/frontend isucon9-final/frontend bash -x build.sh
