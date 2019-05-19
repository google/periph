// Copyright 2017 The Periph Authors. All rights reserved.
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
// per period for write but only one for read.
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

//
// References
//
// DMA channel allocation:
// https://github.com/raspberrypi/linux/issues/1327
//
// DMA location:
// https://www.raspberrypi.org/forums/viewtopic.php?f=71&t=19797

package bcm283x

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio/gpiostream"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/host/pmem"
	"periph.io/x/periph/host/videocore"
)

const (
	periphMask = 0x00FFFFFF
	periphBus  = 0x7E000000
	// maxLite is the maximum transfer allowed by a lite channel.
	maxLite = 65535
)

// Pages 47-50
type dmaStatus uint32

const (
	dmaReset        dmaStatus = 1 << 31 // RESET; Writing a 1 to this bit will reset the DMA
	dmaAbort        dmaStatus = 1 << 30 // ABORT; Writing a 1 to this bit will abort the current DMA CB. The DMA will load the next CB and attempt to continue.
	dmaDisableDebug dmaStatus = 1 << 29 // DISDEBUG; When set to 1, the DMA will not stop when the debug pause signal is asserted.
	// When set to 1, the DMA will keep a tally of the AXI writes going out and
	// the write responses coming in. At the very end of the current DMA transfer
	// it will wait until the last outstanding write response has been received
	// before indicating the transfer is complete. Whilst waiting it will load
	// the next CB address (but will not fetch the CB), clear the active flag (if
	// the next CB address = zero), and it will defer setting the END flag or the
	// INT flag until the last outstanding write response has been received.
	// In this mode, the DMA will pause if it has more than 13 outstanding writes
	// at any one time.
	dmaWaitForOutstandingWrites dmaStatus = 1 << 28 // WAIT_FOR_OUTSTANDING_WRITES
	// 27:24 reserved
	// 23:20 Lowest has higher priority on AXI.
	dmaPanicPriorityShift = 20
	dmaPanicPriorityMask  = 0xF << 20 // PANIC_PRIORITY
	// 19:16 Lowest has higher priority on AXI.
	dmaPriorityShift = 16
	dmaPriorityMask  = 0xF << dmaPriorityShift // PRIORITY
	// 15:9 reserved
	dmaErrorStatus dmaStatus = 1 << 8 // ERROR DMA error was detected; must be cleared manually.
	// 7 reserved
	dmaWaitingForOutstandingWrites dmaStatus = 1 << 6 // WAITING_FOR_OUTSTANDING_WRITES; Indicates if the DMA is currently waiting for any outstanding writes to be received, and is not transferring data.
	dmaDreqStopsDMA                dmaStatus = 1 << 5 // DREQ_STOPS_DMA; Indicates if the DMA is currently paused and not transferring data due to the DREQ being inactive.
	// Indicates if the DMA is currently paused and not transferring data. This
	// will occur if: the active bit has been cleared, if the DMA is currently
	// executing wait cycles or if the debug_pause signal has been set by the
	// debug block, or the number of outstanding writes has exceeded the max
	// count.
	dmaPaused dmaStatus = 1 << 4 // PAUSED
	// Indicates the state of the selected DREQ (Data Request) signal, ie. the
	// DREQ selected by the PERMAP field of the transfer info.
	// 1 = Requesting data. This will only be valid once the DMA has started and
	//     the PERMAP field has been loaded from the CB. It will remain valid,
	//     indicating the selected DREQ signal, until a new CB is loaded. If
	//     PERMAP is set to zero (unpaced transfer) then this bit will read back
	//     as 1.
	// 0 = No data request.
	dmaDreq dmaStatus = 1 << 3 // DREQ
	// This is set when the transfer for the CB ends and INTEN is set to 1. Once
	// set it must be manually cleared down, even if the next CB has INTEN = 0.
	// Write 1 to clear.
	dmaInterrupt dmaStatus = 1 << 2 // INT
	// Set when the transfer described by the current control block is complete.
	// Write 1 to clear.
	dmaEnd dmaStatus = 1 << 1 // END
	// This bit enables the DMA. The DMA will start if this bit is set and the
	// CB_ADDR is non zero. The DMA transfer can be paused and resumed by
	// clearing, then setting it again.
	// This bit is automatically cleared at the end of the complete DMA transfer,
	// ie. after a NEXTCONBK = 0x0000_0000 has been loaded.
	dmaActive dmaStatus = 1 << 0 // ACTIVE
)

var dmaStatusMap = []struct {
	v dmaStatus
	s string
}{
	{dmaReset, "Reset"},
	{dmaAbort, "Abort"},
	{dmaDisableDebug, "DisableDebug"},
	{dmaWaitForOutstandingWrites, "WaitForOutstandingWrites"},
	{dmaErrorStatus, "ErrorStatus"},
	{dmaWaitingForOutstandingWrites, "WaitingForOutstandingWrites"},
	{dmaDreqStopsDMA, "DreqStopsDMA"},
	{dmaPaused, "Paused"},
	{dmaDreq, "Dreq"},
	{dmaInterrupt, "Interrupt"},
	{dmaEnd, "End"},
	{dmaActive, "Active"},
}

func (d dmaStatus) String() string {
	var out []string
	for _, l := range dmaStatusMap {
		if d&l.v != 0 {
			d &^= l.v
			out = append(out, l.s)
		}
	}
	if v := d & dmaPanicPriorityMask; v != 0 {
		out = append(out, fmt.Sprintf("pp%d", v>>dmaPanicPriorityShift))
		d &^= dmaPanicPriorityMask
	}
	if v := d & dmaPriorityMask; v != 0 {
		out = append(out, fmt.Sprintf("p%d", v>>dmaPriorityShift))
		d &^= dmaPriorityMask
	}
	if d != 0 {
		out = append(out, fmt.Sprintf("dmaStatus(0x%x)", uint32(d)))
	}
	if len(out) == 0 {
		return "0"
	}
	return strings.Join(out, "|")
}

// Pages 50-52
type dmaTransferInfo uint32

