FROM golang:1.4.2

RUN go get github.com/tools/godep

RUN CGO_ENABLED=0 go install -a std

MAINTAINER Joel Meyer <joel.meyer@gmail.com>

WORKDIR /go/src/app

COPY go-package-setup /usr/local/bin/
