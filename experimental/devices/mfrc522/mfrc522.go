// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package mfrc522 controls a Mifare RFID card reader.
//
// The code is largely ported from Python library : https://github.com/mxgxw/MFRC522-python
//
// The datasheet is available at : https://www.nxp.com/docs/en/data-sheet/MFRC522.pdf
package mfrc522

import (
	"errors"
	"fmt"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/experimental/devices/mfrc522/commands"
	"time"
)

// Dev is an handle to an MFRC522 RFID reader.
type Dev struct {
	resetPin      gpio.PinOut
	irqPin        gpio.PinIn
	Authenticated bool
	maxSpeedHz    int64
	spiDev        spi.Conn
}

// DefaultKey provides the default bytes for card authentication for method B.
var DefaultKey = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

const maxSpeed = int64(1000000)

/*
NewSPI creates and initializes the RFID card reader attached to SPI.
	spiPort - the SPI device to use.
	resetPin - reset GPIO pin.
	irqPin - irq GPIO pin.
*/
func NewSPI(spiPort spi.Port, resetPin, irqPin string) (*Dev, error) {

	spiDev, err := spiPort.Connect(maxSpeed, spi.Mode0, 8)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	dev := &Dev{
		spiDev:     spiDev,
		maxSpeedHz: maxSpeed,
	}

	pin := gpioreg.ByName(resetPin)
	dev.resetPin = pin
	dev.resetPin.Out(gpio.High)

	pin = gpioreg.ByName(irqPin)
	dev.irqPin = pin
	dev.irqPin.In(gpio.PullUp, gpio.FallingEdge)

	err = dev.Init()

	return dev, nil
}

var initCommands = [][]byte{
	{commands.TModeReg, 0x8D},
	{commands.TPrescalerReg, 0x3E},
	{commands.TReloadRegL, 30},
	{commands.TReloadRegH, 0},
	{commands.TxAutoReg, 0x40},
	{commands.ModeReg, 0x3D},
}

// Init initializes the RFID chip.
func (r *Dev) Init() error {
	err := r.Reset()
	if err != nil {
		return err
	}
	for _, cmdData := range initCommands {
		err = r.devWrite(int(cmdData[0]), cmdData[1])
		if err != nil {
			return err
		}
	}
	err = r.SetAntenna(true)
	if err != nil {
		return err
	}
	return nil
}

func (r *Dev) devWrite(address int, data byte) (err error) {
	newData := []byte{(byte(address) << 1) & 0x7E, data}
	err = r.spiDev.Tx(newData, nil)
	return
}

func (r *Dev) devRead(address int) (result byte, err error) {
	data := []byte{((byte(address) << 1) & 0x7E) | 0x80, 0}
	out := make([]byte, len(data))
	err = r.spiDev.Tx(data, out)
	if err != nil {
		return
	}
	result = out[1]
	return
}

func (r *Dev) setBitmask(address, mask int) (err error) {
	current, err := r.devRead(address)
	if err != nil {
		return
	}
	err = r.devWrite(address, current|byte(mask))
	return
}

func (r *Dev) clearBitmask(address, mask int) (err error) {
	current, err := r.devRead(address)
	if err != nil {
		return
	}
	err = r.devWrite(address, current&^byte(mask))
	return

}

// Reset resets the RFID chip to initial state.
func (r *Dev) Reset() (err error) {
	r.Authenticated = false
	err = r.devWrite(commands.CommandReg, commands.PCD_RESETPHASE)
	return
}

// SetAntenna configures the antenna state, on/off.
func (r *Dev) SetAntenna(state bool) (err error) {
	if state {
		current, err := r.devRead(commands.TxControlReg)
		if err != nil {
			return err
		}
		if current&0x03 == 0 {
			err = r.setBitmask(commands.TxControlReg, 0x03)
		}
	} else {
		err = r.clearBitmask(commands.TxControlReg, 0x03)
	}
	return
}

