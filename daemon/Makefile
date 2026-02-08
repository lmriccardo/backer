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
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BIN) $(CMD_FOLDER)/$(APP_NAME)

run: build
	./$(BIN)

clean:
	rm -f ./$(BIN)

test:
	go test ./...

tidy:
	go mod tidy

dev:
    @echo "Run manually: make run"
	@echo "Or install 'air' and implement hot-reload."
