FROM karalabe/xgo-base

MAINTAINER janbar <janbar@163.com>

ENV GO_VERSION 11701
ENV GOPROXY https://goproxy.cn,direct
ENV GO111MODULE on
WORKDIR /build

RUN \
  export ROOT_DIST=https://studygolang.com/dl/golang/go1.17.1.linux-amd64.tar.gz        && \
  export ROOT_DIST_SHA=dab7d9c34361dc21ec237d584590d72500652e7c909bf082758fb63064fca0ef && \
  sed -i '60,$d' $BOOTSTRAP_PURE && \
  $BOOTSTRAP_PURE

ENTRYPOINT ["./build.sh"]
