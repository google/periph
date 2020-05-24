// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package sysfs

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

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

// eventsListener listens for events for multiple files as a single system call.
//
// One OS thread is needed for all the events. This is more efficient on single
// core system.
type eventsListener struct {
	// Atomic value set to one once fully initialized.
	initialized int32

	// Mapping of file descriptors to wait on with their corresponding channels.
	mu sync.Mutex
	// File descriptor of the epoll handle itself.
	epollFd int
	// Pipes to wake up the EpollWait() system call inside loop().
	r, w *os.File
	// Return channel to confirm that EpollWait() was woken up.
	wakeUp <-chan time.Time
	// Map of file handles to user listening channel.
	fds map[int32]chan<- time.Time
}

// init must be called on a fresh instance.
func (e *eventsListener) init() error {
	if atomic.LoadInt32(&e.initialized) != 0 {
		// Was already initialized.
		return nil
	}
	e.mu.Lock()
	if atomic.LoadInt32(&e.initialized) != 0 {
		// Was already initialized, but this was done concurrently with another
		// thread.
		e.mu.Unlock()
		return nil
	}
	var err error
	e.epollFd, err = syscall.EpollCreate(1)
	switch {
	case err == nil:
		break
	case err.Error() == "function not implemented":
		// Some arch (arm64) do not implement EpollCreate().
		if e.epollFd, err = syscall.EpollCreate1(0); err != nil {
			e.mu.Unlock()
			return err
		}
	default:
		e.mu.Unlock()
		return err
	}
	e.r, e.w, err = os.Pipe()
	if err != nil {
		e.mu.Unlock()
		return err
	}
	// Only need epollIN. epollPRI has no effect on pipes.
	if err = e.addFdInner(e.r.Fd(), epollET|epollIN); err != nil {
		// This object will not be reusable at this point.
		e.mu.Unlock()
		return err
	}
	wakeUp := make(chan time.Time)
	e.wakeUp = wakeUp
	e.fds = map[int32]chan<- time.Time{}
	e.fds[int32(e.r.Fd())] = wakeUp
	// The mutex is still held after this function exits, it's loop() that will
	// release the mutex.
	//
	// This forces loop() to be started before addFd() can be called by users.
	go e.loop()
	// Initialization is now good to go.
	atomic.StoreInt32(&e.initialized, 1)
	return nil
}

// loop is the main event loop.
func (e *eventsListener) loop() {
	var events []syscall.EpollEvent
	type lookup struct {
		c     chan<- time.Time
		event epollEvent
	}
	var lookups []lookup
	for first := true; ; {
		if !first {
			e.mu.Lock()
		}
		if len(events) < len(e.fds) {
			events = make([]syscall.EpollEvent, len(e.fds))
		}
		e.mu.Unlock()
		first = false

		if len(events) == 0 {
			panic("internal error: there's should be at least one pipe")
		}

		// http://man7.org/linux/man-pages/man2/epoll_wait.2.html
		n, err := syscall.EpollWait(e.epollFd, events, -1)
		if n <= 0 {
			// -1 if an error occurred (EBADF, EFAULT, EINVAL) or the call was
			// interrupted by a signal (EINTR).
			// 0 is the timeout occurred. In this case there's no timeout specified.
			// Still handle this explicitly in case a timeout could be triggered by
			// external events, like system sleep.
			continue
		}
		if err != nil {
			// TODO(maruel): It'd be nice to be able to surface this.
			// This may cause a busy loop. Hopefully the user will notice and will
			// fix their code.
			// This can happen when removeFd() is called, in this case silently
			// ignore the error.
			continue
		}

		now := time.Now()
		// Create a look up table with the lock, so that then the channel can be
		// pushed to without the lock.
		if cap(lookups) < n {
			lookups = make([]lookup, 0, n)
		} else {
			lookups = lookups[:0]
		}

		e.mu.Lock()
		for _, ev := range events[:n] {
			ep := epollEvent(ev.Events)
			// Skip over file descriptors that are not present.
			c, ok := e.fds[ev.Fd]
			if !ok {
				// That's a race condition where the file descriptor was removed by
				// removeFd() but it still triggered. Ignore this event.
				continue
			}
			// Look at the event to determine if it's worth sending a pulse. It's
			// maybe not worth it. Ignore epollERR, since it's always set for GPIO
			// sysfs.
			// Pipe and socket trigger epollIN and epollOUT, but GPIO sysfs triggers
			// epollPRI.
			if ep&(epollPRI|epollIN|epollOUT) != 0 {
				lookups = append(lookups, lookup{c: c, event: ep})
			}
		}
		e.mu.Unlock()

		// Once the lock is released, send the timestamps.
		for _, t := range lookups {
			t.c <- now
		}
	}
}

