# Makefile for the Docker image gcr.io/google_containers/kubegateway
# MAINTAINER: Tim Hockin <thockin@google.com>
# If you update this image please bump the tag value before pushing.

.PHONY: all kubegateway container push clean test

APP = kubegateway

TAG = 0.0.2
PREFIX = joelpm


DEVCONTAINER = $(PREFIX)/$(APP)_dev:1.4.2


all: container

.devcontainer:
	docker build -t $(DEVCONTAINER) .
	docker inspect -f '{{.Id}}' $(DEVCONTAINER) > .devcontainer

devcontainer: .devcontainer

binary: devcontainer
	docker run -v $(PWD):/go/src/app --entrypoint /bin/sh $(DEVCONTAINER) -c 'go-package-setup && make $(APP)'

godeps:
	docker run -v $(PWD):/go/src/app --entrypoint /bin/sh $(DEVCONTAINER) -c 'go-package-setup && godeps save'

$(APP): $(APP).go
	go-package-setup
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -a -installsuffix cgo --ldflags '-w' ./$(APP).go

container: kubegateway
	docker build -t $(PREFIX)/$(APP):$(TAG) .

push:
	gcloud docker push $(PREFIX)/$(APP):$(TAG)

clean:
	rm -f $(APP)

test: clean
	godep go test -v --vmodule=*=4
