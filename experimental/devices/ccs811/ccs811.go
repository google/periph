package ccs811

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
)

type Opts struct {
	Addr               uint16
	MeasurementMode    byte
	InterruptWhenReady byte
	UseThreshold       byte
}

func New(bus i2c.Bus, opts *Opts) (*Dev, error) {
	if opts.Addr != 0x5A || opts.Addr != 0x5B {
		return nil, fmt.Errorf("Invalid device address, only 0x5A or 0x5B are allowed")
	}

	if opts.MeasurementMode > 4 {
		return nil, fmt.Errorf("Invalid measurement mode")
	}

	dev := &Dev{
		c:    &i2c.Dev{Bus: bus, Addr: opts.Addr},
		opts: opts,
	}

	// from boot mode to measurement mode
	err := dev.c.Tx([]byte{0xf4}, nil)
	if err != nil {
		return nil, fmt.Errorf("Write error: ", err)
	}

	time.Sleep(20 * time.Millisecond)

	mesModeValue := (opts.MeasurementMode << 4)
	mesModeValue = mesModeValue | ((0x1 & opts.InterruptWhenReady) << 3)
	mesModeValue = mesModeValue | ((0x1 & opts.UseThreshold) << 2)

	dev.SetMeasurementMode(mesModeValue)

	return dev, nil
}

const (
	MeasurementModeIdle         byte = 0
	MeasurementModeConstant1000 byte = 1
	MeasurementModePulse        byte = 2
	MeasurementModeLowPower     byte = 3
	MeasurementModeConstant250  byte = 4
)

type Dev struct {
	c    conn.Conn
	opts *Opts
}

const ( //registers
	statusReg          byte = 0x00
	measurementModeReg byte = 0x01
	algoResultsReg     byte = 0x02
	rawDataReg         byte = 0x03
)

func (d *Dev) SetMeasurementMode(mesModeValue byte) error {
	// set measurement mode
	err := d.c.Tx([]byte{measurementModeReg, mesModeValue}, nil)
	if err != nil {
		fmt.Println("Write error: ", err)
	}
	return err
}

func (d *Dev) Reset() error {
	if err := d.c.Tx([]byte{0x11, 0xE5, 0x72, 0x8A}, nil); err != nil {
		return err
	}
	return nil
}

func (d *Dev) ReadStatus() (byte, error) {
	// r is a read buffer, Tx will try and read len(rx) bytes.
	r := make([]byte, 1)

	if err := d.c.Tx([]byte{statusReg}, r); err != nil {
		return 0, err
	}
	return r[0], nil
}

type ReadData byte

// what data should be read from sensor
const (
	ReadCO2          ReadData = 2
	ReadCO2VOC       ReadData = 4
	ReadCO2VOCStatus ReadData = 5
	ReadAll          ReadData = 8
)

type SensorValues struct {
	ECO2           int
	VOC            int
	Status         byte
	ErrorID        SensorErrorID
	RawDataCurrent int
	RawDataVoltage int
}

func (d *Dev) Sense(mode ReadData) (*SensorValues, error) {
	read := make([]byte, mode)
	err := d.c.Tx([]byte{algoResultsReg}, read)
	if err != nil {
		return nil, err
	}
	sv := &SensorValues{}
	if mode >= ReadCO2 {
		sv.ECO2 = int(uint32(read[0])<<8 | uint32(read[1]))
	}
	if mode >= ReadCO2VOC {
		sv.VOC = int(uint32(read[2])<<8 | uint32(read[3]))
	}
	if mode >= ReadCO2VOCStatus {
		sv.Status = read[4]
	}
	if mode == ReadAll {
		sv.ErrorID = SensorErrorID(read[5])
		sv.RawDataCurrent, sv.RawDataVoltage = valuesFromRawData(read[6:])
	}

	return sv, nil
}

func valuesFromRawData(data []byte) (int, int) {
	current := int(data[0] >> 2)
	voltage := int((uint16(data[0]&0x03) << 8) | uint16(data[1]))
	return current, voltage
}

type SensorErrorID byte

/*
WRITE_REG_INVALID
The CCS811 received an I2C write request addressed to this station but with invalid register address ID
1
READ_REG_INVALID
The CCS811 received an I2C read request to a mailbox ID that is invalid
2
MEASMODE_INVALID
The CCS811 received an I2C request to write an unsupported mode to MEAS_MODE
3
MAX_RESISTANCE
The sensor resistance measurement has reached or exceeded the maximum range
4
HEATER_FAULT
The Heater current in the CCS811 is not in range
5
HEATER_SUPPLY
The Heater voltage is not being applied correctly
*/
const (
	WriteRegInvalid SensorErrorID = 0x1
	ReadRegInvalid  SensorErrorID = 0x2
	MeasModeInvalid SensorErrorID = 0x4
	MaxResistance   SensorErrorID = 0x8
	HeaterFault     SensorErrorID = 0x10
	HeaterSupply    SensorErrorID = 0x20
)
