// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// i2csmoketest verifies that an I²C EEPROM device and a DS2483 device can be accessed on
// an I²C bus. This assumes the presence of the periph-tester board, which includes these two
// devices.
package i2csmoketest

import (
	"flag"
	"fmt"

	"github.com/google/periph/conn/i2c"
)

type SmokeTest struct {
}

// Name
func (s *SmokeTest) Name() string {
	return "i2c-testboard"
}

func (s *SmokeTest) Description() string {
	return "Tests EEPROM and DS2483 on periph-tester board"
}

func (s *SmokeTest) Run(args []string) error {
	f := flag.NewFlagSet("i2c", flag.ExitOnError)
	busNum := f.Int("n", -1, "bus number, -1 for lowest numbered bus")
	f.Parse(args)

	// Open the bus.
	i2cBus, err := i2c.New(*busNum)
	if err != nil {
		return fmt.Errorf("i2c-smoke: %s", err)
	}

	// Run the tests.
	if err := s.ds248x(i2cBus); err != nil {
		return fmt.Errorf("i2c-smoke: %s", err)
	}
	if err := s.eeprom(i2cBus); err != nil {
		return fmt.Errorf("i2c-smoke: %s", err)
	}

	return nil
}

// ds248x tests a Maxim DS248x 1-wire interface chip attached to the I²C bus. Such a chip
// is included on the periph-tester board.
//
// The test performs a reset of the chip and erads its status register using canned command
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

// eeprom tests a 24C08 8Kbit serial EEPROM attached to the I²C bus. Such a chip
// is included on the periph-tester board at addresses 0x50-0x53.
//
// The test performs some longish writes and reads and also tests a write error. It uses some
// ad-hoc command sequences for expediency that should be replaced by a proper driver eventually.
// Only the first 1KB of the chip is tested, thus a smaller EEPROM could be substituted.
//
// On the periph-tester board the write-enable is hooked-up to a gpio pin which must be driven
// low to write to the EEPROM. If no pin is passed to the test only the completion of reads
// will be tested.
//
// Datasheet: http://www.st.com/content/ccc/resource/technical/document/datasheet/cc/f5/a5/01/6f/4b/47/d2/DM00070057.pdf/files/DM00070057.pdf/jcr:content/translations/en.DM00070057.pdf
func (s *SmokeTest) eeprom(bus i2c.Bus) error {
	d := i2c.Dev{Bus: bus, Addr: 0x50}

	// Read byte 10 of the EEPROM and expect to get something.
	var oneByte [1]byte
	if err := d.Tx([]byte{0x12}, oneByte[:]); err != nil {
		return fmt.Errorf("eeprom: error on the first read access")
	}

	// Read page 5 of the EEPROM and expect to get something.
	var onePage [16]byte
	if err := d.Tx([]byte{5 * 16}, onePage[:]); err != nil {
		return fmt.Errorf("eeprom: error reading page 5")
	}

	return nil
}
