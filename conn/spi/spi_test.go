// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package spi

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"periph.io/x/periph/conn"
)

func ExampleAll() {
	fmt.Print("SPI buses available:\n")
	for name := range All() {
		fmt.Printf("- %s\n", name)
	}
}

func Example() {
	// Find a specific device on all available SPI buses:
	for _, opener := range All() {
		bus, err := opener()
		if err != nil {
			log.Fatal(err)
		}
		w := []byte("command to device")
		r := make([]byte, 16)
		if err := bus.Tx(w, r); err != nil {
			log.Fatal(err)
		}
		// Handle 'r'.
		bus.Close()
	}
}

func TestInvalid(t *testing.T) {
	defer reset()
	if _, err := New(-1, -1); err == nil {
		t.FailNow()
	}
}

func TestAll(t *testing.T) {
	defer reset()
	if len(All()) != 0 {
		t.FailNow()
	}
	if err := Register("a", 0, 0, op); err != nil {
		t.Fatal(err)
	}
	if len(All()) != 1 {
		t.FailNow()
	}
	if b, err := New(0, 0); b == nil || err != nil {
		t.Fatal(b, err)
	}
}

func TestRegister_fail(t *testing.T) {
	if err := Register("", 0, 0, op); err == nil {
		t.FailNow()
	}
	if err := Register("a", 0, 0, nil); err == nil {
		t.FailNow()
	}
}

func TestUnregister_fail(t *testing.T) {
	if Unregister("", -1) == nil {
		t.FailNow()
	}
	if err := Register("a", 0, 0, op); err != nil {
		t.Fatal(err)
	}
	if Unregister("a", 1) == nil {
		t.FailNow()
	}
}

//

type basicSPI struct {
}

func (b *basicSPI) String() string {
	return "basicSPI"
}

func (b *basicSPI) Close() error {
	return errors.New("not implemented")
}

func (b *basicSPI) Tx(w, r []byte) error {
	return errors.New("not implemented")
}

func (b *basicSPI) Write(w []byte) (int, error) {
	return 0, errors.New("not implemented")
}

func (b *basicSPI) Duplex() conn.Duplex {
	return conn.DuplexUnknown
}

func (b *basicSPI) Speed(hz int64) error {
	return errors.New("not implemented")
}

func (b *basicSPI) Configure(mode Mode, bits int) error {
	return errors.New("not implemented")
}

func op() (ConnCloser, error) {
	return &basicSPI{}, nil
}

func reset() {
	mu.Lock()
	defer mu.Unlock()
	byNumber = map[int]map[int]Opener{}
	byName = map[string]Opener{}
}
