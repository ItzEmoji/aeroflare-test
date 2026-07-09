BIN ?= aeroflare

.PHONY: build test hash get
.DEFAULT_GOAL := build

build: ## Build both binaries into ./bin
	@scripts/build.sh

test: ## Run the Go tests and the shell tests
	@scripts/test.sh

hash: ## Recompute default.nix's vendorHash (after go.mod/go.sum changes)
	@scripts/hash.sh

get: ## Download a verified prebuilt binary into ./bin (BIN=aeroflare-ci)
	@scripts/get.sh $(BIN)
