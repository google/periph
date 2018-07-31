// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Unlike the bcm283x, the allwinner CPUs do not have a "clear bit" and "set
// bit" registers, they only have the data register. Also, allwinner CPUs do
// not support linked lists of DMA buffers. On the other hand, the Allwinner DMA
// controller supports 8 bits transfers instead of 32-128 bits that the bcm283x
// DMA controller supports.
//
// This means that only 8 bits can be used per sample, and only one stream is
// necessary. This results in 1/8th th memory usage than on the bcm283x. The
// drawback is that a block of 8 contiguous GPIO pins must be dedicated to the
// stream.

package allwinner

import (
	"errors"
	"fmt"
	"log"
	"os"

	"periph.io/x/periph"
	"periph.io/x/periph/host/pmem"
)

// dmaMap represents the DMA memory mapped CPU registers.
//
// This map is specific to the currently supported CPUs and will have to be
// adapted as more CPUs are supported. In particular the number of physical
// channels varies across different CPUs.
//
// Note that we modify the DMA controllers without telling the kernel driver.
// The driver keeps its own table of which DMA channel is available so this
// code could effectively crash the whole system. It practice this works.
// #everythingisfine
type dmaMap struct {
	irqEn       dmaR8Irq                // DMA_IRQ_EN_REG
	irqPendStas dmaR8PendingIrq         // DMA_IRQ_PEND_STAS_REG
	reserved0   [(0x100 - 8) / 4]uint32 //
	normal      [8]dmaR8NormalGroup     // 0x100 The "8" "normal" DMA channels (only one active at a time so there's effectively one)
	reserved1   [0x100 / 4]uint32       //
	dedicated   [8]dmaDedicatedGroup    // 0x300 The 8 "dedicated" (as in actually existing) DMA channels
}

func (d *dmaMap) getDedicated() int {
	for i := len(d.dedicated) - 1; i >= 0; i-- {
		if d.dedicated[i].isAvailable() {
			return i
		}
	}
	return -1
}

// dmaNormalGroup is the control registers for the first block of 8 DMA
// controllers.
//
// They can be intentionally slowed down, unlike the dedicated DMA ones.
//
// The big caveat is that only one controller can be active at a time and the
// execution sequence is in accordance with the priority level. This means that
// two normal DMA cannot be used to do simultaneous read and write. This
// feature is critical for bus bitbanging.
type dmaR8NormalGroup struct {
	cfg         ndmaR8Cfg // NDMA_CTRL_REG
	srcAddr     uint32    // NDMA_SRC_ADDR_REG
	dstAddr     uint32    // NDMA_DEST_ADDR_REG
	byteCounter uint32    // NDMA_BC_REG
	reserved    [4]uint32 //
}

func (d *dmaR8NormalGroup) isAvailable() bool {
	return d.cfg == 0 && d.srcAddr == 0 && d.dstAddr == 0 && d.byteCounter == 0
}

func (d *dmaR8NormalGroup) release() error {
	d.srcAddr = 0
	d.dstAddr = 0
	d.byteCounter = 0
	d.cfg = ndmaLoad
	//drvDMA.dmaMemory.irqEn &^= ...
	//drvDMA.dmaMemory.irqPendStas &^= ...
	return nil
}

// dmaNormalGroup is the control registers for the second block of 8 DMA
// controllers.
//
// They support different DReq and can do non-linear streaming.
type dmaDedicatedGroup struct {
	cfg         ddmaR8Cfg   // DDMA_CTRL_REG
	srcAddr     uint32      // DDMA_SRC_ADDR_REG
	dstAddr     uint32      // DDMA_DEST_ADDR_REG
	byteCounter uint32      // DDMA_BC_REG (24 bits)
	reserved0   [2]uint32   //
	param       ddmaR8Param // DDMA_PARA_REG (dedicated DMA only)
	reserved1   uint32      //
}

func (d *dmaDedicatedGroup) isAvailable() bool {
	return d.cfg == 0 && d.srcAddr == 0 && d.dstAddr == 0 && d.byteCounter == 0 && d.param == 0
}

