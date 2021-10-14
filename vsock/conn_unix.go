// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin || linux

package vsock

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// Conn represents a vsock connection which supported half close.
type Conn interface {
	net.Conn

	FD() (f *os.File, err error)
	CloseRead() error
	CloseWrite() error
}

// conn a wrapper around net.FileConn which supports CloseRead and CloseWrite.
type conn struct {
	vsock  *os.File
	fd     uintptr
	local  *Addr
	remote *Addr
}

var _ Conn = (*conn)(nil)

// Dial connects to the cid and port via virtio socket.
func Dial(cid, port uint32) (Conn, error) {
	fd, err := newSocket()
	if err != nil {
		return nil, fmt.Errorf("create AF_VSOCK socket: %w", err)
	}

	sa := &unix.SockaddrVM{
		CID:  cid,
		Port: port,
	}

	// retry connect in a loop if EINTR is encountered.
	for {
		if err := unix.Connect(fd, sa); err != nil {
			if errno, ok := err.(unix.Errno); ok && errno == unix.EINTR {
				continue
			}
			return nil, fmt.Errorf("connect to %08x.%08x: %w", cid, port, err)
		}
		break
	}

	addr := &Addr{
		CID:  cid,
		Port: port,
	}

	return newConn(uintptr(fd), nil, addr), nil
}

// newConn returns the vsock connection.
func newConn(fd uintptr, local, remote *Addr) Conn {
	vsock := os.NewFile(fd, fmt.Sprintf("%s:%d", network, fd))

	return &conn{
		vsock:  vsock,
		fd:     fd,
		local:  local,
		remote: remote,
	}
}

// FD duplicates the underlying socket descriptor and returns it.
//
// FD implements Conn.FD.
func (c *conn) FD() (*os.File, error) {
	// this is equivalent to dup(2) but creates the new fd with CLOEXEC already set.
	fd, err := fcntl(int(c.vsock.Fd()), unix.F_DUPFD_CLOEXEC, 0)
	if err != nil {
		return nil, os.NewSyscallError("fcntl", err)
	}

	return os.NewFile(uintptr(fd), c.vsock.Name()), nil
}

// CloseRead shuts down the reading side of a vsock connection.
//
// CloseRead implements Conn.CloseRead.
func (c *conn) CloseRead() error {
	return unix.Shutdown(int(c.fd), unix.SHUT_RD)
}

// CloseWrite shuts down the writing side of a vsock connection.
//
// CloseWrite implements Conn.CloseWrite.
func (c *conn) CloseWrite() error {
	return unix.Shutdown(int(c.fd), unix.SHUT_WR)
}

// Read reads data from the connection.
//
// Read implements net.Conn.Read.
func (c *conn) Read(buf []byte) (int, error) {
	return c.vsock.Read(buf)
}

// Write writes data over the connection.
//
// Write implements net.Conn.Write.
func (c *conn) Write(buf []byte) (int, error) {
	return c.vsock.Write(buf)
}

// Close closes the connection.
//
// Close implements net.Conn.Close.
func (c *conn) Close() error {
	return c.vsock.Close()
}

// LocalAddr returns the local address of a connection.
//
// LocalAddr implements net.Conn.LocalAddr.
func (c *conn) LocalAddr() net.Addr {
	return c.local
}

// RemoteAddr returns the remote address of a connection.
//
// RemoteAddr implements net.Conn.RemoteAddr.
func (c *conn) RemoteAddr() net.Addr {
	return c.remote
}

// SetDeadline sets the read and write deadlines associated with the connection.
//
// SetDeadline implements net.Conn.SetDeadline.
func (c *conn) SetDeadline(t time.Time) error {
	return c.vsock.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls.
//
// SetReadDeadline implements net.Conn.SetReadDeadline.
func (c *conn) SetReadDeadline(t time.Time) error {
	return c.vsock.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
//
// SetWriteDeadline implements net.Conn.SetWriteDeadline.
func (c *conn) SetWriteDeadline(t time.Time) error {
	return c.vsock.SetWriteDeadline(t)
}
