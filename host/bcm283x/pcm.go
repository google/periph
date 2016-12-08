// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// pcm means I2S.

package bcm283x

type pcmCS uint32

// Pages 126-129.
const (
	// 31:26 reserved
	pcmStandby      pcmCS = 1 << 25 // STBY
	pcmSync         pcmCS = 1 << 24 // SYNC
	pcmRXSignExtend pcmCS = 1 << 23 // RXSEX
	pcmRXFull       pcmCS = 1 << 22 // RXF
	pcmTXEmpty      pcmCS = 1 << 21 // TXE
	pcmRXData       pcmCS = 1 << 20 // RXD
	pcmTXData       pcmCS = 1 << 19 // TXD
	pcmRXR          pcmCS = 1 << 18 // RXR
	pcmTXW          pcmCS = 1 << 17 // TXW
	pcmRXErr        pcmCS = 1 << 16 // RXERR
	pcmTXErr        pcmCS = 1 << 15 // TXERR
	pcmRXSync       pcmCS = 1 << 14 // RXSYNC
	pcmTXSync       pcmCS = 1 << 13 // TXSYNC
	// 12:10 reserved
	pcmDMAEnable pcmCS = 1 << 9 // DMAEN
	// 8:7
	pcmRXThreshold pcmCS = 1<<8 | 1<<7 // RXTHR
	// 6:5
	pcmTXThreshold pcmCS = 1<<6 | 1<<5 // TXTHR
	pcmRXClear     pcmCS = 1 << 4      // RXCLR
	pcmTXClear     pcmCS = 1 << 3      // TXCLR
	pcmTXEnable    pcmCS = 1 << 2      // TXON
	pcmRXEnable    pcmCS = 1 << 1      // RXON
	pcmEnable      pcmCS = 1 << 0      // EN
)
