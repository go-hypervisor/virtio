// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin || linux

package vsock

import (
	"fmt"
	"net"

	"golang.org/x/sys/unix"
)

type listener struct {
	fd    int
	local Addr
}

var _ net.Listener = (*listener)(nil)

// Listen returns a net.Listener which can accept connections on the given port.
func Listen(cid, port uint32) (net.Listener, error) {
	fd, err := newSocket()
	if err != nil {
		return nil, err
	}

	sa := &unix.SockaddrVM{
		CID:  cid,
		Port: port,
	}
	if err := unix.Bind(fd, sa); err != nil {
		return nil, fmt.Errorf("bind() to %08x.%08x failed: %w", cid, port, err)
	}

	if err := unix.Listen(fd, unix.SOMAXCONN); err != nil {
		return nil, fmt.Errorf("listen() on %08x.%08x failed: %w", cid, port, err)
	}

	return &listener{fd, Addr{cid, port}}, nil
}

// Accept accepts an incoming call and returns the new connection.
func (l *listener) Accept() (net.Conn, error) {
	fd, sa, err := unix.Accept(l.fd)
	if err != nil {
		return nil, err
	}

	savm := sa.(*unix.SockaddrVM)
	addr := &Addr{
		CID:  savm.CID,
		Port: savm.Port,
	}

	return newConn(uintptr(fd), &l.local, addr), nil
}

// Close closes the listening connection.
func (l *listener) Close() error {
	// Note this won't cause the Accept to unblock
	return unix.Close(l.fd)
}

// Addr returns the address the Listener is listening on.
func (l *listener) Addr() net.Addr {
	return l.local
}

// const (
// 	// Operation names which may be returned in net.OpError.
// 	opAccept      = "accept"
// 	opClose       = "close"
// 	opDial        = "dial"
// 	opListen      = "listen"
// 	opRawControl  = "raw-control"
// 	opRawRead     = "raw-read"
// 	opRawWrite    = "raw-write"
// 	opRead        = "read"
// 	opSet         = "set"
// 	opSyscallConn = "syscall-conn"
// 	opWrite       = "write"
// )

// Listen opens a connection-oriented net.Listener for incoming VM sockets
// connections. The port parameter specifies the port for the Listener.
//
// To allow the server to assign a port automatically, specify 0 for port.
// The address of the server can be retrieved using the Addr method.
//
// When the Listener is no longer needed, Close must be called to free resources.
// func Listen(port uint32) (*Listener, error) {
// 	cid, err := ContextID()
// 	if err != nil {
// 		// No addresses available.
// 		return nil, opError(opListen, err, nil, nil)
// 	}
//
// 	l, err := listen(cid, port)
// 	if err != nil {
// 		// No remote address available.
// 		return nil, opError(opListen, err, &Addr{
// 			CID:  cid,
// 			Port: port,
// 		}, nil)
// 	}
//
// 	return l, nil
// }

// // A Listener is a VM sockets implementation of a net.Listener.
// type Listener struct {
// 	l *listener
// }
//
// var _ net.Listener = &Listener{}
//
// // Accept implements the Accept method in the net.Listener interface; it waits
// // for the next call and returns a generic net.Conn. The returned net.Conn will
// // always be of type *Conn.
// func (l *Listener) Accept() (net.Conn, error) {
// 	c, err := l.l.Accept()
// 	if err != nil {
// 		return nil, l.opError(opAccept, err)
// 	}
//
// 	return c, nil
// }
//
// // Addr returns the listener's network address, a *Addr. The Addr returned is
// // shared by all invocations of Addr, so do not modify it.
// func (l *Listener) Addr() net.Addr { return l.l.Addr() }
//
// // Close stops listening on the VM sockets address. Already Accepted connections
// // are not closed.
// func (l *Listener) Close() error {
// 	return l.opError(opClose, l.l.Close())
// }
//
// // SetDeadline sets the deadline associated with the listener. A zero time value
// // disables the deadline.
// //
// // SetDeadline only works with Go 1.12+.
// func (l *Listener) SetDeadline(t time.Time) error {
// 	return l.opError(opSet, l.l.SetDeadline(t))
// }
//
// // opError is a convenience for the function opError that also passes the local
// // address of the Listener.
// func (l *Listener) opError(op string, err error) error {
// 	// No remote address for a Listener.
// 	return opError(op, err, l.Addr(), nil)
// }
//
// // opError unpacks err if possible, producing a net.OpError with the input
// // parameters in order to implement net.Conn. As a convenience, opError returns
// // nil if the input error is nil.
// func opError(op string, err error, local, remote net.Addr) error {
// 	if err == nil {
// 		return nil
// 	}
//
// 	// Unwrap inner errors from error types.
// 	//
// 	// TODO(mdlayher): errors.Cause or similar in Go 1.13.
// 	switch xerr := err.(type) {
// 	// os.PathError produced by os.File method calls.
// 	case *os.PathError:
// 		// Although we could make use of xerr.Op here, we're passing it manually
// 		// for consistency, since some of the Conn calls we are making don't
// 		// wrap an os.File, which would return an Op for us.
// 		//
// 		// As a special case, if the error is related to access to the /dev/vsock
// 		// device, we don't unwrap it, so the caller has more context as to why
// 		// their operation actually failed than "permission denied" or similar.
// 		if xerr.Path != DevVsock {
// 			err = xerr.Err
// 		}
// 	}
//
// 	switch {
// 	case err == io.EOF, errors.Is(err, unix.ENOTCONN):
// 		// We may see a literal io.EOF as happens with x/net/nettest, but
// 		// "transport not connected" also means io.EOF in Go.
// 		return io.EOF
//
// 	case err == os.ErrClosed, errors.Is(err, unix.EBADF), strings.Contains(err.Error(), "use of closed"):
// 		// Different operations may return different errors that all effectively
// 		// indicate a closed file.
// 		//
// 		// To rectify the differences, net.TCPConn uses an error with this text
// 		// from internal/poll for the backing file already being closed.
// 		err = errors.New("use of closed network connection")
//
// 	default:
// 		// Nothing to do, return this directly.
// 	}
//
// 	// Determine source and addr using the rules defined by net.OpError's
// 	// documentation: https://golang.org/pkg/net/#OpError.
// 	var source, addr net.Addr
// 	switch op {
// 	case opClose, opDial, opRawRead, opRawWrite, opRead, opWrite:
// 		if local != nil {
// 			source = local
// 		}
// 		if remote != nil {
// 			addr = remote
// 		}
// 	case opAccept, opListen, opRawControl, opSet, opSyscallConn:
// 		if local != nil {
// 			addr = local
// 		}
// 	}
//
// 	return &net.OpError{
// 		Op:     op,
// 		Net:    network,
// 		Source: source,
// 		Addr:   addr,
// 		Err:    err,
// 	}
// }

