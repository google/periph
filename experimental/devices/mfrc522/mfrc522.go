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
	"periph.io/x/periph/conn/physic"
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

// AuthStatus indicates the authentication response, could be one of AuthOk,
// AuthReadFailure or AuthFailure
type AuthStatus byte

// Card authentication status enum.
const (
	AuthOk AuthStatus = iota
	AuthReadFailure
	AuthFailure
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
	if resetPin == nil {
		return nil, wrapf("reset pin is not set")
	}
	if irqPin == nil {
		return nil, wrapf("IRQ pin is not set")
	}
	spiDev, err := spiPort.Connect(10*physic.MegaHertz, spi.Mode0, 8)
	if err != nil {
		return nil, err
	}
	if err := resetPin.Out(gpio.High); err != nil {
		return nil, err
	}
	if err := irqPin.In(gpio.PullUp, gpio.FallingEdge); err != nil {
		return nil, err
	}
	dev := &Dev{
		spiDev:           spiDev,
		operationTimeout: 30 * time.Second,
		irqPin:           irqPin,
		resetPin:         resetPin,
		antennaGain:      4,
		stop:             make(chan struct{}, 1),
	}
	if err := dev.Init(); err != nil {
		return nil, err
	}
	return dev, nil
}

// Dev is an handle to an MFRC522 RFID reader.
type Dev struct {
	resetPin         gpio.PinOut
	irqPin           gpio.PinIn
	operationTimeout time.Duration
	spiDev           spi.Conn
	stop             chan struct{}

	aMu         sync.Mutex
	antennaGain int

	mu        sync.Mutex
	isWaiting bool
}

// String implements conn.Resource.
func (r *Dev) String() string {
	return fmt.Sprintf("Mifare MFRC522 [bus: %v, reset pin: %s, irq pin: %s]",
		r.spiDev, r.resetPin.Name(), r.irqPin.Name())
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

	return r.devWrite(commands.CommandReg, 16)
}

// SetOperationTimeout updates the device timeout for card operations.
//
// Effectively that sets the maximum time the RFID device will wait for IRQ
// from the proximity card detection.
//
//	timeout the duration to wait for IRQ strobe.
func (r *Dev) SetOperationTimeout(timeout time.Duration) {
	r.operationTimeout = timeout
}

// Init initializes the RFID chip.
func (r *Dev) Init() error {
	if err := r.Reset(); err != nil {
		return err
	}
	if err := r.writeCommandSequence(sequenceCommands.init); err != nil {
		return err
	}

	r.aMu.Lock()
	gain := byte(r.antennaGain)<<4
	r.aMu.Unlock()

	if err := r.devWrite(int(commands.RFCfgReg), gain); err != nil {
		return err
	}

	return r.SetAntenna(true)
}

// Reset resets the RFID chip to initial state.
func (r *Dev) Reset() error {
	return r.devWrite(commands.CommandReg, commands.PCD_RESETPHASE)
}

// SetAntenna configures the antenna state, on/off.
func (r *Dev) SetAntenna(state bool) error {
	if state {
		current, err := r.devRead(commands.TxControlReg)
		if err != nil {
			return err
		}
		if current&0x03 != 0 {
			return wrapf("can not set the bitmask for antenna")
		}
		return r.setBitmask(commands.TxControlReg, 0x03)
	}
	return r.clearBitmask(commands.TxControlReg, 0x03)
}

// SetAntennaGain configures antenna signal strength.
//
//	gain - signal strength from 0 to 7.
func (r *Dev) SetAntennaGain(gain int) {
	r.aMu.Lock()
	defer r.aMu.Unlock()

	if 0 <= gain && gain <= 7 {
		r.antennaGain = gain
	}
}

