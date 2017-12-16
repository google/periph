package cap1188

import "periph.io/x/periph/conn/gpio"

// SamplingTime determines the time to take a single sample
type SamplingTime uint8

// Possible sampling time values (written as 2 bits)
const (
	S320us SamplingTime = 0
	S640us SamplingTime = 1
	// S1_28ms represents 1.28ms sampling time, which is the default.
	S1_28ms SamplingTime = 2 // default
	S2_56ms SamplingTime = 3
)

// AvgSampling set the number of samples per measurement that get
// averaged
type AvgSampling uint8

// possible average sampling values. (written as 3 bits)
const (
	// Avg1 means that 1 sample is taken per measurement
	Avg1   AvgSampling = iota // 0
	Avg2                      // 1
	Avg4                      // 2
	Avg8                      // 3 default
	Avg16                     // 4
	Avg32                     // 5
	Avg64                     // 6
	Avg128                    // 7
)

// CycleTime determines the overall cycle time for all measured channels during
// normal operation.
type CycleTime uint8

// possible cycle time values. (written as 2 bits)
const (
	C35ms CycleTime = iota // 0
	C70ms                  // default
	C105ms
	C140ms
)

// MaxDur is the maximum duration of a touch event before it triggers a
// recalibration.
type MaxDur uint8

// possible touch duration values. (written as 4 bits)
const (
	MaxDur560ms MaxDur = iota
	MaxDur840ms
	MaxDur1120ms
	MaxDur1400ms
	MaxDur1680ms
	MaxDur2240ms
	MaxDur2800ms
	MaxDur3360ms
	MaxDur3920ms
	MaxDur44800ms
	MaxDur5600ms // default
	MaxDur6720ms
	MaxDur7840ms
	MaxDur8906ms
	MaxDur10080ms
	MaxDur11200ms
)

// Opts is optional options to pass to the constructor.
//
// Address is only used on creation of an I²C-device. Its default value is 0x28.
// It can be set to other values (0x29, 0x2a, 0x2b, 0x2c) depending on the HW
// configuration of the ADDR_COMM pin. This has no effect with NewSPI()
type Opts struct {
	// Address is the I2C slave address to use
	Address uint16

	// LinkedLEDs indicates if the LEDs should be activated automatically
	// when their sensors detect a touch event.
	LinkedLEDs bool
	// MaxTouchDuration sets the touch duration threshold. It is possible that a
	// “stuck button” occurs when something is placed on a button which causes a
	// touch to be detected for a long period. By setting this value,
	// a recalibration can be forced when a touch is held on a button for longer
	// than the duration specified.
	MaxTouchDuration MaxDur
	// EnableRecalibration is used to force the recalibration if a touch event lasts
	// longer than MaxTouchDuration.
	EnableRecalibration bool

	// AlertPin is the pin receiving the interrupt when a touch event is detected
	// and optionally if a release event is detected.
	AlertPin gpio.PinIn
	// ResetPin is the pin used to reset the device.
	ResetPin gpio.PinOut

	// InterruptOnRelease indicates if the device should also trigger an
	// interrupt on the AlertPin when a release event is detected.
	InterruptOnRelease bool

	// RetriggerOnHold forces a retrigger of the interrupt when a sensor is pressed
	// for longer than MaxTouchDuration
	RetriggerOnHold bool

	// Averaging and Sampling Configuration Register

	// SamplesPerMeasurement is the number of samples taken per measurement. All
	// samples are taken consecutively on the same channel before the next
	// channel is sampled and the result is averaged over the number of samples
	// measured before updating the measured results.
	// Available options: 1, 2, 4, 8 (default), 16, 32, 64, 128
	SamplesPerMeasurement AvgSampling

	// SamplingTime Determines the time to take a single sample as shown
	SamplingTime SamplingTime

	// CycleTime  determines the overall cycle time for all measured channels
	// during normal operation. All measured channels are sampled at the
	// beginning of the cycle time. If additional time is remaining, then the
	// device is placed into a lower power state for the remaining duration of
	// the cycle.
	CycleTime CycleTime
}

// DefaultOpts returns a pointer to a new Opts with the default option values.
func DefaultOpts() *Opts {
	return &Opts{
		LinkedLEDs:            true,
		MaxTouchDuration:      MaxDur5600ms,
		RetriggerOnHold:       false,
		EnableRecalibration:   false,
		InterruptOnRelease:    false,
		SamplesPerMeasurement: Avg1,
		SamplingTime:          S1_28ms,
		CycleTime:             C35ms,
	}
}
