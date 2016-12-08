// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Unlike the bcm283x, the allwinner CPUs do not have a "clear bit" and "set
// bit" registers, they only have the data register. Also, allwinner CPUs do
// not support linked list of DMA buffers. On the other hand, the Allwinner DMA
// controllers support 8 bits transfers instead of 32-128 bits that the bcm283x
// DMA controllers supports.
//
// This means that only 8 bits can be used per sample, and only one stream is
// necessary. This results in 8 less memory usage than on the bcm283x. The
// drawback is that a block of 8 contiguous GPIO pins must be dedicated to the
// stream.

package allwinner

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
