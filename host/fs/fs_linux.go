// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package fs

import "syscall"
import "golang.org/x/sys/unix"
import "time"
import "errors"

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
	e.event[0].Events = unix.EPOLLPRI | unix.EPOLLET
	e.event[0].Fd = int32(e.fd)
	return syscall.EpollCtl(e.epollFd, epollCTLAdd, e.fd, &e.event[0])
}

func (e *event) wait(timeoutms int) (int, error) {
	// http://man7.org/linux/man-pages/man2/epoll_wait.2.html
	n,err := syscall.EpollWait(e.epollFd, e.event[:], timeoutms)

	// Note: this event occurs only once, so it couldn't be used with multiple calls to WaitForEdge()
	if e.event[0].Events & unix.EPOLLOUT > 0 {
		// an out event indicates an abort and is fired by the deleteEvent method (only by modifying epoll_ctl to listen to this event)
		return n,errors.New("wait aborted")
	}
	return n,err
}

func (e *event) deleteEvent() error {
	// modifying the epoll event to have EPOLLOUT flag set, seems unlocks a pendin epoll_wait (timeout == -1), because
	// an event with EPOLLOUT flag enabled is returned, everytime this EPOLL_CTL_MOD call is made (observed on tests
	// with Raspberry Pi 0)
	// Thus an event with the EPOLLOUT flag set is used, to indicate an abort condition in the wait() method.
	// To distinguish between events created EPOLL_CTL_ADD (done by gpio.In() with a call to makeEvent) and EPOLL_CTL_MOD
	// (donw here, on deletion) the EPOLLOUT flag mustn't be set druing EPOLL_CTL_ADD in makeEvent.
	e.event[0].Events = unix.EPOLLPRI | unix.EPOLLET | unix.EPOLLOUT
	e.event[0].Fd = int32(e.fd)
	syscall.EpollCtl(e.epollFd, unix.EPOLL_CTL_MOD, e.fd, &e.event[0]) //ignore error, couldn't do anything about it

	//fmt.Println("mod err", err)

	// This sleep serves two purposes:
	// 1) 	The gpio.haltEdge() method calls WaitForEdge(0), which result in an epoll_wait with timeout == 0.
	// 		If this epoll_wait would be called before a potentially pending epoll_wait with a timeout > 0, it
	// 		could happen that the generated EPOLLOUT event is consumed by the epoll_wait from haltEdge() and thus
	//		missed by the pending epoll_wait (could be consumed only once). This ultimately would result in a missed
	//		abort condition for the pending wait with timeout > 0.
	//		The delay assures that a pending wait catches the condition first (should run in a seperate go routine).
	// 2)	If there's no delay between the EPOLL_CTL_MOD and the successive EPOLL_CTL_DEL, a pending epoll_wait will
	//		miss the EPOLLOUT event (according to tests on Raspberry Pi0)
	time.Sleep(2*time.Millisecond)

	// aditionally we delete the epoll event, as it is added again with a call to gpio.In(), because gpio.haltEdge()
	// assures that gpio.fEdge is set to nil after calling this method.
	return syscall.EpollCtl(e.epollFd, unix.EPOLL_CTL_DEL, e.fd, nil)
}
