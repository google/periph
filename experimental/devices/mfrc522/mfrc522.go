// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package mfrc522 controls a Mifare RFID card reader.
//
// Datasheet
//
// https://www.nxp.com/docs/en/data-sheet/MFRC522.pdf
package mfrc522

import (
	"fmt"
	"sync"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/experimental/devices/mfrc522/commands"
)

// BlockAccess defines the block access bits.
type BlockAccess byte

// SectorTrailerAccess defines the sector trailing block access bits.
type SectorTrailerAccess byte

// Access bits.
const (
	AnyKeyRWID    BlockAccess = iota
	RAB_WN_IN_DN              = 0x02 // Read (A|B), Write (None), Increment (None), Decrement(None)
	RAB_WB_IN_DN              = 0x04
	RAB_WB_IB_DAB             = 0x06
	RAB_WN_IN_DAB             = 0x01
	RB_WB_IN_DN               = 0x03
	RB_WN_IN_DN               = 0x05
	RN_WN_IN_DN               = 0x07

	KeyA_RN_WA_BITS_RA_WN_KeyB_RA_WA        SectorTrailerAccess = iota
	KeyA_RN_WN_BITS_RA_WN_KeyB_RA_WN                            = 0x02
	KeyA_RN_WB_BITS_RAB_WN_KeyB_RN_WB                           = 0x04
	KeyA_RN_WN_BITS_RAB_WN_KeyB_RN_WN                           = 0x06
	KeyA_RN_WA_BITS_RA_WA_KeyB_RA_WA                            = 0x01
	KeyA_RN_WB_BITS_RAB_WB_KeyB_RN_WB                           = 0x03
	KeyA_RN_WN_BITS_RAB_WB_KeyB_RN_WN                           = 0x05
	KeyA_RN_WN_BITS_RAB_WN_KeyB_RN_WN_EXTRA                     = 0x07
)

// BlocksAccess defines the access structure for first 3 blocks of the sector
// and the access bits for the sector trail.
type BlocksAccess struct {
	B0, B1, B2 BlockAccess
	B3         SectorTrailerAccess
}

func (ba *BlocksAccess) String() string {
	return fmt.Sprintf("B0: %d, B1: %d, B2: %d, B3: %d", ba.B0, ba.B1, ba.B2, ba.B3)
}

// CalculateBlockAccess calculates the block access.
func CalculateBlockAccess(ba *BlocksAccess) []byte {
	res := make([]byte, 4)
	res[0] = ((^ba.getBits(2) & 0x0F) << 4) | (^ba.getBits(1) & 0x0F)
	res[1] = ((ba.getBits(1) & 0x0F) << 4) | (^ba.getBits(3) & 0x0F)
	res[2] = ((ba.getBits(3) & 0x0F) << 4) | (ba.getBits(2) & 0x0F)
	res[3] = res[0] ^ res[1] ^ res[2]
	return res
}

// ParseBlockAccess parses the given byte array into the block access structure.
func ParseBlockAccess(ad []byte) *BlocksAccess {
	return &BlocksAccess{
		B0: BlockAccess(((ad[1] & 0x10) >> 2) | ((ad[2] & 0x01) << 1) | ((ad[2] & 0x10) >> 5)),
		B1: BlockAccess(((ad[1] & 0x20) >> 3) | (ad[2] & 0x02) | ((ad[2] & 0x20) >> 5)),
		B2: BlockAccess(((ad[1] & 0x40) >> 4) | ((ad[2] & 0x04) >> 1) | ((ad[2] & 0x40) >> 6)),
		B3: SectorTrailerAccess(((ad[1] & 0x80) >> 5) | ((ad[2] & 0x08) >> 2) | ((ad[2] & 0x80) >> 7)),
	}
}

// DefaultKey provides the default bytes for card authentication for method B.
var DefaultKey = [...]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

