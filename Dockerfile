FROM golang:alpine

MAINTAINER Knut Ahlers <knut@ahlers.me>

ADD . /go/src/github.com/Luzifer/prometheus-metar
WORKDIR /go/src/github.com/Luzifer/prometheus-metar

RUN set -ex \
 && apk add --update git \
 && go install -ldflags "-X main.version=$(git describe --tags || git rev-parse --short HEAD || echo dev)" \
 && apk del --purge git

EXPOSE 3000

ENTRYPOINT ["/go/bin/prometheus-metar"]
CMD ["--"]
