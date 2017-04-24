// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package i2csmoketest is leveraged by periph-smoketest to verify that an I²C
// EEPROM peripheral and a DS2483 peripheral can be accessed on an I²C bus.
//
// This assumes the presence of the periph-tester board, which includes these
// two peripherals.
//
// See https://github.com/tve/periph-tester
package i2csmoketest

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/mmr"
)

// SmokeTest is imported by periph-smoketest.
type SmokeTest struct {
}

func (s *SmokeTest) String() string {
	return s.Name()
}

// Name implements the SmokeTest interface.
func (s *SmokeTest) Name() string {
	return "i2c-testboard"
}

// Description implements the SmokeTest interface.
func (s *SmokeTest) Description() string {
	return "Tests EEPROM and DS2483 on periph-tester board"
}

// Run implements the SmokeTest interface.
func (s *SmokeTest) Run(args []string) error {
	f := flag.NewFlagSet("i2c", flag.ExitOnError)
	busName := f.String("bus", "", "I²C bus to use")
	wc := f.String("wc", "", "gpio pin for EEPROM write-control pin")
	seed := f.Int64("seed", 0, "random number seed, default is to use the time")
	f.Parse(args)

	// Open the bus.
	i2cBus, err := i2creg.Open(*busName)
	if err != nil {
		return fmt.Errorf("error opening %s: %v", *busName, err)
	}
	defer i2cBus.Close()

	// Open the WC pin.
	var wcPin gpio.PinIO
	if *wc != "" {
		if wcPin = gpioreg.ByName(*wc); wcPin == nil {
			return fmt.Errorf("cannot open gpio pin %s for EEPROM write control", *wc)
		}
	}

	// Init rand.
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	rand.Seed(*seed)
	log.Printf("%s: random number seed %d", s, *seed)

	// Run the tests.
	if err := s.ds248x(i2cBus); err != nil {
		return err
	}
	if err := s.eeprom(i2cBus, wcPin); err != nil {
		return err
	}

	return nil
}

// ds248x tests a Maxim DS248x 1-wire interface chip attached to the I²C bus. Such a chip
// is included on the periph-tester board.
//
// The test performs a reset of the chip and reads its status register using canned command
// sequences gleaned from the ds248x driver in order to avoid introducing a dependency on the
// full driver.
func (s *SmokeTest) ds248x(bus i2c.Bus) error {
	d := i2c.Dev{Bus: bus, Addr: 0x18}

	// Issue a reset command.
	if err := d.Tx([]byte{cmdReset}, nil); err != nil {
		return fmt.Errorf("ds248x: error while resetting: %s", err)
	}

	// Read the status register to confirm that we have a responding ds248x
	var stat [1]byte
	if err := d.Tx([]byte{cmdSetReadPtr, regStatus}, stat[:]); err != nil {
		return fmt.Errorf("ds248x: error while reading status register: %s", err)
	}
	if stat[0] != 0x18 {
		return fmt.Errorf("ds248x: invalid status register value: %#x, expected 0x18\n", stat[0])
	}

	return nil
}

// Constants to support the ds248x test.
const (
	cmdReset      = 0xf0 // reset ds248x
	cmdSetReadPtr = 0xe1 // set the read pointer
	regStatus     = 0xf0 // read ptr for status register
)

