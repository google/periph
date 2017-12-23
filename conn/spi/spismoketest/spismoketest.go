// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package spismoketest is leveraged by periph-smoketest to verify that an
// EEPROM device can be accessed on a SPI port.
//
// This assumes the presence of the periph-tester board, which includes these
// two devices. See https://github.com/tve/periph-tester
package spismoketest

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spireg"
)

// SmokeTest is imported by periph-smoketest.
type SmokeTest struct {
}

func (s *SmokeTest) String() string {
	return s.Name()
}

// Name implements the SmokeTest interface.
func (s *SmokeTest) Name() string {
	return "spi-testboard"
}

// Description implements the SmokeTest interface.
func (s *SmokeTest) Description() string {
	return "Tests EEPROM on periph-tester board"
}

// Run implements the SmokeTest interface.
func (s *SmokeTest) Run(f *flag.FlagSet, args []string) error {
	spiID := f.String("spi", "", "SPI port to use")
	wp := f.String("wp", "", "gpio pin for EEPROM write-protect")
	seed := f.Int64("seed", 0, "random number seed, default is to use the time")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		f.Usage()
		return errors.New("unrecognized arguments")
	}

	// Open the port.
	spiDev, err := spireg.Open(*spiID)
	if err != nil {
		return fmt.Errorf("error opening %s: %v", *spiID, err)
	}
	defer spiDev.Close()

	// Set SPI parameters.
	c, err := spiDev.Connect(4000000, spi.Mode0, 8)
	if err != nil {
		return fmt.Errorf("error setting SPI parameters: %v", err)
	}

	// Open the WC pin.
	var wpPin gpio.PinIO
	if *wp != "" {
		if wpPin = gpioreg.ByName(*wp); wpPin == nil {
			return fmt.Errorf("cannot open gpio pin %s for EEPROM write protect", *wp)
		}
	}

	// Init rand.
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	rand.Seed(*seed)
	log.Printf("%s: random number seed %d", s, *seed)

	// Run the tests.
	return s.eeprom(c, wpPin)
}

