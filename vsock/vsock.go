// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

package vsock

import (
	"fmt"
	"net"
	"os"

	"golang.org/x/sys/unix"
)

// list of context ID.
const (
	// CIDAny binds any guest’s context ID. This seems to work inside VMs only.
	CIDAny = unix.VMADDR_CID_ANY

	// CIDHost well-known host context ID.
	CIDHost = unix.VMADDR_CID_HOST

	// CIDHypervisor reserved hypervisor context ID.
	CIDHypervisor = unix.VMADDR_CID_HYPERVISOR

	// CIDReserved reserved guest’s context ID. This must not be used
	CIDReserved = unix.VMADDR_CID_RESERVED
)

const (
	// PortAny binds a random available port.
	PortAny = unix.VMADDR_PORT_ANY
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