/*
CardWrite the low-level interface to write some raw commands to the card.
	command - the command register
	data - the data to write out to the card using the authenticated sector.
*/
func (r *Dev) CardWrite(command byte, data []byte) (backData []byte, backLength int, err error) {
	backLength = -1
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

	r.devWrite(commands.CommIEnReg, irqEn|0x80)
	r.clearBitmask(commands.CommIrqReg, 0x80)
	r.setBitmask(commands.FIFOLevelReg, 0x80)
	r.devWrite(commands.CommandReg, commands.PCD_IDLE)

	for _, v := range data {
		r.devWrite(commands.FIFODataReg, v)
	}

	r.devWrite(commands.CommandReg, command)

	if command == commands.PCD_TRANSCEIVE {
		r.setBitmask(commands.BitFramingReg, 0x80)
	}

	i := 2000
	n := byte(0)

	for ; i > 0; i-- {
		n, err = r.devRead(commands.CommIrqReg)
		if err != nil {
			return
		}
		if n&(irqWait|1) != 0 {
			break
		}
	}

	r.clearBitmask(commands.BitFramingReg, 0x80)

	if i == 0 {
		err = errors.New("can't read data after 2000 loops")
		return
	}

	if d, err1 := r.devRead(commands.ErrorReg); err1 != nil || d&0x1B != 0 {
		err = err1
		return
	}

	if n&irqEn&0x01 == 1 {
		err = errors.New("IRQ error")
		return
	}

	if command == commands.PCD_TRANSCEIVE {
		n, err = r.devRead(commands.FIFOLevelReg)
		if err != nil {
			return
		}
		lastBits, err1 := r.devRead(commands.ControlReg)
		if err1 != nil {
			err = err1
			return
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
			byteVal, err1 := r.devRead(commands.FIFODataReg)
			if err1 != nil {
				err = err1
				return
			}
			backData[i] = byteVal
		}

	}

	return
}

// Request request the card information. Returns number of blocks available on the card.
func (r *Dev) Request() (backBits int, err error) {
	backBits = 0
	err = r.devWrite(commands.BitFramingReg, 0x07)
	if err != nil {
		return
	}

	_, backBits, err = r.CardWrite(commands.PCD_TRANSCEIVE, []byte{0x26}[:])

	if backBits != 0x10 {
		err = fmt.Errorf("wrong number of bits %d", backBits)
	}

	return
}

// Wait wait for IRQ to strobe on the IRQ pin when the card is detected.
func (r *Dev) Wait() (err error) {
	irqChannel := make(chan bool)
	go func() {
		if r.irqPin.WaitForEdge(20 * time.Second) {
			irqChannel <- true
		}
	}()

	defer func() {
		close(irqChannel)
	}()

	err = r.Init()
	if err != nil {
		return
	}
	err = r.devWrite(commands.CommIrqReg, 0x00)
	if err != nil {
		return
	}
	err = r.devWrite(commands.CommIEnReg, 0xA0)
	if err != nil {
		return
	}

interruptLoop:
	for {
		err = r.devWrite(commands.FIFODataReg, 0x26)
		if err != nil {
			return
		}
		err = r.devWrite(commands.CommandReg, 0x0C)
		if err != nil {
			return
		}
		err = r.devWrite(commands.BitFramingReg, 0x87)
		if err != nil {
			return
		}
		select {
		case _ = <-irqChannel:
			break interruptLoop
		case <-time.After(100 * time.Millisecond):
			// do nothing
		}
	}
	return
}

// AntiColl performs the collision check for different cards.
func (r *Dev) AntiColl() (backData []byte, err error) {

	err = r.devWrite(commands.BitFramingReg, 0x00)

	backData, _, err = r.CardWrite(commands.PCD_TRANSCEIVE, []byte{commands.PICC_ANTICOLL, 0x20}[:])

	if err != nil {
		return
	}

	if len(backData) != 5 {
		err = fmt.Errorf("Back data expected 5, actual %d", len(backData))
		return
	}

	crc := byte(0)

	for _, v := range backData[:4] {
		crc = crc ^ v
	}

	if crc != backData[4] {
		err = errors.New(fmt.Sprintf("CRC mismatch, expected %02x actual %02x", crc, backData[4]))
	}

	return
}

// CRC calculates the CRC of the data using the card chip.
func (r *Dev) CRC(inData []byte) (res []byte, err error) {
	res = []byte{0, 0}
	err = r.clearBitmask(commands.DivIrqReg, 0x04)
	if err != nil {
		return
	}
	err = r.setBitmask(commands.FIFOLevelReg, 0x80)
	if err != nil {
		return
	}
	for _, v := range inData {
		r.devWrite(commands.FIFODataReg, v)
	}
	err = r.devWrite(commands.CommandReg, commands.PCD_CALCCRC)
	if err != nil {
		return
	}
	for i := byte(0xFF); i > 0; i-- {
		n, err1 := r.devRead(commands.DivIrqReg)
		if err1 != nil {
			err = err1
			return
		}
		if n&0x04 > 0 {
			break
		}
	}
	lsb, err := r.devRead(commands.CRCResultRegL)
	if err != nil {
		return
	}
	res[0] = lsb

	msb, err := r.devRead(commands.CRCResultRegM)
	if err != nil {
		return
	}
	res[1] = msb
	return
}

