# Makefile for the Docker image joelpm/kubegateway
# MAINTAINER: Joel Meyer <joel.meyer@gmail.com>

TAG = 0.0.0
PREFIX = joelpm/kubegateway

APP = kubegateway

default: test

setup:  
	go get github.com/mailgun/godebug
	go get github.com/tools/godep

buildgo:  
	CGO_ENABLED=0 GOOS=linux godep go build -ldflags "-s" -a -installsuffix cgo -o $(APP) .

builddocker: buildgo
	docker build --rm -t $(PREFIX)/$(APP):$(TAG) .

push:
	gcloud docker push $(PREFIX)/$(APP):$(TAG)


buildp: buildgo builddocker

run: buildp  
	docker run \
	    -p 8080:80 $(PREFIX)/$(APP)

test:	clean
	godep go test -v --vmodule=*=4 -timeout=5s ./...

debug:  
ifndef $(instrument)  
	godebug run ${gofiles}
else  
	godebug run -instrument=${instrument} ${gofiles}
endif 

.PHONY: all buildgo builddocker push clean test debug

all: builddocker

clean:
	rm -f $(APP)
