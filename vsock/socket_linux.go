// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

//go:build linux

package vsock

import (
	"syscall"

	"golang.org/x/sys/unix"
)

// newSocket invokes unix.Socket with the correct arguments to produce a vsock
// file descriptor.
func newSocket() (int, error) {
	// "Mirror what the standard library does when creating file
	// descriptors: avoid racing a fork/exec with the creation
	// of new file descriptors, so that child processes do not
	// inherit [socket] file descriptors unexpectedly.
	//
	// On Linux, SOCK_CLOEXEC was introduced in 2.6.27. OTOH,
	// Go supports Linux 2.6.23 and above. If we get EINVAL on
	// the first try, it may be that we are running on a kernel
	// older than 2.6.27. In that case, take syscall.ForkLock
	// and try again without SOCK_CLOEXEC.
	//
	// For a more thorough explanation, see similar work in the
	// Go tree: func sysSocket in net/sock_cloexec.go, as well
	// as the detailed comment in syscall/exec_unix.go."
	for {
		fd, err := unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
		switch err {
		case nil:
			return fd, nil
		case unix.EINTR:
			// Retry on interrupted syscalls.
			continue
		case unix.EINVAL:
			syscall.ForkLock.RLock()

			fd, err = unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM, 0)
			if err != nil {
				syscall.ForkLock.RUnlock()
				if err == unix.EINTR {
					// Retry on interrupted syscalls.
					continue
				}

				return 0, err
			}
			unix.CloseOnExec(fd)
			syscall.ForkLock.RUnlock()

			return fd, nil
		default:
			return 0, err
		}
	}
}