// NewSPI creates and initializes the RFID card reader attached to SPI.
//
// 	spiPort - the SPI device to use.
// 	resetPin - reset GPIO pin.
// 	irqPin - irq GPIO pin.
func NewSPI(spiPort spi.Port, resetPin gpio.PinOut, irqPin gpio.PinIn) (*Dev, error) {
	raw, err := commands.NewLowLevelSPI(spiPort, resetPin, irqPin)
	if err != nil {
		return nil, err
	}

	dev := &Dev{
		operationTimeout: 30 * time.Second,
		antennaGain:      4,
		stop:             make(chan struct{}, 1),
		LowLevel:         raw,
	}
	if err := dev.init(); err != nil {
		return nil, err
	}
	return dev, nil
}

// Dev is an handle to an MFRC522 RFID reader.
type Dev struct {
	LowLevel *commands.LowLevel

	stop chan struct{}

	oMu              sync.Mutex
	operationTimeout time.Duration
	antennaGain      int

	mu        sync.Mutex
	isWaiting bool

	xMu         sync.Mutex
	isAccessing bool
}

// String implements conn.Resource.
func (r *Dev) String() string {
	return r.LowLevel.String()
}

// Halt implements conn.Resource.
//
// It soft-stops the chip - PowerDown bit set, command IDLE
func (r *Dev) Halt() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isWaiting {
		select {
		case <-r.stop:
		default:
		}

		r.stop <- struct{}{}
	}

	return r.LowLevel.DevWrite(commands.CommandReg, 16)
}

// SetOperationTimeout updates the device timeout for card operations.
//
// Effectively that sets the maximum time the RFID device will wait for IRQ
// from the proximity card detection.
//
//  timeout the duration to wait for IRQ strobe.
func (r *Dev) SetOperationTimeout(timeout time.Duration) {
	r.oMu.Lock()
	defer r.oMu.Unlock()
	r.operationTimeout = timeout
}

// SetAntennaGain configures antenna signal strength.
//
//  gain - signal strength from 0 to 7.
func (r *Dev) SetAntennaGain(gain int) error {
	if gain < 0 || gain > 7 {
		return wrapf("gain must be in [0..7] interval")
	}

	r.oMu.Lock()
	defer r.oMu.Unlock()
	r.antennaGain = gain
	return nil
}

// ReadCard reads the card sector/block.
//
//  auth - the authentication mode.
//  sector - the sector to authenticate on.
//  block - the block within sector to authenticate.
//  key - the key to be used for accessing the sector data.
func (r *Dev) ReadCard(auth byte, sector int, block int, key [6]byte) (data []byte, err error) {
	if err = r.accessStarted(); err != nil {
		return
	}

	defer func() {
		r.accessFinished()
		if err == nil {
			err = r.LowLevel.StopCrypto()
		}
	}()
	uuid, err := r.selectCard()
	if err != nil {
		return
	}
	state, err := r.LowLevel.Auth(auth, calcBlockAddress(sector, block), key, uuid)
	if err != nil {
		return
	}
	if state != commands.AuthOk {
		err = wrapf("can not authenticate")
		return
	}
	return r.readBlock(sector, block)
}

// ReadAuth - read the card authentication data.
//
// 	sector - the sector to authenticate on.
// 	key - the key to be used for accessing the sector data.
func (r *Dev) ReadAuth(auth byte, sector int, key [6]byte) (data []byte, err error) {
	if err = r.accessStarted(); err != nil {
		return
	}

	defer func() {
		r.accessFinished()
		if err == nil {
			err = r.LowLevel.StopCrypto()
		}
	}()
	uuid, err := r.selectCard()
	if err != nil {
		return
	}

	state, err := r.LowLevel.Auth(auth, calcBlockAddress(sector, 3), key, uuid)
	if err != nil {
		return
	}
	if state != commands.AuthOk {
		return nil, wrapf("can not authenticate")
	}

	return r.read(calcBlockAddress(sector, 3))
}

