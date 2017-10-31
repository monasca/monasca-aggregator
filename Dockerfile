# librdkafka is not available in a stable alpine release
from alpine:edge

env GOPATH=/go

copy . $GOPATH/src/github.com/monasca/monasca-aggregator

run apk add --no-cache openssl py2-jinja2 && \
  apk add --no-cache --virtual build-dep \
    librdkafka-dev git go glide make g++ openssl-dev && \
  cd $GOPATH/src/github.com/monasca/monasca-aggregator && \
  glide install && \
  go build -tags static && \
  cp monasca-aggregator /monasca-aggregator && \
  cp aggregation-specifications.yaml / && \
  apk del build-dep && \
  cd / && \
  rm -rf /go

copy config.yaml start.sh /
expose 8080

cmd ["/start.sh"]

