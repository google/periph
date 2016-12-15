// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// onewiresmoketest verifies that a 1-wire bus search returns two devices, that a ds18b20
// temperature sensor can be read, and that a ds2431 eeprom can be written and read.
//
// This assumes the presence of the periph-tester board, which includes these two devices.
// See https://github.com/tve/periph-tester
package onewiresmoketest

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/periph/conn/i2c"
	"github.com/google/periph/experimental/conn/onewire"
	"github.com/google/periph/experimental/devices/ds18b20"
	"github.com/google/periph/experimental/devices/ds248x"
)

type SmokeTest struct {
}

// Name implements the SmokeTest interface.
func (s *SmokeTest) Name() string {
	return "onewire-testboard"
}

// Description implements the SmokeTest interface.
func (s *SmokeTest) Description() string {
	return "Tests DS18B20 temp sensor and DS2431 EEPROM on periph-tester board"
}

// Run implements the SmokeTest interface.
func (s *SmokeTest) Run(args []string) error {
	f := flag.NewFlagSet("onewire", flag.ExitOnError)
	busNum := f.Int("i2cbus", -1, "bus number for the DS2483 1-wire interface chip, -1 for lowest numbered bus")
	seed := f.Int64("seed", 0, "random number seed, default is to use the time")
	f.Parse(args)

	// Open the i2c bus where the DS2483 is located.
	i2cBus, err := i2c.New(*busNum)
	if err != nil {
		return fmt.Errorf("onewire-smoke: cannot open i2c bus: %v", err)
	}
	defer i2cBus.Close()

	// Open the ds2483 one-wire interface chip.
	onewireBus, err := ds248x.New(i2cBus, nil)
	if err != nil {
		return fmt.Errorf("onewire-smoke: cannot open DS248x: %v", err)
	}
	defer onewireBus.Close()

	// Init rand.
	if *seed == 0 {
		*seed = time.Now().UnixNano()
	}
	rand.Seed(*seed)
	log.Printf("onewire-smoke: random number seed %d", *seed)

	// Run the tests.
	addrs, err := s.search(onewireBus)
	if err != nil {
		return fmt.Errorf("onewire-smoke: %s", err)
	}
	if err := s.ds18b20(onewireBus, addrs[0]); err != nil {
		return fmt.Errorf("onewire-smoke: %s", err)
	}
	if err := s.eeprom(onewireBus, addrs[1]); err != nil {
		return fmt.Errorf("onewire-smoke: %s", err)
	}

	return nil
}

// search performs a search cycle on the bus and verifies that the two expected devices
// are actually found. It returns the two device addresses, ds18b20 first.
func (s *SmokeTest) search(bus onewire.Bus) ([]onewire.Address, error) {
	addrs, err := bus.Search(false)
	if err != nil {
		return nil, fmt.Errorf("search failed: %v", err)
	}

	if len(addrs) != 2 {
		return nil, fmt.Errorf("search expected 2 devices, found %d", len(addrs))
	}

	// Ensure we found devices with the correct family code and return them.
	if addrs[1]&0xff == 0x28 && addrs[0]&0xff == 0x2D {
		// Swap the order so the DS18b20 is first.
		addrs[0], addrs[1] = addrs[1], addrs[0]
	}
	if addrs[0]&0xff == 0x28 && addrs[1]&0xff == 0x2D {
		log.Printf("onewire-smoke: found 2 devices %#x %#x", addrs[0], addrs[1])
		return addrs, nil
	}
	return nil, fmt.Errorf("search expected device families 0x28 and 0x2D, found: %#x %#x", addrs[0], addrs[1])
}

// ds18b20 tests a Maxim DS18B20 (or MAX31820) 1-wire temperature sensor attached to the
// 1-wire bus. Such a chip is included on the periph-tester board.
func (s *SmokeTest) ds18b20(bus onewire.Bus, addr onewire.Address) error {
	dev, err := ds18b20.New(bus, addr, 10)
	if err != nil {
		return err
	}

	t, err := dev.Temperature()
	if err != nil {
		return err
	}
	if t > 50*1000 || t <= 0 {
		return fmt.Errorf("ds18b20: expected temperature in the 0°C..50°C range, got %.2f°C", t.Float64())
	}

	log.Printf("onewire-smoke: temperature is %.2f°C", t.Float64())
	return nil
}

// eeprom tests a ds2431 1Kbit 1-wire EEPROM.
// Such a chip is included on the periph-tester board.
//
// The test currently only writes and reads the scratchpad memory.
// A test of the eeprom itself may be useful if a proper driver is written
// someday. But it's not like that would add any significant additional
// test coverage...
//
// Datasheet: http://datasheets.maximintegrated.com/en/ds/DS2431.pdf
func (s *SmokeTest) eeprom(bus onewire.Bus, addr onewire.Address) error {
	d := onewire.Dev{Bus: bus, Addr: addr}

	// Start by writing some data to the scratchpad
	var data [8]byte
	for i := range data {
		data[i] = byte(rand.Intn(256))
	}
	var buf [13]byte // cmd, target-addr-low, target-addr-hi, data[8], crc16
	buf[0] = 0x0f    // write scratchpad
	copy(buf[3:11], data[:])
	if err := d.Tx(buf[:], nil); err != nil {
		return fmt.Errorf("eeprom: error on the first scratchpad write")
	}

	// Read the scratchpad back
	if err := d.Tx([]byte{0xaa}, buf[:]); err != nil {
		return fmt.Errorf("eeprom: error reading the scratchpad")
	}
	for i := range data {
		if data[i] != buf[i+3] {
			return fmt.Errorf("eeprom: scratchpad data byte %d mismatch, expected %#x got %#x\n",
				i, data[i], buf[i+3])
		}
	}
	log.Printf("onewire-smoke: eeprom test successful")
	return nil
}
