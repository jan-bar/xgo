FROM karalabe/xgo-base

MAINTAINER janbar <janbar@163.com>

ENV G_EXPERIMENTAL true
ENV G_HOME /go
ENV G_MIRROR https://golang.google.cn/dl/
ENV GOPROXY https://goproxy.cn,direct
ENV GOPATH $G_HOME/path
ENV GOROOT $G_HOME/go
ENV PATH $GOROOT/bin:$GOPATH/bin:$PATH
ENV GO111MODULE on

WORKDIR /build

RUN wget https://github.com/voidint/g/releases/download/v1.2.1/g1.2.1.linux-amd64.tar.gz -q -O - | tar -xzC /sbin/ && \
    g install $(g ls-remote stable | tail -n1)

ENTRYPOINT ["./build.sh"]