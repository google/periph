// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package fs

import (
	"strconv"
	"strings"
	"syscall"
)

const isLinux = true

// syscall.EpollCtl() commands.
//
// These are defined here so we don't need to import golang.org/x/sys/unix.
//
// http://man7.org/linux/man-pages/man2/epoll_ctl.2.html
const (
	epollCTLAdd = 1 // EPOLL_CTL_ADD
	epollCTLDel = 2 // EPOLL_CTL_DEL
	epollCTLMod = 3 // EPOLL_CTL_MOD
)

// Bitmask for field syscall.EpollEvent.Events.
//
// These are defined here so we don't need to import golang.org/x/sys/unix.
//
// http://man7.org/linux/man-pages/man2/epoll_ctl.2.html
type epollEvent uint32

const (
	epollIN        epollEvent = 0x1        // EPOLLIN: available for read
	epollOUT       epollEvent = 0x4        // EPOLLOUT: available for write
	epollPRI       epollEvent = 0x2        // EPOLLPRI: exceptional urgent condition
	epollERR       epollEvent = 0x8        // EPOLLERR: error
	epollHUP       epollEvent = 0x10       // EPOLLHUP: hangup
	epollET        epollEvent = 0x80000000 // EPOLLET: Edge Triggered behavior
	epollONESHOT   epollEvent = 0x40000000 // EPOLLONESHOT: One shot
	epollWAKEUP    epollEvent = 0x20000000 // EPOLLWAKEUP: disable system sleep; kernel >=3.5
	epollEXCLUSIVE epollEvent = 0x10000000 // EPOLLEXCLUSIVE: only wake one; kernel >=4.5
)

var bitmaskString = [...]struct {
	e epollEvent
	s string
}{
	{epollIN, "IN"},
	{epollOUT, "OUT"},
	{epollPRI, "PRI"},
	{epollERR, "ERR"},
	{epollHUP, "HUP"},
	{epollET, "ET"},
	{epollONESHOT, "ONESHOT"},
	{epollWAKEUP, "WAKEUP"},
	{epollEXCLUSIVE, "EXCLUSIVE"},
}

// String is useful for debugging.
func (e epollEvent) String() string {
	var out []string
	for _, b := range bitmaskString {
		if e&b.e != 0 {
			out = append(out, b.s)
			e &^= b.e
		}
	}
	if e != 0 {
		out = append(out, "0x"+strconv.FormatUint(uint64(e), 16))
	}
	if len(out) == 0 {
		out = []string{"0"}
	}
	return strings.Join(out, "|")
}

func ioctl(f uintptr, op uint, arg uintptr) error {
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f, uintptr(op), arg); errno != 0 {
		return syscall.Errno(errno)
	}
	return nil
}

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
	switch {
	case err == nil:
		break
	case err.Error() == "function not implemented":
		// Some arch (arm64) do not implement EpollCreate().
		if epollFd, err = syscall.EpollCreate1(0); err != nil {
			return err
		}
	default:
		return err
	}
	e.epollFd = epollFd
	e.fd = int(fd)
	// EPOLLWAKEUP could be used to force the system to not go do sleep while
	// waiting for an edge. This is generally a bad idea, as we'd instead have
	// the system to *wake up* when an edge is triggered. Achieving this is
	// outside the scope of this interface.
	e.event[0].Events = uint32(epollPRI | epollET)
	e.event[0].Fd = int32(e.fd)
	return syscall.EpollCtl(e.epollFd, epollCTLAdd, e.fd, &e.event[0])
}

func (e *event) wait(timeoutms int) (int, error) {
	// http://man7.org/linux/man-pages/man2/epoll_wait.2.html
	return syscall.EpollWait(e.epollFd, e.event[:], timeoutms)
}
