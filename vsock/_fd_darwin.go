// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin

package vsock

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// contextID retrieves the local context ID for this system.
func contextID() (uint32, error) {
	f, err := os.Open(DevVsock)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	v, err := unix.IoctlGetInt(int(f.Fd()), unix.IOCTL_VM_SOCKETS_GET_LOCAL_CID)
	return uint32(v), err
}

// ContextID retrieves the local VM sockets context ID for this system.
// ContextID can be used to directly determine if a system is capable of using
// VM sockets.
//
// If the kernel module is unavailable, access to the kernel module is denied,
// or VM sockets are unsupported on this system, it returns an error.
func ContextID() (uint32, error) {
	return contextID()
}

// A listenFD is a type that wraps a file descriptor used to implement
// net.Listener.
type listenFD interface {
	io.Closer

	EarlyClose() error
	Accept4(flags int) (connFD, unix.Sockaddr, error)
	Bind(sa unix.Sockaddr) error
	Listen(n int) error
	Getsockname() (unix.Sockaddr, error)
	SetNonblocking(name string) error
	SetDeadline(t time.Time) error
}

// A sysListenFD is the system call implementation of listenFD.
type sysListenFD struct {
	// These fields should never be non-zero at the same time.
	fd int      // Used in blocking mode.
	f  *os.File // Used in non-blocking mode.
}

var _ listenFD = &sysListenFD{}

// newListenFD creates a sysListenFD in its default blocking mode.
func newListenFD() (*sysListenFD, error) {
	fd, err := NewSocket()
	if err != nil {
		return nil, err
	}

	return &sysListenFD{
		fd: fd,
	}, nil
}

// Blocking mode methods.

func (lfd *sysListenFD) Bind(sa unix.Sockaddr) error {
	return unix.Bind(lfd.fd, sa)
}

func (lfd *sysListenFD) Getsockname() (unix.Sockaddr, error) {
	return unix.Getsockname(lfd.fd)
}

func (lfd *sysListenFD) Listen(n int) error {
	return unix.Listen(lfd.fd, n)
}

func (lfd *sysListenFD) SetNonblocking(name string) error {
	return lfd.setNonblocking(name)
}

// EarlyClose is a blocking version of Close, only used for cleanup before
// entering non-blocking mode.
func (lfd *sysListenFD) EarlyClose() error {
	return unix.Close(lfd.fd)
}

// Non-blocking mode methods.

func (lfd *sysListenFD) Accept4(flags int) (connFD, unix.Sockaddr, error) {
	// Invoke Go version-specific logic for accept.
	newFD, sa, err := lfd.accept4(flags)
	if err != nil {
		return nil, nil, err
	}

	// Create a non-blocking connFD which will be used to implement net.Conn.
	cfd := &sysConnFD{fd: newFD}
	return cfd, sa, nil
}

func (lfd *sysListenFD) Close() error {
	// In Go 1.12+, *os.File.Close will also close the runtime network poller
	// file descriptor, so that net.Listener.Accept can stop blocking.
	return lfd.f.Close()
}

func (lfd *sysListenFD) SetDeadline(t time.Time) error {
	// Invoke Go version-specific logic for setDeadline.
	return lfd.setDeadline(t)
}

// A connFD is a type that wraps a file descriptor used to implement net.Conn.
type connFD interface {
	io.ReadWriteCloser

	EarlyClose() error
	Connect(sa unix.Sockaddr) error
	Getsockname() (unix.Sockaddr, error)
	Shutdown(how int) error
	SetNonblocking(name string) error
	SetDeadline(t time.Time, typ deadlineType) error
	SyscallConn() (syscall.RawConn, error)
}

var _ connFD = &sysConnFD{}

// newConnFD creates a sysConnFD in its default blocking mode.
func newConnFD() (*sysConnFD, error) {
	fd, err := NewSocket()
	if err != nil {
		return nil, err
	}

	return &sysConnFD{
		fd: fd,
	}, nil
}

// A sysConnFD is the system call implementation of connFD.
type sysConnFD struct {
	// These fields should never be non-zero at the same time.
	fd int      // Used in blocking mode.
	f  *os.File // Used in non-blocking mode.
}

// Blocking mode methods.

func (cfd *sysConnFD) Getsockname() (unix.Sockaddr, error) {
	return unix.Getsockname(cfd.fd)
}

func (cfd *sysConnFD) Connect(sa unix.Sockaddr) error {
	var err error

	for {
		err = unix.Connect(cfd.fd, sa)
		if errors.Is(err, unix.EINTR) {
			// Retry on interrupted syscalls.
			continue
		}
		break
	}

	return err
}

// EarlyClose is a blocking version of Close, only used for cleanup before
// entering non-blocking mode.
func (cfd *sysConnFD) EarlyClose() error {
	return unix.Close(cfd.fd)
}

func (cfd *sysConnFD) SetNonblocking(name string) error {
	return cfd.setNonblocking(name)
}

// Non-blocking mode methods.

func (cfd *sysConnFD) Close() error {
	// *os.File.Close will also close the runtime network poller file descriptor,
	// so that read/write can stop blocking.
	return cfd.f.Close()
}

func (cfd *sysConnFD) Read(b []byte) (int, error) {
	return cfd.f.Read(b)
}

func (cfd *sysConnFD) Write(b []byte) (int, error) {
	return cfd.f.Write(b)
}

func (cfd *sysConnFD) Shutdown(how int) error {
	switch how {
	case unix.SHUT_RD, unix.SHUT_WR:
		return cfd.shutdown(how)
	default:
		panic(fmt.Sprintf("vsock: sysConnFD.Shutdown method invoked with invalid how constant: %d", how))
	}
}

func (cfd *sysConnFD) SetDeadline(t time.Time, typ deadlineType) error {
	return cfd.setDeadline(t, typ)
}

func (cfd *sysConnFD) SyscallConn() (syscall.RawConn, error) {
	return cfd.syscallConn()
}
