.PHONY: build

BUILD_DATE=$(shell date -R)
VERSION=$(shell git describe --tags)
GOVERSION=$(shell go version)

DOCKER_IMAGE=olblak/updatecli
DOCKER_TAG=$(VERSION)

local_bin=./bin/updateCli
build:
	echo $(VERSION)
	# Only build for the current host's OS and arch. Use 'make build.all' for cross-compile
	OS_TARGETS="$(shell go env GOHOSTOS)" \
	ARCH_TARGETS="$(shell go env GOHOSTARCH)" \
	CUSTOM_BINARY="$(local_bin)" \
		./utils/build.sh

build.all:
	./utils/build.sh

diff:
	"$(local_bin)" diff --config ./updateCli.d

show:
	"$(local_bin)" show --config ./updateCli.d

apply:
	"$(local_bin)" apply --config ./updateCli.d

version:
	"$(local_bin)" version

docker.build:
	docker build \
		-t "$(DOCKER_IMAGE):$(DOCKER_TAG)" \
		-t "$(DOCKER_IMAGE):latest" \
		-t "ghcr.io/$(DOCKER_IMAGE):$(DOCKER_TAG)" \
		-t "ghcr.io/$(DOCKER_IMAGE):latest" \
		-f Dockerfile \
		.

docker.run:
	docker run -i -t --rm --name updateCli $(DOCKER_IMAGE):$(DOCKER_TAG) --help
docker.test:
	docker run -i -t \
		-v $$PWD/updateCli.d:/home/updatecli/updateCli.d:ro \
		olblak/updatecli:latest --config /home/updatecli/updateCli.d/pluginsite-api.yaml

docker.push:
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest
	docker push ghcr.io/$(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push ghcr.io/$(DOCKER_IMAGE):latest

display: echo $(DOCKER_TAG)

test:
	go test ./...

lint:
	golangci-lint run
