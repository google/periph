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
	"os"
	"strings"

	"periph.io/x/periph"
	"periph.io/x/periph/host/pmem"
	"periph.io/x/periph/host/videocore"
)

var (
	dmaMemory       *dmaMap
	dmaChannel15    *dmaChannel
	dmaBufAllocator func(s int) (*videocore.Mem, error) = videocore.Alloc
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
// dreq can be dmaFire, dmaPwm, dmaPcmTx, etc. waits is additional wait state
// between clocks.
func (c *controlBlock) initBlock(srcAddr, dstAddr, l uint32, srcIO, dstIO, srcInc, dstInc bool, dreq dmaTransferInfo, waits int) error {
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
	if waits < 0 || waits > dmaWaitcyclesMax {
		return fmt.Errorf("waits must be between 0 and %d", dmaWaitcyclesMax)
	}
	if dreq == dmaFire && waits != 0 {
		return errors.New("using wait cycles without a clock doesn't make sense")
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
		// dmaSrcDReq |
		t |= dmaDstDReq | dreq | dmaTransferInfo(waits<<dmaWaitCyclesShift)
	}
	c.transferInfo = t
	// In bytes.
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
	if dmaMemory != nil {
		// TODO(maruel): Trying to use channel #15 always fails.
		/*
			if dmaChannel15 != nil {
				if dmaChannel15.isAvailable() {
					dmaChannel15.reset()
					return 15, dmaChannel15
				}
			}
		*/
		// TODO(maruel): May as well use a lookup table.
		for i := len(dmaMemory.channels) - 1; i >= 0; i-- {
			for _, exclude := range blacklist {
				if i == exclude {
					goto skip
				}
			}
			if dmaMemory.channels[i].isAvailable() {
				dmaMemory.channels[i].reset()
				return i, &dmaMemory.channels[i]
			}
		skip:
		}
	}
	// Uncomment to understand the state of the DMA channels.
	//fmt.Printf("%#v\n", dmaMemory)
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
	buf, err := dmaBufAllocator((size + 0xFFF) &^ 0xFFF)
	if err != nil {
		return nil, nil, err
	}
	var cb []controlBlock
	if err := buf.AsPOD(&cb); err != nil {
		buf.Close()
		return nil, nil, err
	}
	return cb, buf, nil
}

func startPWMbyDMA(p *Pin, rng, data uint32) (*dmaChannel, *videocore.Mem, error) {
	if dmaMemory == nil {
		return nil, nil, errors.New("bcm283x-dma is not initialized; try running as root?")
	}
	cb, buf, err := allocateCB(4096)
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
		gpioBaseAddr + 0x28 + 4*uint32(p.number/32), // clear
		gpioBaseAddr + 0x1C + 4*uint32(p.number/32), // set
	}
	waits := 0
	// High
	cb[0].initBlock(physBit, dest[1], data*4, false, true, false, false, dmaPWM, waits)
	cb[0].nextCB = physBuf + cbBytes
	// Low
	cb[1].initBlock(physBit, dest[0], (rng-data)*4, false, true, false, false, dmaPWM, waits)
	cb[1].nextCB = physBuf // Loop back to cb[0]

	// OK with lite channels.
	_, ch := pickChannel()
	if ch == nil {
		buf.Close()
		return nil, nil, errors.New("bcm283x-dma: no channel available")
	}
	ch.startIO(physBuf)

	return ch, buf, nil
}

// physToUncachedPhys returns the uncached physical memory address backing a
// physical memory address.
//
// p must be rooted at a page boundary (4096).
func physToUncachedPhys(p uint32) uint32 {
	// http://en.wikibooks.org/wiki/Aros/Platforms/Arm_Raspberry_Pi_support#Framebuffer
	return p | dramBus
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
	if dmaMemory.channels[6].debug&dmaLite != 0 {
		return errors.New("unexpected hardware: DMA channel #6 shouldn't be lite")
	}
	if dmaMemory.channels[7].debug&dmaLite == 0 {
		return errors.New("unexpected hardware: DMA channel #7 should be lite")
	}
	if dmaMemory.enable != 0x7FFF {
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
			_, waits, err := setPWMClockSource(1000000, 10)
			if err != nil {
				return err
			}
			if err := cb.initBlock(uint32(pSrc), uint32(pDst)+holeSize, size-2*holeSize, false, false, true, true, dmaPWM, waits); err != nil {
				return err
			}
		} else {
			// Use maximum performance.
			if err := cb.initBlock(uint32(pSrc), uint32(pDst)+holeSize, size-2*holeSize, false, false, true, true, dmaFire, 0); err != nil {
				return err
			}
		}
		return runIO(pCB, size-2*holeSize > maxLite)
	}

	return pmem.TestCopy(size, holeSize, alloc, copyMem)
}

// driverDMA implements periph.Driver.
//
// It implements much more than the DMA controller, it also exposes the clocks,
// the PWM and PCM controllers.
type driverDMA struct {
}

func (d *driverDMA) String() string {
	return "bcm283x-dma"
}

func (d *driverDMA) Prerequisites() []string {
	return []string{"bcm283x-gpio"}
}

func (d *driverDMA) Init() (bool, error) {
	// baseAddr is initialized by prerequisite driver bcm283x-gpio.
	if err := pmem.MapAsPOD(uint64(baseAddr+0x7000), &dmaMemory); err != nil {
		if os.IsPermission(err) {
			return true, fmt.Errorf("need more access, try as root: %v", err)
		}
		return true, err
	}
	// Channel #15 is "physically removed from the other DMA Channels so it has a
	// different address base".
	if err := pmem.MapAsPOD(uint64(baseAddr+0xE05000), &dmaChannel15); err != nil {
		return true, err
	}
	if err := pmem.MapAsPOD(uint64(baseAddr+0x203000), &pcmMemory); err != nil {
		return true, err
	}
	if err := pmem.MapAsPOD(uint64(baseAddr+0x20C000), &pwmMemory); err != nil {
		return true, err
	}
	if err := pmem.MapAsPOD(uint64(baseAddr+0x101000), &clockMemory); err != nil {
		return true, err
	}
	if err := pmem.MapAsPOD(uint64(baseAddr+0x3000), &timerMemory); err != nil {
		return true, err
	}
	return true, smokeTest()
}

func (d *driverDMA) Close() error {
	// Stop DMA and PWM controllers.
	return nil
}

func resetDMA(ch int) error {
	if ch < len(dmaMemory.channels) {
		dmaMemory.channels[ch].reset()
	} else if ch == 15 {
		dmaChannel15.reset()
	} else {
		return fmt.Errorf("Invalid dma channel %d.", ch)
	}
	return nil
}

func init() {
	if isArm {
		periph.MustRegister(&driverDMA{})
	}
}
