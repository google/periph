// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package allwinner

import (
	"errors"
	"fmt"
	"log"
	"time"
)

const (
	// 31:20 reserved
	// Set this bit to ‘1’ to make the internal read sample point with a delay of
	// half cycle of SPI_CLK. It is used in high speed read operation to reduce
	// the error caused by the time delay of SPI_CLK propagating between master
	// and slave.
	// 1 – delay internal read sample point
	// 0 – normal operation, do not delay internal read sample point
	spiR8HalfDelay       spiR8Ctl = 1 << 19 // Master Sample Data Control
	spiR8TransmitPause   spiR8Ctl = 1 << 18 // Transmit Pause Enable
	spiR8CSLevel         spiR8Ctl = 1 << 17 // SS_LEVEL; Chip Select level
	spiR8CSManual        spiR8Ctl = 1 << 16 // SS_CTRL; Do not switch CS automatically
	spiR8DiscardHash     spiR8Ctl = 1 << 15 // DHB
	spiR8DummyBurst      spiR8Ctl = 1 << 14 // DDB
	spiR8CS0             spiR8Ctl = 0 << 12 // SS; Which CS line to use. For SPI0 only
	spiR8CS1             spiR8Ctl = 1 << 12 //
	spiR8CS2             spiR8Ctl = 2 << 12 //
	spiR8CS3             spiR8Ctl = 3 << 12 //
	spiR8RapidsReadMode  spiR8Ctl = 1 << 11 // RPSM
	spiR8ExchangeBurst   spiR8Ctl = 1 << 10 // XCH
	spiR8RXFIFOReset     spiR8Ctl = 1 << 9  // RXFIFO Reset; Write to reset the FIFO as empty
	spiR8TXFIFOReset     spiR8Ctl = 1 << 8  // TXFIFO Reset; Write to reset the FIFO as empty
	spiR8CSBetweenBursts spiR8Ctl = 1 << 7  // SSCTL
	spiR8LSB             spiR8Ctl = 1 << 6  // LMTF; MSB by default, LSB when set
	spiR8DDMA            spiR8Ctl = 1 << 5  // DMAM; Use dedicated DMA if set, normal DMA otherwise
	spiR8CSActiveLow     spiR8Ctl = 1 << 4  // SSPOL; CS line polarity
	spiR8ClkActiveLow    spiR8Ctl = 1 << 3  // POL; Clock line polarity
	spiR8PHA             spiR8Ctl = 1 << 2  // PHA; Phase 1 if set (leading edge for setup data)
	spiR8Master          spiR8Ctl = 1 << 1  // MODE; Slave mode if not set
	spiR8Enable          spiR8Ctl = 1 << 0  // EN; Enable mode
)

// SPI_CTL
// R8: Page 153-155.  Default: 0x0002001C
type spiR8Ctl uint32

// SPI_INTCTL
// R8: Page 155-156.
type spiR8IntCtl uint32

const (
	spiR8ClearInterrupt spiR8IntStatus = 1 << 31 // Clear interrupt busy flag
	// 30:18 reserved
	spiR8InvalidSS spiR8IntStatus = 1 << 17 // SSI
	spiR8TC        spiR8IntStatus = 1 << 16 // TC; Transfer Completed
)

// SPI_INT_STA
// R8: Page 156-157.
type spiR8IntStatus uint32

const (
	// 31:13 reserved
	spiR8DMATX3Quarter spiR8DMACtl = 1 << 12 // TXFIFO 3/4 empty
	spiR8DMATX1Quarter spiR8DMACtl = 1 << 11 // TXFIFO 1/4 empty
	spiR8DMATXByte     spiR8DMACtl = 1 << 10 // TXFIFO Not Full
	spiR8DMATXHalf     spiR8DMACtl = 1 << 9  // TXFIFO 1/2 empty
	spiR8DMATXEmpty    spiR8DMACtl = 1 << 8  // TXFIFO empty
	// 7:5 reserved
	spiR8DMARX3Quarter spiR8DMACtl = 1 << 4 // RXFIFO 3/4 empty
	spiR8DMARX1Quarter spiR8DMACtl = 1 << 3 // RXFIFO 1/4 empty
	spiR8DMARXByte     spiR8DMACtl = 1 << 2 // RXFIFO Not Full
	spiR8DMARXHalf     spiR8DMACtl = 1 << 1 // RXFIFO 1/2 empty
	spiR8DMARXEmpty    spiR8DMACtl = 1 << 0 // RXFIFO empty
)

// SPI_DMACTL
// R8: Page 158.
type spiR8DMACtl uint32

