// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin

package vsock

import (
	"syscall"

	"golang.org/x/sys/unix"
)

// newSocket invokes unix.Socket with the correct arguments to produce a vsock
// file descriptor.
func newSocket() (fd int, err error) {
	for {
		syscall.ForkLock.RLock()

		fd, err = unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM, 0)
		switch err {
		case nil:
			// set FD_CLOEXEC to fd
			unix.CloseOnExec(fd)
			syscall.ForkLock.RUnlock()

			return fd, nil

		case unix.EINTR:
			syscall.ForkLock.RUnlock()
			// retry on interrupted syscalls
			continue

		default:
			syscall.ForkLock.RUnlock()
			return 0, err
		}
	}
}
