// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin || linux

package vsock

import "io"

// Socket is a connected vsock socket.
type Socket interface {
	io.ReadWriteCloser

	FD() int
	Release() (int, error)
	Shutdown() error
}