// eeprom tests a 24C08 8Kbit serial EEPROM attached to the I²C bus.
// Such a chip is included on the periph-tester board at addresses 0x50-0x53.
//
// The test performs some longish writes and reads and also tests a
// write error. It uses some ad-hoc command sequences for expediency
// that should be replaced by a proper driver eventually.
// Only the first 256 bytes of the chip are tested, thus a smaller
// EEPROM could be substituted.
//
// On the periph-tester board the write-enable is hooked-up to a gpio
// pin which must be driven low to write to the EEPROM. If no pin is
// passed to the test only the completion of reads will be tested.
//
// Datasheet: http://www.st.com/content/ccc/resource/technical/document/datasheet/cc/f5/a5/01/6f/4b/47/d2/DM00070057.pdf/files/DM00070057.pdf/jcr:content/translations/en.DM00070057.pdf
func (s *SmokeTest) eeprom(bus i2c.Bus, wcPin gpio.PinIO) error {
	d := mmr.Dev8{Conn: &i2c.Dev{Bus: bus, Addr: 0x50}, Order: binary.LittleEndian}

	// Read byte 0x12 of the EEPROM and expect to get something.
	if _, err := d.ReadUint8(0x12); err != nil {
		return fmt.Errorf("eeprom: error on the first read access")
	}

	// Read page 5 of the EEPROM and expect to get something.
	var onePage [16]byte
	if err := d.ReadStruct(5*16, onePage[:]); err != nil {
		return fmt.Errorf("eeprom: error reading page 5")
	}

	// Stop here if we don't have write-control for the chip.
	if wcPin == nil {
		log.Printf("%s: no WC pin specified, skipping eeprom write tests", s)
		return nil
	}

	// Enable write-control
	if err := wcPin.Out(gpio.Low); err != nil {
		return fmt.Errorf("eeprom: cannot init WC control pin: %v", err)
	}
	time.Sleep(time.Millisecond)
	wcPin.Out(gpio.High)
	time.Sleep(time.Millisecond)
	wcPin.Out(gpio.Low)

	// Pick a byte in the first 256 bytes and try to write it and read it back a couple
	// of times. Using a random byte for "wear leveling"...
	addr := byte(rand.Intn(256))
	values := []byte{0x55, 0xAA, 0xF0, 0x0F, 0x13}
	log.Printf("%s: writing&reading EEPROM byte %#x", s, addr)
	for _, v := range values {
		// Write byte.
		if err := d.WriteUint8(addr, v); err != nil {
			return fmt.Errorf("eeprom: error writing %#x to byte at %#x: %v", v, addr, err)
		}
		// Read byte back once the peripheral is ready (takes several ms for the write
		// to complete).
		var w byte
		for start := time.Now(); time.Since(start) <= 100*time.Millisecond; {
			var err error
			if w, err = d.ReadUint8(addr); err == nil {
				break
			}
		}
		if w != v {
			return fmt.Errorf("eeprom: wrote %#v but read back %#v", v, w)
		}
	}

	// Pick a page in the first 256 bytes and try to write it and read it back.
	// Using a random page for "wear leveling" and randomizing what gets written
	// so it actually changes from one test run to the next.
	addr = byte(rand.Intn(256)) & 0xf0 // round to page boundary
	r := byte(rand.Intn(256))          // randomizer for value written
	log.Printf("%s: writing&reading EEPROM page %#x", s, addr)
	// val calculates the value for byte i.
	val := func(i int) byte { return byte((i<<4)|(16-i)) ^ r }
	for i := 0; i < 16; i++ {
		onePage[i] = val(i)
	}
	// Write page.
	if err := d.WriteStruct(addr, onePage[:]); err != nil {
		return fmt.Errorf("eeprom: error writing to page %#x: %v", addr, err)
	}
	// Clear the buffer to prep for reading back.
	for i := 0; i < 16; i++ {
		onePage[i] = 0
	}
	// Read page back once the peripheral is ready (takes several ms for the
	// write to complete).
	for start := time.Now(); time.Since(start) <= 100*time.Millisecond; {
		if err := d.ReadStruct(addr, onePage[:]); err == nil {
			break
		}
	}
	// Ensure we got the correct data.
	for i := 0; i < 16; i++ {
		if onePage[i] != val(i) {
			return fmt.Errorf("eeprom: incorrect read of addr %#x: expected %#x got %#x",
				addr+byte(i), val(i), onePage[i])
		}

	}

	// Disable write-control, attempt a write, and expect to get an i2c error.
	// TODO: create a clearly identifiable error.
	wcPin.Out(gpio.High)
	if err := d.WriteUint8(0x10, 0xA5); err == nil {
		return errors.New("eeprom: write with write-control disabled didn't return an error")
	}

	return nil
}
