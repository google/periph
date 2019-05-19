// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

import (
	"errors"
	"fmt"
	"time"

	"periph.io/x/periph/conn/physic"
)

// PWENi is used to enable/disable the corresponding channel. Setting this bit
// to 1 enables the channel and transmitter state machine. All registers and
// FIFO is writable without setting this bit.
//
// MODEi bit is used to determine mode of operation. Setting this bit to 0
// enables PWM mode. In this mode data stored in either PWM_DATi or FIFO is
// transmitted by pulse width modulation within the range defined by PWM_RNGi.
// When this mode is used MSENi defines whether to use PWM algorithm. Setting
// MODEi to 1 enables serial mode, in which data stored in either PWM_DATi or
// FIFO is transmitted serially within the range defined by PWM_RNGi. Data is
// transmitted MSB first and truncated or zeropadded depending on PWM_RNGi.
// Default mode is PWM.
//
// RPTLi is used to enable/disable repeating of the last data available in the
// FIFO just before it empties. When this bit is 1 and FIFO is used, the last
// available data in the FIFO is repeatedly sent. This may be useful in PWM
// mode to avoid duty cycle gaps. If the FIFO is not used this bit does not
// have any effect. Default operation is do-notrepeat.
//
// SBITi defines the state of the output when no transmission takes place. It
// also defines the zero polarity for the zero padding in serialiser mode. This
// bit is padded between two consecutive transfers as well as tail of the data
// when PWM_RNGi is larger than bit depth of data being transferred. this bit
// is zero by default.
//
// POLAi is used to configure the polarity of the output bit. When set to high
// the final output is inverted. Default operation is no inversion.
//
// USEFi bit is used to enable/disable FIFO transfer. When this bit is high
// data stored in the FIFO is used for transmission. When it is low, data
// written to PWM_DATi is transferred. This bit is 0 as default.
//
// CLRF is used to clear the FIFO. Writing a 1 to this bit clears the FIFO.
// Writing 0 has no effect. This is a single shot operation and reading the bit
// always returns 0.
//
// MSENi is used to determine whether to use PWM algorithm or simple M/S ratio
// transmission. When this bit is high M/S transmission is used. This bit is
// zero as default. When MODEi is 1, this configuration bit has no effect.
//
// See page 139-140 for the description of the PWM and M/S ratio algorithms.
const (
	// 31:16 reserved
	pwm2MS pwmControl = 1 << 15 // MSEN2; 0: PWM algorithm is used; 1: M/S transmission is used
	// 14 reserved
	pwm2UseFIFO        pwmControl = 1 << 13 // USEF2; 0: Data register is transmitted; 1: Fifo is used for transmission
	pwm2Polarity       pwmControl = 1 << 12 // POLA2; 0: 0=low 1=high; 1: 1=low 0=high
	pwm2SilenceHigh    pwmControl = 1 << 11 // SBIT2; Defines the state of the output when no transmission takes place
	pwm2RepeatLastData pwmControl = 1 << 10 // RPTL2; 0: Transmission interrupts when FIFO is empty; 1: Last data in FIFO is transmitted repetedly until FIFO is not empty
	pwm2Serialiser     pwmControl = 1 << 9  // MODE2; 0: PWM mode; 1: Serialiser mode
	pwm2Enable         pwmControl = 1 << 8  // PWEN2; Enable channel 2
	pwm2Mask           pwmControl = pwm2MS | pwm2UseFIFO | pwm2Polarity | pwm2SilenceHigh | pwm2RepeatLastData | pwm2Serialiser | pwm2Enable
	pwm1MS             pwmControl = 1 << 7 // MSEN1; 0: PWM algorithm is used; 1: M/S transmission is used
	pwmClearFIFO       pwmControl = 1 << 6 // CLRF1; Clear the fifo
	pwm1UseFIFO        pwmControl = 1 << 5 // USEF1; 0: Data register is transmitted; 1: Fifo is used for transmission
	pwm1Polarity       pwmControl = 1 << 4 // POLA1; 0: 0=low 1=high; 1: 1=low 0=high
	pwm1SilenceHigh    pwmControl = 1 << 3 // SBIT1; Defines the state of the output when no transmission takes place
	pwm1RepeatLastData pwmControl = 1 << 2 // RPTL1; 0: Transmission interrupts when FIFO is empty; 1: Last data in FIFO is transmitted repetedly until FIFO is not empty
	pwm1Serialiser     pwmControl = 1 << 1 // MODE1; 0: PWM mode; 1: Serialiser mode
	pwm1Enable         pwmControl = 1 << 0 // PWEN1; Enable channel 1
	pwm1Mask           pwmControl = pwm1MS | pwm1UseFIFO | pwm1Polarity | pwm1SilenceHigh | pwm1RepeatLastData | pwm1Serialiser | pwm1Enable
)

