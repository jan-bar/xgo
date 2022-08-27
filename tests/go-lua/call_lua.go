package main

/***************************************************
1.下载lua：http://www.lua.org/ftp/lua-5.4.3.tar.gz
2.进入目录: cd ${SRCDIR}/lua-5.4.3
3.在Linux下编译：make linux test
4.在Windows下编译：make mingw test
5.按照下面的cgo使用方法通过go调用lua
6.注意编译好的[libluaLinux.a,libluaWin.a]已经随项目提交,并且在LDFLAGS中配置好
    后续编译可直接用.a文件,无需重复编译lua源码
****************************************************/

/*
#cgo CFLAGS: -I${SRCDIR}/lua-5.4.3/src
#cgo linux LDFLAGS: -L${SRCDIR}/lua-5.4.3/src -lluaLinux -lm -ldl
#cgo windows LDFLAGS: -L${SRCDIR}/lua-5.4.3/src -lluaWin -lm
#include <stdlib.h>
#include <string.h>
#include <errno.h>
#include "lua.h"
#include "lauxlib.h"
#include "lstate.h"
#include "lualib.h"
#include "lundump.h"

static int luac_writer(lua_State *L, const void *b, size_t size, void *ud) {
    UNUSED(L);
    luaL_addlstring((luaL_Buffer *)ud, (const char *)b, size);
    return 0;
}

#define FUNCTION "(function()end)();"
static const char * reader(lua_State * L, void * ud, size_t * size) {
    UNUSED(L);
    if ((*(int *)ud)--) {
        *size = sizeof(FUNCTION) - 1;
        return FUNCTION;
    }
    *size = 0;
    return NULL;
}

#define toproto(L, i) getproto(s2v((L->top)+(i)))
static const Proto *combine(lua_State *L, int n) {
    if (n == 1)
        return toproto(L, -1);

    int i = n;
    if (lua_load(L, reader, &i, "=(luac)", NULL) != LUA_OK)
        return NULL;

    Proto *f = toproto(L, -1);
    for (i = 0; i < n; i++) {
        f->p[i] = toproto(L, i-n-1);
        if (f->p[i]->sizeupvalues > 0)
            f->p[i]->upvalues[0].instack = 0;
    }
    f->sizelineinfo = 0;
    return f;
}
// 上面几个方法都是抄luac.c源码,只是将其中写入文件操作改为写入lua字符串对象
// 最终生成字节码文件数据,返回到Go中使用
static const char *dump_lua_code(int argc, char *argv[], int strip, size_t *outlen) {
    errno = 0; // 默认错误码为0

    lua_State * L = luaL_newstate();
    if (NULL == L) {
        errno = 1;
        return NULL;
    }
    if (!lua_checkstack(L, argc)) {
        errno = 2;
        return NULL;
    }

    int i;
    for (i = 0; i < argc; i++) {
        if (luaL_loadstring(L, argv[i]) != LUA_OK) {
            errno = 3;
            return NULL;
        }
    }

    const Proto *f = combine(L, argc);
    if (f == NULL) {
        errno = 4;
        return NULL;
	}

    luaL_Buffer buf; // 利用lua源码的字符串实现来保存字节码
    luaL_buffinit(L, &buf);

    lua_lock(L);
    if (luaU_dump(L, f, luac_writer, &buf, strip) != LUA_OK) {
        errno = 5;
        return NULL;
    }
    lua_unlock(L);

    luaL_pushresult(&buf);
    const char *dump_code = lua_tolstring(L, -1, outlen);
    lua_close(L);
    return dump_code;
}
*/
import "C"
import (
	"errors"
	"unsafe"
)

/*
func C.CString(string) *C.char
func C.CBytes([]byte) unsafe.Pointer
defer C.free(unsafe.Pointer(cStr))
上面两个将go对象转换为C内存数据,需要显示free

下面是将C数据转换为Go对象
func C.GoString(*C.char) string
func C.GoStringN(*C.char, C.int) string
func C.GoBytes(unsafe.Pointer, C.int) []byte
*/

const (
	LuaTTable    = C.LUA_TTABLE
	LuaTFunction = C.LUA_TFUNCTION
)

