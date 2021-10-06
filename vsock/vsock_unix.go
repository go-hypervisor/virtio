// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin && linux

package vsock

import (
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

	// CIDReserved reserved guest’s context ID. This must not be used.
	CIDReserved = unix.VMADDR_CID_RESERVED
)

const (
	// PortAny binds a random available port.
	PortAny = unix.VMADDR_PORT_ANY
)