// Pages 141-143.
type pwmControl uint32

const (
	// 31:13 reserved
	// STAi bit indicates the current state of the channel which is useful for
	// debugging purposes. The bit set means the channel is currently
	// transmitting data.
	pwmSta4 pwmStatus = 1 << 12 // STA4
	pwmSta3 pwmStatus = 1 << 11 // STA3
	pwmSta2 pwmStatus = 1 << 10 // STA2
	pwmSta1 pwmStatus = 1 << 9  // STA1
	// BERR sets to high when an error has occurred while writing to registers
	// via APB. This may happen if the bus tries to write successively to same
	// set of registers faster than the synchroniser block can cope with.
	// Multiple switching may occur and contaminate the data during
	// synchronisation. Software should clear this bit by writing 1. Writing 0
	// to this bit has no effect.
	pwmBusErr pwmStatus = 1 << 8 // BERR Bus Error flag
	// GAPOi. bit indicates that there has been a gap between transmission of two
	// consecutive data from FIFO. This may happen when FIFO gets empty after
	// state machine has sent a word and waits for the next. If control bit RPTLi
	// is set to high this event will not occur. Software must clear this bit by
	// writing 1. Writing 0 to this bit has no effect.
	pwmGapo4 pwmStatus = 1 << 7 // GAPO4 Channel 4 Gap Occurred flag
	pwmGapo3 pwmStatus = 1 << 6 // GAPO3 Channel 3 Gap Occurred flag
	pwmGapo2 pwmStatus = 1 << 5 // GAPO2 Channel 2 Gap Occurred flag
	pwmGapo1 pwmStatus = 1 << 4 // GAPO1 Channel 1 Gap Occurred flag
	// RERR1 bit sets to high when a read when empty error occurs. Software must
	// clear this bit by writing 1. Writing 0 to this bit has no effect.
	pwmRerr1 pwmStatus = 1 << 3 // RERR1
	// WERR1 bit sets to high when a write when full error occurs. Software must
	// clear this bit by writing 1. Writing 0 to this bit has no effect.
	pwmWerr1 pwmStatus = 1 << 2 // WERR1
	// EMPT1 bit indicates the empty status of the FIFO. If this bit is high FIFO
	// is empty.
	pwmEmpt1 pwmStatus = 1 << 1 // EMPT1
	// FULL1 bit indicates the full status of the FIFO. If this bit is high FIFO
	// is full.
	pwmFull1      pwmStatus = 1 << 0 // FULL1
	pwmStatusMask           = pwmSta4 | pwmSta3 | pwmSta2 | pwmSta1 | pwmBusErr | pwmGapo4 | pwmGapo3 | pwmGapo2 | pwmGapo1 | pwmRerr1 | pwmWerr1 | pwmEmpt1 | pwmFull1
)

// Pages 144-145.
type pwmStatus uint32

const (
	pwmDMAEnable pwmDMACfg = 1 << 31 // ENAB
	// 30:16 reserved
	pwmPanicShift           = 16
	pwmPanicMask  pwmDMACfg = 0xFF << pwmPanicShift // PANIC Default is 7
	pwmDreqMask   pwmDMACfg = 0xFF                  // DREQ Default is 7
)

// Page 145.
type pwmDMACfg uint32