func (d *dmaDedicatedGroup) set(srcAddr, dstAddr, l uint32, srcIO, dstIO bool, src ddmaR8Cfg) {
	d.srcAddr = srcAddr
	d.dstAddr = dstAddr
	d.byteCounter = l
	// TODO(maruel): Slow down the clock by another 2*250x
	//d.param = ddmaR8Param(250 | 250<<16)
	d.param = ddmaR8Param(1<<24 | 1<<8 | 1)
	// All these have value 0. This statement only exist for documentation.
	cfg := ddmaDstWidth8 | ddmaDstBurst1 | ddmaDstLinear | ddmaSrcWidth8 | ddmaSrcLinear | ddmaSrcBurst1
	cfg |= src | ddmaBCRemain
	if srcIO {
		cfg |= ddmaSrcIOMode
	} else if dstIO {
		cfg |= ddmaDstIOMode
	}
	d.cfg = ddmaLoad | cfg
	for i := 0; d.cfg&ddmaLoad != 0 && i < 100000; i++ {
	}
	if d.cfg&ddmaLoad != 0 {
		log.Printf("failed to load DDMA: %# v\n", d)
	}
}

func (d *dmaDedicatedGroup) release() error {
	d.param = 0
	d.srcAddr = 0
	d.dstAddr = 0
	d.byteCounter = 0
	d.cfg = ddmaLoad
	//drvDMA.dmaMemory.irqEn &^= ...
	//drvDMA.dmaMemory.irqPendStas &^= ...
	return nil
}

const (
	// 31 reserved
	dma7QueueEndIrq   dmaA64Irq = 1 << 30 // DMA7_END_IRQ_EN; DMA 7 Queue End Transfer Interrupt Enable.
	dma7PackageEndIrq dmaA64Irq = 1 << 29 // DMA7_PKG_IRQ_EN; DMA 7 Package End Transfer Interrupt Enable.
	dma7HalfIrq       dmaA64Irq = 1 << 28 // DMA7_HLAF_IRQ_EN; DMA 7 Half Package Transfer Interrupt Enable.
	// ...
	// 3 reserved
	dma0QueueEndIrq   dmaA64Irq = 1 << 2 // DMA0_END_IRQ_EN; DMA 0 Queue End Transfer Interrupt Enable.
	dma0PackageEndIrq dmaA64Irq = 1 << 1 // DMA0_PKG_IRQ_EN; DMA 0 Package End Transfer Interrupt Enable.
	dma0HalfIrq       dmaA64Irq = 1 << 0 // DMA0_HLAF_IRQ_EN; DMA 0 Half Package Transfer Interrupt Enable.
)

// DMA_IRQ_EN_REG
// A64: Page 199-201.
type dmaA64Irq uint32

const (
	ddma7EndIrq   dmaR8Irq = 1 << 31 // DDMA7_END_IRQ_EN
	ddma7HalfIreq dmaR8Irq = 1 << 30 // DDMA7_HF_IRQ_EN
	// ...
	ddma0EndIrq   dmaR8Irq = 1 << 17 // DDMA0_END_IRQ_EN
	ddma0HalfIreq dmaR8Irq = 1 << 16 // DDMA0_HF_IRQ_EN
	ndma7EndIrq   dmaR8Irq = 1 << 15 // NDMA7_END_IRQ_EN
	ndma7HalfIreq dmaR8Irq = 1 << 16 // NDDMA7_HF_IRQ_EN
	// ...
	ndma0EndIrq dmaR8Irq = 1 << 1 // NDMA0_END_IRQ_EN
	ndma0HFIreq dmaR8Irq = 1 << 0 // NDMA0_HF_IRQ_EN
)

// DMA_IRQ_EN_REG
// R8: Page 124-126.
type dmaR8Irq uint32

const (
	// 31 reserved
	dma7QueueEndIrqPend   dmaA64PendingIrq = 1 << 30 // DMA7_QUEUE_IRQ_PEND; DMA 7 Queue End Transfer Interrupt Pending. Set 1 to the bit will clear it.
	dma7PackageEndIrqPend dmaA64PendingIrq = 1 << 29 // DMA7_PKG_IRQ_PEND; DMA 7 Package End Transfer Interrupt Pending. Set 1 to the bit will clear it.
	dma7HalfIrqPend       dmaA64PendingIrq = 1 << 28 // DMA7_HLAF_IRQ_PEND; DMA 7 Half Package Transfer Interrupt Pending. Set 1 to the bit will clear it.
	// ...
	// 3 reserved
	dma0QueueEndIrqPend   dmaA64PendingIrq = 1 << 2 // DMA0_QUEUE_IRQ_PEND; DMA 0 Queue End Transfer Interrupt Pending. Set 1 to the bit will clear it.
	dma0PackageEndIrqPend dmaA64PendingIrq = 1 << 1 // DMA0_PKG_IRQ_PEND; DMA 0 Package End Transfer Interrupt Pending. Set 1 to the bit will clear it.
	dma0HalfIrqPend       dmaA64PendingIrq = 1 << 0 // DMA0_HLAF_IRQ_PEND; DMA 0 Half Package Transfer Interrupt Pending. Set 1 to the bit will clear it.
)

