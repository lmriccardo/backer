APP_NAME 		:= backerd
BUILD 			:= ../build
CMD_FOLDER		:= ./cmd
.DEFAULT_GOAL 	:= help

# Default target OS/ARCH (can override with "make GOOS=windows")
GOOS ?= linux
GOARCH ?= amd64

# Add .exe suffix on Windows
ifeq ($(GOOS),windows)
    EXE=.exe
else
    EXE=
endif

GOFILES := $(shell find . -name '*.go' -not -path './vendor/*')
BIN 	:= $(BUILD)/$(APP_NAME)$(EXE)

MODULE 		:= $(shell go list -m)
VERSION 	:= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT 		:= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE 		:= $(shell date -u +%Y-%m-%d)
GOVERSION 	:= $(shell go env GOVERSION)

LDFLAGS := -X '$(MODULE)/internal/platform/version.Version=$(VERSION)' \
           -X '$(MODULE)/internal/platform/version.Commit=$(COMMIT)' \
           -X '$(MODULE)/internal/platform/version.Date=$(DATE)' \
           -X '$(MODULE)/internal/platform/version.GoVersion=$(GOVERSION)'

.PHONY: build run clean fmt lint test tidy dev help

help:
	@echo "Available targets:"
	@echo "  build       Build the application"
	@echo "  run         Run the application"
	@echo "  clean       Clean the build artifacts"
	@echo "  fmt         Format the code"
	@echo "  lint        Lint the code"
	@echo "  test        Run tests"
	@echo "  tidy        Tidy the dependencies"
	@echo "  dev         Development mode"

fmt:
	go fmt ./...

lint:
	@command -v golangci-lint >/dev/null || (echo "golangci-lint not installed" && exit 1)
	golangci-lint run ./...

build: $(BIN)

$(BIN): $(GOFILES) go.mod go.sum
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath \
		-ldflags "$(LDFLAGS)" -o $(BIN) $(CMD_FOLDER)/$(APP_NAME)

run: build
	./$(BIN)

clean:
	rm -f ./$(BIN)

test:
	go test -v ./test/...

tidy:
	go mod tidy

dev:
    @echo "Run manually: make run"
	@echo "Or install 'air' and implement hot-reload."
