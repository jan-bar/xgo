
```shell
# 直接编译该可执行程序依赖sqlite.dll
go build --tags "libsqlite3"

# 后续该可执行程序调用sqlite方式就动态依赖dll
ldd dll.exe 
    ntdll.dll => /c/Windows/SYSTEM32/ntdll.dll (0x7ffb66d60000)
    KERNEL32.DLL => /c/Windows/System32/KERNEL32.DLL (0x7ffb64360000)
    KERNELBASE.dll => /c/Windows/System32/KERNELBASE.dll (0x7ffb637f0000)
    msvcrt.dll => /c/Windows/System32/msvcrt.dll (0x7ffb645f0000)
    sqlite3.dll => ./sqlite/sqlite3.dll (0x7ffb2cb80000)

# 同理可得，linux下面依赖.so也是类似情况，不过一般都是静态编译到可执行程序中会方便些
```
