// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin || linux

package vsock

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

// recvmsg issues a recvmsg unix.
//go:linkname recvmsg golang.org/x/sys/unix.recvmsg
func recvmsg(s int, msg *unix.Msghdr, flags int) (n int, err error)

// sendmsg issues a sendmsg unix.
//go:linkname sendmsg golang.org/x/sys/unix.sendmsg
func sendmsg(s int, msg *unix.Msghdr, flags int) (n int, err error)

type socklen uint32

// getsockopt issues a getsockopt unix.
//go:linkname getsockopt golang.org/x/sys/unix.getsockopt
func getsockopt(s int, level int, name int, val unsafe.Pointer, vallen *socklen) (err error)

// setsockopt issues a setsockopt unix.
//go:linkname setsockopt golang.org/x/sys/unix.setsockopt
func setsockopt(s int, level int, name int, val unsafe.Pointer, vallen uintptr) (err error)

// getpeername issues a getpeername unix.
//go:linkname getpeername golang.org/x/sys/unix.getpeername
func getpeername(fd int, rsa *unix.RawSockaddrAny, addrlen *socklen) (err error)

// getsockname issues a getsockname unix.
//go:linkname getsockname golang.org/x/sys/unix.getsockname
func getsockname(fd int, rsa *unix.RawSockaddrAny, addrlen *socklen) (err error)

//go:linkname fcntl golang.org/x/sys/unix.fcntl
func fcntl(fd int, cmd, arg int) (val int, err error)
