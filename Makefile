BINARY := harness-init
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo "dev")
LDFLAGS := -ldflags "-s -w -X github.com/alvaroeng98/HarnessInit/cmd.version=$(VERSION)"

.PHONY: build build-all test install clean

build:
	go build $(LDFLAGS) -o $(BINARY) .

build-all:
	GOOS=linux   GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64   .
	GOOS=linux   GOARCH=arm64  go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64   .
	GOOS=darwin  GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64  .
	GOOS=darwin  GOARCH=arm64  go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64  .
	GOOS=windows GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64  go build $(LDFLAGS) -o dist/$(BINARY)-windows-arm64.exe .

test:
	go test ./... -v

install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)
	rm -rf dist/
