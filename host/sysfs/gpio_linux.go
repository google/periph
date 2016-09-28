// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"os"
	"syscall"
)

type event [1]syscall.EpollEvent

func (e event) wait(ep, timeoutms int) (int, error) {
	// http://man7.org/linux/man-pages/man2/epoll_wait.2.html
	return syscall.EpollWait(ep, e[:], timeoutms)
}

// makeEvent creates an epoll *edge* triggered event.
//
// An edge triggered event is basically an "auto-reset" event, where waiting on
// the edge resets it. A level triggered event requires manual resetting; this
// could be done via a Read() call but there's no need to require the user to
// call Read(). This is particularly useless in the case of gpio.Rising and
// gpio.Falling.
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
func (e event) makeEvent(f *os.File) (int, error) {
	epollFd, err := syscall.EpollCreate(1)
	if err != nil {
		return 0, err
	}
	// EPOLLWAKEUP could be used to force the system to not go do sleep while
	// waiting for an edge. This is generally a bad idea, as we'd instead have
	// the system to *wake up* when an edge is triggered. Achieving this is
	// outside the scope of this interface.
	const EPOLLPRI = 2
	const EPOLLET = 1 << 31
	const EPOLL_CTL_ADD = 1
	fd := f.Fd()
	e[0].Events = EPOLLPRI | EPOLLET
	e[0].Fd = int32(fd)
	return epollFd, syscall.EpollCtl(epollFd, EPOLL_CTL_ADD, int(fd), &e[0])
}

func isErrBusy(err error) bool {
	e, ok := err.(*os.PathError)
	return ok && e.Err == syscall.EBUSY
}
