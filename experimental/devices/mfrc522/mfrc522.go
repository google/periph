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

// Dev is an handle to an MFRC522 RFID reader.
type Dev struct {
	LowLevel         *commands.LowLevel
	operationTimeout time.Duration
	beforeCall       func()
	afterCall        func()
}

// Key is the access key that consists of 6 bytes. There could be two types of keys - keyA and keyB.
// KeyA and KeyB correspond to the different sector trail and data access. Refer to the datasheet for more details.
type Key [6]byte

// DefaultKey  provides the default bytes for card authentication for method B.
var DefaultKey = Key{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

type config struct {
	defaultTimeout time.Duration
	beforeCall     func()
	afterCall      func()
}

type configF func(*config) *config

// WithTimeout updates the default device-wide configuration timeout.
func WithTimeout(timeout time.Duration) configF {
	return func(c *config) *config {
		c.defaultTimeout = timeout
		return c
	}
}

// WithSyncsets the synchronization for the entire device, using internal mutex.
func WithSync() configF {
	var mu sync.Mutex
	return func(c *config) *config {
		c.beforeCall = mu.Lock
		c.afterCall = mu.Unlock
		return c
	}
}

// noop does nothing
func noop() {}

// NewSPI creates and initializes the RFID card reader attached to SPI.
//
//  spiPort     the SPI device to use.
//  resetPin    reset GPIO pin.
//  irqPin      irq GPIO pin.
func NewSPI(spiPort spi.Port, resetPin gpio.PinOut, irqPin gpio.PinIn, configs ...configF) (*Dev, error) {
	cfg := &config{
		defaultTimeout: 30 * time.Second,
		beforeCall:     noop,
		afterCall:      noop,
	}
	for _, cf := range configs {
		cfg = cf(cfg)
	}
	raw, err := commands.NewLowLevelSPI(spiPort, resetPin, irqPin)
	if err != nil {
		return nil, err
	}
	if err := raw.Init(); err != nil {
		return nil, err
	}

	dev := &Dev{
		LowLevel:         raw,
		operationTimeout: cfg.defaultTimeout,
		beforeCall:       cfg.beforeCall,
		afterCall:        cfg.afterCall,
	}
	return dev, nil
}

// String implements conn.Resource.
func (r *Dev) String() string {
	return r.LowLevel.String()
}

// Halt implements conn.Resource.
//
// It soft-stops the chip - PowerDown bit set, command IDLE
func (r *Dev) Halt() error {
	r.beforeCall()
	defer r.afterCall()
	return r.LowLevel.Halt()
}

// SetAntennaGain configures antenna signal strength.
//
//  gain    signal strength from 0 to 7.
func (r *Dev) SetAntennaGain(gain int) error {
	r.beforeCall()
	defer r.afterCall()
	if gain < 0 || gain > 7 {
		return wrapf("gain must be in [0..7] interval")
	}
	r.LowLevel.SetAntennaGain(gain)
	return nil
}

// ReadCard reads the card sector/block.
//
//  auth     the authentication mode.
//  sector   the sector to authenticate on.
//  block    the block within sector to authenticate.
//  key      the key to be used for accessing the sector data.
func (r *Dev) ReadCard(auth byte, sector int, block int, key Key) (data []byte, err error) {
	return r.ReadCardTimed(r.operationTimeout, auth, sector, block, key)
}

// ReadCardTimed   reads the card sector/block with IRQ event timeout.
//
//  timeout   the operation timeout
//  auth      the authentication mode.
//  sector    the sector to authenticate on.
//  block     the block within sector to authenticate.
//  key       the key to be used for accessing the sector data.
func (r *Dev) ReadCardTimed(timeout time.Duration, auth byte, sector int, block int, key Key) (data []byte, err error) {
	r.beforeCall()
	defer func() {
		r.afterCall()
		if err == nil {
			err = r.LowLevel.StopCrypto()
		}
	}()
	uuid, err := r.selectCard(timeout)
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

// ReadAuth     reads the card authentication data.
//
//  auth    authentication type
//  sector  the sector to authenticate on.
//  key     the key to be used for accessing the sector data.
func (r *Dev) ReadAuth(auth byte, sector int, key Key) (data []byte, err error) {
	return r.ReadAuthTimed(r.operationTimeout, auth, sector, key)
}

// ReadAuthTimed   reads the card authentication data with IRQ event timeout.
//
//  timeout    the operation timeout
//  auth       authentication type
//  sector     the sector to authenticate on.
//  key        the key to be used for accessing the sector data.
func (r *Dev) ReadAuthTimed(timeout time.Duration, auth byte, sector int, key Key) (data []byte, err error) {
	r.beforeCall()
	defer func() {
		r.afterCall()
		if err == nil {
			err = r.LowLevel.StopCrypto()
		}
	}()
	uuid, err := r.selectCard(timeout)
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

// WriteCardTimed   writes the data into the card block.
//
//  auth       the authentiction mode.
//  sector     the sector on the card to write to.
//  block      the block within the sector to write into.
//  data       16 bytes if data to write
//  key        the key used to authenticate the card - depends on the used auth method.
func (r *Dev) WriteCard(auth byte, sector int, block int, data [16]byte, key Key) (err error) {
	return r.WriteCardTimed(r.operationTimeout, auth, sector, block, data, key)
}

// WriteCard    writes the data into the card block with IRQ event timeout.
//
//  timeout     the operation timeout
//  auth        the authentiction mode.
//  sector      the sector on the card to write to.
//  block       the block within the sector to write into.
//  data        16 bytes if data to write
//  key          the key used to authenticate the card - depends on the used auth method.
func (r *Dev) WriteCardTimed(timeout time.Duration, auth byte, sector int, block int, data [16]byte, key Key) (err error) {
	r.beforeCall()
	defer func() {
		r.afterCall()
		if err == nil {
			err = r.LowLevel.StopCrypto()
		}
	}()
	uuid, err := r.selectCard(timeout)
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
// auth     authentication mode.
// sector   sector to set authentication.
// keyA     the key used for AuthA authentication scheme.
// keyB     the key used for AuthB authentication scheme.
// access   the block access structure.
// key      the current key used to authenticate the provided sector.
func (r *Dev) WriteSectorTrail(auth byte, sector int, keyA Key, keyB Key, access *BlocksAccess, key Key) (err error) {
	return r.WriteSectorTrailTimed(r.operationTimeout, auth, sector, keyA, keyB, access, key)
}

// WriteSectorTrailTimed  writes the sector trail with sector access bits with IRQ event timeout.
//
//  timeout   operation timeout
//  auth      authentication mode.
//  sector    sector to set authentication.
//  keyA      the key used for AuthA authentication scheme.
//  keyB      the key used for AuthB authentication scheme.
//  access    the block access structure.
//  key       the current key used to authenticate the provided sector.
func (r *Dev) WriteSectorTrailTimed(timeout time.Duration, auth byte, sector int, keyA Key, keyB Key, access *BlocksAccess, key Key) (err error) {
	r.beforeCall()
	defer func() {
		r.afterCall()
		if err == nil {
			err = r.LowLevel.StopCrypto()
		}
	}()
	uuid, err := r.selectCard(timeout)
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
	var accessData [4]byte
	if err := access.serialize(accessData[:]); err != nil {
		return err
	}
	copy(data[6:], accessData[:])
	copy(data[10:], keyB[:])
	return r.write(calcBlockAddress(sector&0xFF, 3), data[:])
}

//         MFRC522 SPI Dev private/helper functions

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
//  sector - card sector to read from
//  block - the block within the sector (0-3 tor Mifare 4K)
func (r *Dev) readBlock(sector int, block int) ([]byte, error) {
	return r.read(calcBlockAddress(sector, block%3))
}

// selectCard selects the card after the IRQ event was received.
func (r *Dev) selectCard(timeout time.Duration) ([]byte, error) {
	if err := r.LowLevel.WaitForEdge(timeout); err != nil {
		return nil, err
	}
	if err := r.LowLevel.Init(); err != nil {
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

// write  writes the data block into the card at given block address.
//
// blockAddr - the calculated block address
// data - the sector data bytes
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

// preAccess  calculates CRC of the block address to be accessed and sends it to the device for verification.
//
//  blockAddr - the block address to access.
//  cmd - command code to perform on the given block,
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

// read	reads the block
//
//	blockAddr the address to read from the card.
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

func wrapf(format string, a ...interface{}) error {
	return fmt.Errorf("mfrc522: "+format, a...)
}

var _ conn.Resource = &Dev{}
