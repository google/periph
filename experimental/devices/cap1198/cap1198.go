// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package cap1198 controls a Microchip cap1198 device over I²C.

// The device is a 8 Channel Capacitive Touch Sensor with 8 LED Drivers
//
// Datasheet
//
// The official data sheet can be found here:
//
// http://ww1.microchip.com/downloads/en/DeviceDoc/CAP1188%20.pdf
//
package cap1198

import (
	"errors"
	"fmt"
	"sync"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/devices"
)

var (
	inputStatuses [8]bool
)

// Dev is a handle to a cap1198.
type Dev struct {
	d     conn.Conn
	isSPI bool
	opts  *Opts
	mu    sync.Mutex
}

func (d *Dev) String() string {
	return fmt.Sprintf("CAP1198{%s}", d.d)
}

// Halt is a noop for the CAP1198.
func (d *Dev) Halt() error {
	return nil
}

// Reset issues a soft reset to the device.
func (d *Dev) Reset() error {
	panic("not implemented")

	return nil
}

// Standby turns off the capacitive touch sensor inputs. The status registers will
// not be cleared until read. LEDs that are linked to capacitive touch sensor
// inputs will remain linked and active. Sensor inputs that are no longer
// sampled will flag a release and then remain in a non-touched state. LEDs that
// are manually controlled will be unaffected.
func (d *Dev) Standby() error {
	panic("not implemented")

	return nil
}

// WakeUp takes the device out of standby or deep sleep mode.
func (d *Dev) WakeUp() error {
	panic("not implemented")
}

// DeepSleep puts the device in a deep sleep mode. All sensor input scanning is
// disabled. All LEDs are driven to their programmed non-actuated
// state and no PWM operations will be done.
func (d *Dev) DeepSleep() error {
	panic("not implemented")
}

// InputStatus reads and returns the status of the 8 inputs as an array where
// each entry indicates a touch event or not.
func (d *Dev) InputStatus() [8]bool {
	status := make([]byte, 1)
	d.readReg(0x3, status)
	for i := uint8(0); i < 8; i++ {
		inputStatuses[i] = isBitSet(status[0], 7-i)
	}
	return inputStatuses
}

// NewI2C returns a new device that communicates over I²C to CAP1198.
func NewI2C(b i2c.Bus, opts *Opts) (*Dev, error) {
	addr := uint16(0x28)
	if opts != nil {
		switch opts.Address {
		case 0x28, 0x29, 0x2a, 0x2b, 0x2c:
			addr = opts.Address
		case 0x00:
			// do not do anything
		default:
			return nil, errors.New("cap1198: given address not supported by device")
		}
	}
	d := &Dev{d: &i2c.Dev{Bus: b, Addr: addr}, opts: opts, isSPI: false}
	if err := d.makeDev(opts); err != nil {
		return nil, err
	}
	return d, nil
}

// NewSPI returns an object that communicates over SPI to CAP1198 environmental
// sensor.
func NewSPI(p spi.Port, opts *Opts) (*Dev, error) {
	panic("not implemented")
}

