// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"os"
	"syscall"
)

const (
	epollET       = 1 << 31
	epollPRI      = 2
	epoll_CTL_ADD = 1
	epoll_CTL_DEL = 2
	epoll_CTL_MOD = 3
)

type event struct {
	event   [1]syscall.EpollEvent
	epollFd int
	fd      int
}

// makeEvent creates an epoll *edge* triggered event.
//
// An edge triggered event is basically an "auto-reset" event, where waiting on
// the edge resets it. A level triggered event requires manual resetting; this
// could be done via a Read() call but there's no need to require the user to
// call Read(). This is particularly useless in the case of gpio.RisingEdge and
// gpio.FallingEdge.
//
// As per the official doc, edge triggers is still remembered even when no
// epoll_wait() call is running, so no edge is missed. Two edges will be
// coallesced into one if the user mode process can't keep up. There's no
// accumulation of edges.
//
// References:
// behavior and flags: http://man7.org/linux/man-pages/man7/epoll.7.html
// syscall.EpollCreate: http://man7.org/linux/man-pages/man2/epoll_create.2.html
// syscall.EpollCtl: http://man7.org/linux/man-pages/man2/epoll_ctl.2.html
func (e *event) makeEvent(f *os.File) error {
	epollFd, err := syscall.EpollCreate(1)
	if err != nil {
		return err
	}
	e.epollFd = epollFd
	e.fd = int(f.Fd())
	// EPOLLWAKEUP could be used to force the system to not go do sleep while
	// waiting for an edge. This is generally a bad idea, as we'd instead have
	// the system to *wake up* when an edge is triggered. Achieving this is
	// outside the scope of this interface.
	e.event[0].Events = epollPRI | epollET
	e.event[0].Fd = int32(e.fd)
	return syscall.EpollCtl(e.epollFd, epoll_CTL_ADD, e.fd, &e.event[0])
}

func (e *event) wait(timeoutms int) (int, error) {
	// http://man7.org/linux/man-pages/man2/epoll_wait.2.html
	return syscall.EpollWait(e.epollFd, e.event[:], timeoutms)
}

func isErrBusy(err error) bool {
	e, ok := err.(*os.PathError)
	return ok && e.Err == syscall.EBUSY
}