// SelectTag selects the FOB device by device UUID.
func (r *Dev) SelectTag(serial []byte) (blocks byte, err error) {
	dataBuf := make([]byte, len(serial)+2)
	dataBuf[0] = commands.PICC_SElECTTAG
	dataBuf[1] = 0x70
	copy(dataBuf[2:], serial)
	crc, err := r.CRC(dataBuf)
	if err != nil {
		return
	}
	dataBuf = append(dataBuf, crc[0], crc[1])
	backData, backLen, err := r.CardWrite(commands.PCD_TRANSCEIVE, dataBuf)
	if err != nil {
		return
	}

	if backLen == 0x18 {
		blocks = backData[0]
	} else {
		blocks = 0
	}
	return
}

type AuthStatus byte

const (
	AuthOk AuthStatus = iota
	AuthReadFailure
	AuthFailure
)

func (r *Dev) auth(mode byte, blockAddress byte, sectorKey []byte, serial []byte) (authS AuthStatus, err error) {
	buffer := make([]byte, 2)
	buffer[0] = mode
	buffer[1] = blockAddress
	buffer = append(buffer, sectorKey...)
	buffer = append(buffer, serial[:4]...)
	_, _, err = r.CardWrite(commands.PCD_AUTHENT, buffer)
	if err != nil {
		authS = AuthReadFailure
		return
	}
	n, err := r.devRead(commands.Status2Reg)
	if err != nil {
		return
	}
	if n&0x08 != 0 {
		authS = AuthFailure
	}
	authS = AuthOk
	return
}

// StopCrypto stops the crypto chip.
func (r *Dev) StopCrypto() (err error) {
	err = r.clearBitmask(commands.Status2Reg, 0x08)
	return
}

func (r *Dev) preAccess(blockAddr byte, cmd byte) (data []byte, backLen int, err error) {
	send := make([]byte, 4)
	send[0] = cmd
	send[1] = blockAddr

	crc, err := r.CRC(send[:2])
	if err != nil {
		return
	}
	send[2] = crc[0]
	send[3] = crc[1]
	data, backLen, err = r.CardWrite(commands.PCD_TRANSCEIVE, send)
	return
}

func (r *Dev) read(blockAddr byte) (data []byte, err error) {
	data, _, err = r.preAccess(blockAddr, commands.PICC_READ)
	if err != nil {
		return
	}
	if len(data) != 16 {
		err = errors.New(fmt.Sprintf("Expected 16 bytes, actual %d", len(data)))
	}
	return
}

func (r *Dev) write(blockAddr byte, data []byte) (err error) {
	read, backLen, err := r.preAccess(blockAddr, commands.PICC_WRITE)
	if err != nil || backLen != 4 {
		return
	}
	if read[0]&0x0F != 0x0A {
		err = errors.New("can't authorize write")
		return
	}
	newData := make([]byte, 18)
	copy(newData, data[:16])
	crc, err := r.CRC(newData[:16])
	if err != nil {
		return
	}
	newData[16] = crc[0]
	newData[17] = crc[1]
	read, backLen, err = r.CardWrite(commands.PCD_TRANSCEIVE, newData)
	if err != nil {
		return
	}
	if backLen != 4 || read[0]&0x0F != 0x0A {
		err = errors.New("can not write data")
	}
	return
}

func calcBlockAddress(sector int, block int) (addr byte) {
	addr = byte(sector*4 + block)
	return
}

/*
ReadBlock reads the block from the card.
	sector - card sector to read from
	block - the block within the sector (0-3 tor Mifare 4K)
*/
func (r *Dev) ReadBlock(sector int, block int) (res []byte, err error) {
	res, err = r.read(calcBlockAddress(sector, block%3))
	return
}

/*
WriteBlock writes the data into the card block.
	auth - the authentiction mode.
	sector - the sector on the card to write to.
	block - the block within the sector to write into.
	data - 16 bytes if data to write
	key - the key used to authenticate the card - depends on the used auth method.
*/
func (r *Dev) WriteBlock(auth byte, sector int, block int, data [16]byte, key []byte) (err error) {
	defer func() {
		r.StopCrypto()
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
		err = errors.New("Authentication failed")
		return
	}

	err = r.write(calcBlockAddress(sector, block%3), data[:])
	return
}

/*
ReadSectorTrail reads the sector trail (the last sector that contains the sector access bits)
	sector - the sector number to read the data from.
*/
func (r *Dev) ReadSectorTrail(sector int) (res []byte, err error) {
	res, err = r.read(calcBlockAddress(sector&0xFF, 3))
	return
}