const (
	// 31:27 reserved
	// Don't do wide writes as 2 beat burst; only for channels 0 to 6
	dmaNoWideBursts dmaTransferInfo = 1 << 26 // NO_WIDE_BURSTS
	// 25:21 Slows down the DMA throughput by setting the number of dummy cycles
	// burnt after each DMA read or write is completed.
	dmaWaitCyclesShift                 = 21
	dmaWaitcyclesMax                   = 0x1F
	dmaWaitCyclesMask  dmaTransferInfo = dmaWaitcyclesMax << dmaWaitCyclesShift // WAITS
	// 20:16 Peripheral mapping (1-31) whose ready signal shall be used to
	// control the rate of the transfers. 0 means continuous un-paced transfer.
	//
	// It is the source used to pace the data reads and writes operations, each
	// pace being a DReq (Data Request).
	//
	// Page 61
	dmaPerMapShift                   = 16
	dmaPerMapMask    dmaTransferInfo = 31 << dmaPerMapShift
	dmaFire          dmaTransferInfo = 0 << dmaPerMapShift  // PERMAP; Continuous trigger
	dmaDSI           dmaTransferInfo = 1 << dmaPerMapShift  // Display Serial Interface (?)
	dmaPCMTX         dmaTransferInfo = 2 << dmaPerMapShift  //
	dmaPCMRX         dmaTransferInfo = 3 << dmaPerMapShift  //
	dmaSMI           dmaTransferInfo = 4 << dmaPerMapShift  // Secondary Memory Interface (?)
	dmaPWM           dmaTransferInfo = 5 << dmaPerMapShift  //
	dmaSPITX         dmaTransferInfo = 6 << dmaPerMapShift  //
	dmaSPIRX         dmaTransferInfo = 7 << dmaPerMapShift  //
	dmaBscSPIslaveTX dmaTransferInfo = 8 << dmaPerMapShift  //
	dmaBscSPIslaveRX dmaTransferInfo = 9 << dmaPerMapShift  //
	dmaUnused        dmaTransferInfo = 10 << dmaPerMapShift //
	dmaEMMC          dmaTransferInfo = 11 << dmaPerMapShift //
	dmaUARTTX        dmaTransferInfo = 12 << dmaPerMapShift //
	dmaSDHost        dmaTransferInfo = 13 << dmaPerMapShift //
	dmaUARTRX        dmaTransferInfo = 14 << dmaPerMapShift //
	dmaDSI2          dmaTransferInfo = 15 << dmaPerMapShift // Same as DSI
	dmaSlimBusMCTX   dmaTransferInfo = 16 << dmaPerMapShift //
	dmaHDMI          dmaTransferInfo = 17 << dmaPerMapShift // 216MHz; potentially a (216MHz/(26+1)) 8MHz copy rate but it fails if HDMI is disabled
	dmaSlimBusMCRX   dmaTransferInfo = 18 << dmaPerMapShift //
	dmaSlimBusDC0    dmaTransferInfo = 19 << dmaPerMapShift //
	dmaSlimBusDC1    dmaTransferInfo = 20 << dmaPerMapShift //
	dmaSlimBusDC2    dmaTransferInfo = 21 << dmaPerMapShift //
	dmaSlimBusDC3    dmaTransferInfo = 22 << dmaPerMapShift //
	dmaSlimBusDC4    dmaTransferInfo = 23 << dmaPerMapShift //
	dmaScalerFIFO0   dmaTransferInfo = 24 << dmaPerMapShift // Also on SMI; SMI can be disabled with smiDisable
	dmaScalerFIFO1   dmaTransferInfo = 25 << dmaPerMapShift //
	dmaScalerFIFO2   dmaTransferInfo = 26 << dmaPerMapShift //
	dmaSlimBusDC5    dmaTransferInfo = 27 << dmaPerMapShift //
	dmaSlimBusDC6    dmaTransferInfo = 28 << dmaPerMapShift //
	dmaSlimBusDC7    dmaTransferInfo = 29 << dmaPerMapShift //
	dmaSlimBusDC8    dmaTransferInfo = 30 << dmaPerMapShift //
	dmaSlimBusDC9    dmaTransferInfo = 31 << dmaPerMapShift //

	dmaBurstLengthShift                 = 12
	dmaBurstLengthMask  dmaTransferInfo = 0xF << dmaBurstLengthShift // BURST_LENGTH 15:12 0 means a single transfer.
	dmaSrcIgnore        dmaTransferInfo = 1 << 11                    // SRC_IGNORE Source won't be read, output will be zeros.
	dmaSrcDReq          dmaTransferInfo = 1 << 10                    // SRC_DREQ
	dmaSrcWidth128      dmaTransferInfo = 1 << 9                     // SRC_WIDTH 128 bits reads if set, 32 bits otherwise.
	dmaSrcInc           dmaTransferInfo = 1 << 8                     // SRC_INC Increment read pointer by 32/128bits at each read if set.
	dmaDstIgnore        dmaTransferInfo = 1 << 7                     // DEST_IGNORE Do not write.
	dmaDstDReq          dmaTransferInfo = 1 << 6                     // DEST_DREQ
	dmaDstWidth128      dmaTransferInfo = 1 << 5                     // DEST_WIDTH 128 bits writes if set, 32 bits otherwise.
	dmaDstInc           dmaTransferInfo = 1 << 4                     // DEST_INC Increment write pointer by 32/128bits at each read if set.
	dmaWaitResp         dmaTransferInfo = 1 << 3                     // WAIT_RESP DMA waits for AXI write response.
	// 2 reserved
	// 2D mode interpret of txLen; linear if unset; only for channels 0 to 6.
	dmaTransfer2DMode  dmaTransferInfo = 1 << 1 // TDMODE
	dmaInterruptEnable dmaTransferInfo = 1 << 0 // INTEN Generate an interrupt upon completion.
)

