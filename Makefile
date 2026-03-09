BIN := yoink
BUILD_DIR := build
REGISTRY := git.brizzle.dev/bryan/yoink-go
VERSION := $(shell git describe --tags --always --dirty)

.PHONY: all windows linux darwin clean docker-build docker-push

all: windows linux darwin

windows:
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BIN)-windows-amd64.exe

linux:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BIN)-linux-amd64
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BIN)-linux-arm64

darwin:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BIN)-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BIN)-darwin-arm64

docker-build:
	podman build --format docker \
		-t $(REGISTRY):$(VERSION) \
		-t $(REGISTRY):latest \
		.

docker-push: docker-build
	podman push $(REGISTRY):$(VERSION)
	podman push $(REGISTRY):latest

clean:
	rm -rf $(BUILD_DIR)
