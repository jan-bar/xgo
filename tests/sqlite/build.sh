#!/bin/bash -x

# chmod +x build.sh,为脚本加上可执行权限
# 进入sqlite目录,执行如下docker指令
# docker run --rm -v $(pwd):/build karalabe/xgo-1.17.1

go mod download

GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o janbar.linux

CC=x86_64-w64-mingw32-gcc-posix CXX=x86_64-w64-mingw32-g++-posix GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o janbar.exe
