// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

//go:build linux

package vsock

import (
	"golang.org/x/sys/unix"
)

// TODO(zchee): use internal/poll instead of.

func poll(fds *unix.PollFd) (n int, err error) {
	return unix.Ppoll([]unix.PollFd{*fds}, &unix.Timespec{}, &unix.Sigset_t{})
}
