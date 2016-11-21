// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ds18b20

import (
	"github.com/google/periph/devices"
	"github.com/google/periph/experimental/conn/onewire"
)

// Dev is a handle to a Dallas Semi / Maxim DS18B20 temperature sensor on a 1-wire bus.
type Dev struct {
	onewire    onewire.Dev // device on 1-wire bus
	resolution int         // resolution in bits (9..12)
}

// Temperature performs a conversion and returns the temperature.
func (d *Dev) Temperature() (devices.Celsius, error) {
	if err := d.onewire.TxPower([]byte{0x44}, nil); err != nil {
		return 0, err
	}
	conversionSleep(d.resolution)
	return d.LastTemp()
}

// LastTemp reads the temperature resulting from the last conversion from the device.
// It is useful in combination with ConvertAll.
func (d *Dev) LastTemp() (devices.Celsius, error) {
	// Read the scratchpad memory.
	var spad [9]byte
	if err := d.onewire.Tx([]byte{0xbe}, spad[:]); err != nil {
		return 0, err
	}

	// Check the scratchpad CRC.
	if !onewire.CheckCRC(spad[:]) {
		for _, s := range spad {
			if s != 0xff {
				return 0, busError("ds18b20: incorrect scratchpad CRC")
			}
		}
		return 0, busError("ds18b20: device did not respond")
	}

	// spad[1] is MSB, spad[0] is LSB and has 4 fractional bits. Need to do sign extension
	// multiply by 1000 to get devices.Millis, divide by 16 due to 4 fractional bits.
	// Datasheet p.4.
	c := (devices.Celsius(int8(spad[1]))<<8 + devices.Celsius(spad[0])) * 1000 / 16

	// The device powers up with a value of 85 degrees C, so if we read that odds are very high
	// that either no conversion was performed or that the covnersion falied due to lack of
	// power.
	if c == 85000 {
		return 0, busError("ds18b20: has not performed a temperature conversion (insufficient pull-up?)")
	}

	return c, nil
}

//

// busError implements error and onewire.BusError.
type busError string

func (e busError) Error() string  { return string(e) }
func (e busError) BusError() bool { return true }
