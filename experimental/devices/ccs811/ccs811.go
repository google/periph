// Copyright 2019 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ccs811

import (
	"fmt"

	"periph.io/x/periph/conn/physic"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/i2c"
)

// MeasurementMode represents different ways how data is read
type MeasurementMode byte

// Different measurement mode constants:
//
// -  Mode 0: Idle, low current mode.
//
// -  Mode 1: Constant power mode, IAQ measurement every second.
//
// -  Mode 2: Pulse heating mode IAQ measurement every 10 seconds.
//
// -  Mode 3: Low power pulse heating mode IAQ measurement every 60 seconds.
//
// -  Mode 4: Constant power mode, sensor measurement every 250ms.
const (
	MeasurementModeIdle         MeasurementMode = 0
	MeasurementModeConstant1000 MeasurementMode = 1
	MeasurementModePulse        MeasurementMode = 2
	MeasurementModeLowPower     MeasurementMode = 3
	MeasurementModeConstant250  MeasurementMode = 4
)

// NeededData represents set of data read from the sensor.
type NeededData byte

// What data should be read from the sensor.
const (
	ReadCO2          NeededData = 2
	ReadCO2VOC       NeededData = 4
	ReadCO2VOCStatus NeededData = 5
	ReadAll          NeededData = 8
)

// SensorErrorID represents error reported by the sensor.
type SensorErrorID byte

// Error constants, applicable if status registers signals error.
//
// 0 WRITE_REG_INVALID The CCS811 received an I2C write request addressed
// to this station but with invalid register address ID
//
// 1 READ_REG_INVALID The CCS811 received an I2C read request to a mailbox ID that is invalid
//
// 2 MEASMODE_INVALID The CCS811 received an I2C request to write an unsupported mode to MEAS_MODE
//
// 3 MAX_RESISTANCE The sensor resistance measurement has reached or exceeded the maximum range
//
// 4 HEATER_FAULT The Heater current in the CCS811 is not in range
//
// 5 HEATER_SUPPLY The Heater voltage is not being applied correctly
const (
	WriteRegInvalid SensorErrorID = 0x1
	ReadRegInvalid  SensorErrorID = 0x2
	MeasModeInvalid SensorErrorID = 0x4
	MaxResistance   SensorErrorID = 0x8
	HeaterFault     SensorErrorID = 0x10
	HeaterSupply    SensorErrorID = 0x20
)

