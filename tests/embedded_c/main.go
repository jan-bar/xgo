// Go cross compiler (xgo): Test file for embedded C snippets.
// Copyright (c) 2015 Péter Szilágyi. All rights reserved.
//
// Released under the MIT license.

package main

/*
#include <stdio.h>
int sayHi(int argv) {
  printf("Hello, embedded C!\n");
  errno = argv; // 当函数有返回值时,赋值errno会赋值为第二个返回值error
  return argv+1;
}
*/
import "C"
import (
	"fmt"
	"syscall"
)

func main() {
	a, err := C.sayHi(C.int(0)) // errno=0时err=nil
	fmt.Printf("%d,%T,%v\n", int(a), a, err)
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok {
			fmt.Printf("%v,%d\n", err, int(errno))
		}
	}

	a, err = C.sayHi(C.int(5))
	fmt.Printf("%d,%T,%v\n", int(a), a, err)
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok {
			fmt.Printf("%v,%d\n", err, int(errno))
		}
	}
}
