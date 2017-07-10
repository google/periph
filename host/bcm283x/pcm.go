// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// pcm means I2S.

package bcm283x

var pcmMemory *pcmMap

type pcmCS uint32

// Pages 126-129
const (
	// 31:26 reserved
	pcmStandby      pcmCS = 1 << 25 // STBY Allow at least 4 PCM clock cycles to take effect
	pcmSync         pcmCS = 1 << 24 // SYNC Two PCM clocks have occurred since last write
	pcmRXSignExtend pcmCS = 1 << 23 // RXSEX Sign extend RXZ data
	pcmRXFull       pcmCS = 1 << 22 // RXF RX FIFO is full
	pcmTXEmpty      pcmCS = 1 << 21 // TXE TX FIFO is empty
	pcmRXData       pcmCS = 1 << 20 // RXD RX FIFO contains data
	pcmTXData       pcmCS = 1 << 19 // TXD TX FIFO ready to accept data
	pcmRXR          pcmCS = 1 << 18 // RXR RX FIFO needs reading
	pcmTXW          pcmCS = 1 << 17 // TXW TX FIFO needs writing
	pcmRXErr        pcmCS = 1 << 16 // RXERR RX FIFO error
	pcmTXErr        pcmCS = 1 << 15 // TXERR TX FIFO error
	pcmRXSync       pcmCS = 1 << 14 // RXSYNC RX FIFO is out of sync
	pcmTXSync       pcmCS = 1 << 13 // TXSYNC TX FIFO is out of sync
	// 12:10 reserved
	pcmDMAEnable pcmCS = 1 << 9 // DMAEN Generate TX&RX DMA DREQ
	// 8:7 RXTHR controls when pcmRXR is set
	pcmRXThresholdOne  pcmCS = 0 << 7 // One sample in RX FIFO
	pcmRXThreshold1    pcmCS = 1 << 7 // RX FIFO is at least (?) full
	pcmRXThreshold2    pcmCS = 2 << 7 // ?
	pcmRXThresholdFull pcmCS = 3 << 7 // RX is full
	// 6:5 TXTHR controls when pcmTXW is set
	pcmTXThresholdEmpty    pcmCS = 0 << 5 // TX FIFO is empty
	pcmTXThresholdNotFull1 pcmCS = 1 << 5 // At least one sample can be put
	pcmTXThresholdNotFull2 pcmCS = 2 << 5 // At least one sample can be put
	pcmTXThresholdOne      pcmCS = 3 << 5 // One sample can be put
	pcmRXClear             pcmCS = 1 << 4 // RXCLR Clear RX FIFO; takes 2 PCM clock to take effect
	pcmTXClear             pcmCS = 1 << 3 // TXCLR Clear TX FIFO; takes 2 PCM clock to take effect
	pcmTXEnable            pcmCS = 1 << 2 // TXON Enable TX
	pcmRXEnable            pcmCS = 1 << 1 // RXON Enable FX
	pcmEnable              pcmCS = 1 << 0 // EN Enable the PCM
)

type pcmMode uint32

// Page 129-131
const (
	// 31:29 reserved
	pcmClockDisable     pcmMode = 1 << 28                     // CLK_DIS Cleanly disable the PCM clock
	pcmDecimation32     pcmMode = 1 << 27                     // PDMN; 0 is factor 16, 1 is factor 32
	pcmRXPDMFilter      pcmMode = 1 << 26                     // PDME Enable input CIC filter on PDM input
	pcmRXMerge          pcmMode = 1 << 25                     // FRXP Merge both channels as single FIFO entry
	pcmTXMerge          pcmMode = 1 << 24                     // FTXP Merge both channels as singe FIFO entry
	pcmClockSlave       pcmMode = 1 << 23                     // CLKM PCM CLK is input
	pcmClockInverted    pcmMode = 1 << 22                     // CLKI Inverse clock signal
	pcmFSSlave          pcmMode = 1 << 21                     // FSM PCM FS is input
	pcmFSInverted       pcmMode = 1 << 20                     // FSI Invese FS signal
	pcmFrameLengthShift         = 10                          //
	pcmFrameLenghtMask  pcmMode = 0x3F << pcmFrameLengthShift // FLEN Frame length + 1
	pcmFSLenghtMask     pcmMode = 0x3F << 0                   // FSLEN FS pulse clock width
)

type pcmRX uint32

// Page 131-132
const (
	pcmRX1Width     pcmRX = 1 << 31 // CH1WEX Legacy
	pcmRX1Enable    pcmRX = 1 << 30 // CH1EN
	pcmRX1PosShift        = 20
	pcmRX1PosMask   pcmRX = 0x3F << pcmRX1PosShift // CH1POS Clock delay
	pcmRX1Channel16 pcmRX = 8 << 16                // CH1WID (Arbitrary width between 8 and 16 is supported)
	pcmRX2Width     pcmRX = 1 << 15                // CH2WEX Legacy
	pcmRX2Enable    pcmRX = 1 << 14                // CH2EN
	pcmRX2PosShift        = 4
	pcmRX2PosMask   pcmRX = 0x3F << pcmRX2PosShift // CH2POS Clock delay
	pcmRX2Channel16 pcmRX = 8 << 0                 // CH2WID (Arbitrary width between 8 and 16 is supported)
)