// eeprom tests a 5080 8Kbit serial EEPROM attached to the SPI port.
// Such a chip is included on the periph-tester board.
//
// The test performs some longish writes and reads and also tests a
// write error. It uses some ad-hoc command sequences for expediency
// that should be replaced by a proper driver eventually.
// Only the first 256 bytes of the chip are tested, thus a smaller
// EEPROM could be substituted.
//
// On the periph-tester board the write-protect is hooked-up to a gpio
// pin which must be driven low to write to the EEPROM. If no pin is
// passed to the test only the completion of reads will be tested.
//
// Datasheet: http://www.st.com/content/ccc/resource/technical/document/datasheet/28/42/21/c1/13/bf/47/9a/DM00043274.pdf/files/DM00043274.pdf/jcr:content/translations/en.DM00043274.pdf
func (s *SmokeTest) eeprom(d spi.Conn, wpPin gpio.PinIO) error {
	// Can't do anything if we don't have write-protect for the chip.
	if wpPin == nil {
		log.Printf("%s: no WP pin specified, skipping eeprom tests", s)
		return nil
	}

	// Clear write-protect so we can write.
	if err := wpPin.Out(gpio.High); err != nil {
		return fmt.Errorf("eeprom: cannot init WP control pin: %s", err)
	}

	var rBuf [35]byte
	// Read status register of the EEPROM and expect not to get an error (the device
	// can't produce an error with SPI, but the driver/OS could act up).
	if err := d.Tx([]byte{cmdReadStatus, 0}, rBuf[:2]); err != nil {
		return fmt.Errorf("eeprom: error on the first read status access: %v", err)
	}

	// Invert one of the block protect bits and expect to read the modified status reg back.
	sr := 0x02 | ((rBuf[1] & 0x0c) ^ 0x08) // flip BP1
	if err := d.Tx([]byte{cmdWriteEnable}, rBuf[:1]); err != nil {
		return err
	}
	if err := d.Tx([]byte{cmdWriteStatus, sr}, rBuf[:2]); err != nil {
		return err
	}
	if err := waitReady(d); err != nil {
		return err
	}
	if err := d.Tx([]byte{cmdReadStatus, 0}, rBuf[:2]); err != nil {
		return err
	}
	if (rBuf[1] & 0xc) != (sr & 0x0c) {
		return fmt.Errorf("eeprom: wrote %#x to status register but got %#x back", sr, rBuf[1])
	}

	// Clear status register so we can write anywhere.
	if err := d.Tx([]byte{cmdWriteEnable}, rBuf[:1]); err != nil {
		return err
	}
	if err := d.Tx([]byte{cmdWriteStatus, 0x00}, rBuf[:2]); err != nil {
		return err
	}
	if err := waitReady(d); err != nil {
		return err
	}

	// Pick a byte in the first 256 bytes and try to write it and read it back a couple
	// of times. Using a random byte for "wear leveling"...
	addr := [2]byte{0, byte(rand.Intn(256))} // 16-bit big-endian address
	values := []byte{0x55, 0xAA, 0xF0, 0x0F, 0x13}
	log.Printf("%s: writing&reading EEPROM byte %#x", s, addr[1])
	for _, v := range values {
		// Write byte.
		if err := d.Tx([]byte{cmdWriteEnable}, rBuf[:1]); err != nil {
			return err
		}
		if err := d.Tx([]byte{cmdWriteMemory, addr[0], addr[1], v}, rBuf[:4]); err != nil {
			return err
		}
		// Read byte back after the chip is ready.
		if err := waitReady(d); err != nil {
			return err
		}
		if err := d.Tx([]byte{cmdReadMemory, addr[0], addr[1], 0}, rBuf[:4]); err != nil {
			return err
		}
		if rBuf[3] != v {
			return fmt.Errorf("eeprom: wrote %#x but got %#v back", v, rBuf[3])
		}
	}

	// Pick a page in the first 256 bytes and try to write it and read it back.
	// Using a random page for "wear leveling" and randomizing what gets written
	// so it actually changes from one test run to the next.
	addr[1] = byte(rand.Intn(256)) & 0xe0 // round to page boundary
	r := byte(rand.Intn(256))             // randomizer for value written
	log.Printf("%s: writing&reading EEPROM page %#x", s, addr)
	// val calculates the value for byte i.
	val := func(i int) byte { return byte((i<<4)|(16-(i>>1))) ^ r }
	var onePage [32]byte
	for i := 0; i < 32; i++ {
		onePage[i] = val(i)
	}
	// Write page.
	if err := d.Tx([]byte{cmdWriteEnable}, rBuf[:1]); err != nil {
		return err
	}
	if err := d.Tx(append([]byte{cmdWriteMemory, addr[0], addr[1]}, onePage[:]...), rBuf[:35]); err != nil {
		return err
	}
	// Zero buffer in anticipation of read.
	for i := 0; i < 32; i++ {
		onePage[i] = 0
	}
	// Read page back after the chip is ready.
	if err := waitReady(d); err != nil {
		return err
	}
	if err := d.Tx(append([]byte{cmdReadMemory, addr[0], addr[1]}, onePage[:]...), rBuf[:35]); err != nil {
		return err
	}
	// Ensure we got the correct data.
	for i := 0; i < 32; i++ {
		if rBuf[i+3] != val(i) {
			return fmt.Errorf("eeprom: incorrect read of addr %#x: expected %#x got %#x",
				addr[1]+byte(i), val(i), rBuf[i+3])
		}

	}

	// Set write-protect, attempt a write, and expect it not to happen.
	if err := wpPin.Out(gpio.Low); err != nil {
		return err
	}
	if err := d.Tx([]byte{0x10, 0xA5}, nil); err == nil {
		return errors.New("eeprom: write with write-control disabled didn't return an error")
	}
	// Write the value of the second byte in the just-written page into the first byte.
	first := rBuf[0]
	second := rBuf[1]
	if err := d.Tx([]byte{cmdWriteEnable}, rBuf[:1]); err != nil {
		return err
	}
	if err := d.Tx([]byte{cmdWriteMemory, addr[0], addr[1], second}, rBuf[:4]); err != nil {
		return err
	}
	// Read byte back after the chip is ready.
	if err := waitReady(d); err != nil {
		return err
	}
	if err := d.Tx([]byte{cmdReadMemory, addr[0], addr[1], 0}, rBuf[:4]); err != nil {
		return err
	}
	if rBuf[3] != first {
		return fmt.Errorf("eeprom: write protect failed, expected %#x got %#x", first, rBuf[3])
	}

	return nil
}

// waitReady reads the status register until the write is complete or a timeout expires.
func waitReady(d spi.Conn) error {
	for start := time.Now(); time.Since(start) <= 100*time.Millisecond; {
		var rBuf [2]byte
		if err := d.Tx([]byte{cmdReadStatus, 0}, rBuf[:]); err != nil {
			return err
		}
		if rBuf[1]&1 == 0 {
			return nil
		}
	}
	return errors.New("eeprom: write timout")
}

// Constants to support EEPROM tests.
const (
	cmdWriteEnable  = 0x06
	cmdWriteDisable = 0x04
	cmdReadStatus   = 0x05
	cmdWriteStatus  = 0x01
	cmdReadMemory   = 0x03
	cmdWriteMemory  = 0x02
)
