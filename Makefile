VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BIN = saggycli

.PHONY: build test clean

build:
	go build -ldflags "-X saggy.Version=$(VERSION)" -o $(BIN) .

test:
	./tests/test_go

clean:
	rm -f $(BIN)
