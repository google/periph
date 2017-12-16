// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package cap1188 controls a Microchip cap1188 device over I²C.

// The device is a 8 Channel Capacitive Touch Sensor with 8 LED Drivers
//
// Datasheet
//
// The official data sheet can be found here:
//
// http://ww1.microchip.com/downloads/en/DeviceDoc/CAP1188%20.pdf
//
package cap1188

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/mmr"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/devices"
)

var (
	// Debug is a debugging flag to turn on extra logging
	Debug         bool
	inputStatuses []TouchStatus
)

// TouchStatus is the status of an input sensor
type TouchStatus int8

func (t TouchStatus) String() string {
	switch t {
	case OffStatus:
		return "Off    "
	case PressedStatus:
		return "Pressed"
	case HeldStatus:
		return "Held   "
	case ReleasedStatus:
		return "Released"
	default:
		return "Unknown "
	}
}

const (
	// OffStatus indicates that the input sensor isn't being activated
	OffStatus TouchStatus = iota
	// PressedStatus indicates that the input sensor is currently pressed
	PressedStatus
	// HeldStatus indicates that the input sensor was pressed and is still held pressed
	HeldStatus
	// ReleasedStatus indicates that the input sensor was pressed and is being released
	ReleasedStatus
)

const (
	nbrOfLEDs = 8
)

const (
	// reg_LEDLinking - The Sensor Input LED Linking register controls whether a
	// capacitive touch sensor input is linked to an LED output. If the
	// corresponding bit is set, then the appropriate LED output will change
	// states defined by the LED Behavior controls in response to the capacitive
	// touch sensor input.
	reg_LEDLinking = 0x72
	// reg_LEDOutputControl - The LED Output Control Register controls the output
	// state of the LED pins that are not linked to sensor inputs.
	reg_LEDOutputControl = 0x74
)

// Dev is a handle to a cap1188.
type Dev struct {
	d          conn.Conn
	regWrapper mmr.Dev8
	isSPI      bool
	opts       *Opts
	mu         sync.Mutex
}

func (d *Dev) String() string {
	return fmt.Sprintf("cap1188{%s}", d.d)
}

// Halt is a noop for the cap1188.
func (d *Dev) Halt() error {
	return nil
}

// InputStatus reads and returns the status of the 8 inputs as an array where
// each entry indicates a touch event or not.
func (d *Dev) InputStatus() ([]TouchStatus, error) {
	// read inputs
	status, err := d.regWrapper.ReadUint8(0x3)
	if err != nil {
		return inputStatuses, wrap(fmt.Errorf("failed to read the input values - %s", err))
	}

	// read deltas (in two's complement, capped at -128 to 127)
	deltasB := [nbrOfLEDs]byte{}
	if err = d.regWrapper.ReadStruct(0x10, &deltasB); err != nil {
		return inputStatuses, wrap(fmt.Errorf("failed to read the delta values - %s", err))
	}
	deltas := [nbrOfLEDs]int{}
	for i, b := range deltasB {
		deltas[i] = twoComplementsToInt(b)
	}

	// read threshold
	thresholds := [nbrOfLEDs]byte{}
	if err = d.regWrapper.ReadStruct(0x30, &thresholds); err != nil {
		return inputStatuses, wrap(fmt.Errorf("failed to read the threshold values - %s", err))
	}

	// convert the data into a sensor state
	var touched bool
	for i := uint8(0); i < nbrOfLEDs; i++ {
		touched = isBitSet(status, 7-i)

		// if touched {
		// check that we passed the threshold
		// if deltas[i] == 127 && Debug {
		// 	fmt.Println("sensitivity settings are too sensitive")
		// }
		// else {
		// 	if deltas[i] > int(thresholds[i]) {
		// 		if inputStatuses[i] == PressedStatus {
		// 			inputStatuses[i] = HeldStatus
		// 		} else {
		// 			inputStatuses[i] = PressedStatus
		// 		}
		// 		continue
		// 	}
		// }

		if touched {
			if inputStatuses[i] == PressedStatus {
				if d.opts.RetriggerOnHold {
					inputStatuses[i] = HeldStatus
				}
				continue
			}
			inputStatuses[i] = PressedStatus
		} else {
			inputStatuses[i] = OffStatus
		}
	}

	return inputStatuses, nil
}

