BUILD=go build
TEST=go test
TEST_FLAGS=-race -v

DARWIN_TARGET_ENV=GOOS=darwin GOARCH=amd64
LINUX_TARGET_ENV=GOOS=linux GOARCH=amd64


RM=rm -f

.PHONY: all build test clean

all: build

build:
	$(DARWIN_TARGET_ENV) $(BUILD) -o ./bin/bench_darwin ./cmd/bench
	$(LINUX_TARGET_ENV) $(BUILD) -o ./bin/bench_linux ./cmd/bench

	$(DARWIN_TARGET_ENV) $(BUILD) -o ./bin/benchworker_darwin ./cmd/benchworker
	$(LINUX_TARGET_ENV) $(BUILD) -o ./bin/benchworker_linux ./cmd/benchworker

test:
	$(TEST) $(TEST_FLAGS) ./...

clean:
	$(RM) ./bin/bench_darwin
	$(RM) ./bin/bench_linux
	$(RM) ./bin/benchworker_darwin
	$(RM) ./bin/benchworker_linux
