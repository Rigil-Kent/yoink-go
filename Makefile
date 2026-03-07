BIN := yoink
BUILD_DIR := build

.PHONY: all windows linux darwin clean

all: windows linux darwin

windows:
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BIN)-windows-amd64.exe

linux:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BIN)-linux-amd64
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BIN)-linux-arm64

darwin:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BIN)-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BIN)-darwin-arm64

clean:
	rm -rf $(BUILD_DIR)
