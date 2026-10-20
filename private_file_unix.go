//go:build !windows

package main

import (
	"errors"
	"os"
	"syscall"
)

func restrictConfigFile(file *os.File) error { return file.Chmod(0600) }

func validateCredentialFile(path string) error {
	info, e := os.Lstat(path)
	if os.IsNotExist(e) {
		return nil
	}
	if e != nil {
		return errors.New("cannot inspect credential file")
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return errors.New("credential file must be a regular private file (chmod 600); do not use symlinks")
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && int(stat.Uid) != os.Geteuid() {
		return errors.New("credential file must belong to the current user")
	}
	return nil
}
