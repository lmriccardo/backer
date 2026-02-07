APP_NAME 		:= backerd
BIN_FOLDER 		:= build
CMD_FOLDER		:= ./cmd
.DEFAULT_GOAL 	:= help

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

build:
	go build -o $(BIN_FOLDER)/$(APP_NAME) $(CMD_FOLDER)/$(APP_NAME)

run: build
	./$(BIN_FOLDER)/$(APP_NAME)

clean:
	rm -f ./$(BIN_FOLDER)/$(APP_NAME)

test:
	go test ./...

tidy:
	go mod tidy

dev:
    @echo "Run manually: make run"
	@echo "Or install 'air' and implement hot-reload."
