// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin

package vsock

import (
	_ "unsafe" // for go:linkname

	"golang.org/x/sys/unix"
)

// TODO(zchee): use internal/poll instead of.

//go:linkname libc_poll golang.org/x/sys/unix.poll
func libc_poll(fds *unix.PollFd, nfds int, timeout int) (n int, err error)

func poll(fds *unix.PollFd) (n int, err error) {
	return libc_poll(fds, 1, 0)
}