// WriteCard writes the data into the card block.
//
// 	auth - the authentiction mode.
// 	sector - the sector on the card to write to.
// 	block - the block within the sector to write into.
// 	data - 16 bytes if data to write
// 	key - the key used to authenticate the card - depends on the used auth method.
func (r *Dev) WriteCard(auth byte, sector int, block int, data [16]byte, key [6]byte) (err error) {
	if err = r.accessStarted(); err != nil {
		return
	}

	defer func() {
		r.accessFinished()
		if err == nil {
			err = r.LowLevel.StopCrypto()
		}
	}()
	uuid, err := r.selectCard()
	if err != nil {
		return
	}
	state, err := r.LowLevel.Auth(auth, calcBlockAddress(sector, 3), key, uuid)
	if err != nil {
		return
	}
	if state != commands.AuthOk {
		err = wrapf("authentication failed")
		return
	}

	return r.write(calcBlockAddress(sector, block%3), data[:])
}

// WriteSectorTrail writes the sector trail with sector access bits.
//
// 	auth - authentication mode.
// 	sector - sector to set authentication.
// 	keyA - the key used for AuthA authentication scheme.
// 	keyB - the key used for AuthB authentication scheme.
// 	access - the block access structure.
// 	key - the current key used to authenticate the provided sector.
func (r *Dev) WriteSectorTrail(auth byte, sector int, keyA [6]byte, keyB [6]byte, access *BlocksAccess, key [6]byte) (err error) {
	if err = r.accessStarted(); err != nil {
		return
	}

	defer func() {
		r.accessFinished()
		if err == nil {
			err = r.LowLevel.StopCrypto()
		}
	}()
	uuid, err := r.selectCard()
	if err != nil {
		return
	}
	state, err := r.LowLevel.Auth(auth, calcBlockAddress(sector, 3), key, uuid)
	if err != nil {
		return
	}
	if state != commands.AuthOk {
		err = wrapf("failed to authenticate")
		return
	}

	var data [16]byte
	copy(data[:], keyA[:])
	accessData := CalculateBlockAccess(access)
	copy(data[6:], accessData[:4])
	copy(data[10:], keyB[:])
	return r.write(calcBlockAddress(sector&0xFF, 3), data[:])
}

//		MFRC522 SPI Dev private/helper functions

// Checks whether device is already in use.
func (r *Dev) accessStarted() error {
	r.xMu.Lock()
	defer r.xMu.Unlock()

	if r.isAccessing {
		return wrapf("concurrent access is forbidden")
	}

	r.isAccessing = true
	return nil
}

// Marks device as free.
func (r *Dev) accessFinished() {
	r.xMu.Lock()
	defer r.xMu.Unlock()

	r.isAccessing = false
}

// init initializes the RFID chip.
func (r *Dev) init() error {
	if err := r.reset(); err != nil {
		return err
	}
	if err := r.writeCommandSequence(sequenceCommands.init); err != nil {
		return err
	}

	r.oMu.Lock()
	gain := byte(r.antennaGain) << 4
	r.oMu.Unlock()

	if err := r.LowLevel.DevWrite(int(commands.RFCfgReg), gain); err != nil {
		return err
	}

	return r.setAntenna(true)
}

// reset resets the RFID chip to initial state.
func (r *Dev) reset() error {
	return r.LowLevel.DevWrite(commands.CommandReg, commands.PCD_RESETPHASE)
}

// setAntenna configures the antenna state, on/off.
func (r *Dev) setAntenna(state bool) error {
	if state {
		current, err := r.LowLevel.DevRead(commands.TxControlReg)
		if err != nil {
			return err
		}
		if current&0x03 != 0 {
			return wrapf("can not set the bitmask for antenna")
		}
		return r.LowLevel.SetBitmask(commands.TxControlReg, 0x03)
	}
	return r.LowLevel.ClearBitmask(commands.TxControlReg, 0x03)
}

