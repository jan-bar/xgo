
```shell
# c扫描头文件和库位置
gcc -xc -E -v -
# c++扫描头文件和库位置
gcc -xc++ -E -v -

# 通过编译命令指定使用系统环境的sqlite库编译,使用-a时不从缓存中编译,方便观察变化
go build --tags "libsqlite3" -a
# 参考cgo.go文件,该文件指定头文件和库文件引用位置

# 自己下载sqlite源码编译生成.a文件
gcc -c sqlite3.c -lpthread -ldl -o sqlite3.o && ar -cr libsqlite3.a sqlite3.o

# 通过修改sqlite3.c文件的下面两个宏定义,然后执行时查询值,以此确定使用我指定的库编译
select SQLITE_VERSION()
select SQLITE_SOURCE_ID()

# 如果编译C代码
gcc -o test main.c libsqlite3.a -lm -ldl -lpthread -static

```
