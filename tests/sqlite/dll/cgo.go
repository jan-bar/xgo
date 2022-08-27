//go:build libsqlite3
// +build libsqlite3

package main

/*
#cgo CFLAGS: -DUSE_LIBSQLITE3 -I${SRCDIR}/sqlite
#cgo LDFLAGS: -L${SRCDIR}/sqlite
*/
import "C"

// 结合sqlite3_libsqlite3.go文件理解当前文件
