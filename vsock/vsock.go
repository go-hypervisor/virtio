// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

package vsock

import (
	"fmt"
	"net"
	"os"
)

// Conn represents a vsock connection which supported half close.
type Conn interface {
	net.Conn

	FD() (f *os.File, err error)
	CloseRead() error
	CloseWrite() error
}

// Addr represents a address of vsock endpoint.
//
// Addr implements net.Addr interface.
type Addr struct {
	CID  uint32
	Port uint32
}

var _ net.Addr = (*Addr)(nil)

// Network returns the network type for a Addr.
func (Addr) Network() string {
	return "vsock"
}

// String returns a string representation of a Addr.
func (a Addr) String() string {
	return fmt.Sprintf("%08x.%08x", a.CID, a.Port)
}