// CardWrite the low-level interface to write some raw commands to the card.
//
// 	command - the command register
// 	data - the data to write out to the card using the authenticated sector.
func (r *Dev) CardWrite(command byte, data []byte) ([]byte, int, error) {
	var backData []byte
	backLength := -1
	irqEn := byte(0x00)
	irqWait := byte(0x00)

	switch command {
	case commands.PCD_AUTHENT:
		irqEn = 0x12
		irqWait = 0x10
	case commands.PCD_TRANSCEIVE:
		irqEn = 0x77
		irqWait = 0x30
	}

	if err := r.devWrite(commands.CommIEnReg, irqEn|0x80); err != nil {
		return nil, -1, err
	}
	if err := r.clearBitmask(commands.CommIrqReg, 0x80); err != nil {
		return nil, -1, err
	}
	if err := r.setBitmask(commands.FIFOLevelReg, 0x80); err != nil {
		return nil, -1, err
	}
	if err := r.devWrite(commands.CommandReg, commands.PCD_IDLE); err != nil {
		return nil, -1, err
	}

	for _, v := range data {
		if err := r.devWrite(commands.FIFODataReg, v); err != nil {
			return nil, -1, err
		}
	}

	if err := r.devWrite(commands.CommandReg, command); err != nil {
		return nil, -1, err
	}

	if command == commands.PCD_TRANSCEIVE {
		if err := r.setBitmask(commands.BitFramingReg, 0x80); err != nil {
			return nil, -1, err
		}
	}

	i := 2000
	n := byte(0)

	for ; i > 0; i-- {
		n, err := r.devRead(commands.CommIrqReg)
		if err != nil {
			return nil, -1, err
		}
		if n&(irqWait|1) != 0 {
			break
		}
	}

	if err := r.clearBitmask(commands.BitFramingReg, 0x80); err != nil {
		return nil, -1, err
	}

	if i == 0 {
		return nil, -1, wrapf("can't read data after 2000 loops")
	}

	if d, err := r.devRead(commands.ErrorReg); err != nil || d&0x1B != 0 {
		return nil, -1, err
	}

	if n&irqEn&0x01 == 1 {
		return nil, -1, wrapf("IRQ error")
	}

	if command == commands.PCD_TRANSCEIVE {
		n, err := r.devRead(commands.FIFOLevelReg)
		if err != nil {
			return nil, -1, err
		}
		lastBits, err := r.devRead(commands.ControlReg)
		if err != nil {
			return nil, -1, err
		}
		lastBits = lastBits & 0x07
		if lastBits != 0 {
			backLength = (int(n)-1)*8 + int(lastBits)
		} else {
			backLength = int(n) * 8
		}

		if n == 0 {
			n = 1
		}

		if n > 16 {
			n = 16
		}

		backData = make([]byte, n)
		for i := byte(0); i < n; i++ {
			byteVal, err := r.devRead(commands.FIFODataReg)
			if err != nil {
				return nil, -1, err
			}
			backData[i] = byteVal
		}

	}

	return backData, backLength, nil
}

// Request the card information. Returns number of blocks available on the card.
func (r *Dev) Request() (int, error) {
	backBits := -1
	if err := r.devWrite(commands.BitFramingReg, 0x07); err != nil {
		return backBits, err
	}
	_, backBits, err := r.CardWrite(commands.PCD_TRANSCEIVE, []byte{0x26})
	if err != nil {
		return -1, err
	}
	if backBits != 0x10 {
		return -1, wrapf("wrong number of bits %d", backBits)
	}
	return backBits, nil
}