// DMA_IRQ_PEND_REG0
// A64: Page 201-203.
type dmaA64PendingIrq uint32

const (
	ddma7EndIrqPend   dmaR8PendingIrq = 1 << 31 // DDMA7_END_IRQ_PEND
	ddma7HalfIreqPend dmaR8PendingIrq = 1 << 30 // DDMA7_HF_IRQ_PEND
	// ...
	ddma0EndIrqPend   dmaR8PendingIrq = 1 << 17 // DDMA0_END_IRQ_PEND
	ddma0HalfIreqPend dmaR8PendingIrq = 1 << 16 // DDMA0_HF_IRQ_PEND
	ndma7EndIrqPend   dmaR8PendingIrq = 1 << 15 // NDMA7_END_IRQ_PEND
	ndma7HalfIreqPend dmaR8PendingIrq = 1 << 16 // NDDMA7_HF_IRQ_PEND
	// ...
	ndma0EndIrqPend   dmaR8PendingIrq = 1 << 1 // NDMA0_END_IRQ_PEND
	ndma0HalfIreqPend dmaR8PendingIrq = 1 << 0 // NDMA0_HF_IRQ_PEND
)

// DMA_IRQ_PEND_STAS_REG
// R8: Page 126-129.
type dmaR8PendingIrq uint32

