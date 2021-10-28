我的`dockerfile`，主要生成带交叉编译工具链的docker镜像，通过安装`g`工具去安装最新版的go环境。

```dockerfile
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
```

编译指定版本的docker镜像

```sh
docker build -t janboy/cgolang-g .
```

提供两个使用`CGO`交叉编译的示例：

`tests/sqlite`是通过`cgo`操作`sqlite`数据库，将整个目录放到服务器上，然后按照如下脚本执行

```sh
#!/bin/bash -x

# chmod +x build.sh,为脚本加上可执行权限
# 进入sqlite目录,执行如下docker指令
# docker run --rm -v $(pwd):/build janboy/cgolang-g

go mod download

GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o janbar.linux

CC=x86_64-w64-mingw32-gcc-posix CXX=x86_64-w64-mingw32-g++-posix GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o janbar.exe
```

`tests/go-lua`是通过`cgo`操作`lua`，编译方法同上，

```sh
#!/bin/bash -x

# chmod +x build.sh,为脚本加上可执行权限
# 进入sqlite目录,执行如下docker指令
# docker run --rm -v $(pwd):/build janboy/cgolang-g

go mod download

cd lua-5.4.3
make clean
make CC=' gcc -std=gnu99' linux test
cd ..
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o janbar.linux

cd lua-5.4.3
make clean
make CC=' x86_64-w64-mingw32-gcc-posix -std=gnu99' mingw test
cd ..
CC=x86_64-w64-mingw32-gcc-posix CXX=x86_64-w64-mingw32-g++-posix GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o janbar.exe
```

通过复用`xgo`的`docker`环境可以实现我想要的交叉编译带上`cgo`的方法。因为目前只用到了windows和Linux这两个平台，其他平台的编译可以参考`docker/base/build.sh`脚本。

编写一份自己够用的`Dockerfile`只需要编译window和Linux两个平台就行了。基于`ubuntu16.04`是因为使用旧版`GLIBC`，如果使用`ubuntu20.04`编译出来的可执行程序在低版本Linux无法运行。

```dockerfile
FROM ubuntu:16.04

MAINTAINER janbar <janbar@163.com>

ENV GOPROXY https://goproxy.cn,direct
ENV GO111MODULE on
ENV GOPATH /go/path
ENV GOROOT /go/go
ENV PATH $GOROOT/bin:$GOPATH/bin:$PATH
ENV DEBIAN_FRONTEND=noninteractive
WORKDIR /build

RUN \
  apt-get update && \
  apt-get install -y automake autogen build-essential ca-certificates \
    gcc-mingw-w64 g++-mingw-w64 \
    libtool libxml2-dev uuid-dev libssl-dev make wget git curl \
    --no-install-recommends

ENTRYPOINT ["./build.sh"]
```

为了保险起见，还是复用大佬的base镜像，但是我只需要环境，go语言的环境则通过共享宿主机的go环境即可。

```dockerfile
FROM karalabe/xgo-base

MAINTAINER janbar <janbar@163.com>

ENV GOPROXY https://goproxy.cn,direct
ENV GO111MODULE on
ENV GOPATH /go/path
ENV GOROOT /go/go
ENV PATH $GOROOT/bin:$GOPATH/bin:$PATH

WORKDIR /build

ENTRYPOINT ["./build.sh"]
```

执行`docker build -t janbar .`编译出一个镜像。

然后执行`docker run --rm -v $(pwd):/build -v $GOROOT:/go/go janbar`，容器内部使用外部go环境，因此无需为每个版本go单独开一个docker镜像。

## 到dockerhub去操作
地址：<https://hub.docker.com/r/janboy/cgolang-g>

