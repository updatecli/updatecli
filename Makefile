.PHONY: build

BUILD_DATE=$(shell date -R)
VERSION=$(shell git describe --tags)
GOVERSION=$(shell go version)

DOCKER_IMAGE=olblak/updatecli
DOCKER_TAG=$(VERSION)

build:
	echo $(VERSION)
	go build \
		-ldflags "-w -s \
        -X \"github.com/olblak/updateCli/pkg/core/version.BuildTime=$(BUILD_DATE)\" \
        -X \"github.com/olblak/updateCli/pkg/core/version.GoVersion=$(GOVERSION)\" \
        -X \"github.com/olblak/updateCli/pkg/core/version.Version=$(VERSION)\""\
        -o bin/updatecli


build.all:
	./utils/build.sh

diff:
	./bin/updatecli diff --config ./updateCli.d

show:
	./bin/updatecli show --config ./updateCli.d

apply:
	./bin/updatecli apply --config ./updateCli.d

version:
	./bin/updatecli version

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
