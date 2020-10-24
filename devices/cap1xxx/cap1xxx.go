// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package cap1xxx controls a Microchip
// cap1105/cap1106/cap1114/cap1133/cap1126/cap1128/cap1166/cap1188 device over
// I²C.
//
// The cap1xxx devices are a 3/5/6/8/14 channel capacitive touch sensor with
// 2/3/6/8/11 LED drivers.
//
// Datasheet
//
// 3 sensors, 3 LEDs:
// http://ww1.microchip.com/downloads/en/DeviceDoc/CAP1133.pdf
//
// 5-6 sensors, no LED:
// http://ww1.microchip.com/downloads/en/DeviceDoc/CAP1105_CAP1106.pdf
//
// 6 sensors, 2 LEDs:
// http://ww1.microchip.com/downloads/en/DeviceDoc/CAP1126.pdf
//
// 6 sensors, 6 LEDs:
// http://ww1.microchip.com/downloads/en/DeviceDoc/CAP1166.pdf
//
// 8 sensors, 2 LEDs:
// http://ww1.microchip.com/downloads/en/DeviceDoc/CAP1128.pdf
//
// 8 sensors, 8 LEDs:
// http://ww1.microchip.com/downloads/en/DeviceDoc/CAP1188.pdf
//
// 14 sensors, 11 LEDs:
// http://ww1.microchip.com/downloads/en/DeviceDoc/CAP1114.pdf
package cap1xxx

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/mmr"
)

// TouchStatus is the status of an input sensor.
type TouchStatus int8

const (
	// OffStatus indicates that the input sensor isn't being activated.
	OffStatus TouchStatus = iota
	// PressedStatus indicates that the input sensor is currently pressed.
	PressedStatus
	// HeldStatus indicates that the input sensor was pressed and is still held
	// pressed.
	HeldStatus
	// ReleasedStatus indicates that the input sensor was pressed and is being
	// released.
	ReleasedStatus
)

const touchStatusName = "OffStatusPressedStatusHeldStatusReleasedStatus"

var touchStatusIndex = [...]uint8{0, 9, 22, 32, 46}

func (i TouchStatus) String() string {
	if i < 0 || i >= TouchStatus(len(touchStatusIndex)-1) {
		return "TouchStatus(" + strconv.Itoa(int(i)) + ")"
	}
	return touchStatusName[touchStatusIndex[i]:touchStatusIndex[i+1]]
}

// NewI2C returns a new device that communicates over I²C to
// one of the supported cap1xxx device.
//
// Use default options if nil is used.
func NewI2C(b i2c.Bus, opts *Opts) (*Dev, error) {
	if opts == nil {
		opts = &DefaultOpts
	}
	addr, err := opts.i2cAddr()
	if err != nil {
		return nil, wrapf("%v", err)
	}
	d, err := makeDev(&i2c.Dev{Bus: b, Addr: addr}, false, opts)
	if err != nil {
		return nil, err
	}
	return d, nil
}

/*
// NewSPI returns an object that communicates over SPI to cap1xxx touch
// sensor.
//
// TODO(mattetti): Expose once implemented and tested.
func NewSPI(p spi.Port, opts *Opts) (*Dev, error) {
	return nil, fmt.Errorf("cap1xxx: not implemented")
}
*/

// Dev is a handle to a device of the cap1xxx family.
type Dev struct {
	c     mmr.Dev8
	opts  Opts
	isSPI bool

	inputStatuses []TouchStatus
	numLEDs       int
	lastReset     time.Time
}

func (d *Dev) String() string {
	return fmt.Sprintf("cap1xxx{%s}", d.c.Conn)
}

// Halt is a noop for the cap1xxx.
func (d *Dev) Halt() error {
	// TODO(maruel): Turn off the LEDs?
	return nil
}

