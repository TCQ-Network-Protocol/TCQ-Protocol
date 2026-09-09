//  Project: TCQ Network Protocol (Thread Controlled QUIC)
//  Author: Trần Nguyên Hiền (c)
//  Major: Electronic And Communication Engineering
//  Email: trannguyenhien29085@gmail.com
//  Date: 2/3/2026
//	Apache License 2.0
//
// ----------------------------------------------------------------
//go:build linux

package self_test

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// The first sendmsg call on a new UDP socket sometimes errors on Linux.
// It's not clear why this happens.
// See https://github.com/golang/go/issues/63322.
func isPermissionError(err error) bool {
	var serr *os.SyscallError
	if errors.As(err, &serr) {
		return serr.Syscall == "sendmsg" && serr.Err == unix.EPERM
	}
	return false
}
