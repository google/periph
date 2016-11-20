// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package ds18b20 interfaces to Dallas Semi / Maxim DS18B20 and MAX31820
// 1-wire temperature sensors.
//
// Note that both DS18B20 and MAX31820 use family code 0x28.
//
// Both powered sensors and parasitically powered sensors are supported
// as long as the bus driver can provide sufficient power using an active
// pull-up.
//
// The DS18B20 alarm functionality and reading/writing the 2 alarm bytes in
// the EEPROM are not supported. The DS18S20 is also not supported.
//
// Datasheets
//
// https://datasheets.maximintegrated.com/en/ds/DS18B20-PAR.pdf
//
// http://datasheets.maximintegrated.com/en/ds/MAX31820.pdf
package ds18b20

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/periph/experimental/conn/onewire"
)

// New returns an object that communicates over 1-wire to the DS18B20 sensor with the
// specified 64-bit address.
//
// resolutionBits must be in the range 9..12 and determines how many bits of precision
// the readings have. The resolution affects the conversion time: 9bits:94ms, 10bits:188ms,
// 11bits:375ms, 12bits:750ms.
// A resolution of 10 bits corresponds to 0.25C and tends to be a good compromise between
// conversion time and the device's inherent accuracy of +/-0.5C.
func New(o onewire.Bus, addr onewire.Address, resolutionBits int) (*Dev, error) {
	if resolutionBits < 9 || resolutionBits > 12 {
		return nil, errors.New("ds18b20: invalid resolutionBits")
	}

	d := &Dev{onewire: onewire.Dev{Bus: o, Addr: addr}, resolution: resolutionBits}

	// Start by reading the scratchpad memory, this will tell us whether we can talk to the
	// device correctly and also how it's configured.
	spad := make([]byte, 9)
	if err := d.onewire.Tx([]byte{0xbe}, spad); err != nil {
		return nil, err
	}

	// Check the scratchpad CRC.
	if !onewire.CheckCRC(spad) {
		fmt.Fprintf(os.Stderr, "ds18b20: bad CRC %#x != %#x %+v\n",
			onewire.CalcCRC(spad[0:8]), spad[8], spad)
		for _, s := range spad {
			if s != 0xff {
				return nil, errors.New("ds18b20: incorrect scratchpad CRC")
			}
		}
		return nil, errors.New("ds18b20: device did not respond")
	}

	// Change the resolution, if necessary (datasheet p.6).
	if int(spad[4]>>5) != resolutionBits-9 {
		// Set the value in the configuration register.
		d.onewire.Tx([]byte{0x4e, 0, 0, byte((resolutionBits-9)<<5) | 0x1f}, nil)
		// Copy the scratchpad to EEPROM to save the values.
		d.onewire.TxPower([]byte{0x48}, nil)
		// Wait for the write to complete
		time.Sleep(10 * time.Millisecond)
	}

	return d, nil
}

// ConvertAll performs a conversion on all DS18B20 devices on the bus.
//
// During the conversion it places the bus in strong pull-up mode to
// power parasitic devices and returns when the conversions have
// completed. This time period is determined by the maximum
// resolution of all devices on the bus and must be provided.
//
// ConvertAll uses time.Sleep to wait for the conversion to finish,
// which takes from 94ms to 752ms.
func ConvertAll(o onewire.Bus, maxResolutionBits int) error {
	if maxResolutionBits < 9 || maxResolutionBits > 12 {
		return errors.New("ds18b20: invalid maxResolutionBits")
	}
	o.Tx([]byte{0x44}, nil, onewire.StrongPullup)
	conversionSleep(maxResolutionBits)
	return nil
}

//

// conversionSleep sleeps for the time a conversion takes, which depends
// on the resolution:
// 9bits:94ms, 10bits:188ms, 11bits:376ms, 12bits:752ms, datasheet p.6.
func conversionSleep(bits int) {
	time.Sleep((94 << uint(bits-9)) * time.Millisecond)
}