const (
	ndmaLoad       ndmaR8Cfg = 1 << 31 // NDMA_LOAD
	ndmaContinuous ndmaR8Cfg = 1 << 30 // NDMA_CONTI_EN Continuous mode
	ndmaWaitClk0   ndmaR8Cfg = 0 << 27 // NDMA_WAIT_STATE Number of clock to wait for
	ndmaWaitClk2   ndmaR8Cfg = 1 << 27 // 2(n+1)
	ndmaWaitClk6   ndmaR8Cfg = 2 << 27 //
	ndmaWaitClk8   ndmaR8Cfg = 3 << 27 //
	ndmaWaitClk10  ndmaR8Cfg = 4 << 27 //
	ndmaWaitClk12  ndmaR8Cfg = 5 << 27 //
	ndmaWaitClk14  ndmaR8Cfg = 6 << 27 //
	ndmaWaitClk16  ndmaR8Cfg = 7 << 27 //
	ndmaDstWidth32 ndmaR8Cfg = 2 << 25 // NDMA_DST_DATA_WIDTH
	ndmaDstWidth16 ndmaR8Cfg = 1 << 25 //
	ndmaDstWidth8  ndmaR8Cfg = 0 << 25 //
	ndmaDstBurst8  ndmaR8Cfg = 2 << 23 // NDMA_DST_BST_LEN
	ndmaDstBurst4  ndmaR8Cfg = 1 << 23 //
	ndmaDstBurst1  ndmaR8Cfg = 0 << 23 //
	// 22 reserved NDMA_CFG_DST_NON_SECURE ?
	ndmaDstAddrNoInc  ndmaR8Cfg = 1 << 21  // NDMA_DST_ADDR_TYPE
	ndmaDstDrqIRTX    ndmaR8Cfg = 0 << 16  // NDMA_DST_DRQ_TYPE
	ndmaDstDrqUART1TX ndmaR8Cfg = 9 << 16  //
	ndmaDstDrqUART3TX ndmaR8Cfg = 11 << 16 //
	ndmaDstDrqAudio   ndmaR8Cfg = 19 << 16 // 24.576MHz (Page 53)
	ndmaDstDrqSRAM    ndmaR8Cfg = 21 << 16 //
	ndmaDstDrqSPI0TX  ndmaR8Cfg = 24 << 16 //
	ndmaDstDrqSPI1TX  ndmaR8Cfg = 25 << 16 //
	ndmaDstDrqSPI2TX  ndmaR8Cfg = 26 << 16 //
	ndmaDstDrqUSB1    ndmaR8Cfg = 27 << 16 // 480MHz
	ndmaDstDrqUSB2    ndmaR8Cfg = 28 << 16 //
	ndmaDstDrqUSB3    ndmaR8Cfg = 29 << 16 //
	ndmaDstDrqUSB4    ndmaR8Cfg = 30 << 16 //
	ndmaDstDrqUSB5    ndmaR8Cfg = 31 << 16 //
	ndmaBCRemain      ndmaR8Cfg = 1 << 15  // BC_MODE_SEL
	// 14:11 reserved
	ndmaSrcWidth32 ndmaR8Cfg = 2 << 9 // NDMA_SRC_DATA_WIDTH
	ndmaSrcWidth16 ndmaR8Cfg = 1 << 9 //
	ndmaSrcWidth8  ndmaR8Cfg = 0 << 9 //
	ndmaSrcBurst8  ndmaR8Cfg = 2 << 7 // NDMA_SRC_BST_LEN
	ndmaSrcBurst4  ndmaR8Cfg = 1 << 7 //
	ndmaSrcBurst1  ndmaR8Cfg = 0 << 7 //
	// 6 reserved NDMA_CFG_SRC_NON_SECURE ?
	ndmaSrcAddrNoInc  ndmaR8Cfg = 1 << 5  // NDMA_SRC_ADDR_TYPE
	ndmaSrcDrqIRTX    ndmaR8Cfg = 0 << 0  // NDMA_SRC_DRQ_TYPE
	ndmaSrcDrqUART1RX ndmaR8Cfg = 9 << 0  //
	ndmaSrcDrqUART3RX ndmaR8Cfg = 11 << 0 //
	ndmaSrcDrqAudio   ndmaR8Cfg = 19 << 0 // 24.576MHz (Page 53)
	ndmaSrcDrqSRAM    ndmaR8Cfg = 21 << 0 //
	ndmaSrcDrqSDRAM   ndmaR8Cfg = 22 << 0 // 0~400MHz
	ndmaSrcDrqTPAD    ndmaR8Cfg = 23 << 0 //
	ndmaSrcDrqSPI0RX  ndmaR8Cfg = 24 << 0 //
	ndmaSrcDrqSPI1RX  ndmaR8Cfg = 25 << 0 //
	ndmaSrcDrqSPI2RX  ndmaR8Cfg = 26 << 0 //
	ndmaSrcDrqUSB1    ndmaR8Cfg = 27 << 0 // 480MHz
	ndmaSrcDrqUSB2    ndmaR8Cfg = 28 << 0 //
	ndmaSrcDrqUSB3    ndmaR8Cfg = 29 << 0 //
	ndmaSrcDrqUSB4    ndmaR8Cfg = 30 << 0 //
	ndmaSrcDrqUSB5    ndmaR8Cfg = 31 << 0 //
)

// NDMA_CTRL_REG
// R8: Page 129-131.
type ndmaR8Cfg uint32

