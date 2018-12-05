// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package mfrc522 controls a Mifare RFID card reader.
//
// Datasheet
//
// https://www.nxp.com/docs/en/data-sheet/MFRC522.pdf
package mfrc522

import "fmt"

// BlockAccess defines the block access bits.
type BlockAccess byte

// SectorTrailerAccess defines the sector trailing block access bits.
type SectorTrailerAccess byte

// Access bits for the sector data.
const (
	AnyKeyRWID    BlockAccess = 0x0  // Any key (A or B) can read, write, increment and decrement block.
	RAB_WN_IN_DN  BlockAccess = 0x02 // Read (A or B), Write (None), Increment (None), Decrement (None)
	RAB_WB_IN_DN  BlockAccess = 0x04 // Read (A orB), Write (B), Increment (None), Decrement (None)
	RAB_WB_IB_DAB BlockAccess = 0x06 // Read (A or B), Write (B), Icrement (B), Decrement (A or B)
	RAB_WN_IN_DAB BlockAccess = 0x01 // Read (A or B), Write (None), Increment (None), Decrment (A or B)
	RB_WB_IN_DN   BlockAccess = 0x03 // Read (B), Write (B), Increment (None), Decrement (None)
	RB_WN_IN_DN   BlockAccess = 0x05 // Read (B), Write (None), Increment (None), Decrement (None)
	RN_WN_IN_DN   BlockAccess = 0x07 // Read (None), Write (None), Increment (None), Decrement (None)
)

// Access bits for the sector trail.
// Every trail sector has the options for controlling the access to the trailing sector bits.
// For example : KeyA_R[Key]_W[Key]_BITS_R[Key]_W[Key]_KeyB_R[Key]_W[Key]
//
// - KeyA
//   - could be Read by providing [Key] ( where [Key] could be KeyA or KeyB )
//   - could be Written by Providing [Key] ( where [Key] is KeyA or KeyB )
// - access bits for the sector data (see above)
//   - could be Read by providing [Key] ( where [Key] could be KeyA or KeyB )
//   - could be Written by Providing [Key] ( where [Key] is KeyA or KeyB )
// - KeyB
//   - could be Read by providing [Key] ( where [Key] could be KeyA or KeyB )
//   - could be Written by Providing [Key] ( where [Key] is KeyA or KeyB )
//
// example:
//
//  KeyA_RN_WA_BITS_RA_WA_KeyB_RA_WA means
//  - KeyA could not be read but could be overwriten if KeyA is provided
//  - Access bits could be read and overwritten if KeyA is provided during the card authentication
//  - KeyB could be read and overriten if KeyA is provided during the card authentication
// more on the matter: https://www.nxp.com/docs/en/data-sheet/MF1S50YYX_V1.pdf
const (
	KeyA_RN_WA_BITS_RA_WN_KeyB_RA_WA        SectorTrailerAccess = 0x0
	KeyA_RN_WN_BITS_RA_WN_KeyB_RA_WN        SectorTrailerAccess = 0x02
	KeyA_RN_WB_BITS_RAB_WN_KeyB_RN_WB       SectorTrailerAccess = 0x04
	KeyA_RN_WN_BITS_RAB_WN_KeyB_RN_WN       SectorTrailerAccess = 0x06
	KeyA_RN_WA_BITS_RA_WA_KeyB_RA_WA        SectorTrailerAccess = 0x01
	KeyA_RN_WB_BITS_RAB_WB_KeyB_RN_WB       SectorTrailerAccess = 0x03
	KeyA_RN_WN_BITS_RAB_WB_KeyB_RN_WN       SectorTrailerAccess = 0x05
	KeyA_RN_WN_BITS_RAB_WN_KeyB_RN_WN_EXTRA SectorTrailerAccess = 0x07
)

// BlocksAccess defines the access structure for first 3 blocks of the sector and the access bits for the sector trail.
type BlocksAccess struct {
	B0, B1, B2 BlockAccess
	B3         SectorTrailerAccess
}

func (ba *BlocksAccess) String() string {
	return fmt.Sprintf("B0: %d, B1: %d, B2: %d, B3: %d", ba.B0, ba.B1, ba.B2, ba.B3)
}

// serialize calculates the block access and stores it into the passed slice, that must be at least 4 bytes wide.
func (ba *BlocksAccess) serialize(dst []byte) error {
	if len(dst) < 4 {
		return wrapf("serialized array must be of size at least 4")
	}
	dst[0] = ((^ba.getBits(2) & 0x0F) << 4) | (^ba.getBits(1) & 0x0F)
	dst[1] = ((ba.getBits(1) & 0x0F) << 4) | (^ba.getBits(3) & 0x0F)
	dst[2] = ((ba.getBits(3) & 0x0F) << 4) | (ba.getBits(2) & 0x0F)
	dst[3] = dst[0] ^ dst[1] ^ dst[2]
	return nil
}

// Init parses the given byte array into the block access structure.
func (ba *BlocksAccess) Init(ad []byte) {
	ba.B0 = BlockAccess(((ad[1] & 0x10) >> 2) | ((ad[2] & 0x01) << 1) | ((ad[2] & 0x10) >> 5))
	ba.B1 = BlockAccess(((ad[1] & 0x20) >> 3) | (ad[2] & 0x02) | ((ad[2] & 0x20) >> 5))
	ba.B2 = BlockAccess(((ad[1] & 0x40) >> 4) | ((ad[2] & 0x04) >> 1) | ((ad[2] & 0x40) >> 6))
	ba.B3 = SectorTrailerAccess(((ad[1] & 0x80) >> 5) | ((ad[2] & 0x08) >> 2) | ((ad[2] & 0x80) >> 7))
}

func (ba *BlocksAccess) getBits(bitNum uint) byte {
	shift := 3 - bitNum
	bit := byte(1 << shift)
	return (byte(ba.B0)&bit)>>shift | ((byte(ba.B1)&bit)>>shift)<<1 | ((byte(ba.B2)&bit)>>shift)<<2 | ((byte(ba.B3)&bit)>>shift)<<3
}

func calcBlockAddress(sector int, block int) byte {
	return byte(sector*4 + block)
}

var _ fmt.Stringer = &BlocksAccess{}
