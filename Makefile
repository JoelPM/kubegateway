# Makefile for the Docker image gcr.io/google_containers/kubegateway
# MAINTAINER: Tim Hockin <thockin@google.com>
# If you update this image please bump the tag value before pushing.


APP = kubegateway

TAG = 0.0.2
PREFIX = joelpm

.PHONY: all kubegateway container push clean test

all: container

deps:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go get

$(APP): $(APP).go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -a --ldflags '-w' -o $(APP) ./$(APP).go

container: binary
	docker build -t $(PREFIX)/$(APP):$(TAG) .

push:
	gcloud docker push $(PREFIX)/$(APP):$(TAG)

clean:
	rm -f $(APP)

test: clean
	godep go test -v --vmodule=*=4
