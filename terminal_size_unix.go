//go:build darwin || linux

package main

import (
	"os"
	"syscall"
	"unsafe"
)

func terminalColumns(file *os.File) int {
	var size struct{ Rows, Columns, X, Y uint16 }
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, file.Fd(), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&size)))
	if err != 0 {
		return 0
	}
	return int(size.Columns)
}