const (
	// 31:13 reserved
	spiR8DivRateSelect2 spiR8ClockCtl = 1 << 12 // DRS; Use spiDivXX if set, use mask otherwise
	spiR8Div2           spiR8ClockCtl = 0 << 8  // CDR1; Use divisor 2^(n+1)
	spiR8Div4           spiR8ClockCtl = 1 << 8  //
	spiR8Div8           spiR8ClockCtl = 2 << 8  //
	spiR8Div16          spiR8ClockCtl = 3 << 8  //
	spiR8Div32          spiR8ClockCtl = 4 << 8  //
	spiR8Div64          spiR8ClockCtl = 5 << 8  //
	spiR8Div128         spiR8ClockCtl = 6 << 8  //
	spiR8Div256         spiR8ClockCtl = 7 << 8  //
	spiR8Div512         spiR8ClockCtl = 8 << 8  //
	spiR8Div1024        spiR8ClockCtl = 9 << 8  //
	spiR8Div2048        spiR8ClockCtl = 10 << 8 //
	spiR8Div4096        spiR8ClockCtl = 11 << 8 //
	spiR8Div8192        spiR8ClockCtl = 12 << 8 //
	spiR8Div16384       spiR8ClockCtl = 13 << 8 //
	spiR8Div32768       spiR8ClockCtl = 14 << 8 //
	spiR8Div65536       spiR8ClockCtl = 15 << 8 //
	spiR8Div1Mask       spiR8ClockCtl = 0xFF    // CDR2; Use divisor 2*(n+1)
)

// SPI_CCTL
// R8: Page 159.
type spiR8ClockCtl uint32

const (
	// 31:25 reserved
	spiR8FIFOTXShift = 16 // 0 to 64
	// 15:7 reserved
	spiR8FIFORXShift = 0 // 0 to 64
)

// SPI_FIFO_STA
// R8: Page 160.
type spiR8FIFOStatus uint32

func (s spiR8FIFOStatus) tx() uint8 {
	return uint8((uint32(s) >> 16) & 127)
}

func (s spiR8FIFOStatus) rx() uint8 {
	return uint8(uint32(s) & 127)
}

// spiR8Group is the mapping of SPI registers for one SPI controller.
// R8: Page 152-153.
type spiR8Group struct {
	rx              uint32          // 0x00 SPI_RX_DATA RX Data
	tx              uint32          // 0x04 SPI_TX_DATA TX Data
	ctl             spiR8Ctl        // 0x08 SPI_CTL Control
	intCtl          spiR8IntCtl     // 0x0C SPI_INTCTL Interrupt Control
	status          spiR8IntStatus  // 0x10 SPI_ST Status
	dmaCtl          spiR8DMACtl     // 0x14 SPI_DMACTL DMA Control
	wait            uint32          // 0x18 SPI_WAIT Clock Counter; 16 bits
	clockCtl        spiR8ClockCtl   // 0x1C SPI_CCTL Clock Rate Control
	burstCounter    uint32          // 0x20 SPI_BC Burst Counter; 24 bits
	transmitCounter uint32          // 0x24 SPI_TC Transmit Counter; 24 bits
	fifoStatus      spiR8FIFOStatus // 0x28 SPI_FIFO_STA FIFO Status
	reserved        [(0x1000 - 0x02C) / 4]uint32
}

func (s *spiR8Group) setup() {
	s.intCtl = 0
	s.status = 0
	//s.dmaCtl = spiR8DMARXByte
	s.dmaCtl = 0
	s.wait = 2
	s.clockCtl = spiR8DivRateSelect2 | spiR8Div1024
	// spiR8DDMA
	s.ctl = spiR8CSManual | spiR8LSB | spiR8Master | spiR8Enable
}

// spiMap is the mapping of SPI registers.
// R8: Page 152-153.
type spiMap struct {
	groups [3]spiR8Group
}

// spi2Write do a write on SPI2_MOSI via polling.
func spi2Write(w []byte) error {
	if drvDMA.clockMemory == nil || drvDMA.spiMemory == nil {
		return errors.New("subsystem not initialized")
	}
	// Make sure the source clock is disabled. Set it at 250kHz.
	//drvDMA.clockMemory.spi2Clk &^= clockSPIEnable
	drvDMA.clockMemory.spi2Clk |= clockSPIEnable
	drvDMA.clockMemory.spi2Clk = clockSPIDiv8a | clockSPIDiv12b
	ch := &drvDMA.spiMemory.groups[2]
	ch.setup()
	fmt.Printf("Setup done\n")
	for i := 0; i < len(w)/4; i++ {
		// TODO(maruel): Access it in 8bit mode.
		ch.tx = uint32(w[0])
		for ch.fifoStatus.tx() == 0 {
			log.Printf("Waiting for bit %# v\n", ch)
			time.Sleep(time.Second)
		}
	}
	fmt.Printf("Done\n")
	return nil
}

// spi2Read do a read on SPI2_MISO via polling.
func spi2Read(r []byte) error {
	if drvDMA.clockMemory == nil || drvDMA.spiMemory == nil {
		return errors.New("subsystem not initialized")
	}
	// Make sure the source clock is disabled. Set it at 250kHz.
	//drvDMA.clockMemory.spi2Clk &^= clockSPIEnable
	drvDMA.clockMemory.spi2Clk |= clockSPIEnable
	drvDMA.clockMemory.spi2Clk = clockSPIDiv8a | clockSPIDiv12b
	ch := &drvDMA.spiMemory.groups[2]
	ch.setup()
	for i := 0; i < len(r)/4; i++ {
		ch.tx = 0
		for ch.status&spiR8TC == 0 {
		}
		// TODO(maruel): Access it in 8bit mode.
		r[i] = uint8(ch.rx)
	}
	fmt.Printf("Done\n")
	return nil
}