// Wait wait for IRQ to strobe on the IRQ pin when the card is detected.
func (r *Dev) Wait() error {
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
		result := r.irqPin.WaitForEdge(r.operationTimeout)
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

	if err := r.Init(); err != nil {
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

// AntiColl performs the collision check for different cards.
func (r *Dev) AntiColl() ([]byte, error) {

	if err := r.devWrite(commands.BitFramingReg, 0x00); err != nil {
		return nil, err
	}

	backData, _, err := r.CardWrite(commands.PCD_TRANSCEIVE, []byte{commands.PICC_ANTICOLL, 0x20}[:])

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

// CRC calculates the CRC of the data using the card chip.
func (r *Dev) CRC(inData []byte) ([]byte, error) {
	if err := r.clearBitmask(commands.DivIrqReg, 0x04); err != nil {
		return nil, err
	}
	if err := r.setBitmask(commands.FIFOLevelReg, 0x80); err != nil {
		return nil, err
	}
	for _, v := range inData {
		if err := r.devWrite(commands.FIFODataReg, v); err != nil {
			return nil, err
		}
	}
	if err := r.devWrite(commands.CommandReg, commands.PCD_CALCCRC); err != nil {
		return nil, err
	}
	for i := byte(0xFF); i > 0; i-- {
		n, err := r.devRead(commands.DivIrqReg)
		if err != nil {
			return nil, err
		}
		if n&0x04 > 0 {
			break
		}
	}
	lsb, err := r.devRead(commands.CRCResultRegL)
	if err != nil {
		return nil, err
	}

	msb, err := r.devRead(commands.CRCResultRegM)
	if err != nil {
		return nil, err
	}
	return []byte{lsb, msb}, nil
}

// SelectTag selects the FOB device by device UUID.
func (r *Dev) SelectTag(serial []byte) (byte, error) {
	dataBuf := make([]byte, len(serial)+2)
	dataBuf[0] = commands.PICC_SElECTTAG
	dataBuf[1] = 0x70
	copy(dataBuf[2:], serial)
	crc, err := r.CRC(dataBuf)
	if err != nil {
		return 0, err
	}
	dataBuf = append(dataBuf, crc[0], crc[1])
	backData, backLen, err := r.CardWrite(commands.PCD_TRANSCEIVE, dataBuf)
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

// StopCrypto stops the crypto chip.
func (r *Dev) StopCrypto() error {
	return r.clearBitmask(commands.Status2Reg, 0x08)
}

// ReadBlock reads the block from the card.
//
// 	sector - card sector to read from
// 	block - the block within the sector (0-3 tor Mifare 4K)
func (r *Dev) ReadBlock(sector int, block int) ([]byte, error) {
	return r.read(calcBlockAddress(sector, block%3))
}

// WriteBlock writes the data into the card block.
//
// 	auth - the authentiction mode.
// 	sector - the sector on the card to write to.
// 	block - the block within the sector to write into.
// 	data - 16 bytes if data to write
// 	key - the key used to authenticate the card - depends on the used auth method.
func (r *Dev) WriteBlock(auth byte, sector int, block int, data [16]byte, key [6]byte) (err error) {
	defer func() {
		if err == nil {
			err = r.StopCrypto()
		}
	}()
	uuid, err := r.selectCard()
	if err != nil {
		return
	}
	state, err := r.Auth(auth, sector, 3, key, uuid)
	if err != nil {
		return
	}
	if state != AuthOk {
		err = wrapf("authentication failed")
		return
	}

	return r.write(calcBlockAddress(sector, block%3), data[:])
}

// ReadSectorTrail reads the sector trail (the last sector that contains the
// sector access bits)
//
// 	sector - the sector number to read the data from.
func (r *Dev) ReadSectorTrail(sector int) ([]byte, error) {
	return r.read(calcBlockAddress(sector&0xFF, 3))
}

// WriteSectorTrail writes the sector trait with sector access bits.
//
// 	auth - authentication mode.
// 	sector - sector to set authentication.
// 	keyA - the key used for AuthA authentication scheme.
// 	keyB - the key used for AuthB authentication schemd.
// 	access - the block access structure.
// 	key - the current key used to authenticate the provided sector.
func (r *Dev) WriteSectorTrail(auth byte, sector int, keyA [6]byte, keyB [6]byte, access *BlocksAccess, key [6]byte) (err error) {
	defer func() {
		if err == nil {
			err = r.StopCrypto()
		}
	}()
	uuid, err := r.selectCard()
	if err != nil {
		return
	}
	state, err := r.Auth(auth, sector, 3, key, uuid)
	if err != nil {
		return
	}
	if state != AuthOk {
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

// Auth authenticate the card fof the sector/block using the provided data.
//
// 	mode - the authentication mode.
// 	sector - the sector to authenticate on.
// 	block - the block within sector to authenticate.
// 	sectorKey - the key to be used for accessing the sector data.
// 	serial - the serial of the card.
func (r *Dev) Auth(mode byte, sector, block int, sectorKey [6]byte, serial []byte) (AuthStatus, error) {
	return r.auth(mode, calcBlockAddress(sector, block), sectorKey, serial)
}

// ReadCard reads the card sector/block.
//
// 	auth - the authentication mode.
// 	sector - the sector to authenticate on.
// 	block - the block within sector to authenticate.
// 	key - the key to be used for accessing the sector data.
func (r *Dev) ReadCard(auth byte, sector int, block int, key [6]byte) (data []byte, err error) {
	defer func() {
		if err == nil {
			err = r.StopCrypto()
		}
	}()
	uuid, err := r.selectCard()
	if err != nil {
		return
	}
	state, err := r.Auth(auth, sector, block, key, uuid)
	if err != nil {
		return
	}
	if state != AuthOk {
		err = wrapf("can not authenticate")
		return
	}
	return r.ReadBlock(sector, block)
}

// ReadAuth - read the card authentication data.
//
// 	sector - the sector to authenticate on.
// 	key - the key to be used for accessing the sector data.
func (r *Dev) ReadAuth(auth byte, sector int, key [6]byte) (data []byte, err error) {
	defer func() {
		if err == nil {
			err = r.StopCrypto()
		}
	}()
	uuid, err := r.selectCard()
	if err != nil {
		return
	}
	state, err := r.Auth(auth, sector, 3, key, uuid)
	if err != nil {
		return
	}
	if state != AuthOk {
		return nil, wrapf("can not authenticate")
	}

	return r.read(calcBlockAddress(sector, 3))
}

//		MFRC522 SPI Dev private/helper functions

func (ba *BlocksAccess) getBits(bitNum uint) byte {
	shift := 3 - bitNum
	bit := byte(1 << shift)
	return (byte(ba.B0)&bit)>>shift | ((byte(ba.B1)&bit)>>shift)<<1 | ((byte(ba.B2)&bit)>>shift)<<2 | ((byte(ba.B3)&bit)>>shift)<<3
}

func (r *Dev) writeCommandSequence(commands [][]byte) error {
	for _, cmdData := range commands {
		if err := r.devWrite(int(cmdData[0]), cmdData[1]); err != nil {
			return err
		}
	}
	return nil
}

func (r *Dev) selectCard() ([]byte, error) {
	if err := r.Wait(); err != nil {
		return nil, err
	}
	if err := r.Init(); err != nil {
		return nil, err
	}
	if _, err := r.Request(); err != nil {
		return nil, err
	}
	uuid, err := r.AntiColl()
	if err != nil {
		return nil, err
	}
	if _, err := r.SelectTag(uuid); err != nil {
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
	crc, err := r.CRC(newData[:16])
	if err != nil {
		return err
	}
	newData[16] = crc[0]
	newData[17] = crc[1]
	read, backLen, err = r.CardWrite(commands.PCD_TRANSCEIVE, newData[:])
	if err != nil {
		return err
	}
	if backLen != 4 || read[0]&0x0F != 0x0A {
		err = wrapf("can't write data")
	}
	return nil
}

func (r *Dev) auth(mode byte, blockAddress byte, sectorKey [6]byte, serial []byte) (AuthStatus, error) {
	buffer := make([]byte, 2)
	buffer[0] = mode
	buffer[1] = blockAddress
	buffer = append(buffer, sectorKey[:]...)
	buffer = append(buffer, serial[:4]...)
	_, _, err := r.CardWrite(commands.PCD_AUTHENT, buffer)
	if err != nil {
		return AuthReadFailure, err
	}
	if n, err := r.devRead(commands.Status2Reg); err != nil || n&0x08 == 0 {
		return AuthFailure, err
	}
	return AuthOk, nil
}

func (r *Dev) devWrite(address int, data byte) error {
	newData := []byte{(byte(address) << 1) & 0x7E, data}
	return r.spiDev.Tx(newData, nil)
}

func (r *Dev) devRead(address int) (byte, error) {
	data := []byte{((byte(address) << 1) & 0x7E) | 0x80, 0}
	out := make([]byte, len(data))
	if err := r.spiDev.Tx(data, out); err != nil {
		return 0, err
	}
	return out[1], nil
}

func (r *Dev) setBitmask(address, mask int) error {
	current, err := r.devRead(address)
	if err != nil {
		return err
	}
	return r.devWrite(address, current|byte(mask))
}

func (r *Dev) clearBitmask(address, mask int) error {
	current, err := r.devRead(address)
	if err != nil {
		return err
	}
	return r.devWrite(address, current&^byte(mask))
}

func (r *Dev) preAccess(blockAddr byte, cmd byte) ([]byte, int, error) {
	send := make([]byte, 4)
	send[0] = cmd
	send[1] = blockAddr

	crc, err := r.CRC(send[:2])
	if err != nil {
		return nil, -1, err
	}
	send[2] = crc[0]
	send[3] = crc[1]
	return r.CardWrite(commands.PCD_TRANSCEIVE, send)
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