// LinkLeds() link the behavior of the LEDs to the touch sensors.
// Doing so, disabled the option for the host to set specific LEDs on/off.
func (d *Dev) LinkLeds() error {
	if err := d.regWrapper.WriteUint8(reg_LEDLinking, 0xff); err != nil {
		return wrap(fmt.Errorf("failed to link LEDs - %s", err))
	}
	d.opts.LinkedLEDs = true
	return nil
}

// UnlinkLeds disassociate the LEDs from the input sensors allowing the host to
// control the LEDs.
func (d *Dev) UnlinkLeds() error {
	if err := d.regWrapper.WriteUint8(reg_LEDLinking, 0x00); err != nil {
		return wrap(fmt.Errorf("failed to unlink LEDs - %s", err))
	}
	d.opts.LinkedLEDs = false
	return nil
}

// AllLedsOn turns all the LEDs on.
// This is quite more efficient than looping through each led and turn them on.
func (d *Dev) AllLedsOn() error {
	if d.opts.LinkedLEDs {
		return wrap(fmt.Errorf("can't manually set LEDs when they are linked to sensors"))
	}
	return d.regWrapper.WriteUint8(reg_LEDOutputControl, 0xff)
}

// AllLedsOff turns all the LEDs off.
// This is quite more efficient than looping through each led and turn them off.
func (d *Dev) AllLedsOff() error {
	if d.opts.LinkedLEDs {
		return wrap(fmt.Errorf("can't manually set LEDs when they are linked to sensors"))
	}
	return d.regWrapper.WriteUint8(reg_LEDOutputControl, 0x00)
}

// SetLed sets the state of a LED as on or off
// Only works if the LEDs are not linked to the sensors
func (d *Dev) SetLed(idx int, state bool) error {
	if d.opts.LinkedLEDs {
		return wrap(fmt.Errorf("can't manually set LEDs when they are linked to sensors"))
	}
	if idx > 7 || idx < 0 {
		return wrap(fmt.Errorf("invalid led idx %d", idx))
	}
	if Debug {
		fmt.Printf("Set LED state %d - %t\n", idx, state)
	}
	if state {
		return d.setBit(reg_LEDOutputControl, idx)
	}
	return d.clearBit(reg_LEDOutputControl, idx)
}

// Reset issues a soft reset to the device using the reset pin
// if available.
func (d *Dev) Reset() (err error) {
	d.ClearInterrupt()
	if d != nil && d.opts != nil && d.opts.ResetPin != nil {
		if Debug {
			fmt.Println("cap1188: Resetting the device using the reset pin")
		}
		if err = d.opts.ResetPin.Out(gpio.Low); err != nil {
			return err
		}
		if err = d.opts.ResetPin.Out(gpio.High); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
		if err = d.opts.ResetPin.Out(gpio.Low); err != nil {
			return err
		}
	}

	return nil
}

