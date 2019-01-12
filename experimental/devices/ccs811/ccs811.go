package ccs811

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
)

// Different measurement modes constants
const (
	MeasurementModeIdle         byte = 0
	MeasurementModeConstant1000 byte = 1
	MeasurementModePulse        byte = 2
	MeasurementModeLowPower     byte = 3
	MeasurementModeConstant250  byte = 4
)

// NeededData represents set of data read from the sensor
type NeededData byte

// What data should be read from the sensor
const (
	ReadCO2          NeededData = 2
	ReadCO2VOC       NeededData = 4
	ReadCO2VOCStatus NeededData = 5
	ReadAll          NeededData = 8
)

// SensorErrorID represents error reported by the sensor
type SensorErrorID byte

/*
Error constants, applicable if status registers signals error
0
WRITE_REG_INVALID
The CCS811 received an I2C write request addressed to this station
but with invalid register address ID
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

// Opts holds the configuration options.
type Opts struct {
	Addr               uint16
	MeasurementMode    byte
	InterruptWhenReady bool
	UseThreshold       bool
}

// DefaultOpts are the safe default options.
var DefaultOpts = Opts{
	Addr:               0x5A,
	MeasurementMode:    MeasurementModeConstant1000,
	InterruptWhenReady: false,
	UseThreshold:       false,
}

// New creates a new driver for CCS811 VOC sensor
func New(bus i2c.Bus, opts *Opts) (*Dev, error) {
	if opts.Addr != 0x5A && opts.Addr != 0x5B {
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
	err := dev.StartSensorApp()
	if err != nil {
		return nil, fmt.Errorf("Error transitioning from boot do app mode: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	dev.SetMeasurementMode(MeasurementModeConstant1000, opts.InterruptWhenReady, opts.UseThreshold)

	return dev, nil
}

// Dev is an handle to an CCS811 sensor.
type Dev struct {
	c    conn.Conn
	opts *Opts
}

const ( //registers
	statusReg          byte = 0x00
	measurementModeReg byte = 0x01
	algoResultsReg     byte = 0x02
	rawDataReg         byte = 0x03
	environmentReg     byte = 0x05
	baselineReg        byte = 0x11
	resetReg           byte = 0xFF
)

func (d *Dev) String() string {
	return "CCS811"
}

// StartSensorApp initializes sensor to application mode
func (d *Dev) StartSensorApp() error {
	err := d.c.Tx([]byte{0xf4}, nil)
	if err != nil {
		return err
	}
	return nil
}

// SetMeasurementMode sets one of the 5 measurement modes, interrupt generation
// and interrupt threshold
//
// generateInterrupt:
//		if true, CCS811 will trigger interrupt when new data is available
// useThreshold:
// 		if true, you have to set Threshold register with appropriate values
func (d *Dev) SetMeasurementMode(measurementMode byte, generateInterrupt, useThreshold bool) error {
	mesModeValue := (measurementMode << 4)
	if generateInterrupt {
		mesModeValue = mesModeValue | (0x1 << 3)
	}
	if useThreshold {
		mesModeValue = mesModeValue | (0x1 << 2)
	}

	// set measurement mode
	err := d.c.Tx([]byte{measurementModeReg, mesModeValue}, nil)
	if err != nil {
		return err
	}
	return nil
}

type MeasurementMode struct {
	MeasurementMode   byte
	GenerateInterrupt bool
	UseThreshold      bool
}

func (d *Dev) GetMeasurementMode() (mm MeasurementMode, err error) {
	r := make([]byte, 1)
	err = d.c.Tx([]byte{measurementModeReg}, r)
	if err != nil {
		return mm, err
	}
	mode := r[0] >> 4
	threshold := (r[0]&4 == 1)
	interrupt := (r[0]&8 == 1)

	return MeasurementMode{MeasurementMode: mode, GenerateInterrupt: interrupt, UseThreshold: threshold}, nil
}

// Reset sets device into the BOOT mode
func (d *Dev) Reset() error {
	if err := d.c.Tx([]byte{resetReg, 0x11, 0xE5, 0x72, 0x8A}, nil); err != nil {
		return err
	}
	return nil
}

// ReadStatus returns value of status register
func (d *Dev) ReadStatus() (byte, error) {
	r := make([]byte, 1)
	if err := d.c.Tx([]byte{statusReg}, r); err != nil {
		return 0, err
	}
	return r[0], nil
}

// ReadRawData provides current and voltage on the sensor
// current in uA
// voltage from 0-1023, where 1023 = 1.65V
func (d *Dev) ReadRawData() (current int, voltage int, err error) {
	r := make([]byte, 2)
	if err = d.c.Tx([]byte{rawDataReg}, r); err != nil {
		return 0, 0, err
	}
	current, voltage = valuesFromRawData(r)
	return
}

// SetEnvironmentData allows to provide temperature and humidity so
// sensor can compensate it's measurement
func (d *Dev) SetEnvironmentData(temp, humidity float32) error {
	rawTemp := uint16((temp + 25) * 512)
	rawHum := uint16(humidity * 512)
	w := []byte{environmentReg,
		byte(rawHum >> 8),
		byte(rawHum),
		byte(rawTemp >> 8),
		byte(rawTemp)}

	if err := d.c.Tx(w, nil); err != nil {
		return err
	}
	return nil
}

// GetBaseline provides current baseline used by measurement alogrithm
func (d *Dev) GetBaseline() ([]byte, error) {
	r := make([]byte, 2)
	if err := d.c.Tx([]byte{baselineReg}, r); err != nil {
		return r, err
	}
	return r, nil
}

/*
SetBaseline sets current baseline for measurement algorithm.
For mor detail check sensor's specification

Manual Baseline Correction
There is a mechanism within CCS811 to manually save and restore a previously
saved baseline value using the BASELINE register. The correct time to save the baseline
will depend on the customer use-case and application.

• For devices which are powered for >24 hours at a time:
• During the first 500 hours – save the baseline every
24-48 hours.
• After the first 500 hours – save the baseline every 5-7 days.

• For devices which are powered <24 hours at a time:
• If the device is run in, save the baseline before power down
• If multiple operating modes are used, a separate baseline should be stored for each
• The baseline should only be restored when the resistance is stable (typically 20-30 minutes)
• If changing from a low to high power mode (without spending at least 10 minutes in idle),
  the sensor resistance should be allowed to settle again before restoring the baseline

Note(s):
1) If a value is written to the BASELINE register while the sensor is stabilising, the output of the TVOC and eCO2 calculations may be higher than expected.
2) The baseline must be written after the conditioning period
*/
func (d *Dev) SetBaseline(baseline []byte) error {
	w := []byte{baselineReg, baseline[0], baseline[1]}
	if err := d.c.Tx(w, nil); err != nil {
		return err
	}
	return nil
}

