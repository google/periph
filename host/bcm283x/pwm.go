// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bcm283x

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
	msen2 pwmControl = 1 << 15 // MSEN2 if set, M/S transmission is used; else PWM algo is used
	// 14 reserved
	usef2    pwmControl = 1 << 13 // USER2 if set, fifo is used for transmission; else data register is used
	pola2    pwmControl = 1 << 12 // POLA2
	sbit2    pwmControl = 1 << 11 // SBIT2
	rptl2    pwmControl = 1 << 10 // RPTL2
	mode2    pwmControl = 1 << 9  // MODE2
	pwen2    pwmControl = 1 << 8  // PWEN2 Enable channel 2
	pwm2Mask pwmControl = msen2 | usef2 | pola2 | sbit2 | rptl2 | mode2 | pwen2
	msen1    pwmControl = 1 << 7 // MSEN1
	clrf     pwmControl = 1 << 6 // CLRF1 clear the fifo
	usef1    pwmControl = 1 << 5 // USEF1
	pola1    pwmControl = 1 << 4 // POLA1
	sbit1    pwmControl = 1 << 3 // SBIT1
	rptl1    pwmControl = 1 << 2 // RPTL1
	mode1    pwmControl = 1 << 1 // MODE1
	pwen1    pwmControl = 1 << 0 // PWEN1 Enable channel 1
	pwm1Mask pwmControl = msen1 | usef1 | pola1 | sbit1 | rptl1 | mode1 | pwen1
)

// Pages 141-143.
type pwmControl uint32

const (
	// 31:13 reserved
	// STAi bit indicates the current state of the channel which is useful for
	// debugging purposes. The bit set means the channel is currently
	// transmitting data.
	sta4 pwmStatus = 1 << 12 // STA4
	sta3 pwmStatus = 1 << 11 // STA3
	sta2 pwmStatus = 1 << 10 // STA2
	sta1 pwmStatus = 1 << 9  // STA1
	// BERR sets to high when an error has occurred while writing to registers
	// via APB. This may happen if the bus tries to write successively to same
	// set of registers faster than the synchroniser block can cope with.
	// Multiple switching may occur and contaminate the data during
	// synchronisation. Software should clear this bit by writing 1. Writing 0
	// to this bit has no effect.
	busErr pwmStatus = 1 << 8 // BERR Bus Error flag
	// GAPOi. bit indicates that there has been a gap between transmission of two
	// consecutive data from FIFO. This may happen when FIFO gets empty after
	// state machine has sent a word and waits for the next. If control bit RPTLi
	// is set to high this event will not occur. Software must clear this bit by
	// writing 1. Writing 0 to this bit has no effect.
	gapo4 pwmStatus = 1 << 7 // GAPO4 Channel 4 Gap Occurred flag
	gapo3 pwmStatus = 1 << 6 // GAPO3 Channel 3 Gap Occurred flag
	gapo2 pwmStatus = 1 << 5 // GAPO2 Channel 2 Gap Occurred flag
	gapo1 pwmStatus = 1 << 4 // GAPO1 Channel 1 Gap Occurred flag
	// RERR1 bit sets to high when a read when empty error occurs. Software must
	// clear this bit by writing 1. Writing 0 to this bit has no effect.
	rerr1 pwmStatus = 1 << 3 // RERR1
	// WERR1 bit sets to high when a write when full error occurs. Software must
	// clear this bit by writing 1. Writing 0 to this bit has no effect.
	werr1 pwmStatus = 1 << 2 // WERR1
	// EMPT1 bit indicates the empty status of the FIFO. If this bit is high FIFO
	// is empty.
	empt1 pwmStatus = 1 << 1 // EMPT1
	// FULL1 bit indicates the full status of the FIFO. If this bit is high FIFO
	// is full.
	full1 pwmStatus = 1 << 0 // FULL1
)

// Pages 144-145.
type pwmStatus uint32

const (
	enab pwmDMACfg = 1 << 31 // ENAB
	// 30:16 reserved
	panicMask pwmDMACfg = 0xFF00 // PANIC Default is 7
	dreqMask  pwmDMACfg = 0xFF   // DREQ Default is 7
)

// Page 145.
type pwmDMACfg uint32
