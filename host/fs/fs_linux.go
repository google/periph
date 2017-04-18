// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package fs

import "syscall"

const isLinux = true

func ioctl(f uintptr, op uint, arg uintptr) error {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f, uintptr(op), arg); errno != 0 {
		return syscall.Errno(errno)
	}
	return nil
}

const (
	epollET     = 1 << 31
	epollPRI    = 2
	epollCTLAdd = 1
	epollCTLDel = 2
	epollCTLMod = 3
)

type event struct {
	event   [1]syscall.EpollEvent
	epollFd int
	fd      int
}

// makeEvent creates an epoll *edge* triggered event.
//
// References:
// behavior and flags: http://man7.org/linux/man-pages/man7/epoll.7.html
// syscall.EpollCreate: http://man7.org/linux/man-pages/man2/epoll_create.2.html
// syscall.EpollCtl: http://man7.org/linux/man-pages/man2/epoll_ctl.2.html
func (e *event) makeEvent(fd uintptr) error {
	epollFd, err := syscall.EpollCreate(1)
	if err != nil {
		return err
	}
	e.epollFd = epollFd
	e.fd = int(fd)
	// EPOLLWAKEUP could be used to force the system to not go do sleep while
	// waiting for an edge. This is generally a bad idea, as we'd instead have
	// the system to *wake up* when an edge is triggered. Achieving this is
	// outside the scope of this interface.
	e.event[0].Events = epollPRI | epollET
	e.event[0].Fd = int32(e.fd)
	return syscall.EpollCtl(e.epollFd, epollCTLAdd, e.fd, &e.event[0])
}

func (e *event) wait(timeoutms int) (int, error) {
	// http://man7.org/linux/man-pages/man2/epoll_wait.2.html
	return syscall.EpollWait(e.epollFd, e.event[:], timeoutms)
}
