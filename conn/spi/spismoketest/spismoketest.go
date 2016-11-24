// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// spismoketest verifies that an I²C EEPROM device can be accessed on
// an SPI bus. This assumes the presence of the periph-tester board,
// which includes these two devices.
package spismoketest

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/periph/conn/gpio"
	"github.com/google/periph/conn/spi"
)

type SmokeTest struct {
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
func (s *SmokeTest) Run(args []string) error {
	f := flag.NewFlagSet("spi", flag.ExitOnError)
	busNum := f.Int("bus", -1, "bus number, -1 for lowest numbered bus")
	csNum := f.Int("cs", -1, "chip select number, -1 for the lowest numbered")
	wp := f.Int("wp", 0, "gpio pin number for EEPROM write-protect pin")
	f.Parse(args)

	// Open the bus.
	spiDev, err := spi.New(*busNum, *csNum)
	if err != nil {
		return fmt.Errorf("spi-smoke: %s", err)
	}
	defer spiDev.Close()

	// Open the WC pin.
	var wpPin gpio.PinIO
	if *wp != 0 {
		wpPin = gpio.ByNumber(*wp)
		if wpPin == nil {
			return fmt.Errorf("spi-smoke: cannot open gpio pin %d for EEPROM write protect", *wp)
		}
	}

	// Run the tests.
	if err := s.eeprom(spiDev, wpPin); err != nil {
		return fmt.Errorf("spi-smoke: %s", err)
	}

	return nil
}

// eeprom tests a 5080 8Kbit serial EEPROM attached to the SPI bus.
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
func (s *SmokeTest) eeprom(d spi.Dev, wpPin gpio.PinIO) error {
	rand.Seed(time.Now().UnixNano())

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

	// Stop here if we don't have write-control for the chip.
	if wpPin == nil {
		log.Println("spi-smoke: no WC pin specified, skipping eeprom write tests")
		return nil
	}

	// Enable write-control
	if err := wpPin.Out(gpio.High); err != nil {
		return fmt.Errorf("eeprom: cannot init WC control pin: %s", err)
	}
	time.Sleep(time.Millisecond)
	wpPin.Out(gpio.Low)
	time.Sleep(time.Millisecond)
	wpPin.Out(gpio.High)

	// Pick a byte in the first 256 bytes and try to write it and read it back a couple
	// of times. Using a random byte for "wear leveling"...
	addr := byte(rand.Intn(256))
	values := []byte{0x55, 0xAA, 0xF0, 0x0F, 0x13}
	log.Printf("spi-smoke writing&reading EEPROM byte %#x", addr)
	for _, v := range values {
		// Write byte.
		if err := d.Tx([]byte{addr, v}, nil); err != nil {
			return fmt.Errorf("eeprom: error writing %#x to byte at %#x: %s", v, addr, err)
		}
		// Read byte back, looping until the chip is ready (it takes a while to complete
		// the write (5ms max).
		for {
			err := d.Tx([]byte{addr}, oneByte[:])
			if err == nil {
				break
			}
			return fmt.Errorf("eeprom: error reading byte written: %s", err)
		}
	}

	// Pick a page in the first 256 bytes and try to write it and read it back.
	// Using a random page for "wear leveling" and randomizing what gets written
	// so it actually changes from one test run to the next.
	addr = byte(rand.Intn(256)) & 0xf0 // round to page boundary
	r := byte(rand.Intn(256))          // randomizer for value written
	log.Printf("spi-smoke writing&reading EEPROM page %#x with randomizer %#x", addr, r)
	// val calculates the value for byte i.
	val := func(i int) byte { return byte((i<<4)|(16-i)) ^ r }
	for i := 0; i < 16; i++ {
		onePage[i] = val(i)
	}
	// Write page.
	if err := d.Tx(append([]byte{addr}, onePage[:]...), nil); err != nil {
		return fmt.Errorf("eeprom: error writing to page %#x: %s", addr, err)
	}
	// Read page back, looping until the chip is ready (it takes a while to complete
	// the write. Start by zeroing the buffer.
	for i := 0; i < 16; i++ {
		onePage[i] = 0
	}
	for {
		err := d.Tx([]byte{addr}, onePage[:])
		if err == nil {
			break
		}
		return fmt.Errorf("eeprom: error reading page written: %s", err)
	}
	// Ensure we got the correct data.
	for i := 0; i < 16; i++ {
		if onePage[i] != val(i) {
			return fmt.Errorf("eeprom: incorrect read of addr %#x: expected %#x got %#x",
				addr+byte(i), val(i), onePage[i])
		}

	}

	return nil
}