// pwmMap is the block to control the PWM generator.
//
// Note that pins are named PWM0 and PWM1 but the mapping uses channel numbers
// 1 and 2.
//   - PWM0: GPIO12, GPIO18, GPIO40, GPIO52.
//   - PWM1: GPIO13, GPIO19, GPIO41, GPIO45, GPIO53.
//
// Each channel works independently. They can either output a bitstream or a
// serialised version of up to eight 32 bits words.
//
// The default base PWM frequency is 100Mhz.
//
// Description at page 138-139.
//
// Page 140-141.
type pwmMap struct {
	ctl    pwmControl // CTL
	status pwmStatus  // STA
	dmaCfg pwmDMACfg  // DMAC
	// This register is used to define the range for the corresponding channel.
	// In PWM mode evenly distributed pulses are sent within a period of length
	// defined by this register. In serial mode serialised data is transmitted
	// within the same period. If the value in PWM_RNGi is less than 32, only the
	// first PWM_RNGi bits are sent resulting in a truncation. If it is larger
	// than 32 excess zero bits are padded at the end of data. Default value for
	// this register is 32.
	dummy1 uint32 // Padding
	rng1   uint32 // RNG1
	// This register stores the 32 bit data to be sent by the PWM Controller when
	// USEFi is 0. In PWM mode data is sent by pulse width modulation: the value
	// of this register defines the number of pulses which is sent within the
	// period defined by PWM_RNGi. In serialiser mode data stored in this
	// register is serialised and transmitted.
	dat1 uint32 // DAT1
	// This register is the FIFO input for the all channels. Data written to this
	// address is stored in channel FIFO and if USEFi is enabled for the channel
	// i it is used as data to be sent. This register is write only, and reading
	// this register will always return bus default return value, pwm0.
	// When more than one channel is enabled for FIFO usage, the data written
	// into the FIFO is shared between these channels in turn. For example if the
	// word series A B C D E F G H I .. is written to FIFO and two channels are
	// active and configured to use FIFO then channel 1 will transmit words A C E
	// G I .. and channel 2 will transmit words B D F H .. .  Note that
	// requesting data from the FIFO is in locked-step manner and therefore
	// requires tight coupling of state machines of the channels. If any of the
	// channel range (period) value is different than the others this will cause
	// the channels with small range values to wait between words hence resulting
	// in gaps between words. To avoid that, each channel sharing the FIFO should
	// be configured to use the same range value. Also note that RPTLi are not
	// meaningful when the FIFO is shared between channels as there is no defined
	// channel to own the last data in the FIFO. Therefore sharing channels must
	// have their RPTLi set to zero.
	//
	// If the set of channels to share the FIFO has been modified after a
	// configuration change, FIFO should be cleared before writing new data.
	fifo   uint32 // FIF1
	dummy2 uint32 // Padding
	rng2   uint32 // RNG2 Equivalent of rng1 for channel 2
	dat2   uint32 // DAT2 Equivalent of dat1 for channel 2
}

// reset stops the PWM.
func (p *pwmMap) reset() {
	p.dmaCfg = 0
	p.ctl |= pwmClearFIFO
	p.ctl &^= pwm1Enable | pwm2Enable
	Nanospin(100 * time.Microsecond) // Cargo cult copied. Probably not necessary.
	p.status = pwmBusErr | pwmGapo1 | pwmGapo2 | pwmGapo3 | pwmGapo4 | pwmRerr1 | pwmWerr1
	Nanospin(100 * time.Microsecond)
	// Use the full 32 bits of DATi.
	p.rng1 = 32
	p.rng2 = 32
}

// setPWMClockSource sets the PWM clock for use by the DMA controller for
// pacing.
//
// It may select an higher frequency than the one requested.
//
// Other potentially good clock sources are PCM, SPI and UART.
func setPWMClockSource() (physic.Frequency, error) {
	if drvDMA.pwmMemory == nil {
		return 0, errors.New("subsystem PWM not initialized")
	}
	if drvDMA.clockMemory == nil {
		return 0, errors.New("subsystem Clock not initialized")
	}
	if drvDMA.pwmDMACh != nil {
		// Already initialized
		return drvDMA.pwmDMAFreq, nil
	}

	// divs * div must fit in rng1 registor.
	div := uint32(drvDMA.pwmBaseFreq / drvDMA.pwmDMAFreq)
	actual, divs, err := drvDMA.clockMemory.pwm.set(drvDMA.pwmBaseFreq, div)
	if err != nil {
		return 0, err
	}

	if e := actual / physic.Frequency(divs*div); drvDMA.pwmDMAFreq != e {
		return 0, fmt.Errorf("unexpected DMA frequency %s != %s (%d/%d/%d)", drvDMA.pwmDMAFreq, e, actual, divs, div)
	}
	// It acts as a clock multiplier, since this amount of data is sent per
	// clock tick.
	drvDMA.pwmMemory.rng1 = divs * div
	Nanospin(10 * time.Microsecond)
	// Periph data (?)

	// Use low priority.
	drvDMA.pwmMemory.dmaCfg = pwmDMAEnable | pwmDMACfg(15<<pwmPanicShift|15)
	Nanospin(10 * time.Microsecond)
	drvDMA.pwmMemory.ctl |= pwmClearFIFO
	Nanospin(10 * time.Microsecond)
	old := drvDMA.pwmMemory.ctl
	drvDMA.pwmMemory.ctl = (old & ^pwmControl(0xff)) | pwm1UseFIFO | pwm1Enable

	// Start DMA
	if drvDMA.pwmDMACh, drvDMA.pwmDMABuf, err = dmaWritePWMFIFO(); err != nil {
		return 0, err
	}

	return drvDMA.pwmDMAFreq, nil
}

func resetPWMClockSource() error {
	if drvDMA.pwmDMACh != nil {
		drvDMA.pwmDMACh.reset()
		drvDMA.pwmDMACh = nil
	}
	if drvDMA.pwmDMABuf != nil {
		if err := drvDMA.pwmDMABuf.Close(); err != nil {
			return err
		}
		drvDMA.pwmDMABuf = nil
	}
	_, _, err := drvDMA.clockMemory.pwm.set(0, 0)
	return err
}
