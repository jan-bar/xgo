package main

import (
	"os"
	"syscall"
)

func main() {
	L, err := NewLuaState()
	if err != nil {
		panic(err)
	}
	defer L.Close()
	L.OpenLibs()

	script := `
function test(a, b, c, ...)
    print(a, b, c, ...)
end

print(os.time())
test(123,'456',true,666,777)
`
	if err = L.DoString(script); err != nil {
		panic(err)
	}

	code, err := DumpLuaCode(true, script, "a b c")
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok {
			// 编码lua代码报错, 用返回值捕获C里面的错误
			println("DumpLuaCode errno",int(errno))
		} else {
			println("DumpLuaCode", err.Error())
		}
	}

	code, err = DumpLuaCode(true, "a=math.pow(2,10)",
		script, "print('end', a)", "local b=a;print(type(b))")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile("luac.out", code, 0666)
	if err != nil {
		panic(err)
	}
}
