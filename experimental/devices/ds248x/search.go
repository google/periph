// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ds248x

import (
	"fmt"
	"os"

	"github.com/google/periph/experimental/conn/onewire"
)

// Search performs a "search" cycle on the 1-wire bus and returns
// the addresses of all devices on the bus if alarmOnly is false and of all devices in
// alarm state if alarmOnly is true.
//
// The addresses are returned as 64-bit integers with the family code in the lower byte and the CRC
// byte in the top byte.
//
// If an error occurs during the search the already-discovered devices are returned with the error.
//
// For a description of the search algorithm, see Maxim's AppNote 187
// https://www.maximintegrated.com/en/app-notes/index.mvp/id/187
func (d *Dev) Search(alarmOnly bool) ([]uint64, error) {
	// Loop to do the search. Each iteration detects one device
	var devices []uint64  // devices we're finding
	lastDiscrepancy := -1 // how far we need to repeat the same ID in the next iteration
	var lastDevice uint64 // ID of last device found
	for {
		// Reset the bus and if there is no device present (or an error) just return.
		if present, err := d.reset(); !present {
			return nil, err
		}

		// Issue a search command.
		cmd := byte(0xf0) // plain search
		if alarmOnly {
			cmd = 0xec // alarm search
		}
		d.i2cTx([]byte{cmd1WWrite, cmd}, nil)
		d.waitIdle(8 * d.tSlot)

		// Loop to accumulate the 64 bits of an ID.
		discrepancy := -1   // how far we need to repeat the same ID in this iteration
		var device uint64   // ID of current device
		var idBytes [8]byte // ID of device as bytes
		for bit := 0; bit < 64; bit++ {
			//fmt.Fprintf(os.Stderr, "*** bit loop %d\n", bit)
			// Decide which direction to search into: 0 or 1.
			var dir byte
			if bit < lastDiscrepancy {
				// We haven't reached the last discrepancy yet, so we need to
				// repeat the bits of the last device.
				dir = byte((lastDevice >> uint8(bit)) & 1)
			} else if bit == lastDiscrepancy {
				// We reached the bit where we picked 0 last time and now we need 1.
				dir = 1
			}

			// Perform triplet operation and do an explicit error check so we abort the
			// search on error.
			status := d.searchTriplet(dir)
			if d.err != nil {
				return devices, d.err
			}
			gotZero := (status & 0x20) == 0 // some device with 0 in its ID responded
			gotOne := (status & 0x40) == 0  // some device with 1 in its ID responded
			taken := status >> 7            // direction taken

			// Check for the absence of devices on the bus. This is a 1-wire bus error
			// condition and we return a partial result.
			if !gotZero && !gotOne {
				return devices,
					OneWireBusError("1-wire: devices disappeared during search")
			}

			// Check whether we have devices responding for 0 and 1 and we picked 0.
			if gotZero && gotOne && taken == 0 {
				discrepancy = bit
			}

			// Shift a bit into the current device's ID
			device = device | (uint64(taken) << uint(bit))

			// If we got a full byte then save it for CRC calculation.
			if bit&7 == 7 {
				idBytes[bit>>3] = byte(device >> uint(bit-7))
			}

			//fmt.Fprintf(os.Stderr, "Bit %2d: dir=%d 0=%t 1=%t ID=%#x\n",
			//	bit, dir, gotZero, gotOne, device)
		}

		// Verify the CRC and record device if we got it right.
		if onewire.CheckCRC(idBytes[:]) {
			devices = append(devices, device)
			lastDevice = device
			lastDiscrepancy = discrepancy
			if lastDiscrepancy == -1 {
				return devices, nil // we reached the last device
			}
		} else {
			// CRC error: return partial result. This is a transient error.
			fmt.Fprintf(os.Stderr, "CRC failed on %+v\n", idBytes)
			return devices, OneWireBusError("1-wire CRC error during search")
		}
	}
}

// searchTriplet performs a single bit search triplet command on the bus, waits for it to complete
// and returs the status register, which contains the result. It uses the persistent error model and
// returns a status of 0 on error.
func (d *Dev) searchTriplet(direction byte) byte {
	// Send one-wire triplet command.
	var dir byte
	if direction != 0 {
		dir = 0x80
	}
	d.i2cTx([]byte{cmd1WTriplet, dir}, nil)
	return d.waitIdle(0 * d.tSlot) // in theory 3*tSlot but it's actually overlapped
}