// // listener is the net.Listener implementation for connection-oriented VM sockets.
// type listener struct {
// 	fd   listenFD
// 	addr *Addr
// }
//
// var _ net.Listener = &listener{}
//
// // Addr and Close implement the net.Listener interface for listener.
// func (l *listener) Addr() net.Addr {
// 	return l.addr
// }
//
// func (l *listener) Close() error {
// 	return l.fd.Close()
// }
//
// func (l *listener) SetDeadline(t time.Time) error {
// 	return l.fd.SetDeadline(t)
// }
//
// // Accept accepts a single connection from the listener, and sets up
// // a net.Conn backed by conn.
// func (l *listener) Accept() (net.Conn, error) {
// 	// Mimic what internal/poll does and close on exec, but leave it up to
// 	// newConn to set non-blocking mode.
// 	// See: https://golang.org/src/internal/poll/sock_cloexec.go.
// 	//
// 	// TODO(mdlayher): acquire syscall.ForkLock.RLock here once the Go 1.11
// 	// code can be removed and we're fully using the runtime network poller in
// 	// non-blocking mode.
// 	cfd, sa, err := l.fd.Accept4(unix.SOCK_CLOEXEC)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	savm := sa.(*unix.SockaddrVM)
// 	remote := &Addr{
// 		CID:  savm.CID,
// 		Port: savm.Port,
// 	}
//
// 	return newConn(cfd, l.addr, remote)
// }
//
// // listen is the entry point for Listen on Linux.
// func listen(cid, port uint32) (*Listener, error) {
// 	lfd, err := newListenFD()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return listenLinux(lfd, cid, port)
// }
//
// // listenLinux is the entry point for tests on Linux.
// func listenLinux(lfd listenFD, cid, port uint32) (l *Listener, err error) {
// 	defer func() {
// 		if err != nil {
// 			// If any system calls fail during setup, the socket must be closed
// 			// to avoid file descriptor leaks.
// 			_ = lfd.EarlyClose()
// 		}
// 	}()
//
// 	// Zero-value for "any port" is friendlier in Go than a constant.
// 	if port == 0 {
// 		port = PortAny
// 	}
//
// 	sa := &unix.SockaddrVM{
// 		CID:  cid,
// 		Port: port,
// 	}
//
// 	if err := lfd.Bind(sa); err != nil {
// 		return nil, err
// 	}
//
// 	if err := lfd.Listen(unix.SOMAXCONN); err != nil {
// 		return nil, err
// 	}
//
// 	lsa, err := lfd.Getsockname()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Done with blocking mode setup, transition to non-blocking before the
// 	// caller has a chance to start calling things concurrently that might make
// 	// the locking situation tricky.
// 	//
// 	// Note: if any calls fail after this point, lfd.Close should be invoked
// 	// for cleanup because the socket is now non-blocking.
// 	if err := lfd.SetNonblocking("vsock-listen"); err != nil {
// 		return nil, err
// 	}
//
// 	lsavm := lsa.(*unix.SockaddrVM)
// 	addr := &Addr{
// 		CID:  lsavm.CID,
// 		Port: lsavm.Port,
// 	}
//
// 	return &Listener{
// 		l: &listener{
// 			fd:   lfd,
// 			addr: addr,
// 		},
// 	}, nil
// }
