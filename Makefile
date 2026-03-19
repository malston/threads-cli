BINARY := threads
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X github.com/malston/threads-cli/cmd.Version=$(VERSION) -X github.com/malston/threads-cli/cmd.BuildTime=$(BUILD_TIME)

.PHONY: build test test-integration lint install clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:
	go test ./...

test-integration:
	go test -tags integration ./...

lint:
	golangci-lint run ./...

install:
	go install -ldflags "$(LDFLAGS)" .

clean:
	rm -f $(BINARY)
