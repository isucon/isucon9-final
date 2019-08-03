BUILD=go build
TEST=go test
TEST_FLAGS=-race -v

RM=rm -f

.PHONY: all build test clean

all: build

build:
	GOOS=darwin GOARCH=amd64 $(BUILD) -o ./bin/isutrain_darwin_amd64 .
	GOOS=linux GOARCH=amd64 $(BUILD) -o ./bin/isutrain_linux_amd64 .

test:
	$(TEST) $(TEST_FLAGS) ./...

clean:
	$(RM) ./bin/isutrain