// Opts holds the configuration options. The address must be 0x5A or 0x5B.
type Opts struct {
	Addr               uint16
	MeasurementMode    MeasurementMode
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

// New creates a new driver for CCS811 VOC sensor.
func New(bus i2c.Bus, opts *Opts) (*Dev, error) {
	if opts.Addr != 0x5A && opts.Addr != 0x5B {
		return nil, fmt.Errorf("Invalid device address, only 0x5A or 0x5B are allowed")
	}

	if opts.MeasurementMode > MeasurementModeConstant250 {
		return nil, fmt.Errorf("Invalid measurement mode")
	}

	dev := &Dev{
		c:    &i2c.Dev{Bus: bus, Addr: opts.Addr},
		opts: *opts,
	}

	// From boot mode to measurement mode.
	if err := dev.StartSensorApp(); err != nil {
		return nil, fmt.Errorf("Error transitioning from boot do app mode: %v", err)
	}
	mmp := &MeasurementModeParams{MeasurementMode: opts.MeasurementMode,
		GenerateInterrupt: opts.InterruptWhenReady,
		UseThreshold:      opts.UseThreshold}

	if err := dev.SetMeasurementModeRegister(*mmp); err != nil {
		return nil, fmt.Errorf("Error setting measurement mode: %v", err)
	}

	return dev, nil
}

// Dev is an handle to an CCS811 sensor.
type Dev struct {
	c    conn.Conn
	opts Opts
}

const ( //Sensor's registers.
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

// StartSensorApp initializes sensor to application mode.
func (d *Dev) StartSensorApp() error {
	return d.c.Tx([]byte{0xf4}, nil)
}

// SetMeasurementModeRegister sets one of the 5 measurement modes, interrupt generation
// and interrupt threshold.
func (d *Dev) SetMeasurementModeRegister(mmp MeasurementModeParams) error {
	mesModeValue := (mmp.MeasurementMode << 4)
	if mmp.GenerateInterrupt {
		mesModeValue = mesModeValue | (0x1 << 3)
	}
	if mmp.UseThreshold {
		mesModeValue = mesModeValue | (0x1 << 2)
	}

	return d.c.Tx([]byte{measurementModeReg, byte(mesModeValue)}, nil)
}

// MeasurementModeParams is a structure representing Measuremode register of the sensor.
type MeasurementModeParams struct {
	MeasurementMode   MeasurementMode
	GenerateInterrupt bool // True if sensor should generate interrupts on new measurement.
	UseThreshold      bool // True if sensor should use thresholds from threshold register.
}

// GetMeasurementModeRegister returns current measurement mode of the sensor.
func (d *Dev) GetMeasurementModeRegister() (MeasurementModeParams, error) {
	r := make([]byte, 1)

	if err := d.c.Tx([]byte{measurementModeReg}, r); err != nil {
		return MeasurementModeParams{}, err
	}
	mode := MeasurementMode(r[0] >> 4)
	threshold := (r[0]&4 == 1)
	interrupt := (r[0]&8 == 1)

	return MeasurementModeParams{MeasurementMode: mode, GenerateInterrupt: interrupt, UseThreshold: threshold}, nil
}

// Reset sets device into the BOOT mode.
func (d *Dev) Reset() error {
	if err := d.c.Tx([]byte{resetReg, 0x11, 0xE5, 0x72, 0x8A}, nil); err != nil {
		return err
	}
	return nil
}

// ReadStatus returns value of status register.
func (d *Dev) ReadStatus() (byte, error) {
	r := make([]byte, 1)
	if err := d.c.Tx([]byte{statusReg}, r); err != nil {
		return 0, err
	}
	return r[0], nil
}

// ReadRawData provides current and voltage on the sensor.
// Current is in range of 0-63uA. Voltage is in range 0-1.65V.
func (d *Dev) ReadRawData() (physic.ElectricCurrent, physic.ElectricPotential, error) {
	r := make([]byte, 2)
	if err := d.c.Tx([]byte{rawDataReg}, r); err != nil {
		return 0, 0, err
	}
	current, voltage := valuesFromRawData(r)
	return current, voltage, nil
}

// SetEnvironmentData allows to provide temperature and humidity so
// sensor can compensate it's measurement.
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

// GetBaseline provides current baseline used by internal measurement alogrithm.
// For better understanding how to use this value, check the SetBaseline and documentation.
func (d *Dev) GetBaseline() ([]byte, error) {
	r := make([]byte, 2)
	if err := d.c.Tx([]byte{baselineReg}, r); err != nil {
		return nil, err
	}
	return r, nil
}

// SetBaseline sets current baseline for internal measurement algorithm.
// For more details check sensor's specification.
//
// Manual Baseline Correction.
//
// There is a mechanism within CCS811 to manually save and restore a previously
// saved baseline value using the BASELINE register. The correct time to save
// the baseline will depend on the customer use-case and application.
//
// For devices which are powered for >24 hours at a time:
//
// - During the first 500 hours – save the baseline every 24-48 hours.
//
// - After the first 500 hours – save the baseline every 5-7 days.
//
// For devices which are powered <24 hours at a time:
//
// - If the device is run in, save the baseline before power down.
//
// - If multiple operating modes are used, a separate baseline should be stored for each.
//
// - The baseline should only be restored when the resistance is stable
// (typically 20-30 minutes).
//
// - If changing from a low to high power mode (without spending at least
// 10 minutes in idle), the sensor resistance should be allowed to settle again
// before restoring the baseline.
//
// Note(s):
//
// 1) If a value is written to the BASELINE register while the sensor
// is stabilising, the output of the TVOC and eCO2 calculations may be higher
// than expected.
//
// 2) The baseline must be written after the conditioning period
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
	RawDataCurrent physic.ElectricCurrent
	RawDataVoltage physic.ElectricPotential
}

// Sense provides data from the sensor.
// This function read all 8 available bytes including error, raw data etc.
// If you want just eCO2 and/or VOC, use SensePartial.
func (d *Dev) Sense(values *SensorValues) error {
	return d.SensePartial(ReadAll, values)
}

// SensePartial provides marginaly more efficient reading from the sensor.
// You can specify what subset of data you want through NeededData constants.
func (d *Dev) SensePartial(requested NeededData, values *SensorValues) error {
	read := make([]byte, requested)
	if err := d.c.Tx([]byte{algoResultsReg}, read); err != nil {
		return err
	}
	if requested >= ReadCO2 {
		// Exptected range: 400ppm to 8192ppm.
		// 0x3F is used to erase randomly set top bits,
		// causing value out of range given by specs.
		values.ECO2 = int(uint32(read[0]&0x3F)<<8 | uint32(read[1]))
	}
	if requested >= ReadCO2VOC {
		// Expected range: 0ppb to 1187ppb.
		// 0x7 is used to erase randomly set top bits
		// causing value out of range given by specs.
		values.VOC = int(uint32(read[2]&0x7)<<8 | uint32(read[3]))
	}
	if requested >= ReadCO2VOCStatus {
		values.Status = read[4]
	}
	if requested == ReadAll {
		values.ErrorID = SensorErrorID(read[5])
		values.RawDataCurrent, values.RawDataVoltage = valuesFromRawData(read[6:])
	}

	return nil
}

// Parse current and voltage from raw data.
func valuesFromRawData(data []byte) (physic.ElectricCurrent, physic.ElectricPotential) {
	c := physic.ElectricCurrent(int64(data[0]>>2) * 1000)
	sensorsVoltageUnits := int64((uint16(data[0]&0x03) << 8) | uint16(data[1]))
	// 1.65V = 1023
	// sensorsVoltageUnits is converted to V, and after that to nV.
	// 165 is used instead of 1.65 to prevent types truncation.
	p := physic.ElectricPotential((sensorsVoltageUnits * 165 * (1000 * 1000 * 1000) / 102300))
	return c, p
}

// FwVersions is a strcutre which aggregates all different versions of sensors features.
type FwVersions struct {
	HWIdentifier       byte
	HWVersion          byte
	BootVersion        string
	ApplicationVersion string
}

// GetFirmwareData populates FwVersions structure with data.
func (d *Dev) GetFirmwareData() (*FwVersions, error) {
	version := &FwVersions{}
	buffer1 := make([]byte, 1)

	if err := d.c.Tx([]byte{0x20}, buffer1); err != nil {
		return version, err
	}
	version.HWIdentifier = buffer1[0]

	if err := d.c.Tx([]byte{0x21}, buffer1); err != nil {
		return version, err
	}
	version.HWVersion = buffer1[0]

	buffer2 := make([]byte, 2)
	if err := d.c.Tx([]byte{0x23}, buffer2); err != nil {
		return version, err
	}
	minor := buffer2[0] & 0x0F
	major := (buffer2[0] & 0xF0) >> 4
	trivial := buffer2[1]
	version.BootVersion = fmt.Sprintf("%d.%d.%d", major, minor, trivial)

	if err := d.c.Tx([]byte{0x24}, buffer2); err != nil {
		return version, err
	}
	minor = buffer2[0] & 0x0F
	major = (buffer2[0] & 0xF0) >> 4
	trivial = buffer2[1]
	version.ApplicationVersion = fmt.Sprintf("%d.%d.%d", major, minor, trivial)

	return version, nil
}
