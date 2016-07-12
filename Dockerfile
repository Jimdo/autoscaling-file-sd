FROM golang:alpine

MAINTAINER Knut Ahlers <knut@ahlers.me>

ADD . /go/src/github.com/Jimdo/autoscaling-file-sd
WORKDIR /go/src/github.com/Jimdo/autoscaling-file-sd

RUN set -ex \
 && apk add --update git ca-certificates \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" \
 && apk del --purge git

VOLUME ["/config"]

ENTRYPOINT ["/go/bin/autoscaling-file-sd"]
CMD ["--"]
