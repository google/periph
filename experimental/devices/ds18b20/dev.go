// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ds18b20

import (
	"errors"

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
	if err := d.onewire.TxPup([]byte{0x44}, nil); err != nil {
		return 0, err
	}
	conversionSleep(d.resolution)
	return d.LastTemp()
}

// TemperatureFloat performs a conversion and returns the temperature.
func (d *Dev) TemperatureFloat() (float64, error) {
	t, err := d.Temperature()
	if err != nil {
		return 0.0, err
	}
	return t.Float64(), nil
}

// LastTemp reads the temperature resulting from the last conversion from the device.
// It is useful in combination with ConvertAll.
func (d *Dev) LastTemp() (devices.Celsius, error) {
	// Read the scratchpad memory.
	spad := make([]byte, 9)
	if err := d.onewire.Tx([]byte{0xbe}, spad); err != nil {
		return 0, err
	}

	// Check the scratchpad CRC.
	if !onewire.CheckCRC(spad) {
		return 0, errors.New("incorrect DS18B20 scratchpad CRC")
	}

	// spad[1] is MSB, spad[0] is LSB and has 4 fractional bits. Need to do sign extension
	// multiply by 1000 to get devices.Millis, divide by 16 due to 4 fractional bits.
	// Datasheet p.4.
	return (devices.Celsius(int8(spad[1]))<<8 + devices.Celsius(spad[0])) * 1000 / 16, nil
}

// LastTempFloat reads the temperature resulting from the last conversion from the device.
// It is useful in combination with ConvertAll.
func (d *Dev) LastTempFloat() (float64, error) {
	t, err := d.LastTemp()
	if err != nil {
		return 0.0, err
	}
	return t.Float64(), nil
}
