
BUILD_DATE=$(shell date -R)
VERSION=$(shell git describe --tags)
GOVERSION=$(shell go version)

DOCKER_IMAGE=olblak/updatecli
DOCKER_TAG=$(VERSION)

local_bin=./bin/updateCli

.PHONY: build
build: ## Build updateCli for the host OS and architecture
	echo $(VERSION)
	# Only build for the current host's OS and arch. Use 'make build.all' for cross-compile
	OS_TARGETS="$(shell go env GOHOSTOS)" \
	ARCH_TARGETS="$(shell go env GOHOSTARCH)" \
	CUSTOM_BINARY="$(local_bin)" \
		./utils/build.sh

.PHONY: build.all
build.all: ## Build updateCli for all supported OSes and architectures
	./utils/build.sh

.PHONY: diff
diff: ## Run the "diff" updateCli's subcommand for smoke test
	"$(local_bin)" diff --config ./updateCli.d

.PHONY: show
show: ## Run the "show" updateCli's subcommand for smoke test
	"$(local_bin)" show --config ./updateCli.d

.PHONY: apply
apply: ## Run the "apply" updateCli's subcommand for smoke test
	"$(local_bin)" apply --config ./updateCli.d

.PHONY: version
version: ## Run the "version" updateCli's subcommand for smoke test
	"$(local_bin)" version

.PHONY: docker.build
docker.build: ## Build the updateCli's Docker image
	docker build \
		-t "$(DOCKER_IMAGE):$(DOCKER_TAG)" \
		-t "$(DOCKER_IMAGE):latest" \
		-t "ghcr.io/$(DOCKER_IMAGE):$(DOCKER_TAG)" \
		-t "ghcr.io/$(DOCKER_IMAGE):latest" \
		-f Dockerfile \
		.

.PHONY: docker.run
docker.run: docker.build ## Execute the updateCli's Docker image
	docker run -i -t --rm --name updateCli $(DOCKER_IMAGE):$(DOCKER_TAG) --help

.PHONY: docker.test
docker.test: docker.build ## Smoke Test the updateCli's Docker image
	docker run -i -t \
		-v $$PWD/updateCli.d:/home/updatecli/updateCli.d:ro \
		"$(DOCKER_IMAGE):$(DOCKER_TAG)" --config /home/updatecli/updateCli.d/pluginsite-api.yaml

.PHONY: docker.push
docker.push: docker.build ## Push the updateCli's Docker image to remote registry
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest
	docker push ghcr.io/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push ghcr.io/$(DOCKER_IMAGE):latest

.PHONY: display
display: ## Prints the current DOCKER_TAG
	echo $(DOCKER_TAG)

.PHONY: test
test: ## Execute the Golang's tests for updateCli
	go test ./...

.PHONY: lint
lint: ## Execute the Golang's linters on updateCli's source code
	golangci-lint run

.PHONY: help
help: ## Show this Makefile's help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