var dmaTransferInfoMap = []struct {
	v dmaTransferInfo
	s string
}{
	{dmaNoWideBursts, "NoWideBursts"},
	{dmaSrcIgnore, "SrcIgnore"},
	{dmaSrcDReq, "SrcDReq"},
	{dmaSrcWidth128, "SrcWidth128"},
	{dmaSrcInc, "SrcInc"},
	{dmaDstIgnore, "DstIgnore"},
	{dmaDstDReq, "DstDReq"},
	{dmaDstWidth128, "DstWidth128"},
	{dmaDstInc, "DstInc"},
	{dmaWaitResp, "WaitResp"},
	{dmaTransfer2DMode, "Transfer2DMode"},
	{dmaInterruptEnable, "InterruptEnable"},
}

var dmaPerMap = []string{
	"Fire",
	"DSI",
	"PCMTX",
	"PCMRX",
	"SMI",
	"PWM",
	"SPITX",
	"SPIRX",
	"BscSPISlaveTX",
	"BscSPISlaveRX",
	"Unused",
	"EMMC",
	"UARTTX",
	"SDHOST",
	"UARTRX",
	"DSI2",
	"SlimBusMCTX",
	"HDMI",
	"SlimBusMCRX",
	"SlimBusDC0",
	"SlimBusDC1",
	"SlimBusDC2",
	"SlimBusDC3",
	"SlimBusDC4",
	"ScalerFIFO0",
	"ScalerFIFO1",
	"ScalerFIFO2",
	"SlimBusDC5",
	"SlimBusDC6",
	"SlimBusDC7",
	"SlimBusDC8",
	"SlimBusDC9",
}

func (d dmaTransferInfo) String() string {
	var out []string
	for _, l := range dmaTransferInfoMap {
		if d&l.v != 0 {
			d &^= l.v
			out = append(out, l.s)
		}
	}
	if v := d & dmaWaitCyclesMask; v != 0 {
		out = append(out, fmt.Sprintf("waits=%d", v>>dmaWaitCyclesShift))
		d &^= dmaWaitCyclesMask
	}
	if v := d & dmaBurstLengthMask; v != 0 {
		out = append(out, fmt.Sprintf("burst=%d", v>>dmaBurstLengthShift))
		d &^= dmaBurstLengthMask
	}
	out = append(out, dmaPerMap[(d&dmaPerMapMask)>>dmaPerMapShift])
	d &^= dmaPerMapMask
	if d != 0 {
		out = append(out, fmt.Sprintf("dmaTransferInfo(0x%x)", uint32(d)))
	}
	return strings.Join(out, "|")
}

// Page 55
type dmaDebug uint32

const (
	// 31:29 reserved
	dmaLite dmaDebug = 1 << 28 // LITE RO set for lite DMA controllers
	// 27:25 version
	dmaVersionShift          = 25
	dmaVersionMask  dmaDebug = 7 << dmaVersionShift // VERSION
	// 24:16 dmaState
	dmaStateShift          = 16
	dmaStateMask  dmaDebug = 0x1FF << dmaStateShift // DMA_STATE; the actual states are not documented
	// 15:8  dmaID
	dmaIDShift = 8
	dmaIDMask  = 0xFF << dmaIDShift // DMA_ID; the index of the DMA controller
	// 7:4   outstandingWrites
	dmaOutstandingWritesShift = 4
	dmaOutstandingWritesMask  = 0xF << dmaOutstandingWritesShift // OUTSTANDING_WRITES
	// 3     reserved
	dmaReadError           dmaDebug = 1 << 2 // READ_ERROR slave read error; clear by writing a 1
	dmaFIFOError           dmaDebug = 1 << 1 // FIF_ERROR fifo error; clear by writing a 1
	dmaReadLastNotSetError dmaDebug = 1 << 0 // READ_LAST_NOT_SET_ERROR last AXI read signal was not set when expected
)

var dmaDebugMap = []struct {
	v dmaDebug
	s string
}{
	{dmaLite, "Lite"},
	{dmaReadError, "ReadError"},
	{dmaFIFOError, "FIFOError"},
	{dmaReadLastNotSetError, "ReadLastNotSetError"},
}

func (d dmaDebug) String() string {
	var out []string
	for _, l := range dmaDebugMap {
		if d&l.v != 0 {
			d &^= l.v
			out = append(out, l.s)
		}
	}
	if v := d & dmaVersionMask; v != 0 {
		out = append(out, fmt.Sprintf("v%d", uint32(v>>dmaVersionShift)))
		d &^= dmaVersionMask
	}
	if v := d & dmaStateMask; v != 0 {
		out = append(out, fmt.Sprintf("state(%x)", uint32(v>>dmaStateShift)))
		d &^= dmaStateMask
	}
	if v := d & dmaIDMask; v != 0 {
		out = append(out, fmt.Sprintf("#%x", uint32(v>>dmaIDShift)))
		d &^= dmaIDMask
	}
	if v := d & dmaOutstandingWritesMask; v != 0 {
		out = append(out, fmt.Sprintf("OutstandingWrites=%d", uint32(v>>dmaOutstandingWritesShift)))
		d &^= dmaOutstandingWritesMask
	}
	if d != 0 {
		out = append(out, fmt.Sprintf("dmaDebug(0x%x)", uint32(d)))
	}
	if len(out) == 0 {
		return "0"
	}
	return strings.Join(out, "|")
}

// 31:30 0
// 29:16 yLength (only for channels #0 to #6)
// 15:0  xLength
type dmaTransferLen uint32

// 31:16 dstStride byte increment to apply at the end of each row in 2D mode
// 15:0  srcStride byte increment to apply at the end of each row in 2D mode
type dmaStride uint32

func (d dmaStride) String() string {
	y := (d >> 16) & 0xFFFF
	if y != 0 {
		return fmt.Sprintf("0x%x,0x%x", uint32(y), uint32(d&0xFFFF))
	}
	return fmt.Sprintf("0x%x", uint32(d&0xFFFF))
}

// controlBlock is 256 bits (32 bytes) in length.
//
// https://www.raspberrypi.org/wp-content/uploads/2012/02/BCM2835-ARM-Peripherals.pdf
// Page 40.
type controlBlock struct {
	transferInfo dmaTransferInfo // 0x00 TI
	srcAddr      uint32          // 0x04 SOURCE_AD pointer to source in physical address space
	dstAddr      uint32          // 0x08 DEST_AD pointer to destination in physical address space
	txLen        dmaTransferLen  // 0x0C TXFR_LEN length in bytes
	stride       dmaStride       // 0x10 STRIDE
	// Pointer to the next chained controlBlock; must be 32 bytes aligned.
	// Set it to 0 to stop.
	nextCB   uint32    // 0x14 NEXTCONBK
	reserved [2]uint32 // 0x18+0x1C
}