// InputStatus reads and returns the status of the inputs.
//
// The slice t will have the sensed inputs updated upon successful read. If the
// slice is too long, extraneous elements are ignored. If the slice is too
// short, only the provided subset is updated without error.
func (d *Dev) InputStatus(t []TouchStatus) error {
	d.resetSinceAtLeast(200 * time.Millisecond)
	// Read inputs.
	status, err := d.c.ReadUint8(0x3)
	if err != nil {
		return wrapf("failed to read the input values: %v", err)
	}

	// Read deltas (in two's complement, capped at -128 to 127).
	//deltasB := [len(d.inputStatuses)]byte{}
	//if err = d.c.ReadStruct(0x10, &deltasB); err != nil {
	//	return d.inputStatuses, wrapf("failed to read the delta values: %v", err)
	//}
	//deltas := [len(d.inputStatuses)]int{}
	//for i, b := range deltasB {
	//	deltas[i] = int(int8(b))
	//}
	// Read thresholds.
	//thresholds := [len(d.inputStatuses)]byte{}
	//if err = d.c.ReadStruct(0x30, &thresholds); err != nil {
	//	return d.inputStatuses, wrapf("failed to read the threshold values: %v", err)
	//}

	// Convert the data into a sensor state.
	for i := uint8(0); i < uint8(len(d.inputStatuses)); i++ {
		// Check if the bit is set.
		// TODO(mattetti): check if the event is passed the threshold:
		// deltas[i] > int(thresholds[i])

		// If the bit is set, it was touched.
		idx := len(d.inputStatuses) - 1
		if status&(1<<(uint8(idx)-i)) != 0 {
			if d.inputStatuses[i] == PressedStatus {
				if d.opts.RetriggerOnHold {
					d.inputStatuses[i] = HeldStatus
				}
				continue
			}
			d.inputStatuses[i] = PressedStatus
		} else {
			d.inputStatuses[i] = OffStatus
		}
	}
	copy(t, d.inputStatuses[:])
	return nil
}

// LinkLEDs links the behavior of the LEDs to the touch sensors.
// Doing so, disabled the option for the host to set specific LEDs on/off.
func (d *Dev) LinkLEDs(on bool) error {
	if on {
		if err := d.c.WriteUint8(regLEDLinking, 0xff); err != nil {
			return wrapf("failed to link LEDs: %v", err)
		}
	} else {
		if err := d.c.WriteUint8(regLEDLinking, 0x00); err != nil {
			return wrapf("failed to unlink LEDs: %v", err)
		}
	}
	d.opts.LinkedLEDs = on
	return nil
}

// AllLEDs turns all the LEDs on or off.
//
// This is quite more efficient than looping through each LED and turn them
// on/off.
func (d *Dev) AllLEDs(on bool) error {
	if d.opts.LinkedLEDs {
		return wrapf("can't manually set LEDs when they are linked to sensors")
	}
	// TODO(maruel): support > 8 LEDs.
	var v byte
	if on {
		v = 0xFF
	}
	if err := d.c.WriteUint8(regLEDOutputControl, v); err != nil {
		return wrapf("failed to set all LEDs: %v", err)
	}
	return nil
}

// SetLED sets the state of a LED as on or off.
//
// Only works if the LEDs are not linked to the sensors
func (d *Dev) SetLED(idx int, state bool) error {
	if d.opts.LinkedLEDs {
		return wrapf("can't manually set LEDs when they are linked to sensors")
	}
	if idx >= d.numLEDs || idx < 0 {
		return wrapf("invalid led idx %d", idx)
	}
	if d.opts.Debug {
		log.Printf("cap1xxx: Set LED state %d - %t", idx, state)
	}
	// TODO(maruel): support > 8 LEDs.
	if state {
		if err := d.setBit(regLEDOutputControl, idx); err != nil {
			return wrapf("failed to set LED #%d to %t: %v", idx, state, err)
		}
	}
	if err := d.clearBit(regLEDOutputControl, idx); err != nil {
		return wrapf("failed to set LED #%d to %t: %v", idx, state, err)
	}
	return nil
}

// Reset issues a soft reset to the device using the reset pin if available.
func (d *Dev) Reset() error {
	if err := d.ClearInterrupt(); err != nil {
		return err
	}
	if d.opts.ResetPin != nil {
		if d.opts.Debug {
			log.Println("cap1xxx: Resetting the device using the reset pin")
		}
		if err := d.opts.ResetPin.Out(gpio.Low); err != nil {
			return wrapf("failed to set reset pin low: %v", err)
		}
		sleep(1 * time.Microsecond)
		if err := d.opts.ResetPin.Out(gpio.High); err != nil {
			return wrapf("failed to set reset pin high: %v", err)
		}
		sleep(10 * time.Millisecond)
		if err := d.opts.ResetPin.Out(gpio.Low); err != nil {
			return wrapf("failed to set reset pin low: %v", err)
		}
	}
	// Track the reset time since the device won't be ready for up to 15ms and
	// won't be ready for first conversion for up to 200ms.
	d.lastReset = time.Now()
	// Time to communications is 15ms.
	sleep(15 * time.Millisecond)
	return nil
}

// ClearInterrupt resets the interrupt flag.
func (d *Dev) ClearInterrupt() error {
	// Clear the main control bit.
	if err := d.clearBit(0x0, 0); err != nil {
		return wrapf("failed to clean interrupt: %v")
	}
	return nil
}

