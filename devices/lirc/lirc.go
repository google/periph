// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package lirc

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/google/pio"
	"github.com/google/pio/conn/gpio"
	"github.com/google/pio/conn/ir"
)

// Conn is an open port to lirc.
type Conn struct {
	w           net.Conn
	c           chan ir.Message
	lock        sync.Mutex
	list        map[string][]string // list of remotes and associated keys
	pendingList map[string][]string // list of remotes and associated keys being created.
}

// New returns a IR receiver / emitter handle.
func New() (*Conn, error) {
	w, err := net.Dial("unix", "/var/run/lirc/lircd")
	if err != nil {
		return nil, err
	}
	c := &Conn{w: w, c: make(chan ir.Message), list: map[string][]string{}}
	// Inconditionally retrieve the list of all known keys at start.
	if _, err := w.Write([]byte("LIST\n")); err != nil {
		w.Close()
		return nil, err
	}
	go c.loop(bufio.NewReader(w))
	return c, nil
}

// Close closes the socket to lirc. It is not a requirement to close before
// process termination.
func (c *Conn) Close() error {
	return c.w.Close()
}

// Emit implements ir.IR.
func (c *Conn) Emit(remote string, key ir.Key) error {
	// http://www.lirc.org/html/lircd.html#lbAH
	_, err := fmt.Fprintf(c.w, "SEND_ONCE %s %s", remote, key)
	return err
}

// Channel implements ir.IR.
func (c *Conn) Channel() <-chan ir.Message {
	return c.c
}

// Codes returns all the known codes.
//
// Empty if the list was not retrieved yet.
func (c *Conn) Codes() map[string][]string {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.list
}

//

func (c *Conn) loop(r *bufio.Reader) {
	defer func() {
		close(c.c)
		c.c = nil
	}()
	for {
		line, err := read(r)
		if line == "BEGIN" {
			err = c.readData(r)
		} else if len(line) != 0 {
			// Format is: <code> <repeat count> <button name> <remote control name>
			// http://www.lirc.org/html/lircd.html#lbAG
			if parts := strings.SplitN(line, " ", 5); len(parts) != 4 {
				log.Printf("ir: corrupted line: %v", line)
			} else {
				if i, err2 := strconv.Atoi(parts[1]); err2 != nil {
					log.Printf("ir: corrupted line: %v", line)
				} else if len(parts[2]) != 0 && len(parts[3]) != 0 {
					c.c <- ir.Message{Key: ir.Key(parts[2]), RemoteType: parts[3], Repeat: i != 0}
				}
			}
		}
		if err != nil {
			break
		}
	}
}

func (c *Conn) readData(r *bufio.Reader) error {
	// Format is:
	// BEGIN
	// <original command>
	// SUCCESS
	// DATA
	// <number of entries 1 based>
	// <entries>
	// ...
	// END
	cmd, err := read(r)
	if err != nil {
		return err
	}
	switch cmd {
	case "SIGHUP":
		_, err = c.w.Write([]byte("LIST\n"))
	default:
		// In case of any error, ignore the rest.
		line, err := read(r)
		if err != nil {
			return err
		}
		if line != "SUCCESS" {
			log.Printf("ir: unexpected line: %v, expected SUCCESS", line)
			return nil
		}
		if line, err = read(r); err != nil {
			return err
		}
		if line != "DATA" {
			log.Printf("ir: unexpected line: %v, expected DATA", line)
			return nil
		}
		if line, err = read(r); err != nil {
			return err
		}
		nbLines, err := strconv.Atoi(line)
		if err != nil {
			return err
		}
		list := make([]string, nbLines)
		for i := 0; i < nbLines; i++ {
			if list[i], err = read(r); err != nil {
				return err
			}
		}
		switch {
		case cmd == "LIST":
			// Request the codes for each remote.
			c.pendingList = map[string][]string{}
			for _, l := range list {
				if _, ok := c.pendingList[l]; ok {
					log.Printf("ir: unexpected %s", cmd)
				} else {
					c.pendingList[l] = []string{}
					if _, err = fmt.Fprintf(c.w, "LIST %s\n", l); err != nil {
						return err
					}
				}
			}
		case strings.HasPrefix(line, "LIST "):
			if c.pendingList == nil {
				log.Printf("ir: unexpected %s", cmd)
			} else {
				remote := cmd[5:]
				c.pendingList[remote] = list
				all := true
				for _, v := range c.pendingList {
					if len(v) == 0 {
						all = false
						break
					}
				}
				if all {
					c.lock.Lock()
					c.list = c.pendingList
					c.pendingList = nil
					c.lock.Unlock()
				}
			}
		default:
		}
	}
	line, err := read(r)
	if err != nil {
		return err
	}
	if line != "END" {
		log.Printf("ir: unexpected line: %v, expected END", line)
	}
	return nil
}

func read(r *bufio.Reader) (string, error) {
	raw, err := r.ReadBytes('\n')
	if err != nil {
		return "", err
	}
	if len(raw) != 0 {
		raw = raw[:len(raw)-1]
	}
	return string(raw), nil
}

// driver implements pio.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "lirc"
}

func (d *driver) Type() pio.Type {
	// Return the lowest priority, which is Functional.
	return pio.Functional
}

func (d *driver) Init() (bool, error) {
	in, out := getPins()
	if in == -1 && out == -1 {
		return false, nil
	}
	if in != -1 {
		if pin := gpio.ByNumber(in); pin != nil {
			gpio.MapFunction("IR_IN", pin)
		} else {
			gpio.MapFunction("IR_IN", gpio.INVALID)
		}
	} else {
		gpio.MapFunction("IR_IN", gpio.INVALID)
	}
	if out != -1 {
		if pin := gpio.ByNumber(out); pin != nil {
			gpio.MapFunction("IR_OUT", pin)
		} else {
			gpio.MapFunction("IR_OUT", gpio.INVALID)
		}
	} else {
		gpio.MapFunction("IR_OUT", gpio.INVALID)
	}
	return true, nil
}

// getPins queries the kernel module to determine which GPIO pins are taken by
// the driver.
//
// The return values can be converted to bcm238x.Pin. Return (-1, -1) on
// failure.
func getPins() (int, int) {
	// This is configured in /boot/config.txt as:
	// dtoverlay=gpio_in_pin=23,gpio_out_pin=22
	bytes, err := ioutil.ReadFile("/sys/module/lirc_rpi/parameters/gpio_in_pin")
	if err != nil || len(bytes) == 0 {
		return -1, -1
	}
	in, err := strconv.Atoi(strings.TrimRight(string(bytes), "\n"))
	if err != nil {
		return -1, -1
	}
	bytes, err = ioutil.ReadFile("/sys/module/lirc_rpi/parameters/gpio_out_pin")
	if err != nil || len(bytes) == 0 {
		return -1, -1
	}
	out, err := strconv.Atoi(strings.TrimRight(string(bytes), "\n"))
	if err != nil {
		return -1, -1
	}
	return in, out
}

var _ ir.Conn = &Conn{}
