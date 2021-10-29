package main

/***************************************************
1.下载lua：http://www.lua.org/ftp/lua-5.4.3.tar.gz
2.进入目录: cd ${SRCDIR}/lua-5.4.3
3.在Linux下编译：make linux test
4.在Windows下编译：make mingw test
5.按照下面的cgo使用方法通过go调用lua
****************************************************/

/*
#cgo CFLAGS: -I ${SRCDIR}/lua-5.4.3/src
#cgo linux LDFLAGS: -L ${SRCDIR}/lua-5.4.3/src -llua -lm -ldl
#cgo windows LDFLAGS: -L ${SRCDIR}/lua-5.4.3/src -llua -lm
#include <stdlib.h>
#include <string.h>
#include "lua.h"
#include "lauxlib.h"
#include "lstate.h"
#include "lualib.h"
#include "lundump.h"

static const char *fatal(const char *msg, lua_State *L, size_t *outlen) {
    luaL_Buffer buf;
    luaL_buffinit(L, &buf);
    luaL_addstring(&buf, msg);
    luaL_pushresult(&buf);
    return lua_tolstring(L, -1, outlen);
}

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

static const char *dump_lua_code(int argc, char *argv[], int strip, size_t *outlen) {
    lua_State * L = luaL_newstate();
    if (NULL == L) {
        const char *s = "cannot create state: not enough memory";
        *outlen = strlen(s);
        return s;
    }
    if (!lua_checkstack(L, argc))
        return fatal("too many input files", L, outlen);

    int i;
    for (i = 0; i < argc; i++) {
        if (luaL_loadstring(L, argv[i]) != LUA_OK)
            return fatal(lua_tostring(L, -1), L, outlen);
    }

    const Proto *f = combine(L, argc);
    if (f == NULL)
        return fatal(lua_tostring(L, -1), L, outlen);

    luaL_Buffer buf;
    luaL_buffinit(L, &buf);

    lua_lock(L);
    if (luaU_dump(L, f, luac_writer, &buf, strip) != LUA_OK)
        return fatal("unable to dump lua code", L, outlen);
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
	res := C.dump_lua_code(C.int(argc), &cStr[0], stripInt, &size)

	data := C.GoBytes(unsafe.Pointer(res), C.int(size))
	if data[0] != 0x1b || data[1] != 'L' || data[2] != 'u' || data[3] != 'a' {
		return nil, errors.New(string(data)) // 不是头部,则返回错误字符串
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
	var size C.size_t
	r := C.lua_tolstring(L.s, C.int(index), &size)
	return C.GoStringN(r, C.int(size))
}

func (L *LuaState) ToBytes(index int) []byte {
	var size C.size_t
	b := C.lua_tolstring(L.s, C.int(index), &size)
	return C.GoBytes(unsafe.Pointer(b), C.int(size))
}

func (L *LuaState) Close() {
	C.lua_close(L.s)
}
