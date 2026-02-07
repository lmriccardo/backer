.DEFAULT_GOAL := help

CLI_DIR 	:= cli
DAEMON_DIR  := daemon
UI_DIR 	 	:= ui

.PHONY: help

help:
	@echo "Targets:"
	@echo "  make fmt        - format all"
	@echo "  make lint       - lint all"
	@echo "  make test       - test all"
	@echo "  make build      - build all"
	@echo "  make dev        - start dev stack (daemon + ui)"
	@echo "  make api-gen    - regenerate API clients"
	@echo ""
	@echo "Component targets:"
	@echo "  make cli/<t>    - run target <t> in CLI"
	@echo "  make daemon/<t> - run target <t> in daemon"
	@echo "  make ui/<t>     - run target <t> in UI"

# Generic delegator: make cli/test => (cd product/cli && make test)
.PHONY: cli/% daemon/% ui/% api/%
cli/%:
	@$(MAKE) -C $(CLI_DIR) $*
daemon/%:
	@$(MAKE) -C $(DAEMON_DIR) $*
ui/%:
	@$(MAKE) -C $(UI_DIR) $*
api/%:
	@$(MAKE) -C $(API_DIR) $*

.PHONY: fmt lint test build clean
fmt:    cli/fmt daemon/fmt ui/fmt
lint:   cli/lint daemon/lint ui/lint
test:   cli/test daemon/test ui/test
build:  cli/build daemon/build ui/build
clean:  cli/clean daemon/clean ui/clean

# “Dev” is usually daemon + UI watch mode; CLI doesn’t need a long-running process
.PHONY: dev
dev:
	@$(MAKE) daemon/dev & \
	$(MAKE) ui/dev