func DumpLuaCode(strip bool, script ...string) ([]byte, error) {
	argc := len(script)
	if argc == 0 {
		return nil, errors.New("no script")
	}

	// cmd/cgo/doc.go#256,去这个文件学习cgo
	var (
		cStr = make([]*C.char, argc)
		size C.size_t
		i    int
	)
	for i = 0; i < argc; i++ {
		cStr[i] = C.CString(script[i])
	}
	defer func() {
		for i = 0; i < argc; i++ {
			C.free(unsafe.Pointer(cStr[i]))
		}
	}()

	stripInt := C.int(0)
	if strip {
		stripInt = C.int(1)
	}
	res, err := C.dump_lua_code(C.int(argc), &cStr[0], stripInt, &size)
	if err != nil {
		// errno, ok := err.(syscall.Errno); ok // 第二个返回值来源errno
		return nil, err
	}

	// C中的[const char *]数据用如下方式转换到Go中,经过查看多方代码,发现都没有手动free内存
	// 而且一旦执行[defer C.free(unsafe.Pointer(res))]释放内存会导致程序异常退出
	// 所以该问题需要长时间运行观察观察,目前先按这种方式不手动释放内存使用
	data := C.GoBytes(unsafe.Pointer(res), C.int(size))
	if string(data[:4]) != "\x1bLua" {
		return nil, errors.New("not luac file")
	}
	return data, nil
}

type LuaState struct {
	s *C.lua_State
}

func NewLuaState() (*LuaState, error) {
	ls := C.luaL_newstate()
	if ls == nil {
		return nil, errors.New("new lua is error")
	}
	return &LuaState{s: ls}, nil
}

func (L *LuaState) OpenLibs() {
	C.luaL_openlibs(L.s)
}

func (L *LuaState) PCall(nArgs, nResults int) error {
	ctx := C.lua_KContext(0)
	if 0 != int(C.lua_pcallk(L.s, C.int(nArgs), C.int(nResults), 0, ctx, nil)) {
		return errors.New(L.ToString(-1))
	}
	return nil
}

func (L *LuaState) DoFile(filename string) error {
	cLuaFile := C.CString(filename)
	defer C.free(unsafe.Pointer(cLuaFile))
	if 0 != int(C.luaL_loadfilex(L.s, cLuaFile, nil)) {
		return errors.New(L.ToString(-1))
	}
	return L.PCall(0, int(C.LUA_MULTRET))
}

func (L *LuaState) DoString(script string) error {
	cScript := C.CString(script)
	defer C.free(unsafe.Pointer(cScript))
	if 0 != int(C.luaL_loadstring(L.s, cScript)) {
		return errors.New(L.ToString(-1))
	}
	return L.PCall(0, int(C.LUA_MULTRET))
}

func (L *LuaState) GetGlobal(name string) {
	cName := C.CString(name)
	C.lua_getglobal(L.s, cName)
	C.free(unsafe.Pointer(cName))
}

func (L *LuaState) LuaType(index int, lType C.int) bool {
	return C.lua_type(L.s, C.int(index)) == lType
}

func (L *LuaState) IsString(index int) bool {
	return C.lua_isstring(L.s, C.int(index)) == 1
}

func (L *LuaState) PushByte(b byte) {
	C.lua_pushinteger(L.s, C.lua_Integer(b))
}

func (L *LuaState) GetTable(index int) {
	C.lua_gettable(L.s, C.int(index))
}

func (L *LuaState) PushBytes(b []byte) {
	C.lua_pushlstring(L.s, (*C.char)(unsafe.Pointer(&b[0])), C.size_t(len(b)))
}

func (L *LuaState) Pop(n int) {
	C.lua_settop(L.s, C.int(-n-1))
}

func (L *LuaState) ToString(index int) string {
	// https://github.com/aarzilli/golua/blob/11106aa577653365582edb61e6cb9f7edeb81eed/lua/lua.go#L503
	var size C.size_t
	r := C.lua_tolstring(L.s, C.int(index), &size)
	return C.GoStringN(r, C.int(size))
}

func (L *LuaState) ToBytes(index int) []byte {
	// https://github.com/aarzilli/golua/blob/11106aa577653365582edb61e6cb9f7edeb81eed/lua/lua.go#L509
	var size C.size_t
	b := C.lua_tolstring(L.s, C.int(index), &size)
	return C.GoBytes(unsafe.Pointer(b), C.int(size))
}

func (L *LuaState) Close() {
	C.lua_close(L.s)
}