// request the card information. Returns number of blocks available on the card.
func (r *Dev) request() (int, error) {
	backBits := -1
	if err := r.LowLevel.DevWrite(commands.BitFramingReg, 0x07); err != nil {
		return backBits, err
	}
	_, backBits, err := r.LowLevel.CardWrite(commands.PCD_TRANSCEIVE, []byte{0x26})
	if err != nil {
		return -1, err
	}
	if backBits != 0x10 {
		return -1, wrapf("wrong number of bits %d", backBits)
	}
	return backBits, nil
}

// wait wait for IRQ to strobe on the IRQ pin when the card is detected.
func (r *Dev) wait() error {
	r.mu.Lock()
	waitCancelled := false
	if r.isWaiting {
		r.mu.Unlock()
		return wrapf("concurrent access is forbidden")
	}

	r.isWaiting = true
	r.mu.Unlock()

	irqChannel := make(chan bool)

	go func() {
		r.oMu.Lock()
		timeout := r.operationTimeout
		r.oMu.Unlock()

		result := r.LowLevel.WaitForEdge(timeout)
		r.mu.Lock()
		defer r.mu.Unlock()
		if waitCancelled {
			return
		}

		irqChannel <- result
	}()

	defer func() {
		r.mu.Lock()
		r.isWaiting = false
		waitCancelled = true
		r.mu.Unlock()

		close(irqChannel)
	}()

	if err := r.init(); err != nil {
		return err
	}
	if err := r.writeCommandSequence(sequenceCommands.waitInit); err != nil {
		return err
	}

	for {
		if err := r.writeCommandSequence(sequenceCommands.waitLoop); err != nil {
			return err
		}
		select {
		case <-r.stop:
			return wrapf("halt")
		case irqResult := <-irqChannel:
			if !irqResult {
				return wrapf("timeout waitinf for IRQ edge: %v", r.operationTimeout)
			}
			return nil
		case <-time.After(100 * time.Millisecond):
			// do nothing
		}
	}
}

// antiColl performs the collision check for different cards.
func (r *Dev) antiColl() ([]byte, error) {
	if err := r.LowLevel.DevWrite(commands.BitFramingReg, 0x00); err != nil {
		return nil, err
	}

	backData, _, err := r.LowLevel.CardWrite(commands.PCD_TRANSCEIVE, []byte{commands.PICC_ANTICOLL, 0x20}[:])

	if err != nil {
		return nil, err
	}

	if len(backData) != 5 {
		return nil, wrapf("back data expected 5, actual %d", len(backData))
	}

	crc := byte(0)

	for _, v := range backData[:4] {
		crc = crc ^ v
	}

	if crc != backData[4] {
		return nil, wrapf("CRC mismatch, expected %02x actual %02x", crc, backData[4])
	}

	return backData, nil
}

// selectTag selects the FOB device by device UUID.
func (r *Dev) selectTag(serial []byte) (byte, error) {
	dataBuf := make([]byte, len(serial)+2)
	dataBuf[0] = commands.PICC_SElECTTAG
	dataBuf[1] = 0x70
	copy(dataBuf[2:], serial)
	crc, err := r.LowLevel.CRC(dataBuf)
	if err != nil {
		return 0, err
	}
	dataBuf = append(dataBuf, crc[0], crc[1])
	backData, backLen, err := r.LowLevel.CardWrite(commands.PCD_TRANSCEIVE, dataBuf)
	if err != nil {
		return 0, err
	}

	var blocks byte

	if backLen == 0x18 {
		blocks = backData[0]
	} else {
		blocks = 0
	}
	return blocks, nil
}

// readBlock reads the block from the card.
//
// 	sector - card sector to read from
// 	block - the block within the sector (0-3 tor Mifare 4K)
func (r *Dev) readBlock(sector int, block int) ([]byte, error) {
	return r.read(calcBlockAddress(sector, block%3))
}

