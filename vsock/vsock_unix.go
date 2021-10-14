// Copyright 2021 The Go Hypervisor Authors
// SPDX-License-Identifier: BSD-3-Clause

//go:build darwin || linux

package vsock

import (
	"golang.org/x/sys/unix"
)

// list of context ID.
const (
	// VMAddrCIDAny binds any guest’s context ID. This seems to work inside VMs only.
	VMAddrCIDAny = unix.VMADDR_CID_ANY

	// VMAddrCIDHost well-known host context ID.
	VMAddrCIDHost = unix.VMADDR_CID_HOST

	// VMAddrCIDHypervisor reserved hypervisor context ID.
	VMAddrCIDHypervisor = unix.VMADDR_CID_HYPERVISOR

	// VMAddrCIDReserved reserved guest’s context ID. This must not be used.
	VMAddrCIDReserved = unix.VMADDR_CID_RESERVED
)

const (
	// VMAddrPortAny binds a random available port.
	VMAddrPortAny = unix.VMADDR_PORT_ANY
)

// ErrClosing is returned by wait if the Socket is in the process of closing.
// var ErrClosing = errors.New("Socket is closing")

// ErrMessageTruncated indicates that data was lost because the provided buffer
// was too small.
// var ErrMessageTruncated = errors.New("message truncated")

