FROM golang:1.12

RUN apt-get update && apt-get -y install mariadb-client

ENV GO111MODULE=on

WORKDIR /go/src/webapp
CMD ["go", "run", "main.go", "utils.go"]
