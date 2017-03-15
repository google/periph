// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// The DMA controller can be used for two functionality:
// - implement zero-CPU continuous PWM.
// - bitbang a large stream of bits over a GPIO pin, for example for WS2812b
//   support.
//
// The way it works under the hood is that the bcm283x has two registers, one
// to set a bit and one to clear a bit.
//
// So two DMA controllers are used, one writing a "clear bit" stream and one
// for the "set bit" stream. This requires two independent 32 bits wide streams
// per period.
//
// References
//
// Page 7:
// " Software accessing RAM directly must use physical addresses (based at
// 0x00000000). Software accessing RAM using the DMA engines must use bus
// addresses (based at 0xC0000000) " ... to skip the L1 cache.
//
// " The BCM2835 DMA Controller provides a total of 16 DMA channels. Each
// channel operates independently from the others and is internally arbitrated
// onto one of the 3 system buses. This means that the amount of bandwidth that
// a DMA channel may consume can be controlled by the arbiter settings. "
//
// The CPU has 16 DMA channels but only the first 7 (#0 to #6) can do strides.
// 7~15 have half the bandwidth.

package bcm283x

// Pages 47-50
type dmaStatus uint32

const (
	reset                    dmaStatus = 1 << 31 // RESET
	abort                    dmaStatus = 1 << 30 // ABORT
	disDebug                 dmaStatus = 1 << 29 // DISDEBUG
	waitForOutstandingWrites dmaStatus = 1 << 28 // WAIT_FOR_OUTSTANDING_WRITES
	// 27:24 reserved
	// 23:20 Lowest has higher priority on AXI.
	panicPriorityShift = 20 // PANIC_PRIORITY
	// 19:16 Lowest has higher priority on AXI.
	priorityShift = 16 // PRIORITY
	// 15:9 reserved
	errorStatus dmaStatus = 1 << 8 // ERROR DMA error was detected; must be cleared manually.
	// 7 reserved
	waitingForOutstandingWrites dmaStatus = 1 << 6 // WAITING_FOR_OUTSTANDING_WRITES
	dreqStopsDMA                dmaStatus = 1 << 5 // DREQ_STOPS_DMA
	paused                      dmaStatus = 1 << 4 // PAUSED
	dreq                        dmaStatus = 1 << 3 // DREQ
	interrupt                   dmaStatus = 1 << 2 // INT
	end                         dmaStatus = 1 << 1 // END
	active                      dmaStatus = 1 << 0 // ACTIVE
)

// Pages 50-52
type dmaTransferInfo uint32

const (
	// 31:27 reserved
	// Don't do wide writes as 2 beat burst; only for channels 0 to 6
	noWideBursts dmaTransferInfo = 1 << 26 // NO_WIDE_BURSTS
	// 25:21 Slows down the DMA throughput by setting the numbre of dummy cycles
	// burnt after each DMA read or write is completed.
	waitCyclesShift = 21 // WAITS
	// 20:16 Peripheral mapping (1-31) whose ready signal shall be used to
	// control the rate of the transfers. 0 means continuous un-paced transfer.
	//
	// It is the source used to pace the data reads and writes operations, each
	// pace being a DReq (Data Request).
	//
	// Page 61
	fire          dmaTransferInfo = iota << 16 // PERMAP; Continuous trigger
	dsi                                        //
	pcmTX                                      //
	pcmRX                                      //
	smi                                        //
	pwm                                        //
	spiTX                                      //
	spiRX                                      //
	bscSPIslaveTX                              //
	bscSPIslaveRX                              //
	unused                                     //
	eMMC                                       //
	uartTX                                     //
	sdHost                                     //
	uartRX                                     //
	dsi2                                       // Same as dsi
	slimBusMCTX                                //
	hdmi                                       //
	slimBusMCRX                                //
	slimBusDC0                                 //
	slimBusDC1                                 //
	slimBusDC2                                 //
	slimBusDC3                                 //
	slimBusDC4                                 //
	scalerFifo0                                // Also on SMI; SMI can be disabled with smiDisable
	scalerFifo1                                //
	scalerFifo2                                //
	slimBusDC5                                 //
	slimBusDC6                                 //
	slimBusDC7                                 //
	slimBusDC8                                 //
	slimBusDC9                                 //

	burstLengthShift                 = 12      // BURST_LENGTH 15:12 0 means a single transfer.
	srcIgnore        dmaTransferInfo = 1 << 11 // SRC_IGNORE Source won't be read, output will be zeros.
	srcDReq          dmaTransferInfo = 1 << 10 // SRC_DREQ
	srcWidth128      dmaTransferInfo = 1 << 9  // SRC_WIDTH 128 bits reads if set, 32 bits otherwise.
	srcInc           dmaTransferInfo = 1 << 8  // SRC_INC Increment read pointer by 32/128bits at each read if set.
	dstIgnore        dmaTransferInfo = 1 << 7  // DEST_IGNORE Do not write.
	dstDReq          dmaTransferInfo = 1 << 6  // DEST_DREQ
	dstWidth         dmaTransferInfo = 1 << 5  // DEST_WIDTH 128 bits writes if set, 32 bits otherwise.
	dstInc           dmaTransferInfo = 1 << 4  // DEST_INC Increment write pointer by 32/128bits at each read if set.
	waitResp         dmaTransferInfo = 1 << 3  // WAIT_RESP DMA waits for AXI write response.
	// 2 reserved
	// 2D mode interpret of txLen; linear if unset; only for channels 0 to 6.
	transfer2DMode  dmaTransferInfo = 1 << 1 // TDMODE
	interruptEnable dmaTransferInfo = 1 << 0 // INTEN Generate an interrupt upon completion.
)

// Page 55
type dmaDebug uint32

const (
	// 31:29 reserved
	lite dmaDebug = 28 << 1 // LITE RO set for lite DMA controllers
	// 27:25 version
	version dmaDebug = 7 << 25 // VERSION
	// 24:16 dmaState
	stateShift = 16 // DMA_STATE
	// 15:8  dmaID
	idShift = 8 // DMA_ID
	// 7:4   outstandingWrites
	outstandingWritesShift = 4 // OUTSTANDING_WRITES
	// 3     reserved
	readError           dmaDebug = 1 << 2 // READ_ERROR slave read error; clear by writing a 1
	fifoError           dmaDebug = 1 << 1 // FIF_ERROR fifo error; clear by writing a 1
	readLastNotSetError dmaDebug = 1 << 0 // READ_LAST_NOT_SET_ERROR last AXI read signal was not set when expected
)

// 31:30 0
// 29:16 yLength (only for channels #0 to #6)
// 15:0  xLength
type dmaTransferLen uint32

// 31:16 dstStride byte increment to apply at the end of each row in 2D mode
// 15:0  srcStride byte increment to apply at the end of each row in 2D mode
type dmaStride uint32
