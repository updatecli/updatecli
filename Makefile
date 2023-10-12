
GOVERSION=$(shell go version)
export GOVERSION

DOCKER_BUILDKIT=1
export DOCKER_BUILDKIT

# Used by the test-e2e system
VENOM_VAR_binpath ?= $(CURDIR)/dist/updatecli_$(shell go env GOOS)_$(shell go env GOARCH)_v1
export VENOM_VAR_binpath
VENOM_VAR_rootpath ?= $(CURDIR)
export VENOM_VAR_rootpath

local_bin=./dist/updatecli_$(shell go env GOHOSTOS)_$(shell go env GOHOSTARCH)/updatecli

clean: ## Clean go test cache
	go clean -testcache

.PHONY: build
build: ## Build updatecli as a "dirty snapshot" (no tag, no release, but all OS/arch combinations)
	goreleaser build --snapshot --clean

.PHONY: build.all
build.all: ## Build updatecli for "release" (tag or release and all OS/arch combinations)
	goreleaser --clean --skip=publish,sign

.PHONY: release ## Create a new updatecli release including packages
release: ## release generate a release
	goreleaser release --clean --timeout=2h

.PHONY: release.snapshot ## Create a new snapshot release without publishing assets
release.snapshot: ## release.snapshot generate a snapshot release but do not published it (no tag, but all OS/arch combinations)
	goreleaser release --snapshot --clean --skip=publish,sign

.PHONY: diff
diff: ## Run the "diff" updatecli's subcommand for smoke test
	"$(local_bin)" diff --config ./updatecli.d

.PHONY: show
show: ## Run the "show" updatecli's subcommand for smoke test
	"$(local_bin)" show --config ./updatecli.d

.PHONY: apply
apply: ## Run the "apply" updatecli's subcommand for smoke test
	"$(local_bin)" apply --config ./updatecli.d

.PHONY: version
version: ## Run the "version" updatecli's subcommand for smoke test
	"$(local_bin)" version

.PHONY: test
test: ## Execute the Golang's tests for updatecli
	go test ./... -race -coverprofile=coverage.txt -covermode=atomic

test-short: ## Execute the Golang's tests for updatecli
	go test ./... -short

.PHONY: test-e2e
test-e2e: ## Execute updatecli end to end tests
	@echo "==\nUsing the following updatecli binary (from variable 'VENOM_VAR_binpath'): $(VENOM_VAR_binpath)/updatecli\n=="
	@test -n "$$GITHUB_TOKEN" || { echo "Undefined required variable 'GITHUB_TOKEN'"; exit 1; }
	@test -n "$$GITHUB_ACTOR" || { echo "Undefined required variable 'GITHUB_ACTOR'"; exit 1; }
	time venom run e2e/venom.d/* --output-dir ./e2e --format yaml

.PHONY: lint
lint: ## Execute the Golang's linters on updatecli's source code
	golangci-lint run

.PHONY: help
help: ## Show this Makefile's help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
