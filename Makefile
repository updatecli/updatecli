.PHONY: build

ENV=updateCli.env

include $(ENV)

BUILD_DATE=`date -R`
VERSION=$(shell git describe --tags)
GOVERSION=$(shell go version)

DOCKER_IMAGE=olblak/updatecli
DOCKER_TAG=$(shell git describe --exact-match --tags $$(git log -n1 --pretty='%h'))

build:
	echo $(VERSION)
	go build \
		-ldflags "-w -s \
			-X \"github.com/olblak/updateCli/pkg/version.BuildTime=$(BUILD_DATE)\" \
			-X \"github.com/olblak/updateCli/pkg/version.GoVersion=$(GOVERSION)\" \
			-X \"github.com/olblak/updateCli/pkg/version.Version=$(VERSION)\""\
		-o bin/updatecli

run:
	./bin/updatecli --config ./updateCli.d

docker.build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -t $(DOCKER_IMAGE):latest .

docker.run:
	docker run -i -t --rm --name updateCli $(DOCKER_IMAGE):$(DOCKER_TAG) --help
docker.test:
	docker run -i -t \
		-v $$PWD/updateCli.d:/home/updatecli/updateCli.d:ro \
		olblak/updatecli:latest --config /home/updatecli/updateCli.d/pluginsite-api.yaml

docker.push:
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

display: echo $(DOCKER_TAG)