// initBlock initializes a controlBlock for any valid DMA operation.
//
// l is in bytes, not in words.
//
// dreq can be dmaFire, dmaPwm, dmaPcmTx, etc. waits is additional wait state
// between clocks.
func (c *controlBlock) initBlock(srcAddr, dstAddr, l uint32, srcIO, dstIO, srcInc, dstInc bool, dreq dmaTransferInfo) error {
	if srcIO && dstIO {
		return errors.New("only one of src and dst can be I/O")
	}
	if srcAddr == 0 && dstAddr == 0 {
		return errors.New("at least one source or destination is required")
	}
	if srcAddr == 0 && srcIO {
		return errors.New("using src as I/O requires src")
	}
	if dstAddr == 0 && dstIO {
		return errors.New("using dst as I/O requires dst")
	}
	if dreq&^dmaPerMapMask != 0 {
		return errors.New("dreq must be one of the clock source, nothing else")
	}

	t := dmaNoWideBursts | dmaWaitResp
	if srcAddr == 0 {
		t |= dmaSrcIgnore
		c.srcAddr = 0
	} else {
		if srcIO {
			// Memory mapped register
			c.srcAddr = physToBus(srcAddr)
		} else {
			// Normal memory
			c.srcAddr = physToUncachedPhys(srcAddr)
		}
		if srcInc {
			t |= dmaSrcInc
		}
	}
	if dstAddr == 0 {
		t |= dmaDstIgnore
		c.dstAddr = 0
	} else {
		if dstIO {
			// Memory mapped register
			c.dstAddr = physToBus(dstAddr)
		} else {
			// Normal memory
			c.dstAddr = physToUncachedPhys(dstAddr)
		}
		if dstInc {
			t |= dmaDstInc
		}
	}
	if dreq != dmaFire {
		// Inserting a wait prevents multiple transfers in a single DReq cycle.
		waits := 1
		t |= dreq | dmaTransferInfo(waits<<dmaWaitCyclesShift)
		if srcIO {
			t |= dmaSrcDReq
		}
		if dstIO {
			t |= dmaDstDReq
		}
	}
	c.transferInfo = t
	c.txLen = dmaTransferLen(l)
	c.stride = 0
	c.nextCB = 0
	return nil
}

func (c *controlBlock) GoString() string {
	return fmt.Sprintf(
		"{\n  transferInfo: %s,\n  srcAddr:      0x%x,\n  dstAddr:      0x%x,\n  txLen:        %d,\n  stride:       %s,\n  nextCB:       0x%x,\n}",
		&c.transferInfo, c.srcAddr, c.dstAddr, c.txLen, &c.stride, c.nextCB)
}

// DMAChannel is the memory mapped registers for one DMA channel.
//
// Page 39.
type dmaChannel struct {
	cs           dmaStatus                  // 0x00 CS
	cbAddr       uint32                     // 0x04 CONNBLK_AD *controlBlock in physical address space; rounded to 32 bytes
	transferInfo dmaTransferInfo            // 0x08 TI (RO) Copyied by DMA on start from cbAddr
	srcAddr      uint32                     // 0x0C SOURCE_AD (RO) Copyied by DMA on start from cbAddr
	dstAddr      uint32                     // 0x10 DEST_AD (RO) Copyied by DMA on start from cbAddr
	txLen        dmaTransferLen             // 0x14 TXFR_LEN (RO) Copyied by DMA on start from cbAddr
	stride       dmaStride                  // 0x18 STRIDE (RO) Copyied by DMA on start from cbAddr
	nextCB       uint32                     // 0x1C NEXTCONBK Only safe to edit when DMA is paused
	debug        dmaDebug                   // 0x20 DEBUG
	reserved     [(0x100 - 0x24) / 4]uint32 // 0x24
}

func (d *dmaChannel) isAvailable() bool {
	return (d.cs&^dmaDreq) == 0 && d.cbAddr == 0
}

// reset resets the DMA channel in a way that makes it directly available.
//
// It doesn't clear the local controlBlock cached values.
func (d *dmaChannel) reset() {
	d.cs = dmaReset
	d.cbAddr = 0
}

// startIO initializes the DMA channel to start a transmission.
//
// The channel must have been reseted before.
func (d *dmaChannel) startIO(cb uint32) {
	d.cbAddr = cb
	d.cs = dmaWaitForOutstandingWrites | 8<<dmaPanicPriorityShift | 8<<dmaPriorityShift | dmaActive
}

// wait waits for a DMA channel transmission to complete.
//
// It must have started successfully before.
func (d *dmaChannel) wait() error {
	// TODO(maruel): Calculate the number of bytes remaining, the clock rate and
	// do a short sleep instead of a spin. To do so, it'll need the clock rate.
	// Spin until the the bit is reset, to release the DMA controller channel.
	for d.cs&dmaActive != 0 && d.debug&(dmaReadError|dmaFIFOError|dmaReadLastNotSetError) == 0 {
	}
	if d.debug&dmaReadError != 0 {
		return errors.New("DMA read error")
	}
	if d.debug&dmaFIFOError != 0 {
		return errors.New("DMA FIFO error")
	}
	if d.debug&dmaReadLastNotSetError != 0 {
		return errors.New("DMA AIX read error")
	}
	return nil
}

func (d *dmaChannel) GoString() string {
	return fmt.Sprintf(
		"{\n  cs:           %s,\n  cbAddr:       0x%x,\n  transferInfo: %s,\n  srcAddr:      0x%x,\n  dstAddr:      0x%x,\n  txLen:        %v,\n  stride:       %s,\n  nextCB:       0x%x,\n  debug:        %s,\n  reserved:     {...},\n}",
		d.cs, d.cbAddr, d.transferInfo, d.srcAddr, d.dstAddr, d.txLen, d.stride, d.nextCB, d.debug)
}