// ClearInterrupt resets the interrupt flag
func (d *Dev) ClearInterrupt() error {
	// clear the main control bit
	return d.clearBit(0x0, 0)
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

// NewI2C returns a new device that communicates over I²C to cap1188.
func NewI2C(b i2c.Bus, opts *Opts) (*Dev, error) {
	addr := uint16(0x28) // default address
	if opts != nil {
		switch opts.Address {
		case 0x28, 0x29, 0x2a, 0x2b, 0x2c:
			addr = opts.Address
		case 0x00:
			// do not do anything
		default:
			return nil, wrap(errors.New("given address not supported by device"))
		}
	}
	if Debug {
		fmt.Printf("cap1188: Connecting via I2C address: %#X\n", addr)
	}
	d := &Dev{d: &i2c.Dev{Bus: b, Addr: addr}, opts: opts, isSPI: false}
	resetMemInputStatus()
	if err := d.makeDev(opts); err != nil {
		return nil, err
	}
	return d, nil
}

// NewSPI returns an object that communicates over SPI to cap1188 environmental
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
	d.regWrapper = mmr.Dev8{Conn: d.d, Order: binary.LittleEndian}

	var productID byte
	var err error
	// Read the product id to confirm it matches our expectations.
	if productID, err = d.regWrapper.ReadUint8(0xFD); err != nil {
		return fmt.Errorf("failed to read product id - %s", err)
	}
	if productID != 0x50 {
		return fmt.Errorf("cap1188: unexpected chip id %x; is this a cap1188?", productID)
	}
	// manufacturer ID on 0xFE, should be 0x5D
	// revision ID on 0xFF, should be 0x83

	// reset the device
	if err = d.Reset(); err != nil {
		return fmt.Errorf("failed to reset the device - %s", err)
	}

	var recalFlag byte
	if opts.EnableRecalibration {
		recalFlag = 1
	}
	var intOnRel byte
	if !opts.InterruptOnRelease {
		intOnRel = 1 // 0 = trigger on release
	}

	// enable all inputs
	if err = d.regWrapper.WriteUint8(0x21, 0xff); err != nil {
		return wrap(fmt.Errorf("failed to enable all inputs - %s", err))
	}
	// enable interrupts
	if err = d.regWrapper.WriteUint8(0x27, 0xff); err != nil {
		return wrap(fmt.Errorf("failed to enable interrupts - %s", err))
	}
	// enable/disable repeats (TODO: make it an option)
	if err = d.regWrapper.WriteUint8(0x28, 0xff); err != nil {
		return fmt.Errorf("failed to disable repeats - %s", err)
	}
	// enable multitouch
	multitouchConfig := (
	// Enables the multiple button blocking circuitry.
	//  ‘0’ - The multiple touch circuitry is disabled. The device will not
	// block multiple touches.
	//  ‘1’ (default) - The multiple touch circuitry is enabled. The device
	// will flag the number of touches equal to programmed multiple touch
	// threshold and block all others. It will remember which sensor inputs
	// are valid and block all others until that sensor pad has been
	// released. Once a sensor pad has been released, the N detected touches
	// (determined via the cycle order of CS1 - CS8) will be flagged and all
	// others blocked.
	byte(0)<<7 |
		byte(0)<<6 | byte(0)<<5 | byte(0)<<4 |
		// Determines the number of simultaneous touches on all sensor pads
		// before a Multiple Touch Event is detected and sensor inputs are blocked.
		// set to 2
		byte(0)<<3 | byte(1)<<2 |
		byte(0)<<1 | byte(0)<<0)
	if err = d.regWrapper.WriteUint8(0x2a, multitouchConfig); err != nil {
		return wrap(fmt.Errorf("failed to enable multitouch - %s", err))
	}
	// Averaging and Sampling Config
	samplingConfig := (byte(0)<<7 |
		// number of samples taken per measurement
		// TODO: use opts.SamplesPerMeasurement
		byte(0)<<6 |
		byte(0)<<5 |
		byte(0)<<4 |
		// sample time
		// TODO: use opts.SamplingTime
		byte(1)<<3 |
		byte(0)<<2 |
		// overall cycle time
		// TODO: use opts.CycleTime
		byte(0)<<1 |
		byte(0)<<0)
	if Debug {
		fmt.Printf("cap1188: Sampling config mask: %08b\n", samplingConfig)
	}
	if err = d.regWrapper.WriteUint8(0x24, samplingConfig); err != nil {
		return fmt.Errorf("failed to enable multitouch - %s", err)
	}

	// customize sensitivity (TODO)
	sensitivity := (byte(0)<<7 |
		// Controls the sensitivity of a touch detection. The sensitivity settings act
		// to scale the relative delta count value higher or lower based on the system parameters. A setting of
		// 000b is the most sensitive while a setting of 111b is the least sensitive. At the more sensitive settings,
		// touches are detected for a smaller delta capacitance corresponding to a “lighter” touch. These settings
		// are more sensitive to noise, however, and a noisy environment may flag more false touches with higher
		// sensitivity levels.
		// Set to 4x: TODO: make that configurable.
		byte(1)<<6 | byte(0)<<5 | byte(1)<<4 |
		byte(0)<<3 | byte(0)<<2 | byte(0)<<1 | byte(0)<<0)
	if Debug {
		fmt.Printf("cap1188: Sensitivity mask: %08b\n", sensitivity)
	}
	if err = d.regWrapper.WriteUint8(0x1F, sensitivity); err != nil {
		return fmt.Errorf("failed to set sensitivity - %s", err)
	}

	if opts.LinkedLEDs {
		if err = d.LinkLeds(); err != nil {
			return err
		}
	}

	if opts.RetriggerOnHold {
		if err = d.regWrapper.WriteUint8(0x28, 0xff); err != nil {
			return fmt.Errorf("failed to set retrigger on hold - %s", err)
		}
	} else {
		if err = d.regWrapper.WriteUint8(0x28, 0x00); err != nil {
			return fmt.Errorf("failed to turn off retrigger on hold - %s", err)
		}
	}

	// page 47 - configuration registers
	config := (
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
		byte(1)<<4 |
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
		recalFlag<<3 |
		byte(0)<<2 |
		byte(0)<<1 |
		byte(0)<<0)
	if Debug {
		fmt.Printf("cap1188: Config mask: %08b\n", config)
	}
	if err = d.regWrapper.WriteUint8(0x20, config); err != nil {
		return fmt.Errorf("failed to set the device configuration - %s", err)
	}

	config2 := (
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
		byte(1)<<5 |
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
		byte(0)<<1 |
		// Controls the interrupt behavior when a release is detected on a button.
		// when 0: An interrupt is generated when a press is detected and
		// again when a release is detected and at the repeat rate (if enabled)
		// when 1:  An interrupt is generated when a press is detected and
		// at the repeat rate but not when a release is detected.
		intOnRel<<0)
	if Debug {
		fmt.Printf("cap1188: Config2 mask: %08b\n", config2)
	}
	if err = d.regWrapper.WriteUint8(0x44, config2); err != nil {
		return fmt.Errorf("failed to set the device configuration 2 - %s", err)
	}

	return nil
}