func (d *Dev) makeDev(opts *Opts) error {
	// Use default options if none are passed.
	if opts == nil {
		opts = DefaultOpts()
	}
	d.opts = opts

	var productID [1]byte
	// Read register 0xFD to read the product id.
	if err := d.readReg(0xFD, productID[:]); err != nil {
		return fmt.Errorf("failed to read product id - %s", err)
	}
	if productID[0] != 0x50 {
		return fmt.Errorf("cap1198: unexpected chip id %x; is this a CAP1198?", productID[0])
	}
	// manufacturer ID on 0xFE, should be 0x5D
	// revision ID on 0xFF, should be 0x83

	var recalFlag byte
	if opts.EnableRecalibration {
		recalFlag = 1
	}
	var intOnRel byte
	if !opts.InterruptOnRelease {
		intOnRel = 1
	}
	// page 47 - configuration registers
	config := []byte{
		// config 1
		0x20, (
		// Timeout: Enables the timeout and idle functionality of the SMBus protocol.
		// default 0: The SMBus timeout and idle functionality are disabled. The
		// SMBus interface will not time out if the clock line is held low.
		// Likewise, it will not reset if both data and clock lines are held
		// high for longer than 200us
		byte(0)<<7 |
			// Configures the operation of the WAKE pin.
			// default 0: The WAKE pin is not asserted when a touch is detected while the
			// device is in Standby. It will still be used to wake the device from
			// Deep Sleep when driven high.
			byte(0)<<6 |
			// digital noise threshold
			// Determines whether the digital noise threshold is used by the device.
			// default 1:  The noise threshold is disabled. Any delta count that
			// is less than the touch threshold is used for the automatic
			// re-calibration routine.
			byte(1)<<5 |
			// analog noise filter
			// default 0: Determines whether the analog noise filter is enabled. Setting this
			// bit disables the feature.
			byte(0)<<4 |
			// maximum duration recalibration
			// Determines whether the maximum duration recalibration is enabled.
			//
			// if 0, the maximum duration recalibration functionality is
			// disabled. A touch may be held indefinitely and no re-calibration
			// will be performed on any sensor input.
			//
			// if 1, The maximum duration recalibration functionality is
			// enabled. If a touch is held for longer than the
			// opts.MaxTouchDuration, then the re-calibration routine will be
			// restarted
			recalFlag<<3),
		//config 2
		0x44, (
		// Linked LED Transition controls
		// default 0: The Linked LED Transition controls set the min duty
		// cycle equal to the max duty cycle
		byte(0)<<7 |
			// Determines the ALERT# pin polarity and behavior.
			// default 1: the ALERT# pin is active low and open drain.
			byte(1)<<6 |
			// Determines whether the device will reduce power consumption
			// while waiting between conversion time completion and the end of
			// the polling cycle.
			// default 0: The device will always power down as much as possible
			// during the time between the end of the last conversion and the
			// end of the polling cycle.
			byte(0)<<5 |
			//  Determines whether the LED Mirror Control register bits are
			//  linked to the LED Polarity bits. Setting this bit blocks the
			//  normal behavior which is to automatically set and clear the LED
			//  Mirror Control bits when the LED Polarity bits are set or
			//  cleared.
			//  default 0: When the LED Polarity controls are set, the
			//  corresponding LED Mirror control is automatically set. Likewise,
			//  when the LED Polarity controls are cleared, the corresponding
			//  LED Mirror control is also cleared.
			byte(0)<<4 |
			// Determines whether the Noise Status bits will show RF Noise as
			// the only input source.
			// default 0: The Noise Status registers will show both RF noise and
			// low frequency EMI noise if either is detected on a capacitive
			// touch sensor input.
			byte(0)<<3 |
			// Determines whether the RF noise filter is enabled.
			// default 0:  If RF noise is detected by the analog block, the
			// delta count on the corresponding channel is set to 0. Note that
			// this does not require that Noise Status bits be set.
			byte(0)<<2 |
			// Controls the interrupt behavior when a release is detected on a button.
			// when 0: An interrupt is generated when a press is detected and
			// again when a release is detected and at the repeat rate (if enabled)
			// when 1:  An interrupt is generated when a press is detected and
			// at the repeat rate but not when a release is detected.
			intOnRel<<0),
		// enable all inputs
		0x21, 1,
		// enable interrupts
		0x27, 1,
		// customize multi touch (TODO)
		// 0x2a, TODO
		// customize sensitivity (TODO)
		0x1F, (byte(0)<<7 |
			// Controls the sensitivity of a touch detection. The sensitivity settings act
			// to scale the relative delta count value higher or lower based on the system parameters. A setting of
			// 000b is the most sensitive while a setting of 111b is the least sensitive. At the more sensitive settings,
			// touches are detected for a smaller delta capacitance corresponding to a “lighter” touch. These settings
			// are more sensitive to noise, however, and a noisy environment may flag more false touches with higher
			// sensitivity levels.
			// Set to 2x: TODO: make that configurable.
			byte(1)<<6 | byte(1)<<5 | byte(0)<<4 |
			byte(1)<<3 | byte(1)<<2 | byte(1)<<1 | byte(1)<<0),
		// TODO: Averaging and Sampling Config
	}

	if err := d.writeCommands(config[:]); err != nil {
		return err
	}

	return nil
}

func (d *Dev) readReg(reg uint8, b []byte) error {
	if d.isSPI {
		return fmt.Errorf("SPI isn't implemented yet")
	}
	return d.d.Tx([]byte{reg}, b)
}

// writeCommands writes a command to the device.
func (d *Dev) writeCommands(b []byte) error {
	if d.isSPI {
		return fmt.Errorf("SPI isn't implemented yet")
	}
	return d.d.Tx(b, nil)
}

// b is the byte to check and position is the bit position
// index 0 where 7 is the "most left bit".
func isBitSet(b byte, pos uint8) bool {
	return (b>>pos)&1 == 1
}

var _ devices.Device = &Dev{}