type pcmTX uint32

// Page 133-134
const (
	pcmTX1Width     pcmTX = 1 << 31 // CH1WX Legacy
	pcmTX1Enable    pcmTX = 1 << 30 // CH1EN Enable channel 1
	pcmTX1PosShift        = 20
	pcmTX1PosMask   pcmTX = 0x3F << pcmTX1PosShift // CH1POS Clock delay
	pcmTX1Channel16 pcmTX = 8 << 16                // CH1WID (Arbitrary width between 8 and 16 is supported)
	pcmTX2Width     pcmTX = 1 << 15                // CH2WEX Legacy
	pcmTX2Enable    pcmTX = 1 << 14                // CH2EN
	pcmTX2PosShift        = 4
	pcmTX2PosMask   pcmTX = 0x3F << pcmTX2PosShift // CH2POS Clock delay
	pcmTX2Channel16 pcmTX = 8 << 0                 // CH2WID (Arbitrary width between 8 and 16 is supported)
)

type pcmDreq uint32

// Page 134-135
const (
	// 31 reserved
	pcmDreqTXPanicShift         = 24
	pcmDreqTXPanicMask  pcmDreq = 0x7F << pcmDreqTXPanicShift // TX_PANIC Panic level
	// 23 reserved
	pcmDreqRXPanicShift         = 16
	pcmDreqRXPanicMask  pcmDreq = 0x7F << pcmDreqRXPanicShift // RX_PANIC Panic level
	// 15 reserved
	pcmDreqTXLevelShift         = 8
	pcmDreqTXLevelMask  pcmDreq = 0x7F << pcmDreqTXPanicShift // TX Request Level
	// 7 reserved
	pcmDreqRXLevelShift         = 0
	pcmDreqRXLevelMask  pcmDreq = 0x7F << pcmDreqRXPanicShift // RX Request Level
)

type pcmInterrupt uint32

// Page 135
const (
	// 31:4 reserved
	pcmIntRXErr    pcmInterrupt = 1 << 3 // RXERR RX error interrupt enable
	pcmIntTXErr    pcmInterrupt = 1 << 2 // TXERR TX error interrupt enable
	pcmIntRXEnable pcmInterrupt = 1 << 1 // RXR RX Read interrupt enable
	pcmIntTXEnable pcmInterrupt = 1 << 0 // TXW TX Write interrupt enable
)

type pcmIntStatus uint32

// Page 135-136
const (
	// 31:4 reserved
	pcmIntStatRXErr    pcmIntStatus = 1 << 3 // RXERR RX error occurred / clear
	pcmIntStatTXErr    pcmIntStatus = 1 << 2 // TXERR TX error occurred / clear
	pcmIntStatRXEnable pcmIntStatus = 1 << 1 // RXR RX Read interrupt occurred / clear
	pcmIntStatTXEnable pcmIntStatus = 1 << 0 // TXW TX Write interrupt occurred / clear
)

// pcmGray puts it into a special data/strobe mode that is under 'best effort'
// contract.
type pcmGray uint32

// Page 136-137
const (
	// 31:22 reserved
	pcmGrayRXFIFOLevelShift         = 16
	pcmGrayRXFIFOLevelMask  pcmGray = 0x3F << pcmGrayRXFIFOLevelShift // RXFIFOLEVEL How many words in RXFIFO
	pcmGrayFlushShift               = 10
	pcmGrayFlushMask                = 0x3F << pcmGrayFlushShift // FLUSHED How many bits were valid when flush occurred
	pcmGrayRXLevelShift             = 4
	pcmGrayRXLevelMask      pcmGray = 0x3F << pcmGrayRXLevelShift // RXLEVEL How many GRAY coded bits received
	pcmGrayFlush            pcmGray = 1 << 2                      // FLUSH
	pcmGrayClear            pcmGray = 1 << 1                      // CLR
	pcmGrayEnable           pcmGray = 1 << 0                      // EN
)

// Page 119
type pcmMap struct {
	controlStatus pcmCS        // CS_A
	fifo          uint32       // FIFO_A
	mode          pcmMode      // MODE_A
	rxc           pcmRX        // RXC_A
	txc           pcmTX        // TXC_A
	dreq          pcmDreq      // DREQ_A
	inten         pcmInterrupt // INTEN_A
	intstc        pcmIntStatus // INTSTC_A
	gray          pcmGray      // GRAY
}