// SensorValues represents data read from the sensor.
// Data are populated based on NeededData parameter.
type SensorValues struct {
	ECO2           int
	VOC            int
	Status         byte
	ErrorID        SensorErrorID
	RawDataCurrent int // current in uA
	RawDataVoltage int // voltage from 0-1023, where 1023 = 1.65V
}

// Sense provides data from the sensor.
// ReadData parameter specifies which data should be read.
func (d *Dev) Sense(mode NeededData) (*SensorValues, error) {
	read := make([]byte, mode)
	err := d.c.Tx([]byte{algoResultsReg}, read)
	if err != nil {
		return nil, err
	}
	sv := &SensorValues{}
	if mode >= ReadCO2 {
		// exptected range: 400ppm to 8192ppm
		// 0x3F is used to erase randomly set top bits causing value out of range given by specs
		sv.ECO2 = int(uint32(read[0]&0x3F)<<8 | uint32(read[1]))
	}
	if mode >= ReadCO2VOC {
		// expected range: 0ppb to 1187ppb
		// 0x7 is used to erase randomly set top bits causing value out of range given by specs
		sv.VOC = int(uint32(read[2]&0x7)<<8 | uint32(read[3]))
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

// parse current and voltage from raw data
func valuesFromRawData(data []byte) (current int, voltage int) {
	current = int(data[0] >> 2)
	voltage = int((uint16(data[0]&0x03) << 8) | uint16(data[1]))
	return current, voltage
}

type FwVersions struct {
	HWIdentifier       byte
	HWVersion          byte
	BootVersion        string
	ApplicationVersion string
}

func (d *Dev) GetFirmwareData() (version *FwVersions, err error) {
	version = &FwVersions{}
	hwid := make([]byte, 1)

	if err := d.c.Tx([]byte{0x20}, hwid); err != nil {
		return version, err
	} else {
		version.HWIdentifier = hwid[0]
	}

	hwver := make([]byte, 1)
	if err := d.c.Tx([]byte{0x21}, hwver); err != nil {
		return version, err
	} else {
		version.HWVersion = hwver[0]
	}

	bootver := make([]byte, 2)
	if err := d.c.Tx([]byte{0x23}, bootver); err != nil {
		return version, err
	} else {
		minor := bootver[0] & 0x0F
		major := (bootver[0] & 0xF0) >> 4
		trivial := bootver[1]
		version.BootVersion = fmt.Sprintf("%d.%d.%d", major, minor, trivial)
	}

	appver := make([]byte, 2)
	if err := d.c.Tx([]byte{0x24}, appver); err != nil {
		return version, err
	} else {
		minor := appver[0] & 0x0F
		major := (appver[0] & 0xF0) >> 4
		trivial := appver[1]
		version.ApplicationVersion = fmt.Sprintf("%d.%d.%d", major, minor, trivial)
	}
	return version, nil
}