// dmaMap is the block for the first 15 channels and control registers.
//
// Note that we modify the DMA controllers without telling the kernel driver.
// The driver keeps its own table of which DMA channel is available so this
// code could effectively crash the whole system. It practice this works.
// #everythingisfine
//
// Page 40.
type dmaMap struct {
	channels  [15]dmaChannel
	padding0  [0xE0]byte //
	intStatus uint32     // 0xFE0 INT_STATUS bits 15:0 mapped to controllers #15 to #0
	padding1  [0xC]byte  //
	enable    uint32     // 0xFF0 ENABLE bits 14:0 mapped to controllers #14 to #0
}

func indent(s, indent string) string {
	var out []string
	for _, x := range strings.Split(s, "\n") {
		if len(x) != 0 {
			out = append(out, indent+x)
		} else {
			out = append(out, "")
		}
	}
	return strings.Join(out, "\n")
}

func (d *dmaMap) GoString() string {
	out := []string{"{"}
	for i := range d.channels {
		out = append(out, indent(fmt.Sprintf("%d: %s", i, d.channels[i].GoString()+","), "  "))
	}
	out = append(out, fmt.Sprintf("  intStatus: 0x%x,", d.intStatus))
	out = append(out, fmt.Sprintf("  enable:    0x%x,", d.enable))
	out = append(out, "}")
	return strings.Join(out, "\n")
}

// pickChannel searches for a free DMA channel.
func pickChannel(blacklist ...int) (int, *dmaChannel) {
	// Try the lite ones first.
	if drvDMA.dmaMemory != nil {
		// TODO(maruel): Trying to use channel #15 always fails.
		/*
			if drvDMA.dmaChannel15 != nil {
				if drvDMA.dmaChannel15.isAvailable() {
					drvDMA.dmaChannel15.reset()
					return 15, drvDMA.dmaChannel15
				}
			}
		*/
		// TODO(maruel): May as well use a lookup table.
		for i := len(drvDMA.dmaMemory.channels) - 1; i >= 0; i-- {
			for _, exclude := range blacklist {
				if i == exclude {
					goto skip
				}
			}
			if drvDMA.dmaMemory.channels[i].isAvailable() {
				drvDMA.dmaMemory.channels[i].reset()
				return i, &drvDMA.dmaMemory.channels[i]
			}
		skip:
		}
	}
	// Uncomment to understand the state of the DMA channels.
	//log.Printf("%#v", drvDMA.dmaMemory)
	return -1, nil
}

// runIO picks a DMA channel, initialize it and runs a transfer.
//
// It tries to release the channel as soon as it can.
func runIO(pCB pmem.Mem, liteOk bool) error {
	var blacklist []int
	if !liteOk {
		blacklist = []int{7, 8, 9, 10, 11, 12, 13, 14, 15}
	}
	_, ch := pickChannel(blacklist...)
	if ch == nil {
		return errors.New("bcm283x-dma: no channel available")
	}
	defer ch.reset()
	ch.startIO(uint32(pCB.PhysAddr()))
	return ch.wait()
}

func allocateCB(size int) ([]controlBlock, *videocore.Mem, error) {
	buf, err := drvDMA.dmaBufAllocator((size + 0xFFF) &^ 0xFFF)
	if err != nil {
		return nil, nil, err
	}
	var cb []controlBlock
	if err := buf.AsPOD(&cb); err != nil {
		_ = buf.Close()
		return nil, nil, err
	}
	return cb, buf, nil
}

// dmaWriteStreamPCM streams data to a PCM enabled pin as a half-duplex I²S
// channel.
func dmaWriteStreamPCM(p *Pin, w gpiostream.Stream) error {
	d := w.Duration()
	if d == 0 {
		return nil
	}
	f := w.Frequency()
	_, _, _, actualfreq, err := calcSource(f, 1)
	if err != nil {
		return err
	}
	if actualfreq != f {
		return errors.New("TODO(maruel): handle oversampling")
	}

	// Start clock earlier.
	drvDMA.pcmMemory.reset()
	_, _, err = setPCMClockSource(f)
	if err != nil {
		return err
	}

	// Calculate the number of bytes needed.
	l := (int(w.Frequency()/f) + 7) / 8 // Bytes
	buf, err := drvDMA.dmaBufAllocator((l + 0xFFF) &^ 0xFFF)
	if err != nil {
		return err
	}
	defer buf.Close()
	if err := copyStreamToDMABuf(w, buf.Uint32()); err != nil {
		return err
	}

	cb, pCB, err := allocateCB(4096)
	if err != nil {
		return err
	}
	defer pCB.Close()
	reg := drvDMA.pcmBaseAddr + 0x4 // pcmMap.fifo
	if err = cb[0].initBlock(uint32(buf.PhysAddr()), reg, uint32(l), false, true, true, false, dmaPCMTX); err != nil {
		return err
	}

	defer drvDMA.pcmMemory.reset()
	// Start transfer
	drvDMA.pcmMemory.set()
	err = runIO(pCB, l <= maxLite)
	// We have to wait PCM to be finished even after DMA finished.
	for drvDMA.pcmMemory.cs&pcmTXErr == 0 {
		Nanospin(10 * time.Nanosecond)
	}
	return err
}

func dmaWritePWMFIFO() (*dmaChannel, *videocore.Mem, error) {
	if drvDMA.dmaMemory == nil {
		return nil, nil, errors.New("bcm283x-dma is not initialized; try running as root?")
	}
	cb, buf, err := allocateCB(32 + 4) // CB + data
	if err != nil {
		return nil, nil, err
	}
	u := buf.Uint32()
	offsetBytes := uint32(32)
	u[offsetBytes/4] = 0x0
	physBuf := uint32(buf.PhysAddr())
	physBit := physBuf + offsetBytes
	dest := drvDMA.pwmBaseAddr + 0x18 // PWM FIFO
	if err := cb[0].initBlock(physBit, dest, 4, false, true, false, false, dmaPWM); err != nil {
		_ = buf.Close()
		return nil, nil, err
	}
	cb[0].nextCB = physBuf // Loop back to self.

	_, ch := pickChannel()
	if ch == nil {
		_ = buf.Close()
		return nil, nil, errors.New("bcm283x-dma: no channel available")
	}
	ch.startIO(physBuf)

	return ch, buf, nil
}