//

func makeDev(c conn.Conn, isSPI bool, opts *Opts) (*Dev, error) {
	d := &Dev{
		opts:  *opts,
		isSPI: isSPI,
		c:     mmr.Dev8{Conn: c, Order: binary.LittleEndian},
	}

	// Read the product id to confirm it matches our expectations.
	productID, err := d.c.ReadUint8(0xFD)
	if err != nil {
		return nil, wrapf("failed to read product id: %v", err)
	}
	switch productID {
	case 0x3A: // cap1114
		d.inputStatuses = make([]TouchStatus, 14)
		d.numLEDs = 11
		return nil, errors.New("cap1xxx: cap1114 is not yet supported")
	case 0x50: // cap1188
		d.inputStatuses = make([]TouchStatus, 8)
		d.numLEDs = 8
	case 0x51: // cap1166
		d.inputStatuses = make([]TouchStatus, 6)
		d.numLEDs = 6
	case 0x52: // cap1128
		d.inputStatuses = make([]TouchStatus, 8)
		d.numLEDs = 2
	case 0x53: // cap1126
		d.inputStatuses = make([]TouchStatus, 6)
		d.numLEDs = 2
	case 0x54: // cap1133
		d.inputStatuses = make([]TouchStatus, 3)
		d.numLEDs = 3
	case 0x55: // cap1106
		// No LED.
		d.inputStatuses = make([]TouchStatus, 6)
	case 0x56: // cap1105
		// No LED.
		d.inputStatuses = make([]TouchStatus, 5)
	case 0x67: // cap1206
		// http://ww1.microchip.com/downloads/en/DeviceDoc/00001567B.pdf
		d.inputStatuses = make([]TouchStatus, 6)
		return nil, errors.New("cap1xxx: cap1206 is not yet supported")
	case 0x69: // cap1296
		// http://ww1.microchip.com/downloads/en/DeviceDoc/00001569B.pdf
		d.inputStatuses = make([]TouchStatus, 6)
		return nil, errors.New("cap1xxx: cap1296 is not yet supported")
	case 0x6B: // cap1208
		// http://ww1.microchip.com/downloads/en/DeviceDoc/00001570C.pdf
		d.inputStatuses = make([]TouchStatus, 8)
		return nil, errors.New("cap1xxx: cap1208 is not yet supported")
	case 0x6D: // cap1203
		// http://ww1.microchip.com/downloads/en/DeviceDoc/00001572B.pdf
		d.inputStatuses = make([]TouchStatus, 3)
		return nil, errors.New("cap1xxx: cap1203 is not yet supported")
	case 0x6F: // cap1293
		// http://ww1.microchip.com/downloads/en/DeviceDoc/00001566B.pdf
		d.inputStatuses = make([]TouchStatus, 3)
		return nil, errors.New("cap1xxx: cap1293 is not yet supported")
	case 0x71: // cap1298
		// http://ww1.microchip.com/downloads/en/DeviceDoc/00001571B.pdf
		d.inputStatuses = make([]TouchStatus, 8)
		return nil, errors.New("cap1xxx: cap1298 is not yet supported")
	default:
		return nil, wrapf("unexpected chip id %x; is this a cap1xxx?", productID)
	}
	// manufacturer ID on 0xFE, should be 0x5D
	// revision ID on 0xFF, should be 0x83

	// Reset the device.
	if err := d.Reset(); err != nil {
		return nil, err
	}

	var recalFlag byte
	if d.opts.EnableRecalibration {
		recalFlag = 1
	}
	var intOnRel byte
	if !d.opts.InterruptOnRelease {
		intOnRel = 1 // 0 = trigger on release
	}

	// Enable all inputs.
	if err := d.c.WriteUint8(0x21, 0xff); err != nil {
		return nil, wrapf("failed to enable all inputs: %v", err)
	}
	// Enable interrupts.
	if err := d.c.WriteUint8(0x27, 0xff); err != nil {
		return nil, wrapf("failed to enable interrupts: %v", err)
	}
	// Enable/disable repeats.
	// TODO(mattetti): make it an option.
	if err := d.c.WriteUint8(0x28, 0xff); err != nil {
		return nil, wrapf("failed to disable repeats: %v", err)
	}
	// Enable multitouch.
	multitouchConfig := (
	// Enables the multiple button blocking circuitry.
	// - 0: The multiple touch circuitry is disabled. The device will not block
	//   multiple touches.
	// - 1 (default): The multiple touch circuitry is enabled. The device will
	//   flag the number of touches equal to programmed multiple touch threshold
	//   and block all others. It will remember which sensor inputs are valid and
	//   block all others until that sensor pad has been released. Once a sensor
	//   pad has been released, the N detected touches (determined via the cycle
	//   order of CS1 - CS8) will be flagged and all others blocked.
	byte(0)<<7 |
		byte(0)<<6 | byte(0)<<5 | byte(0)<<4 |
		// Determines the number of simultaneous touches on all sensor pads before
		// a Multiple Touch Event is detected and sensor inputs are blocked.
		// Set to 2.
		byte(0)<<3 | byte(1)<<2 |
		byte(0)<<1 | byte(0)<<0)
	if err := d.c.WriteUint8(0x2a, multitouchConfig); err != nil {
		return nil, fmt.Errorf("failed to enable multitouch: %v", err)
	}
	// Averaging and Sampling Config.
	samplingConfig := (byte(0)<<7 |
		// Number of samples taken per measurement.
		// TODO(mattetti): use d.opts.SamplesPerMeasurement
		byte(0)<<6 |
		byte(0)<<5 |
		byte(0)<<4 |
		// Sample time.
		// TODO(mattetti): use d.opts.SamplingTime
		byte(1)<<3 |
		byte(0)<<2 |
		// Overall cycle time.
		// TODO(mattetti): use d.opts.CycleTime
		byte(0)<<1 |
		byte(0)<<0)
	if d.opts.Debug {
		log.Printf("cap1xxx: Sampling config mask: %08b", samplingConfig)
	}
	if err := d.c.WriteUint8(0x24, samplingConfig); err != nil {
		return nil, wrapf("failed to enable multitouch: %v", err)
	}

	// Customize sensitivity.
	sensitivity := (byte(0)<<7 |
		// Controls the sensitivity of a touch detection. The sensitivity settings
		// act to scale the relative delta count value higher or lower based on the
		// system parameters. A setting of 000b is the most sensitive while a
		// setting of 111b is the least sensitive. At the more sensitive settings,
		// touches are detected for a smaller delta capacitance corresponding to a
		// “lighter” touch. These settings are more sensitive to noise, however,
		// and a noisy environment may flag more false touches with higher
		// sensitivity levels.
		// Set to 4x.
		// TODO(mattetti): make that configurable.
		byte(1)<<6 | byte(0)<<5 | byte(1)<<4 |
		byte(0)<<3 | byte(0)<<2 | byte(0)<<1 | byte(0)<<0)
	if d.opts.Debug {
		log.Printf("cap1xxx: Sensitivity mask: %08b", sensitivity)
	}
	if err := d.c.WriteUint8(0x1F, sensitivity); err != nil {
		return nil, wrapf("failed to set sensitivity: %v", err)
	}

	if d.opts.LinkedLEDs {
		if err := d.LinkLEDs(true); err != nil {
			return nil, err
		}
	}

	if d.opts.RetriggerOnHold {
		if err := d.c.WriteUint8(0x28, 0xff); err != nil {
			return nil, wrapf("failed to set retrigger on hold: %v", err)
		}
	} else {
		if err := d.c.WriteUint8(0x28, 0x00); err != nil {
			return nil, wrapf("failed to turn off retrigger on hold: %v", err)
		}
	}

	// page 47 - configuration registers
	config := (
	// Timeout: Enables the timeout and idle functionality of the SMBus protocol.
	// - 0 (default): The SMBus timeout and idle functionality are disabled. The
	//   SMBus interface will not time out if the clock line is held low.
	//   Likewise, it will not reset if both data and clock lines are held high
	//   for longer than 200us
	byte(0)<<7 |
		// Configures the operation of the WAKE pin.
		// - 0 (default): The WAKE pin is not asserted when a touch is detected
		//   while the device is in Standby. It will still be used to wake the
		//   device from Deep Sleep when driven high.
		byte(0)<<6 |
		// Digital noise threshold.
		// Determines whether the digital noise threshold is used by the device.
		// - 1 (default): The noise threshold is disabled. Any delta count that is
		//   less than the touch threshold is used for the automatic re-calibration
		//   routine.
		byte(1)<<5 |
		// Analog noise filter.
		// - 0 (default): The analog noise filter is enabled.
		// - 1: Disables the feature.
		byte(1)<<4 |
		// Maximum duration recalibration.
		// Determines whether the maximum duration recalibration is enabled.
		// - 0: The maximum duration recalibration functionality is disabled. A
		//   touch may be held indefinitely and no re-calibration will be performed
		//   on any sensor input.
		// - 1: The maximum duration recalibration functionality is enabled. If a
		//   touch is held for longer than the d.opts.MaxTouchDuration, then the
		//   re-calibration routine will be restarted.
		recalFlag<<3 |
		byte(0)<<2 |
		byte(0)<<1 |
		byte(0)<<0)
	if d.opts.Debug {
		log.Printf("cap1xxx: Config mask: %08b", config)
	}
	if err := d.c.WriteUint8(0x20, config); err != nil {
		return nil, wrapf("failed to set the device configuration: %v", err)
	}

	config2 := (
	// Linked LED Transition controls.
	// - 0 (default): The Linked LED Transition controls set the min duty cycle
	//   equal to the max duty cycle.
	byte(0)<<7 |
		// Determines the ALERT# pin polarity and behavior.
		// - 1 (default): The ALERT# pin is active low and open drain.
		byte(1)<<6 |
		// Determines whether the device will reduce power consumption while
		// waiting between conversion time completion and the end of the polling
		// cycle.
		// - 0 (default): The device will always power down as much as possible
		//   during the time between the end of the last conversion and the end of
		//   the polling cycle.
		byte(1)<<5 |
		// Determines whether the LED Mirror Control register bits are linked to
		// the LED Polarity bits. Setting this bit blocks the normal behavior which
		// is to automatically set and clear the LED Mirror Control bits when the
		// LED Polarity bits are set or cleared.
		// - 0 (default): When the LED Polarity controls are set, the corresponding
		//   LED Mirror control is automatically set. Likewise, when the LED
		//   Polarity controls are cleared, the corresponding LED Mirror control is
		//   also cleared.
		byte(0)<<4 |
		// Determines whether the Noise Status bits will show RF Noise as the only
		// input source.
		// - 0 (default): The Noise Status registers will show both RF noise and low
		//   frequency EMI noise if either is detected on a capacitive touch sensor
		//   input.
		byte(0)<<3 |
		// Determines whether the RF noise filter is enabled.
		// - 0 (default): If RF noise is detected by the analog block, the delta
		//   count on the corresponding channel is set to 0. Note that this does
		//   not require that Noise Status bits be set.
		byte(0)<<2 |
		byte(0)<<1 |
		// Controls the interrupt behavior when a release is detected on a button.
		// - 0: An interrupt is generated when a press is detected and again when a
		//   release is detected and at the repeat rate (if enabled)
		// - 1: An interrupt is generated when a press is detected and at the
		//   repeat rate but not when a release is detected.
		intOnRel<<0)
	if d.opts.Debug {
		log.Printf("cap1xxx: Config2 mask: %08b", config2)
	}
	if err := d.c.WriteUint8(0x44, config2); err != nil {
		return nil, wrapf("failed to set the device configuration 2: %v", err)
	}

	return d, nil
}