// setBit sets a specific bit on a register
// TODO: avoid reading before writing, keep states in memory
func (d *Dev) setBit(regID uint8, idx int) error {
	v, err := d.regWrapper.ReadUint8(regID)
	if err != nil {
		return err
	}
	v |= (1 << uint8(idx))
	return d.regWrapper.WriteUint8(regID, v)
}

// clearBit clears a specific bit on a register
// TODO: avoid reading before writing, keep states in memory
func (d *Dev) clearBit(regID uint8, idx int) error {
	v, err := d.regWrapper.ReadUint8(regID)
	if err != nil {
		return err
	}
	v &= ^(1 << uint8(idx))
	return d.regWrapper.WriteUint8(regID, v)
}

// resetMemInputStatus resets the in memory slice containing the state of the
// inputs.
func resetMemInputStatus() {
	inputStatuses = make([]TouchStatus, nbrOfLEDs)
}

// b is the byte to check and position is the bit position
// index 0 where 7 is the "most left bit".
func isBitSet(b byte, pos uint8) bool {
	return (b>>pos)&1 == 1
}

// twoComplementsToInt returns the int value of a standard 2’s complement number
func twoComplementsToInt(v byte) int {
	if (v & (1 << (8 - 1))) == 0 {
		return int(v)
	}
	return int(v) - (1 << 8)
}

func wrap(err error) error {
	return fmt.Errorf("cap1188: %v", err)
}

var _ devices.Device = &Dev{}
