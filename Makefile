MAIN_FILE=main.go
TEST_FILE=main_test.go
BINARY_NAME=proxy
BINARY_LINUX=$(BINARY_NAME)_linux
BINARY_WINDOWS=$(BINARY_NAME)_windows
BINARY_BSD=$(BINARY_NAME)_freebsd

GO=go
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOTEST=$(GO) test
GOGET=$(GO) get

.DEFAULT_GOAL := all
.PHONY: test

DIST_DIR=dist
TEST_DIR=
all: test build
build:
	$(GOBUILD) -ldflags "-s -w" -o $(DIST_DIR)/$(BINARY_NAME)   $(MAIN_FILE)
test:
	$(GOTEST) -v $(MAIN_FILE) $(TEST_FILE)
clean:
	$(GOCLEAN)
	rm -f $(DIST_DIR)/$(BINARY_NAME)
	rm -f $(DIST_DIR)/$(BINARY_LINUX)
	rm -f $(DIST_DIR)/$(BINARY_WINDOWS)
	rm -f $(DIST_DIR)/$(BINARY_BSD)

deps:
	$(GOGET) ./..


build_linux: test linux
build_linux_xgo: test linux
build_windows: test windows
build_windows_xgo: test windows
build_freebsd: test freebsd

linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "-s -w" -o $(DIST_DIR)/$(BINARY_LINUX)  $(MAIN_FILE)

windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "-s -w" -o $(DIST_DIR)/$(BINARY_WINDOWS)  $(MAIN_FILE)

freebsd:
	GOOS=freebsd GOARCH=amd64 $(GOBUILD) -ldflags "-s -w" -o $(DIST_DIR)/$(BINARY_BSD)  $(MAIN_FILE)

linux_amd64_xgo:
	xgo --targets=linux/amd64  --dest $(DIST_DIR)/ ./

linux_386_xgo:
	xgo --targets=linux/386  --dest $(DIST_DIR)/ ./

windows_xgo:
	xgo --targets=windows/amd64  --dest $(DIST_DIR)/ ./

freebsd_xgo:
	xgo --targets=freebsd/amd64  --dest $(DIST_DIR)/ ./


