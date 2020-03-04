.PHONY: build

ENV=updateCli.env

include $(ENV)

DOCKER_IMAGE=olblak/updatecli
DOCKER_TAG=$(shell git describe --exact-match --tags $$(git log -n1 --pretty='%h'))

build:
	go build -o bin/updatecli

run:
	./bin/updatecli --config ./updateCli.d

docker.build:
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -t $(DOCKER_IMAGE):latest .

docker.run:
	docker run -i -t --rm --name updateCli $(DOCKER_IMAGE):$(DOCKER_TAG) --help
docker.test:
	docker run -i -t \
		-v $$PWD/updateCli.yaml:/home/updatecli/updateCli.yaml:ro \
		olblak/updatecli:latest --config /home/updatecli/updateCli.yaml

docker.push:
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

display:
	echo $(DOCKER_TAG)
