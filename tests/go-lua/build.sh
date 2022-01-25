#!/bin/bash -x

# chmod +x build.sh,为脚本加上可执行权限
# 进入sqlite目录,执行如下docker指令
# docker run --rm -v $(pwd):/build -v $GOROOT:/go/go janboy/cgolang-g

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