func startPWMbyDMA(p *Pin, rng, data uint32) (*dmaChannel, *videocore.Mem, error) {
	if drvDMA.dmaMemory == nil {
		return nil, nil, errors.New("bcm283x-dma is not initialized; try running as root?")
	}
	cb, buf, err := allocateCB(2*32 + 4) // 2 CBs + mask
	if err != nil {
		return nil, nil, err
	}
	u := buf.Uint32()
	cbBytes := uint32(32)
	offsetBytes := cbBytes * 2
	u[offsetBytes/4] = uint32(1) << uint(p.number&31)
	physBuf := uint32(buf.PhysAddr())
	physBit := physBuf + offsetBytes
	dest := [2]uint32{
		drvGPIO.gpioBaseAddr + 0x28 + 4*uint32(p.number/32), // clear
		drvGPIO.gpioBaseAddr + 0x1C + 4*uint32(p.number/32), // set
	}
	// High
	if err := cb[0].initBlock(physBit, dest[1], data*4, false, true, false, false, dmaPWM); err != nil {
		_ = buf.Close()
		return nil, nil, err
	}
	cb[0].nextCB = physBuf + cbBytes
	// Low
	if err := cb[1].initBlock(physBit, dest[0], (rng-data)*4, false, true, false, false, dmaPWM); err != nil {
		_ = buf.Close()
		return nil, nil, err
	}
	cb[1].nextCB = physBuf // Loop back to cb[0]

	var blacklist []int
	if data*4 >= 1<<16 || (rng-data)*4 >= 1<<16 {
		// Don't use lite channels.
		blacklist = []int{7, 8, 9, 10, 11, 12, 13, 14, 15}
	}
	_, ch := pickChannel(blacklist...)

	if ch == nil {
		_ = buf.Close()
		return nil, nil, errors.New("bcm283x-dma: no channel available")
	}
	ch.startIO(physBuf)

	return ch, buf, nil
}

// overSamples calculates the skip value which are the values that are read but
// discarded as the clock is too fast.
func overSamples(s gpiostream.Stream) (int, error) {
	desired := s.Frequency()
	skip := drvDMA.pwmDMAFreq / desired
	if skip < 1 {
		return 0, fmt.Errorf("frequency is too high(%s)", desired)
	}
	actualFreq := drvDMA.pwmDMAFreq / skip
	errorPercent := 100 * (actualFreq - desired) / desired
	if errorPercent < -10 || errorPercent > 10 {
		return 0, fmt.Errorf("actual resolution differs more than 10%%(%s vs %s)", desired, actualFreq)
	}
	return int(skip), nil
}

// dmaReadStream streams input from a pin.
func dmaReadStream(p *Pin, b *gpiostream.BitStream) error {
	skip, err := overSamples(b)
	if err != nil {
		return err
	}
	if _, err := setPWMClockSource(); err != nil {
		return err
	}

	// Needs 32x the memory since each read is one full uint32. On the other
	// hand one could read 32 contiguous pins simultaneously at no cost.
	// TODO(simokawa): Implement a function to get number of bits for all type of
	// Stream
	l := len(b.Bits) * 8 * 4 * int(skip)
	// TODO(simokawa): Allocate multiple pages and CBs for huge buffer.
	buf, err := drvDMA.dmaBufAllocator((l + 0xFFF) &^ 0xFFF)
	if err != nil {
		return err
	}
	defer buf.Close()
	cb, pCB, err := allocateCB(4)
	if err != nil {
		return err
	}
	defer pCB.Close()

	reg := drvGPIO.gpioBaseAddr + 0x34 + 4*uint32(p.number/32) // GPIO Pin Level 0
	if err := cb[0].initBlock(reg, uint32(buf.PhysAddr()), uint32(l), true, false, false, true, dmaPWM); err != nil {
		return err
	}
	err = runIO(pCB, l <= maxLite)
	uint32ToBitLSBF(b.Bits, buf.Bytes(), uint8(p.number&31), skip*4)
	return err
}

// dmaWriteStreamEdges streams data to a pin as a half-duplex one controlBlock
// per bit toggle DMA stream.
//
// Memory usage is 32 bytes x number of bit changes rounded up to nearest
// 4Kb, so an arbitrary stream of 1s or 0s only takes 4Kb but a stream of
// 101010s will takes 256x the memory.
//
// TODO(maruel): Use huffman-coding-like repeated patterns detection to
// "compress" the bitstream. This trades off upfront computation for lower
// memory usage. The "compressing" function should be public, so the user can
// call it only once yet stream multiple times.
//
// TODO(maruel): Mutate the program as it goes to reduce duplication by having
// the DMA controller write in a following controlBlock.nextCB.
// handling gpiostream.Program explicitly.
func dmaWriteStreamEdges(p *Pin, w gpiostream.Stream) error {
	d := w.Duration()
	if d == 0 {
		return nil
	}
	var bits []byte
	var msb bool
	switch v := w.(type) {
	case *gpiostream.BitStream:
		bits = v.Bits
		msb = !v.LSBF
	default:
		return fmt.Errorf("unknown type: %T", v)
	}
	skip, err := overSamples(w)
	if err != nil {
		return err
	}

	// Calculate the number of controlBlock needed.
	count := 1
	stride := uint32(skip)
	last := getBit(bits[0], 0, msb)
	l := int(int64(d) * int64(w.Frequency()) / int64(physic.Hertz)) // Bits
	for i := 1; i < l; i++ {
		if v := getBit(bits[i/8], i%8, msb); v != last || stride == maxLite {
			last = v
			count++
			stride = 0
		}
		stride += uint32(skip)
	}
	// 32 bytes for each CB and 4 bytes for the mask.
	bufBytes := count*32 + 4
	cb, buf, err := allocateCB((bufBytes + 0xFFF) &^ 0xFFF)
	if err != nil {
		return err
	}
	defer buf.Close()

	// Setup the single mask buffer of 4Kb.
	mask := uint32(1) << uint(p.number&31)
	u := buf.Uint32()
	offset := (len(buf.Bytes()) - 4)
	u[offset/4] = mask
	physBit := uint32(buf.PhysAddr()) + uint32(offset)

	// Other constants during the loop.
	// Waits does not seem to work as expected. Not counted as DREQ pulses?
	// Use PWM's rng1 instead for this.
	//waits := divs - 1
	dest := [2]uint32{
		drvGPIO.gpioBaseAddr + 0x28 + 4*uint32(p.number/32), // clear
		drvGPIO.gpioBaseAddr + 0x1C + 4*uint32(p.number/32), // set
	}

	// Render the controlBlock's to trigger the bit trigger for either Set or
	// Clear GPIO memory registers.
	last = getBit(bits[0], 0, msb)
	index := 0
	stride = uint32(skip)
	for i := 1; i < l; i++ {
		if v := getBit(bits[i/8], i%8, msb); v != last || stride == maxLite {
			if err := cb[index].initBlock(physBit, dest[last], stride*4, false, true, false, false, dmaPWM); err != nil {
				return err
			}
			// Hardcoded len(controlBlock) == 32. It is not necessary to use
			// physToUncachedPhys() here.
			cb[index].nextCB = uint32(buf.PhysAddr()) + uint32(32*(index+1))
			index++
			stride = 0
			last = v
		}
		stride += uint32(skip)
	}
	if err := cb[index].initBlock(physBit, dest[last], stride*4, false, true, false, false, dmaPWM); err != nil {
		return err
	}

	// Start clock before DMA
	_, err = setPWMClockSource()
	if err != nil {
		return err
	}
	return runIO(buf, true)
}