/*
WriteSectorTrail writes the sector trait with sector access bits.
	auth - authentication mode.
	sector - sector to set authentication.
	keyA - the key used for AuthA authentication scheme.
	keyB - the key used for AuthB authentication schemd.
	access - the block access structure.
	key - the current key used to authenticate the provided sector.
*/
func (r *Dev) WriteSectorTrail(auth byte, sector int, keyA [6]byte, keyB [6]byte, access *BlocksAccess, key []byte) (err error) {
	defer func() {
		r.StopCrypto()
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
		err = errors.New("Failed to authenticate")
		return
	}

	data := make([]byte, 16)
	copy(data, keyA[:])
	accessData := CalculateBlockAccess(access)
	copy(data[6:], accessData[:4])
	copy(data[10:], keyB[:])
	err = r.write(calcBlockAddress(sector&0xFF, 3), data)
	return
}

/*
Auth authenticate the card fof the sector/block using the provided data.
	mode - the authentication mode.
	sector - the sector to authenticate on.
	block - the block within sector to authenticate.
	sectorKey - the key to be used for accessing the sector data.
	serial - the serial of the card.
*/
func (r *Dev) Auth(mode byte, sector, block int, sectorKey []byte, serial []byte) (authS AuthStatus, err error) {
	authS, err = r.auth(mode, calcBlockAddress(sector, block), sectorKey, serial)
	return
}

func (r *Dev) selectCard() (uuid []byte, err error) {
	err = r.Wait()
	if err != nil {
		return
	}
	err = r.Init()
	if err != nil {
		return
	}
	_, err = r.Request()
	if err != nil {
		return
	}
	uuid, err = r.AntiColl()
	if err != nil {
		return
	}
	_, err = r.SelectTag(uuid)
	if err != nil {
		return
	}
	return
}

/*
ReadCard reads the card sector/block.
	auth - the authentication mode.
	sector - the sector to authenticate on.
	block - the block within sector to authenticate.
	key - the key to be used for accessing the sector data.
*/
func (r *Dev) ReadCard(auth byte, sector int, block int, key []byte) (data []byte, err error) {
	defer func() {
		r.StopCrypto()
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
		err = errors.New("Can not authenticate")
		return
	}

	data, err = r.ReadBlock(sector, block)

	return
}

/*
ReadAuth - read the card authentication data.
	sector - the sector to authenticate on.
	key - the key to be used for accessing the sector data.
*/
func (r *Dev) ReadAuth(auth byte, sector int, key []byte) (data []byte, err error) {
	defer func() {
		r.StopCrypto()
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
		err = errors.New("Can not authenticate")
		return
	}

	data, err = r.read(calcBlockAddress(sector, 3))
	return
}

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

// BlocksAccess defines the access structure for first 3 blocks of the sector and the access bits for the
// sector trail.
type BlocksAccess struct {
	B0, B1, B2 BlockAccess
	B3         SectorTrailerAccess
}

func (ba *BlocksAccess) getBits(bitNum uint) (res byte) {
	shift := bitNum - 1
	bit := byte(1 << shift)
	res = (byte(ba.B0)&bit)>>shift | ((byte(ba.B1)&bit)>>shift)<<1 | ((byte(ba.B2)&bit)>>shift)<<2 | ((byte(ba.B3)&bit)>>shift)<<3
	return
}

// CalculateBlockAccess calculates the block access.
func CalculateBlockAccess(ba *BlocksAccess) (res []byte) {
	res = make([]byte, 4)
	res[0] = (^ba.getBits(1) & 0x0F) | ((^ba.getBits(2) & 0x0F) << 4)
	res[1] = (^ba.getBits(3) & 0x0F) | (ba.getBits(1) & 0x0F << 4)
	res[2] = (ba.getBits(2) & 0x0F) | (ba.getBits(3) & 0x0F << 4)
	res[3] = res[0] ^ res[1] ^ res[2]
	return
}

// ParseBlockAccess parses the given byte array into the block access structure.
func ParseBlockAccess(ad []byte) (ba *BlocksAccess) {
	ba = new(BlocksAccess)
	ba.B0 = BlockAccess(ad[1]&0x10>>4 | ad[2]&0x01<<1 | ad[2]&0x10>>2)
	ba.B1 = BlockAccess(ad[1]&0x20>>5 | ad[2]&0x02 | ad[2]&0x20>>3)
	ba.B2 = BlockAccess(ad[1]&0x40>>6 | ad[2]&0x04>>1 | ad[2]&0x40>>4)
	ba.B3 = SectorTrailerAccess(ad[1]&0x80>>7 | ad[2]&0x08>>2 | ad[2]&0x80>>5)
	return
}