// // A deadlineType specifies the type of deadline to set for a Conn.
// type deadlineType int
//
// // Possible deadlineType values.
// const (
// 	deadline deadlineType = iota
// 	readDeadline
// 	writeDeadline
// )
//
// // socket is a connected unix domain socket.
// type socket struct {
// 	// fd is the bound socket.
// 	//
// 	// fd must be read atomically, and only remains valid if read while
// 	// within gate.
// 	fd int32
//
// 	// race is an atomic variable used to avoid triggering the race
// 	// detector.
// 	race *int32
// }
//
// var _ Socket = (*socket)(nil)
//
// // vsocket creates a new host vsock socket.
// func vsocket() (int, error) {
// 	fd, err := unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM, 0)
// 	if err != nil {
// 		return 0, err
// 	}
//
// 	return fd, nil
// }
//
// // NewSocketFromFD returns a socket from an existing fd.
// //
// // NewSocketFromFD takes ownership of fd.
// func NewSocketFromFD(fd int) (Socket, error) {
// 	// fd must be non-blocking for non-blocking unix.Accept in
// 	// ServerSocket.Accept.
// 	if err := unix.SetNonblock(fd, true); err != nil {
// 		return nil, err
// 	}
//
// 	return &socket{
// 		fd: int32(fd),
// 	}, nil
// }
//
// // FD returns the FD for this Socket.
// //
// // The FD is non-blocking and must not be made blocking.
// //
// // N.B. os.File.Fd makes the FD blocking. Use of Release instead of FD is
// // strongly preferred.
// //
// // The returned FD cannot be used safely if there may be concurrent callers to
// // Close or Release.
// //
// // Use Release to take ownership of the FD.
// //
// // FD implements Socket.FD.
// func (s *socket) FD() int {
// 	return int(atomic.LoadInt32(&s.fd))
// }
//
// // Release releases ownership of the socket FD.
// //
// // The returned FD is non-blocking.
// //
// // Any concurrent or future callers of Socket methods will receive EBADF.
// //
// // Release implements Socket.Release.
// func (s *socket) Release() (int, error) {
// 	// Set the FD in the socket to -1, to ensure that all future calls to
// 	// FD/Release get nothing and Close calls return immediately.
// 	fd := int(atomic.SwapInt32(&s.fd, -1))
// 	if fd < 0 {
// 		// Already closed or closing.
// 		return -1, unix.EBADF
// 	}
//
// 	return fd, nil
// }
//
// // enterFD enters the FD gate and returns the FD value.
// //
// // If enterFD returns ok, s.gate.Leave must be called when done with the FD.
// // Callers may only block while within the gate using s.wait.
// //
// // The returned FD is guaranteed to remain valid until s.gate.Leave.
// func (s *socket) enterFD() (int, bool) {
// 	fd := int(atomic.LoadInt32(&s.fd))
// 	if fd < 0 {
// 		return -1, false
// 	}
//
// 	return fd, true
// }
//
// func (s *socket) shutdown(fd int) error {
// 	// Shutdown the socket to cancel any pending accepts.
// 	return unix.Shutdown(fd, unix.SHUT_RDWR)
// }
//
// // Shutdown closes the socket for read and write.
// //
// // Shutdown implements Socket.Shutdown.
// func (s *socket) Shutdown() error {
// 	fd, ok := s.enterFD()
// 	if !ok {
// 		return unix.EBADF
// 	}
//
// 	return s.shutdown(fd)
// }
//
// // Close closes the socket.
// //
// // Close implements Socket.Close.
// func (s *socket) Close() error {
// 	// Set the FD in the socket to -1, to ensure that all future calls to
// 	// FD/Release get nothing and Close calls return immediately.
// 	fd := int(atomic.SwapInt32(&s.fd, -1))
// 	if fd < 0 {
// 		// Already closed or closing.
// 		return unix.EBADF
// 	}
//
// 	// Shutdown the socket to cancel any pending accepts.
// 	s.shutdown(fd)
//
// 	return unix.Close(fd)
// }
//
// // wait blocks until the socket FD is ready for reading or writing, depending
// // on the value of write.
// //
// // Returns errClosing if the Socket is in the process of closing.
// func (s *socket) wait(write bool) error {
// 	for {
// 		// Checking the FD on each loop is not strictly necessary, it
// 		// just avoids an extra poll call.
// 		fd := atomic.LoadInt32(&s.fd)
// 		if fd < 0 {
// 			return ErrClosing
// 		}
//
// 		events := &unix.PollFd{
// 			// The actual socket FD.
// 			Fd:     fd,
// 			Events: unix.POLLIN,
// 		}
// 		if write {
// 			events.Events = unix.POLLOUT
// 		}
//
// 		_, err := poll(events)
// 		if err == unix.EINTR {
// 			continue
// 		}
// 		if err != nil {
// 			return err
// 		}
//
// 		return nil
// 	}
// }
//
// // buildIovec builds an iovec slice from the given []byte slice.
// //
// // iovecs is used as an initial slice, to avoid excessive allocations.
// func buildIovec(bufs [][]byte, iovecs []unix.Iovec) ([]unix.Iovec, int) {
// 	var length int
//
// 	for i := range bufs {
// 		if l := len(bufs[i]); l > 0 {
// 			iovecs = append(iovecs, unix.Iovec{
// 				Base: &bufs[i][0],
// 				Len:  uint64(l),
// 			})
// 			length += l
// 		}
// 	}
//
// 	return iovecs, length
// }
//
// // SocketReader wraps an individual receive operation.
// //
// // This may be used for doing vectorized reads and/or sending additional
// // control messages (e.g. FDs). The normal entrypoint is ReadVec.
// //
// // One of ExtractFDs or DisposeFDs must be called if EnableFDs is used.
// type SocketReader struct {
// 	socket   *socket
// 	source   []byte
// 	blocking bool
// 	race     *int32
// }
//
// // ReadVec reads into the pre-allocated bufs. Returns bytes read.
// //
// // The pre-allocatted space used by ReadVec is based upon slice lengths.
// //
// // This function is not guaranteed to read all available data, it
// // returns as soon as a single recvmsg call succeeds.
// func (r *SocketReader) ReadVec(bufs [][]byte) (int, error) {
// 	iovecs, length := buildIovec(bufs, make([]unix.Iovec, 0, 2))
//
// 	var msg unix.Msghdr
// 	if len(r.source) != 0 {
// 		msg.Name = &r.source[0]
// 		msg.Namelen = uint32(len(r.source))
// 	}
//
// 	if len(iovecs) != 0 {
// 		msg.Iov = &iovecs[0]
// 		msg.SetIovlen(len(iovecs))
// 	}
//
// 	// n is the bytes received.
// 	var n int
//
// 	fd, ok := r.socket.enterFD()
// 	if !ok {
// 		return 0, unix.EBADF
// 	}
//
// 	// Leave on returns below.
// 	for {
// 		var e error
//
// 		// Try a non-blocking recv first, so we don't give up the go runtime M.
// 		n, e = recvmsg(fd, &msg, unix.MSG_DONTWAIT|unix.MSG_TRUNC)
// 		if e == nil {
// 			break
// 		}
// 		if e == unix.EINTR {
// 			continue
// 		}
// 		if !r.blocking {
// 			return 0, e
// 		}
// 		if e != unix.EAGAIN && e != unix.EWOULDBLOCK {
// 			return 0, e
// 		}
//
// 		// Wait for the socket to become readable.
// 		err := r.socket.wait(false)
// 		if errors.Is(err, ErrClosing) {
// 			err = unix.EBADF
// 		}
// 		if err != nil {
// 			return 0, err
// 		}
// 	}
//
// 	if msg.Namelen < uint32(len(r.source)) {
// 		r.source = r.source[:msg.Namelen]
// 	}
//
// 	// All unet sockets are SOCK_STREAM or SOCK_SEQPACKET, both of which
// 	// indicate that the other end is closed by returning a 0 length read
// 	// with no error.
// 	if n == 0 {
// 		return 0, io.EOF
// 	}
//
// 	if r.race != nil {
// 		// See comments on Socket.race.
// 		atomic.AddInt32(r.race, 1)
// 	}
//
// 	if int(n) > length {
// 		return length, ErrMessageTruncated
// 	}
//
// 	return int(n), nil
// }
//
// // reader returns a reader for this socket.
// func (s *socket) reader(blocking bool) SocketReader {
// 	return SocketReader{socket: s, blocking: blocking, race: s.race}
// }
//
// // Read implements io.Reader.Read.
// //
// // Read implements Socket.Read.
// func (s *socket) Read(p []byte) (int, error) {
// 	r := s.reader(true)
// 	return r.ReadVec([][]byte{p})
// }
//
// // SocketWriter wraps an individual send operation.
// //
// // The normal entrypoint is WriteVec.
// type SocketWriter struct {
// 	socket   *socket
// 	to       []byte
// 	blocking bool
// 	race     *int32
// }
//
// // WriteVec writes the bufs to the socket. Returns bytes written.
// //
// // This function is not guaranteed to send all data, it returns
// // as soon as a single sendmsg call succeeds.
// func (w *SocketWriter) WriteVec(bufs [][]byte) (int, error) {
// 	iovecs, _ := buildIovec(bufs, make([]unix.Iovec, 0, 2))
//
// 	if w.race != nil {
// 		// See comments on Socket.race.
// 		atomic.AddInt32(w.race, 1)
// 	}
//
// 	var msg unix.Msghdr
// 	if len(w.to) != 0 {
// 		msg.Name = &w.to[0]
// 		msg.Namelen = uint32(len(w.to))
// 	}
//
// 	if len(iovecs) > 0 {
// 		msg.Iov = &iovecs[0]
// 		msg.SetIovlen(len(iovecs))
// 	}
//
// 	fd, ok := w.socket.enterFD()
// 	if !ok {
// 		return 0, unix.EBADF
// 	}
//
// 	// Leave on returns below.
// 	for {
// 		// Try a non-blocking send first, so we don't give up the go runtime M.
// 		n, e := sendmsg(fd, &msg, unix.MSG_DONTWAIT|unix.MSG_NOSIGNAL)
// 		if e == nil {
// 			return int(n), nil
// 		}
// 		if e == unix.EINTR {
// 			continue
// 		}
// 		if !w.blocking {
// 			return 0, e
// 		}
// 		if e != unix.EAGAIN && e != unix.EWOULDBLOCK {
// 			return 0, e
// 		}
//
// 		// Wait for the socket to become writeable.
// 		err := w.socket.wait(true)
// 		if err == ErrClosing {
// 			err = unix.EBADF
// 		}
// 		if err != nil {
// 			return 0, err
// 		}
// 	}
// }
//
// // writer returns a writer for this socket.
// func (s *socket) writer(blocking bool) SocketWriter {
// 	return SocketWriter{socket: s, blocking: blocking, race: s.race}
// }
//
// // Write implements io.Writer.Write.
// //
// // Write implements Socket.Write.
// func (s *socket) Write(p []byte) (int, error) {
// 	r := s.writer(true)
// 	return r.WriteVec([][]byte{p})
// }