const (
	ddmaLoad       ddmaR8Cfg = 1 << 31 // DDMA_LOAD
	ddmaBusy       ddmaR8Cfg = 1 << 30 // DDMA_BSY_STA
	ddmaContinuous ddmaR8Cfg = 1 << 29 // DDMA_CONTI_MODE_EN
	// 28:27 reserved  28 = DDMA_CFG_DST_NON_SECURE ?
	ddmaDstWidth32     ddmaR8Cfg = 2 << 25  // DDMA_DST_DATA_WIDTH
	ddmaDstWidth16     ddmaR8Cfg = 1 << 25  //
	ddmaDstWidth8      ddmaR8Cfg = 0 << 25  //
	ddmaDstBurst8      ddmaR8Cfg = 2 << 23  // DDMA_DST_BST_LEN
	ddmaDstBurst4      ddmaR8Cfg = 1 << 23  //
	ddmaDstBurst1      ddmaR8Cfg = 0 << 23  //
	ddmaDstVertical    ddmaR8Cfg = 3 << 21  // DDMA_ADDR_MODE; no idea what it's use it. It's not explained in the datasheet ...
	ddmaDstHorizontal  ddmaR8Cfg = 2 << 21  // ... and the official drivers/dma/sun6i-dma.c driver doesn't use it
	ddmaDstIOMode      ddmaR8Cfg = 1 << 21  // Non incrementing
	ddmaDstLinear      ddmaR8Cfg = 0 << 21  // Normal incrementing position
	ddmaDstDrqSRAM     ddmaR8Cfg = 0 << 16  // DDMA_DST_DRQ_SEL
	ddmaDstDrqSDRAM    ddmaR8Cfg = 1 << 16  // DDR ram speed
	ddmaDstDrqNAND     ddmaR8Cfg = 3 << 16  //
	ddmaDstDrqUSB0     ddmaR8Cfg = 4 << 16  //
	ddmaDstDrqSPI1TX   ddmaR8Cfg = 8 << 16  //
	ddmaDstDrqCryptoTX ddmaR8Cfg = 10 << 16 //
	ddmaDstDrqTCON0    ddmaR8Cfg = 14 << 16 //
	ddmaDstDrqSPI0TX   ddmaR8Cfg = 26 << 16 //
	ddmaDstDrqSPI2TX   ddmaR8Cfg = 28 << 16 //
	ddmaBCRemain       ddmaR8Cfg = 1 << 15  // BC_MODE_SEL
	// 14:11 reserved
	ddmaSrcWidth32    ddmaR8Cfg = 2 << 9 // DDMA_SRC_DATA_WIDTH
	ddmaSrcWidth16    ddmaR8Cfg = 1 << 9 //
	ddmaSrcWidth8     ddmaR8Cfg = 0 << 9 //
	ddmaSrcBurst8     ddmaR8Cfg = 2 << 7 // DDMA_SRC_BST_LEN
	ddmaSrcBurst4     ddmaR8Cfg = 1 << 7 //
	ddmaSrcBurst1     ddmaR8Cfg = 0 << 7 //
	ddmaSrcVertical   ddmaR8Cfg = 3 << 5 // DDMA_SRC_ADDR_MODE
	ddmaSrcHorizontal ddmaR8Cfg = 2 << 5 //
	ddmaSrcIOMode     ddmaR8Cfg = 1 << 5 // Non incrementing
	ddmaSrcLinear     ddmaR8Cfg = 0 << 5 // Normal incrementing position
	// 4:0 drq
	ddmaSrcDrqSRAM     ddmaR8Cfg = 0 << 0  // DDMA_SRC_DRQ_TYPE
	ddmaSrcDrqSDRAM    ddmaR8Cfg = 1 << 0  //
	ddmaSrcDrqNAND     ddmaR8Cfg = 3 << 0  //
	ddmaSrcDrqUSB0     ddmaR8Cfg = 4 << 0  //
	ddmaSrcDrqSPI1RX   ddmaR8Cfg = 9 << 0  //
	ddmaSrcDrqCryptoRX ddmaR8Cfg = 11 << 0 //
	ddmaSrcDrqSPI0RX   ddmaR8Cfg = 27 << 0 //
	ddmaSrcDrqSPI2RX   ddmaR8Cfg = 29 << 0 //
)

// DDMA_CFG_REG
// R8: Page 131-134.
type ddmaR8Cfg uint32

const (
	// For each value, N+1 is actually used.
	ddmaDstBlkSizeMask      ddmaR8Param = 0xFF << 24 // DEST_DATA_BLK_SIZE
	ddmaDstWaitClkCycleMask ddmaR8Param = 0xFF << 16 // DEST_WAIT_CLK_CYC
	ddmaSrcBlkSizeMask      ddmaR8Param = 0xFF << 8  // SRC_DATA_BLK_SIZE
	ddmaSrcWaitClkCycleMask ddmaR8Param = 0xFF << 0  // SRC_WAIT_CLK_CYC
)

// DDMA_PARA_REG
// R8: Page 134.
type ddmaR8Param uint32