func (d *Dev) resetSinceAtLeast(t time.Duration) {
	readyAt := d.lastReset.Add(t)
	if now := time.Now(); now.Before(readyAt) {
		sleep(readyAt.Sub(now))
	}
}

// setBit sets a specific bit on a register.
//
// TODO(mattetti): avoid reading before writing, keep states in memory.
func (d *Dev) setBit(regID uint8, idx int) error {
	v, err := d.c.ReadUint8(regID)
	if err != nil {
		return err
	}
	v |= (1 << uint8(idx))
	return d.c.WriteUint8(regID, v)
}

// clearBit clears a specific bit on a register.
//
// TODO(mattetti): avoid reading before writing, keep states in memory.
func (d *Dev) clearBit(regID uint8, idx int) error {
	v, err := d.c.ReadUint8(regID)
	if err != nil {
		return err
	}
	v &= ^(1 << uint8(idx))
	return d.c.WriteUint8(regID, v)
}

//

const (
	// regLEDLinking is the Sensor Input LED Linking register controls whether a
	// capacitive touch sensor input is linked to an LED output. If the
	// corresponding bit is set, then the appropriate LED output will change
	// states defined by the LED Behavior controls in response to the capacitive
	// touch sensor input.
	regLEDLinking = 0x72
	// regLEDOutputControl is the LED Output Control Register controls the output
	// state of the LED pins that are not linked to sensor inputs.
	regLEDOutputControl = 0x74
)

var sleep = time.Sleep

func wrapf(format string, a ...interface{}) error {
	return fmt.Errorf("cap1xxx: "+format, a...)
}

var _ conn.Resource = &Dev{}
