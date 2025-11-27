BINARY ?= mangashelf
CMD ?= ./cmd/mangashelf

.PHONY: build dev test clean build-all

build:
@mkdir -p bin
GO111MODULE=on go build -o bin/$(BINARY) $(CMD)

dev:
GO111MODULE=on go run $(CMD)

test:
GO111MODULE=on go test ./...

clean:
rm -rf bin dist

build-all:
@mkdir -p dist
GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o dist/$(BINARY)-linux-amd64 $(CMD)
GOOS=linux GOARCH=arm64 GO111MODULE=on go build -o dist/$(BINARY)-linux-arm64 $(CMD)
GOOS=darwin GOARCH=amd64 GO111MODULE=on go build -o dist/$(BINARY)-darwin-amd64 $(CMD)
GOOS=darwin GOARCH=arm64 GO111MODULE=on go build -o dist/$(BINARY)-darwin-arm64 $(CMD)
GOOS=windows GOARCH=amd64 GO111MODULE=on go build -o dist/$(BINARY)-windows-amd64.exe $(CMD)