// dmaWriteStreamDualChannel streams data to a pin using two DMA channels.
//
// In practice this leads to a glitchy stream.
func dmaWriteStreamDualChannel(p *Pin, w gpiostream.Stream) error {
	// TODO(maruel): Analyse 'w' to figure out the programs to load, and create
	// the number of controlBlock needed to reduce memory usage.
	// TODO(maruel): When only one channel is needed, it is much more memory
	// efficient to use DMA to write to PWM FIFO.
	skip, err := overSamples(w)
	if err != nil {
		return err
	}
	// Calculates the number of needed bytes.
	l := int(int64(w.Duration())*int64(w.Frequency())/int64(physic.Hertz)) * skip * 4
	bufLen := (l + 0xFFF) &^ 0xFFF
	bufSet, err := drvDMA.dmaBufAllocator(bufLen)
	if err != nil {
		return err
	}
	defer bufSet.Close()
	bufClear, err := drvDMA.dmaBufAllocator(bufLen)
	if err != nil {
		return err
	}
	defer bufClear.Close()
	cb, pCB, err := allocateCB(4096)
	if err != nil {
		return err
	}
	defer pCB.Close()

	// Needs 64x the memory since each write is 2 full uint32. On the other
	// hand one could write 32 contiguous pins simultaneously at no cost.
	mask := uint32(1) << uint(p.number&31)
	if err := raster32(w, skip, bufClear.Uint32(), bufSet.Uint32(), mask); err != nil {
		return err
	}

	// Start clock before DMA start
	_, err = setPWMClockSource()
	if err != nil {
		return err
	}

	regSet := drvGPIO.gpioBaseAddr + 0x1C + 4*uint32(p.number/32)
	if err := cb[0].initBlock(uint32(bufSet.PhysAddr()), regSet, uint32(l), false, true, true, false, dmaPWM); err != nil {
		return err
	}
	regClear := drvGPIO.gpioBaseAddr + 0x28 + 4*uint32(p.number/32)
	if err := cb[1].initBlock(uint32(bufClear.PhysAddr()), regClear, uint32(l), false, true, true, false, dmaPWM); err != nil {
		return err
	}

	// The first channel must be a full bandwidth one. The "light" ones are
	// effectively a single one, which means that they are interleaved. If both
	// are "light" then the jitter is largely increased.
	x, chSet := pickChannel(6, 7, 8, 9, 10, 11, 12, 13, 14, 15)
	if chSet == nil {
		return errors.New("bcm283x-dma: no channel available")
	}
	defer chSet.reset()
	_, chClear := pickChannel(x)
	if chClear == nil {
		return errors.New("bcm283x-dma: no secondary channel available")
	}
	defer chClear.reset()

	// Two channel need to be synchronized but there is not such a mechanism.
	chSet.startIO(uint32(pCB.PhysAddr()))        // cb[0]
	chClear.startIO(uint32(pCB.PhysAddr()) + 32) // cb[1]

	err1 := chSet.wait()
	err2 := chClear.wait()
	if err1 == nil {
		return err2
	}
	return err1
}

// physToUncachedPhys returns the uncached physical memory address backing a
// physical memory address.
//
// p must be rooted at a page boundary (4096).
func physToUncachedPhys(p uint32) uint32 {
	// http://en.wikibooks.org/wiki/Aros/Platforms/Arm_Raspberry_Pi_support#Framebuffer
	return p | drvGPIO.dramBus
}

func physToBus(p uint32) uint32 {
	return (p & periphMask) | periphBus
}

