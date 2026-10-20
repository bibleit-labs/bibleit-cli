package main

import (
	"os"
	"syscall"
	"unsafe"
)

var getConsoleScreenBufferInfo = syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleScreenBufferInfo")

func terminalColumns(file *os.File) int {
	var info struct {
		Size, Cursor struct{ X, Y int16 }
		Attributes   uint16
		Window       struct{ Left, Top, Right, Bottom int16 }
		MaximumSize  struct{ X, Y int16 }
	}
	ok, _, _ := getConsoleScreenBufferInfo.Call(file.Fd(), uintptr(unsafe.Pointer(&info)))
	if ok == 0 {
		return 0
	}
	return int(info.Window.Right-info.Window.Left) + 1
}
