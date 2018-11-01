// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package commands

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
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

// NewSPI creates and initializes the RFID card reader attached to SPI.
//
// 	spiPort - the SPI device to use.
// 	resetPin - reset GPIO pin.
// 	irqPin - irq GPIO pin.
func NewLowLevelSPI(spiPort spi.Port, resetPin gpio.PinOut, irqPin gpio.PinIn) (*LowLevel, error) {
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

	dev := &LowLevel{
		spiDev:   spiDev,
		irqPin:   irqPin,
		resetPin: resetPin,
	}

	return dev, nil
}

// LowLevel is a low-level handler of a MFRC522 RFID reader.
type LowLevel struct {
	resetPin gpio.PinOut
	irqPin   gpio.PinIn
	spiDev   spi.Conn
}

// String implements conn.Resource.
func (r *LowLevel) String() string {
	return fmt.Sprintf("Mifare MFRC522 [bus: %v, reset pin: %s, irq pin: %s]",
		r.spiDev, r.resetPin.Name(), r.irqPin.Name())
}

// DevWrite sends data to a device.
func (r *LowLevel) DevWrite(address int, data byte) error {
	newData := []byte{(byte(address) << 1) & 0x7E, data}
	return r.spiDev.Tx(newData, nil)
}

// DevRead gets data from a device.
func (r *LowLevel) DevRead(address int) (byte, error) {
	data := []byte{((byte(address) << 1) & 0x7E) | 0x80, 0}
	out := make([]byte, len(data))
	if err := r.spiDev.Tx(data, out); err != nil {
		return 0, err
	}
	return out[1], nil
}

// CRC calculates the CRC of the data using the card chip.
func (r *LowLevel) CRC(inData []byte) ([]byte, error) {
	if err := r.ClearBitmask(DivIrqReg, 0x04); err != nil {
		return nil, err
	}
	if err := r.SetBitmask(FIFOLevelReg, 0x80); err != nil {
		return nil, err
	}
	for _, v := range inData {
		if err := r.DevWrite(FIFODataReg, v); err != nil {
			return nil, err
		}
	}
	if err := r.DevWrite(CommandReg, PCD_CALCCRC); err != nil {
		return nil, err
	}
	for i := byte(0xFF); i > 0; i-- {
		n, err := r.DevRead(DivIrqReg)
		if err != nil {
			return nil, err
		}
		if n&0x04 > 0 {
			break
		}
	}
	lsb, err := r.DevRead(CRCResultRegL)
	if err != nil {
		return nil, err
	}

	msb, err := r.DevRead(CRCResultRegM)
	if err != nil {
		return nil, err
	}
	return []byte{lsb, msb}, nil
}

// SetBitmask sets register bit.
func (r *LowLevel) SetBitmask(address, mask int) error {
	current, err := r.DevRead(address)
	if err != nil {
		return err
	}
	return r.DevWrite(address, current|byte(mask))
}

// ClearBitmask clears register bit.
func (r *LowLevel) ClearBitmask(address, mask int) error {
	current, err := r.DevRead(address)
	if err != nil {
		return err
	}
	return r.DevWrite(address, current&^byte(mask))
}

// stopCrypto stops the crypto chip.
func (r *LowLevel) StopCrypto() error {
	return r.ClearBitmask(Status2Reg, 0x08)
}

// WaitForEdge waits for an IRQ pin to strobe.
func (r *LowLevel) WaitForEdge(timeout time.Duration) bool {
	return r.irqPin.WaitForEdge(timeout)
}

// Auth authenticate the card fof the sector/block using the provided data.
//
// 	mode - the authentication mode.
// 	sector - the sector to authenticate on.
// 	block - the block within sector to authenticate.
// 	sectorKey - the key to be used for accessing the sector data.
// 	serial - the serial of the card.
func (r *LowLevel) Auth(mode byte, blockAddress byte, sectorKey [6]byte, serial []byte) (AuthStatus, error) {
	buffer := make([]byte, 2)
	buffer[0] = mode
	buffer[1] = blockAddress
	buffer = append(buffer, sectorKey[:]...)
	buffer = append(buffer, serial[:4]...)
	_, _, err := r.CardWrite(PCD_AUTHENT, buffer)
	if err != nil {
		return AuthReadFailure, err
	}
	if n, err := r.DevRead(Status2Reg); err != nil || n&0x08 == 0 {
		return AuthFailure, err
	}
	return AuthOk, nil
}

// CardWrite the low-level interface to write some raw commands to the card.
//
// 	command - the command register
// 	data - the data to write out to the card using the authenticated sector.
func (r *LowLevel) CardWrite(command byte, data []byte) ([]byte, int, error) {
	var backData []byte
	backLength := -1
	irqEn := byte(0x00)
	irqWait := byte(0x00)

	switch command {
	case PCD_AUTHENT:
		irqEn = 0x12
		irqWait = 0x10
	case PCD_TRANSCEIVE:
		irqEn = 0x77
		irqWait = 0x30
	}

	if err := r.DevWrite(CommIEnReg, irqEn|0x80); err != nil {
		return nil, -1, err
	}
	if err := r.ClearBitmask(CommIrqReg, 0x80); err != nil {
		return nil, -1, err
	}
	if err := r.SetBitmask(FIFOLevelReg, 0x80); err != nil {
		return nil, -1, err
	}
	if err := r.DevWrite(CommandReg, PCD_IDLE); err != nil {
		return nil, -1, err
	}

	for _, v := range data {
		if err := r.DevWrite(FIFODataReg, v); err != nil {
			return nil, -1, err
		}
	}

	if err := r.DevWrite(CommandReg, command); err != nil {
		return nil, -1, err
	}

	if command == PCD_TRANSCEIVE {
		if err := r.SetBitmask(BitFramingReg, 0x80); err != nil {
			return nil, -1, err
		}
	}

	i := 2000
	n := byte(0)

	for ; i > 0; i-- {
		n, err := r.DevRead(CommIrqReg)
		if err != nil {
			return nil, -1, err
		}
		if n&(irqWait|1) != 0 {
			break
		}
	}

	if err := r.ClearBitmask(BitFramingReg, 0x80); err != nil {
		return nil, -1, err
	}

	if i == 0 {
		return nil, -1, wrapf("can't read data after 2000 loops")
	}

	if d, err := r.DevRead(ErrorReg); err != nil || d&0x1B != 0 {
		return nil, -1, err
	}

	if n&irqEn&0x01 == 1 {
		return nil, -1, wrapf("IRQ error")
	}

	if command == PCD_TRANSCEIVE {
		n, err := r.DevRead(FIFOLevelReg)
		if err != nil {
			return nil, -1, err
		}
		lastBits, err := r.DevRead(ControlReg)
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
			byteVal, err := r.DevRead(FIFODataReg)
			if err != nil {
				return nil, -1, err
			}
			backData[i] = byteVal
		}

	}

	return backData, backLength, nil
}

func wrapf(format string, a ...interface{}) error {
	return fmt.Errorf("mfrc522: "+format, a...)
}