// addFd starts listening to events generated by file descriptor |fd|.
//
// fd is the OS file descriptor. In practice, it must fit a int32 value. It
// works on pipes, sockets and sysfs objects like GPIO but not on standard
// files.
//
// c is the channel to send events to. Unbuffered channel will block the event
// loop, which may mean lost events, especially if multiple files are listened
// to simultaneously.
//
// flags is the events to listen to. No need to specify epollERR and epollHUP,
// they are sent anyway.
//
// addFd lazy initializes eventsListener if it was not initialized yet.
//
// It can fail due to various reasons, a few are:
//   ENOSPC: /proc/sys/fs/epoll/max_user_watches limit was exceeded
//   ENOMEM: No memory available
//   EPERM: fd is a regular file or directory
func (e *eventsListener) addFd(fd uintptr, c chan<- time.Time, flags epollEvent) error {
	if c == nil {
		return errors.New("fd: addFd requires a valid channel")
	}
	if err := e.init(); err != nil {
		return err
	}
	if err := e.addFdInner(fd, flags); err != nil {
		return err
	}
	e.mu.Lock()
	e.fds[int32(fd)] = c
	e.mu.Unlock()
	// Wake up the poller so it notices there's one new file.
	e.wakeUpLoop(nil)
	return nil
}

func (e *eventsListener) addFdInner(fd uintptr, flags epollEvent) error {
	ev := syscall.EpollEvent{Events: uint32(flags), Fd: int32(fd)}
	return syscall.EpollCtl(e.epollFd, epollCTLAdd, int(fd), &ev)
}

// removeFd stop listening to events on this file descriptor.
func (e *eventsListener) removeFd(fd uintptr) error {
	if err := syscall.EpollCtl(e.epollFd, epollCTLDel, int(fd), nil); err != nil {
		return err
	}
	e.mu.Lock()
	delete(e.fds, int32(fd))
	e.mu.Unlock()
	// Wake up the poller so it notices there's one less file.
	e.wakeUpLoop(nil)
	return nil
}

// wakeUpLoop wakes up the poller and waits for it.
//
// Must not be called with the lock held.
func (e *eventsListener) wakeUpLoop(c <-chan time.Time) time.Time {
	if atomic.LoadInt32(&e.initialized) == 0 {
		return time.Time{}
	}
	// TODO(maruel): Figure out a way to wake up that doesn't require emptying.
	var b [1]byte
	_, _ = e.w.Write(b[:])
	var t time.Time
	if c != nil {
		// To prevent deadlock, also empty c.
		for {
			select {
			case <-c:
			case t = <-e.wakeUp:
				goto out
			}
		}
	out:
	} else {
		t = <-e.wakeUp
	}
	// Don't forget to empty the pipe. Sadly, this will wake up the loop a second
	// time.
	_, _ = e.r.Read(b[:])
	return t
}

// events is the global events listener.
//
// It uses a single global goroutine lazily initialized to call
// syscall.EpollWait() to listen to many file descriptors at once.
var events eventsListener
