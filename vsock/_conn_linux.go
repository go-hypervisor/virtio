// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

//go:build linux

package vsock

import (
	"net"
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// A Conn is a VM sockets implementation of a net.Conn.
type connLinux struct {
	fd     connFD
	local  *Addr
	remote *Addr
}

// FD duplicates the underlying socket descriptor and returns it.
//
// FD implements Conn.FD.
func (v *connLinux) FD() (*os.File, error) {
	return nil, nil
	// 	// this is equivalent to dup(2) but creates the new fd with CLOEXEC already set.
	// 	fd, err := fcntl(int(v.fd), unix.F_DUPFD_CLOEXEC, 0)
	// 	if err != nil {
	// 		return nil, os.NewSyscallError("fcntl", err)
	// 	}
	//
	// 	return os.NewFile(uintptr(fd), v.fd.Name()), nil
}

// Close closes the connection.
func (c *connLinux) Close() error {
	return c.opError(opClose, c.fd.Close())
}

// CloseRead shuts down the reading side of the VM sockets connection. Most
// callers should just use Close.
//
// CloseRead only works with Go 1.12+.
func (c *connLinux) CloseRead() error {
	return c.opError(opClose, c.fd.Shutdown(unix.SHUT_RD))
}

// CloseWrite shuts down the writing side of the VM sockets connection. Most
// callers should just use Close.
//
// CloseWrite only works with Go 1.12+.
func (c *connLinux) CloseWrite() error {
	return c.opError(opClose, c.fd.Shutdown(unix.SHUT_WR))
}

// LocalAddr returns the local network address. The Addr returned is shared by
// all invocations of LocalAddr, so do not modify it.
func (c *connLinux) LocalAddr() net.Addr { return c.local }

// RemoteAddr returns the remote network address. The Addr returned is shared by
// all invocations of RemoteAddr, so do not modify it.
func (c *connLinux) RemoteAddr() net.Addr { return c.remote }

// Read implements the net.Conn Read method.
func (c *connLinux) Read(b []byte) (int, error) {
	n, err := c.fd.Read(b)
	if err != nil {
		return n, c.opError(opRead, err)
	}

	return n, nil
}

// Write implements the net.Conn Write method.
func (c *connLinux) Write(b []byte) (int, error) {
	n, err := c.fd.Write(b)
	if err != nil {
		return n, c.opError(opWrite, err)
	}

	return n, nil
}

// SetDeadline implements the net.Conn SetDeadline method.
func (c *connLinux) SetDeadline(t time.Time) error {
	return c.opError(opSet, c.fd.SetDeadline(t, deadline))
}

// SetReadDeadline implements the net.Conn SetReadDeadline method.
func (c *connLinux) SetReadDeadline(t time.Time) error {
	return c.opError(opSet, c.fd.SetDeadline(t, readDeadline))
}

// SetWriteDeadline implements the net.Conn SetWriteDeadline method.
func (c *connLinux) SetWriteDeadline(t time.Time) error {
	return c.opError(opSet, c.fd.SetDeadline(t, writeDeadline))
}

// SyscallConn returns a raw network connection. This implements the
// syscall.Conn interface.
func (c *connLinux) SyscallConn() (syscall.RawConn, error) {
	rc, err := c.fd.SyscallConn()
	if err != nil {
		return nil, c.opError(opSyscallConn, err)
	}

	return &rawConn{
		rc:     rc,
		local:  c.local,
		remote: c.remote,
	}, nil
}

// opError is a convenience for the function opError that also passes the local
// and remote addresses of the Conn.
func (c *connLinux) opError(op string, err error) error {
	return opError(op, err, c.local, c.remote)
}

// A rawConn is a syscall.RawConn that wraps an internal syscall.RawConn in order
// to produce net.OpError error values.
type rawConn struct {
	rc     syscall.RawConn
	local  *Addr
	remote *Addr
}

var _ syscall.RawConn = &rawConn{}

// Control implements the syscall.RawConn Control method.
func (rc *rawConn) Control(fn func(fd uintptr)) error {
	return rc.opError(opRawControl, rc.rc.Control(fn))
}

// Control implements the syscall.RawConn Read method.
func (rc *rawConn) Read(fn func(fd uintptr) (done bool)) error {
	return rc.opError(opRawRead, rc.rc.Read(fn))
}

// Control implements the syscall.RawConn Write method.
func (rc *rawConn) Write(fn func(fd uintptr) (done bool)) error {
	return rc.opError(opRawWrite, rc.rc.Write(fn))
}

// opError is a convenience for the function opError that also passes the local
// and remote addresses of the rawConn.
func (rc *rawConn) opError(op string, err error) error {
	return opError(op, err, rc.local, rc.remote)
}

// newConn creates a Conn using a connFD, immediately setting the connFD to
// non-blocking mode for use with the runtime network poller.
func newConn(cfd connFD, local, remote *Addr) (Conn, error) {
	// Note: if any calls fail after this point, cfd.Close should be invoked
	// for cleanup because the socket is now non-blocking.
	if err := cfd.SetNonblocking(local.name()); err != nil {
		return nil, err
	}

	return &connLinux{
		fd:     cfd,
		local:  local,
		remote: remote,
	}, nil
}

// dial is the entry point for Dial on Linux.
func dial(cid, port uint32) (Conn, error) {
	cfd, err := newConnFD()
	if err != nil {
		return nil, err
	}

	return dialLinux(cfd, cid, port)
}

// dialLinux is the entry point for tests on Linux.
func dialLinux(cfd connFD, cid, port uint32) (c Conn, err error) {
	defer func() {
		if err != nil {
			// If any system calls fail during setup, the socket must be closed
			// to avoid file descriptor leaks.
			_ = cfd.EarlyClose()
		}
	}()

	rsa := &unix.SockaddrVM{
		CID:  cid,
		Port: port,
	}

	if err := cfd.Connect(rsa); err != nil {
		return nil, err
	}

	lsa, err := cfd.Getsockname()
	if err != nil {
		return nil, err
	}

	lsavm := lsa.(*unix.SockaddrVM)

	local := &Addr{
		CID:  lsavm.CID,
		Port: lsavm.Port,
	}

	remote := &Addr{
		CID:  cid,
		Port: port,
	}

	return newConn(cfd, local, remote)
}
