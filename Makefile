VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BIN = saggycli
LDFLAGS = -X saggy.Version=$(VERSION)
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

.PHONY: build test clean ci-format ci-test ci-build ci-release

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BIN) .

test:
	./tests/test_go

ci-format:
	@test -z "$$(gofmt -l .)" || (echo "gofmt found unformatted files:" && gofmt -l . && exit 1)

ci-test: test

ci-build:
	$(foreach platform,$(PLATFORMS),\
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		go build -ldflags "$(LDFLAGS)" \
			-o bin/$(BIN)-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform))) . ;)

ci-release: ci-build

clean:
	rm -rf bin