// smokeTest allocates two physical pages, ask the DMA controller to copy the
// data from one page to another and make sure the content is as expected.
//
// This should take a fraction of a second and will make sure the driver is
// usable. This ensures there's at least one DMA channel available.
func smokeTest() error {
	// If these are commented out due to a new processor having different
	// characteristics, the corresponding code needs to be updated.
	if drvDMA.dmaMemory.channels[6].debug&dmaLite != 0 {
		return errors.New("unexpected hardware: DMA channel #6 shouldn't be lite")
	}
	if drvDMA.dmaMemory.channels[7].debug&dmaLite == 0 {
		return errors.New("unexpected hardware: DMA channel #7 should be lite")
	}
	if drvDMA.dmaMemory.enable != 0x7FFF {
		return errors.New("unexpected hardware: DMA enable is not fully set")
	}

	const size = 4096 * 4 // 16kb
	const holeSize = 1    // Minimum DMA alignment

	alloc := func(s int) (pmem.Mem, error) {
		return videocore.Alloc(s)
	}

	copyMem := func(pDst, pSrc uint64) error {
		// Allocate a control block and initialize it.
		pCB, err2 := videocore.Alloc(4096)
		if err2 != nil {
			return err2
		}
		defer pCB.Close()
		var cb *controlBlock
		if err := pCB.AsPOD(&cb); err != nil {
			return err
		}
		if false {
			// This code is not run by default because it resets the PWM clock on
			// process startup, which may cause undesirable glitches.

			// Initializes the PWM clock right away to 1MHz.
			_, err := setPWMClockSource()
			if err != nil {
				return err
			}
			if err := cb.initBlock(uint32(pSrc), uint32(pDst)+holeSize, size-2*holeSize, false, false, true, true, dmaPWM); err != nil {
				return err
			}
		} else {
			// Use maximum performance.
			if err := cb.initBlock(uint32(pSrc), uint32(pDst)+holeSize, size-2*holeSize, false, false, true, true, dmaFire); err != nil {
				return err
			}
		}
		return runIO(pCB, size-2*holeSize <= maxLite)
	}

	return pmem.TestCopy(size, holeSize, alloc, copyMem)
}

// driverDMA implements periph.Driver.
//
// It implements much more than the DMA controller, it also exposes the clocks,
// the PWM and PCM controllers.
type driverDMA struct {
	pcmBaseAddr uint32
	pwmBaseAddr uint32

	dmaMemory     *dmaMap
	dmaChannel15  *dmaChannel
	pcmMemory     *pcmMap
	clockMemory   *clockMap
	timerMemory   *timerMap
	gpioPadMemory *gpioPadMap
	// Page 138
	// - Two independent bit-streams
	// - Each channel either a PWM or serialised version of a 32-bit word
	// - Variable input and output resolutions.
	// - Load data from a FIFO storage block, to extent to 8 32-bit words (256
	//   bits).
	//
	// Author note: 100Mhz base resolution with a 256 bits 1-bit stream is
	// actually good enough to generate a DAC.
	pwmMemory *pwmMap

	// These clocks are shared with hardware PWM, DMA driven PWM and BitStream.
	pwmBaseFreq physic.Frequency
	pwmDMAFreq  physic.Frequency
	pwmDMACh    *dmaChannel
	pwmDMABuf   *videocore.Mem

	// dmaBufAllocator is overridden for unit testing.
	dmaBufAllocator func(s int) (*videocore.Mem, error) // Set to videocore.Alloc
}

func (d *driverDMA) Close() error {
	// TODO(maruel): Stop DMA and PWM controllers.
	d.pcmBaseAddr = 0
	d.pwmBaseAddr = 0
	d.dmaMemory = nil
	d.dmaChannel15 = nil
	d.pcmMemory = nil
	d.clockMemory = nil
	d.timerMemory = nil
	d.pwmMemory = nil
	d.pwmBaseFreq = 0
	d.pwmDMAFreq = 0
	d.pwmDMACh = nil
	d.pwmDMABuf = nil
	d.dmaBufAllocator = nil
	return nil
}

func (d *driverDMA) String() string {
	return "bcm283x-dma"
}

func (d *driverDMA) Prerequisites() []string {
	return []string{"bcm283x-gpio"}
}

func (d *driverDMA) After() []string {
	return nil
}

func (d *driverDMA) Init() (bool, error) {
	d.dmaBufAllocator = videocore.Alloc
	d.pwmBaseFreq = 25 * physic.MegaHertz
	d.pwmDMAFreq = 200 * physic.KiloHertz
	// baseAddr is initialized by prerequisite driver bcm283x-gpio.
	if err := pmem.MapAsPOD(uint64(drvGPIO.baseAddr+0x7000), &d.dmaMemory); err != nil {
		if os.IsPermission(err) {
			return true, fmt.Errorf("need more access, try as root: %v", err)
		}
		return true, err
	}
	// Channel #15 is "physically removed from the other DMA Channels so it has a
	// different address base".
	if err := pmem.MapAsPOD(uint64(drvGPIO.baseAddr+0xE05000), &d.dmaChannel15); err != nil {
		return true, err
	}
	d.pcmBaseAddr = drvGPIO.baseAddr + 0x203000
	if err := pmem.MapAsPOD(uint64(d.pcmBaseAddr), &d.pcmMemory); err != nil {
		return true, err
	}
	d.pwmBaseAddr = drvGPIO.baseAddr + 0x20C000
	if err := pmem.MapAsPOD(uint64(d.pwmBaseAddr), &d.pwmMemory); err != nil {
		return true, err
	}
	if err := pmem.MapAsPOD(uint64(drvGPIO.baseAddr+0x101000), &d.clockMemory); err != nil {
		return true, err
	}
	if err := pmem.MapAsPOD(uint64(drvGPIO.baseAddr+0x3000), &d.timerMemory); err != nil {
		return true, err
	}
	if err := pmem.MapAsPOD(uint64(drvGPIO.baseAddr+0x100000), &d.gpioPadMemory); err != nil {
		return true, err
	}
	// Do not run smokeTest() unless it's clear it is not dangerous.
	return true, nil
}

func debugDMA() {
	for i, ch := range drvDMA.dmaMemory.channels {
		log.Println(i, ch.cs.String())
		if ch.cs&dmaActive != 0 {
			log.Printf("%x: %s", ch.cbAddr, ch.GoString())
		}
	}
	log.Println(15, drvDMA.dmaChannel15.cs.String())
}

func resetDMA(ch int) error {
	if ch < len(drvDMA.dmaMemory.channels) {
		drvDMA.dmaMemory.channels[ch].reset()
	} else if ch == 15 {
		drvDMA.dmaChannel15.reset()
	} else {
		return fmt.Errorf("invalid dma channel %d", ch)
	}
	return nil
}

func init() {
	if isArm {
		periph.MustRegister(&drvDMA)
	}
}

var drvDMA driverDMA
