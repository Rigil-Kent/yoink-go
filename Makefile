BIN := yoink
BUILD_DIR := build
REGISTRY := git.brizzle.dev/bryan/yoink-go
VERSION ?= $(shell git describe --tags --always --dirty)
NOTES ?= ""

.PHONY: all windows linux darwin clean docker-build docker-push tag gitea-release release

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

tag:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make tag VERSION=1.2.0"; exit 1; fi
	git tag $(VERSION)
	git tag -f latest
	git push origin $(VERSION)
	git push origin -f latest

gitea-release:
	tea release create \
		--tag $(VERSION) \
		--title "$(VERSION)" \
		$(if $(NOTES),--note $(NOTES),) \
		--asset $(BUILD_DIR)/$(BIN)-windows-amd64.exe \
		--asset $(BUILD_DIR)/$(BIN)-linux-amd64 \
		--asset $(BUILD_DIR)/$(BIN)-linux-arm64 \
		--asset $(BUILD_DIR)/$(BIN)-darwin-amd64 \
		--asset $(BUILD_DIR)/$(BIN)-darwin-arm64

release:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make release VERSION=1.3.0 NOTES='...'"; exit 1; fi
	$(MAKE) tag VERSION=$(VERSION)
	$(MAKE) clean all
	$(MAKE) gitea-release VERSION=$(VERSION) NOTES=$(NOTES)
	$(MAKE) docker-push VERSION=$(VERSION)

clean:
	rm -rf $(BUILD_DIR)