// smokeTest allocates two physical pages, ask the DMA controller to copy the
// data from one page to another (with a small offset) and make sure the
// content is as expected.
//
// This should take a fraction of a second and will make sure the driver is
// usable.
func smokeTest() error {
	const size = 4096  // 4kb
	const holeSize = 1 // Minimum DMA alignment.

	alloc := func(s int) (pmem.Mem, error) {
		return pmem.Alloc(s)
	}

	copyMem := func(pDst, pSrc uint64) error {
		n := drvDMA.dmaMemory.getDedicated()
		if n == -1 {
			return errors.New("no channel available")
		}
		drvDMA.dmaMemory.irqEn &^= 3 << uint(2*n+16)
		drvDMA.dmaMemory.irqPendStas = 3 << uint(2*n+16)
		ch := &drvDMA.dmaMemory.dedicated[n]
		defer func() {
			_ = ch.release()
		}()
		ch.set(uint32(pSrc), uint32(pDst)+holeSize, 4096-2*holeSize, false, false, ddmaDstDrqSDRAM|ddmaSrcDrqSDRAM)

		for ch.cfg&ddmaBusy != 0 {
		}
		return nil
	}

	return pmem.TestCopy(size, holeSize, alloc, copyMem)
}

// driverDMA implements periph.Driver.
//
// It implements much more than the DMA controller, it also exposes the clocks,
// the PWM and PCM controllers.
type driverDMA struct {
	// dmaMemory is the memory map of the CPU DMA registers.
	dmaMemory *dmaMap
	// pwmMemory is the memory map of the CPU PWM registers.
	pwmMemory *pwmMap
	// spiMemory is the memory mapping for the spi CPU registers.
	spiMemory *spiMap
	// clockMemory is the memory mapping for the clock CPU registers.
	clockMemory *clockMap
	// timerMemory is the memory mapping for the timer CPU registers.
	timerMemory *timerMap
}

func (d *driverDMA) String() string {
	return "allwinner-dma"
}

func (d *driverDMA) Prerequisites() []string {
	return []string{"allwinner-gpio"}
}

func (d *driverDMA) After() []string {
	return nil
}

func (d *driverDMA) Init() (bool, error) {
	// dmaBaseAddr is the physical base address of the DMA registers.
	var dmaBaseAddr uint32
	// pwmBaseAddr is the physical base address of the PWM registers.
	var pwmBaseAddr uint32
	// spiBaseAddr is the physical base address of the clock registers.
	var spiBaseAddr uint32
	// clockBaseAddr is the physical base address of the clock registers.
	var clockBaseAddr uint32
	// timerBaseAddr is the physical base address of the timer registers.
	var timerBaseAddr uint32
	if IsA64() {
		// Page 198.
		dmaBaseAddr = 0x1C02000
		// Page 194.
		pwmBaseAddr = 0x1C21400
		// Page 161.
		timerBaseAddr = 0x1C20C00
		// Page 81.
		clockBaseAddr = 0x1C20000
		// Page Page 545.
		spiBaseAddr = 0x01C68000
	} else if IsR8() {
		// Page 124.
		dmaBaseAddr = 0x1C02000
		// Page 83.
		pwmBaseAddr = 0x1C20C00 + 0x200
		// Page 85.
		timerBaseAddr = 0x1C20C00
		// Page 57.
		clockBaseAddr = 0x1C20000
		// Page 151.
		spiBaseAddr = 0x01C05000
	} else {
		// H3
		// Page 194.
		//dmaBaseAddr = 0x1C02000
		// Page 187.
		//pwmBaseAddr = 0x1C21400
		// Page 154.
		//timerBaseAddr = 0x1C20C00
		return false, errors.New("unsupported CPU architecture")
	}

	if err := pmem.MapAsPOD(uint64(dmaBaseAddr), &d.dmaMemory); err != nil {
		if os.IsPermission(err) {
			return true, fmt.Errorf("need more access, try as root: %v", err)
		}
		return true, err
	}

	if err := pmem.MapAsPOD(uint64(pwmBaseAddr), &d.pwmMemory); err != nil {
		return true, err
	}
	if err := pmem.MapAsPOD(uint64(timerBaseAddr), &d.timerMemory); err != nil {
		return true, err
	}
	if err := pmem.MapAsPOD(uint64(clockBaseAddr), &d.clockMemory); err != nil {
		return true, err
	}
	if err := pmem.MapAsPOD(uint64(spiBaseAddr), &d.spiMemory); err != nil {
		return true, err
	}

	return true, smokeTest()
}

func (d *driverDMA) Close() error {
	// Stop DMA and PWM controllers.
	return nil
}

func init() {
	if false && isArm {
		// TODO(maruel): This is intense, wait to be sure it works.
		periph.MustRegister(&drvDMA)
	}
}

var drvDMA driverDMA
