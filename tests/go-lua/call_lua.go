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
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
#include <stdlib.h>
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
	if 0 != C.lua_pcallk(L.s, C.int(nArgs), C.int(nResults), 0, ctx, nil) {
		return errors.New(L.ToString(-1))
	}
	return nil
}

func (L *LuaState) DoFile(filename string) error {
	cLuaFile := C.CString(filename)
	defer C.free(unsafe.Pointer(cLuaFile))
	if 0 != C.luaL_loadfilex(L.s, cLuaFile, nil) {
		return errors.New(L.ToString(-1))
	}
	return L.PCall(0, int(C.LUA_MULTRET))
}

func (L *LuaState) DoString(script string) error {
	cScript := C.CString(script)
	defer C.free(unsafe.Pointer(cScript))
	if 0 != C.luaL_loadstring(L.s, cScript) {
		return errors.New(L.ToString(-1))
	}
	return L.PCall(0, int(C.LUA_MULTRET))
}

func (L *LuaState) GetGlobal(name string) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	C.lua_getglobal(L.s, cName)
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