func (ba *BlocksAccess) getBits(bitNum uint) byte {
	shift := 3 - bitNum
	bit := byte(1 << shift)
	return (byte(ba.B0)&bit)>>shift | ((byte(ba.B1)&bit)>>shift)<<1 | ((byte(ba.B2)&bit)>>shift)<<2 | ((byte(ba.B3)&bit)>>shift)<<3
}

func (r *Dev) writeCommandSequence(commands [][]byte) error {
	for _, cmdData := range commands {
		if err := r.LowLevel.DevWrite(int(cmdData[0]), cmdData[1]); err != nil {
			return err
		}
	}
	return nil
}

func (r *Dev) selectCard() ([]byte, error) {
	if err := r.wait(); err != nil {
		return nil, err
	}
	if err := r.init(); err != nil {
		return nil, err
	}
	if _, err := r.request(); err != nil {
		return nil, err
	}
	uuid, err := r.antiColl()
	if err != nil {
		return nil, err
	}
	if _, err := r.selectTag(uuid); err != nil {
		return nil, err
	}
	return uuid, nil
}

func calcBlockAddress(sector int, block int) byte {
	return byte(sector*4 + block)
}

func (r *Dev) write(blockAddr byte, data []byte) error {
	read, backLen, err := r.preAccess(blockAddr, commands.PICC_WRITE)
	if err != nil || backLen != 4 {
		return err
	}
	if read[0]&0x0F != 0x0A {
		return wrapf("can't authorize write")
	}
	var newData [18]byte
	copy(newData[:], data[:16])
	crc, err := r.LowLevel.CRC(newData[:16])
	if err != nil {
		return err
	}
	newData[16] = crc[0]
	newData[17] = crc[1]
	read, backLen, err = r.LowLevel.CardWrite(commands.PCD_TRANSCEIVE, newData[:])
	if err != nil {
		return err
	}
	if backLen != 4 || read[0]&0x0F != 0x0A {
		err = wrapf("can't write data")
	}
	return nil
}

func (r *Dev) preAccess(blockAddr byte, cmd byte) ([]byte, int, error) {
	send := make([]byte, 4)
	send[0] = cmd
	send[1] = blockAddr

	crc, err := r.LowLevel.CRC(send[:2])
	if err != nil {
		return nil, -1, err
	}
	send[2] = crc[0]
	send[3] = crc[1]
	return r.LowLevel.CardWrite(commands.PCD_TRANSCEIVE, send)
}

func (r *Dev) read(blockAddr byte) ([]byte, error) {
	data, _, err := r.preAccess(blockAddr, commands.PICC_READ)
	if err != nil {
		return nil, err
	}
	if len(data) != 16 {
		return nil, wrapf("expected 16 bytes, actual %d", len(data))
	}
	return data, nil
}

// the command batches for card init and wait loop.
var sequenceCommands = struct {
	init     [][]byte
	waitInit [][]byte
	waitLoop [][]byte
}{
	init: [][]byte{
		{commands.TModeReg, 0x8D},
		{commands.TPrescalerReg, 0x3E},
		{commands.TReloadRegL, 30},
		{commands.TReloadRegH, 0},
		{commands.TxAutoReg, 0x40},
		{commands.ModeReg, 0x3D},
	},
	waitInit: [][]byte{
		{commands.CommIrqReg, 0x00},
		{commands.CommIEnReg, 0xA0},
	},
	waitLoop: [][]byte{
		{commands.FIFODataReg, 0x26},
		{commands.CommandReg, 0x0C},
		{commands.BitFramingReg, 0x87},
	},
}

func wrapf(format string, a ...interface{}) error {
	return fmt.Errorf("mfrc522: "+format, a...)
}

var _ conn.Resource = &Dev{}
var _ fmt.Stringer = &BlocksAccess{}
