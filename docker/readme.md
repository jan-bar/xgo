我的`dockerfile`，主要生成带交叉编译工具链的docker镜像。

```dockerfile
FROM karalabe/xgo-base

MAINTAINER janbar <janbar@163.com>

ENV GO_VERSION 11701
ENV GOPROXY https://goproxy.cn,direct
ENV GO111MODULE on
WORKDIR /build # 指定默认目录,用于映射编译路径

# 由于后续版本不支持darwin/386,所以用sed去掉
# 我的编译脚本不使用xgo因此无需安装xgo
RUN \
  export ROOT_DIST=https://studygolang.com/dl/golang/go1.17.1.linux-amd64.tar.gz        && \
  export ROOT_DIST_SHA=dab7d9c34361dc21ec237d584590d72500652e7c909bf082758fb63064fca0ef && \
  sed -i '60,$d' $BOOTSTRAP_PURE && \
  $BOOTSTRAP_PURE
# 在默认目录/build中包含build.sh脚本
ENTRYPOINT ["./build.sh"]
```

编译指定版本的docker镜像

```sh
docker build -t karalabe/xgo-1.17.1 ./go-1.17.1
```

提供两个使用`CGO`交叉编译的示例：

`tests/sqlite`是通过`cgo`操作`sqlite`数据库，将整个目录放到服务器上，然后按照如下脚本执行

```sh
#!/bin/bash -x

# chmod +x build.sh,为脚本加上可执行权限
# 进入sqlite目录,执行如下docker指令
# docker run --rm -v $(pwd):/build karalabe/xgo-1.17.1

go mod download

GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o janbar.linux

CC=x86_64-w64-mingw32-gcc-posix CXX=x86_64-w64-mingw32-g++-posix GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o janbar.exe
```

`tests/go-lua`是通过`cgo`操作`lua`，编译方法同上，

```sh
#!/bin/bash -x

# chmod +x build.sh,为脚本加上可执行权限
# 进入sqlite目录,执行如下docker指令
# docker run --rm -v $(pwd):/build karalabe/xgo-1.17.1

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

执行`docker build -t janbar .`编译出一个镜像。

然后执行`docker run --rm -v $(pwd):/build -v $GOROOT:/go/go janbar`，容器内部使用外部go环境，因此无需为每个版本go单独开一个docker镜像。