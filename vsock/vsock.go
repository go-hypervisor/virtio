// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

package vsock

import (
	"fmt"
	"net"
)

const (
	// devVsock is the location of /dev/vsock.
	// It is exposed on both the hypervisor and on virtual machines.
	devVsock = "/dev/vsock"

	// network is the vsock network.
	network = "vsock"
)

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
	return network
}

// String returns a string representation of a Addr.
func (a Addr) String() string {
	return fmt.Sprintf("%08x.%08x", a.CID, a.Port)
}

// name returns a name of file for use with os.NewFile for Addr.
func (a Addr) name() string {
	return fmt.Sprintf("%s:%s", a.Network(), a.String())
}
