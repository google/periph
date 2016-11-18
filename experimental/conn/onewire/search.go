// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package onewire

import (
	"fmt"
	"os"
)

// BusSearcher provides the basic bus transaction necessary to search a 1-wire
// bus for devices. Buses that implement this interface can be searched with the
// Search function.
type BusSearcher interface {
	Bus
	// Triplet performs a single bit search triplet command on the bus,
	// waits for it to complete and returns the result.
	SearchTriplet(direction byte) (TripletResult, error)
}

type TripletResult struct {
	GotZero bool  // a device with a zero in the current bit position responded
	GotOne  bool  // a device with a one in the current bit position responded
	Taken   uint8 // direction taken: 0 or 1
}

// Search performs a "search" cycle on the 1-wire bus and returns
// the addresses of all devices on the bus if alarmOnly is false and of all
// devices in alarm state if alarmOnly is true.
//
// If an error occurs during the search the already-discovered devices are
// returned with the error.
//
// For a description of the search algorithm, see Maxim's AppNote 187
// https://www.maximintegrated.com/en/app-notes/index.mvp/id/187
//
// This function is defined here so the implementation of buses that support the
// BusSearcher interface can call it. Applications should call Bus.Search.
func Search(bus BusSearcher, alarmOnly bool) ([]Address, error) {
	var devices []Address // devices we're finding
	lastDiscrepancy := -1 // how far we need to repeat the same ID in the next iteration
	var lastDevice uint64 // ID of last device found

	// Loop to do the search. Each iteration detects one device.
	for {
		// Issue a search command.
		cmd := byte(0xf0) // plain search
		if alarmOnly {
			cmd = 0xec // alarm search
		}
		err := bus.Tx([]byte{cmd}, nil, WeakPullup)
		// Expect an NoDevicesError if no device is present on the bus and
		// pass that back.
		if err != nil {
			return devices, err
		}

		// Loop to accumulate the 64 bits of an ID.
		discrepancy := -1   // how far we need to repeat the same ID in this iteration
		var device uint64   // ID of current device
		var idBytes [8]byte // ID of device as bytes
		for bit := 0; bit < 64; bit++ {
			//fmt.Fprintf(os.Stderr, "*** bit loop %d\n", bit)
			// Decide which direction to search into: 0 or 1.
			var dir byte
			if bit < lastDiscrepancy {
				// We haven't reached the last discrepancy yet, so we
				// need to repeat the bits of the last device.
				dir = byte((lastDevice >> uint8(bit)) & 1)
			} else if bit == lastDiscrepancy {
				// We reached the bit where we picked 0 last time and
				// now we need 1.
				dir = 1
			}

			// Perform triplet operation and do an explicit error check so
			// we abort the search on error.
			result, err := bus.SearchTriplet(dir)
			if err != nil {
				return devices, err
			}

			// Check for the absence of devices on the bus. This is a 1-wire
			// bus error condition and we return a partial result.
			if !result.GotZero && !result.GotOne {
				return devices, fmt.Errorf("1-wire: devices disappeared during search")
			}

			// Check whether we have devices responding for 0 and 1 and we
			// picked 0.
			if result.GotZero && result.GotOne && result.Taken == 0 {
				discrepancy = bit
			}

			// Shift a bit into the current device's ID
			device |= uint64(result.Taken) << uint(bit)

			// If we got a full byte then save it for CRC calculation.
			if bit&7 == 7 {
				idBytes[bit>>3] = byte(device >> uint(bit-7))
			}

			//fmt.Fprintf(os.Stderr, "Bit %2d: dir=%d 0=%t 1=%t ID=%#x\n",
			//	bit, dir, gotZero, gotOne, device)
		}

		// Verify the CRC and record device if we got it right.
		if !CheckCRC(idBytes[:]) {
			// CRC error: return partial result.  This is a transient error.
			e := fmt.Sprintf("1-wire: CRC error during search, addr=%+v", idBytes)
			fmt.Fprintln(os.Stderr, e)
			return devices, busError(e)
		}
		devices = append(devices, Address(device))
		lastDevice = device
		lastDiscrepancy = discrepancy
		if lastDiscrepancy == -1 {
			return devices, nil // we reached the last device
		}
	}
}
