// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package mpu9250 MPU-9250 is a 9-axis MotionTracking device that combines a 3-axis gyroscope, 3-axis accelerometer, 3-axis magnetometer and a Digital Motion Processor™ (DMP)
//
// Datasheet
//
// https://www.invensense.com/wp-content/uploads/2015/02/PS-MPU-9250A-01-v1.1.pdf
// https://www.invensense.com/wp-content/uploads/2015/02/MPU-9250-Register-Map.pdf
package mpu9250

import (
	"fmt"
	"math"
	"time"

	"periph.io/x/periph/experimental/devices/mpu9250/reg"
)

const (
	slaveNumberError = "slave number 0 .. 3"
	registers        = 12
	accelsenSitivity = 16384
)

type (
	// Proto defines the low-level methods used by different transports.
	Proto interface {
		writeMaskedReg(address byte, mask byte, value byte) error
		readMaskedReg(address byte, mask byte) (byte, error)
		readByte(address byte) (byte, error)
		writeByte(address byte, value byte) error
		readUint16(address ...byte) (uint16, error)
		writeMagReg(address byte, value byte) error
	}

	// AccelerometerData the values for x/y/z axises.
	AccelerometerData struct {
		X, Y, Z int16
	}

	// GyroscopeData the values for x/y/z axises.
	GyroscopeData struct {
		X, Y, Z int16
	}

	// RotationData the rotation around X/Y/Z axises.
	RotationData struct {
		X, Y, Z int16
	}

	// Deviation defines the standard deviation for major axises.
	Deviation struct {
		X, Y, Z float64
	}

	// SelfTestResult defines the results for self-test for accelerometer, gyroscope.
	SelfTestResult struct {
		AccelDeviation Deviation
		GyroDeviation  Deviation
	}

	// MPU9250 defines the structure to keep reference to the transport.
	MPU9250 struct {
		transport Proto
		debug     func(string, ...interface{})
	}
)

// New Creates the new instance of the driver.
//
// transport the transport interface.
func New(transport Proto) (*MPU9250, error) {
	return &MPU9250{transport: transport, debug: noop}, nil
}

// Debug sets the debug logger implementation.
func (m *MPU9250) Debug(f DebugF) {
	m.debug = f
}

// Init initializes the device
func (m *MPU9250) Init() error {
	if err := m.transferBatch(initSequence, "error initializing %d: [%x:%x] => %v"); err != nil {
		return err
	}
	return nil
}

// Calibrate Calibrates the device using maximum precision for both Gyroscope and Accelerometer.
func (m *MPU9250) Calibrate() error {

	if err := m.transferBatch(calibrateSequence, "error calibrating %d: [%x:%x] => %v"); err != nil {
		return err
	}
	reads, err := m.GetFIFOCount() // read FIFO sample count
	if err != nil {
		return wrapf("can't get FIFO => %v", err)
	}

	m.debug("Read %d packets\n", reads)

	packets := reads / registers

	var buffer [registers]byte

	toUint16 := func(offset int) int16 {
		return int16(buffer[offset])<<8 | int16(buffer[offset+1])
	}

	writeGyroOffset := func(v int16, h, l byte) error {
		o := (-v) >> 2
		if err := m.transport.writeByte(h, byte(o>>8)); err != nil {
			return wrapf("can't write Gyro offset %x(h):%x => %v", h, o, err)
		}
		if err := m.transport.writeByte(l, byte(o&0xFF)); err != nil {
			return wrapf("can't write Gyro offset %x(l):%x => %v", l, o, err)
		}
		return nil
	}

	writeAccelOffset := func(o uint16, h, l byte) error {
		if err := m.transport.writeByte(h, byte(o>>8)); err != nil {
			return wrapf("can't write Accelerator %x(h):%x => %v", h, o, err)
		}
		if err := m.transport.writeByte(l, byte(o&0xFF)); err != nil {
			return wrapf("can't write Accelerator %x(l):%x => %v", l, o, err)
		}
		return nil
	}

	var (
		accelX, accelY, accelZ, gyroX, gyroY, gyroZ                         int64
		accelXBias, accelYBias, accelZBias, gyroXBias, gyroYBias, gyroZBias int16
	)

	for i := 0; i < int(packets); i++ {
		for j := 0; j < registers; j++ {
			b, err := m.GetFIFOByte()
			if err != nil {
				return wrapf("can't read data byte %d of packet %d => %v", j, i, err)
			}
			buffer[j] = b
		}
		accelX += int64(toUint16(0))
		accelY += int64(toUint16(2))
		accelZ += int64(toUint16(4))
		gyroX += int64(toUint16(6))
		gyroY += int64(toUint16(8))
		gyroZ += int64(toUint16(10))
	}

	accelXBias = int16(accelX / int64(packets))
	accelYBias = int16(accelY / int64(packets))
	accelZBias = int16(accelZ / int64(packets))
	gyroXBias = int16(gyroX / int64(packets))
	gyroYBias = int16(gyroY / int64(packets))
	gyroZBias = int16(gyroZ / int64(packets))

	m.debug("Raw accelerometer bias: X:%d, Y:%d, Z:%d\n", accelXBias, accelYBias, accelZBias)
	m.debug("Raw gyroscope bias X:%d, Y:%d, Z:%d\n", gyroXBias, gyroYBias, gyroZBias)

	if accelZBias > 0 {
		accelZBias -= accelsenSitivity
	} else {
		accelZBias += accelsenSitivity
	}

	var factoryGyroBiasX, factoryGyroBiasY, factoryGyroBiasZ int16
	factoryGyroBiasX, err = m.ReadSignedWord(reg.MPU9250_XG_OFFSET_H, reg.MPU9250_XG_OFFSET_L)
	if err != nil {
		return err
	}
	factoryGyroBiasY, err = m.ReadSignedWord(reg.MPU9250_YG_OFFSET_H, reg.MPU9250_YG_OFFSET_L)
	if err != nil {
		return err
	}
	factoryGyroBiasZ, err = m.ReadSignedWord(reg.MPU9250_ZG_OFFSET_H, reg.MPU9250_ZG_OFFSET_L)
	if err != nil {
		return err
	}
	m.debug("Factory gyroscope bias: X:%d, Y:%d, Z:%d\n", int16(factoryGyroBiasX), int16(factoryGyroBiasY), int16(factoryGyroBiasZ))

	if err := writeGyroOffset(gyroXBias, reg.MPU9250_GYRO_XOUT_H, reg.MPU9250_GYRO_XOUT_L); err != nil {
		return err
	}
	if err := writeGyroOffset(gyroYBias, reg.MPU9250_GYRO_YOUT_H, reg.MPU9250_GYRO_YOUT_L); err != nil {
		return err
	}
	if err := writeGyroOffset(gyroZBias, reg.MPU9250_GYRO_ZOUT_H, reg.MPU9250_GYRO_ZOUT_L); err != nil {
		return err
	}

	// Construct the accelerometer biases for push to the hardware accelerometer bias registers. These registers contain
	// factory trim values which must be added to the calculated accelerometer biases; on boot up these registers will hold
	// non-zero values.
	var factoryBiasX, factoryBiasY, factoryBiasZ int16
	factoryBiasX, err = m.ReadSignedWord(reg.MPU9250_XA_OFFSET_H, reg.MPU9250_XA_OFFSET_L)
	if err != nil {
		return err
	}
	factoryBiasY, err = m.ReadSignedWord(reg.MPU9250_YA_OFFSET_H, reg.MPU9250_YA_OFFSET_L)
	if err != nil {
		return err
	}
	factoryBiasZ, err = m.ReadSignedWord(reg.MPU9250_ZA_OFFSET_H, reg.MPU9250_ZA_OFFSET_L)
	if err != nil {
		return err
	}

	// In addition, bit 0 of the lower byte must be preserved since it is used for temperature
	// compensation calculations.
	maskX := factoryBiasX & 1
	maskY := factoryBiasY & 1
	maskZ := factoryBiasZ & 1

	m.debug("Factory accelerometer bias: X:%d, Y:%d, Z:%d\n", int16(factoryBiasX), int16(factoryBiasY), int16(factoryBiasZ))

	// Accelerometer bias registers expect bias input as 2048 LSB per g, so that
	// the accelerometer biases calculated above must be divided by 8.
	factoryBiasX -= accelXBias >> 3
	factoryBiasY -= accelYBias >> 3
	factoryBiasZ -= accelZBias >> 3

	// restore the temperature preserve bit
	factoryBiasX |= maskX
	factoryBiasY |= maskY
	factoryBiasZ |= maskZ

	if err := writeAccelOffset(uint16(factoryBiasX), reg.MPU9250_XA_OFFSET_H, reg.MPU9250_XA_OFFSET_L); err != nil {
		return err
	}
	if err := writeAccelOffset(uint16(factoryBiasY), reg.MPU9250_YA_OFFSET_H, reg.MPU9250_YA_OFFSET_L); err != nil {
		return err
	}
	return writeAccelOffset(uint16(factoryBiasZ), reg.MPU9250_ZA_OFFSET_H, reg.MPU9250_ZA_OFFSET_L)
}

var selftTestSequence = [][]byte{
	{reg.MPU9250_SMPLRT_DIV, 0x00},
	{reg.MPU9250_CONFIG, 0x02},
	{reg.MPU9250_PWR_MGMT_2, 0x00},
	{reg.MPU9250_GYRO_CONFIG, 0x00},
	{reg.MPU9250_ACCEL_CONFIG2, 0x02},
	{reg.MPU9250_ACCEL_CONFIG, 0x00},
}

// SelfTest runs the self test on the device.
//
//	returns the accelerator and gyroscope deviations from the factory defaults.
func (m *MPU9250) SelfTest() (*SelfTestResult, error) {
	if err := m.transferBatch(selftTestSequence, "error initializing self-test sequence %d: [%x:%x] => %v"); err != nil {
		return nil, err
	}
	const iterations = 200
	var (
		avgAccelX, avgAccelY, avgAccelZ       int32
		avgGyroX, avgGyroY, avgGyroZ          int32
		avgSTAccelX, avgSTAccelY, avgSTAccelZ int32
		avgSTGyroX, avgSTGyroY, avgSTGyroZ    int32
		stGyroX, stGyroY, stGyroZ             byte
		stAccelX, stAccelY, stAccelZ          byte
	)

	add := func(f func() (int16, error), dest *int32) error {
		v, err := f()
		if err != nil {
			return err
		}
		*dest += int32(v)
		return nil
	}

	collectData := func(ax, ay, az, gx, gy, gz *int32) error {
		for i := 0; i < iterations; i++ {
			if err := add(m.GetAccelerationX, ax); err != nil {
				return err
			}
			if err := add(m.GetAccelerationY, ay); err != nil {
				return err
			}
			if err := add(m.GetAccelerationZ, az); err != nil {
				return err
			}
			if err := add(m.GetRotationX, gx); err != nil {
				return err
			}
			if err := add(m.GetRotationY, gy); err != nil {
				return err
			}
			if err := add(m.GetRotationZ, gz); err != nil {
				return err
			}
		}
		*ax /= iterations
		*ay /= iterations
		*az /= iterations
		*gx /= iterations
		*gy /= iterations
		*gz /= iterations
		return nil
	}

	// Collect the average measurement for the registers data
	if err := collectData(&avgAccelX, &avgAccelY, &avgAccelZ, &avgGyroX, &avgGyroY, &avgGyroZ); err != nil {
		return nil, wrapf("selftest: error reading register data %v", err)
	}

	m.debug("Avg accelerometer: X: %d, Y: %d, Z: %d\n    gyroscope:     X: %d, Y: %d, Z: %d\n", avgAccelX, avgAccelY, avgAccelZ, avgGyroX, avgGyroY, avgGyroZ)

	// Collect the self-test data
	if err := m.transport.writeByte(reg.MPU9250_ACCEL_CONFIG, 0xe0); err != nil {
		return nil, wrapf("selftest: error setting selftest for accelerometer: %v", err)
	}
	if err := m.transport.writeByte(reg.MPU9250_GYRO_CONFIG, 0xe0); err != nil {
		return nil, wrapf("selftest: error setting selftest for gyroscope: %v", err)
	}
	if err := collectData(&avgSTAccelX, &avgSTAccelY, &avgSTAccelZ, &avgSTGyroX, &avgSTGyroY, &avgSTGyroZ); err != nil {
		return nil, wrapf("selftest: error reading self-test register data %v", err)
	}
	m.debug("Avg Trim accelerometer : X: %d, Y: %d, Z: %d\n         gyroscope:      X: %d, Y: %d, Z: %d\n", avgSTAccelX, avgSTAccelY, avgSTAccelZ, avgSTGyroX, avgSTGyroY, avgSTGyroZ)
	if err := m.transport.writeByte(reg.MPU9250_ACCEL_CONFIG, 0x00); err != nil {
		return nil, wrapf("selftest: error resetting accelerometer: %v", err)
	}
	if err := m.transport.writeByte(reg.MPU9250_GYRO_CONFIG, 0x00); err != nil {
		return nil, wrapf("selftest: error resetting gyroscope: %v", err)
	}
	time.Sleep(25 * time.Millisecond)

	// get the self-test trim values
	if stX, err := m.transport.readByte(reg.MPU9250_SELF_TEST_X_ACCEL); err == nil {
		stAccelX = stX
	} else {
		return nil, wrapf("selftest: error getting self-test accelerometer value X: %v", err)
	}
	if stY, err := m.transport.readByte(reg.MPU9250_SELF_TEST_Y_ACCEL); err == nil {
		stAccelY = stY
	} else {
		return nil, wrapf("selftest: error getting self-test accelerometer value Y: %v", err)
	}
	if stZ, err := m.transport.readByte(reg.MPU9250_SELF_TEST_Z_ACCEL); err == nil {
		stAccelZ = stZ
	} else {
		return nil, wrapf("selftest: error getting self-test accelerometer value Z: %v", err)
	}
	if gyX, err := m.transport.readByte(reg.MPU9250_SELF_TEST_X_GYRO); err == nil {
		stGyroX = gyX
	} else {
		return nil, wrapf("selftest: error getting self-test gyroscope value X: %v", err)
	}
	if gyY, err := m.transport.readByte(reg.MPU9250_SELF_TEST_Y_GYRO); err == nil {
		stGyroY = gyY
	} else {
		return nil, wrapf("selftest: error getting self-test gyroscope value Y: %v", err)
	}
	if gyZ, err := m.transport.readByte(reg.MPU9250_SELF_TEST_Z_GYRO); err == nil {
		stGyroZ = gyZ
	} else {
		return nil, wrapf("selftest: error getting self-test gyroscope value Z: %v", err)
	}

	m.debug("Self-Test accelerometer: X: %d, Y: %d, Z: %d\n          gyroscope:     X: %d, Y: %d, Z: %d\n", stAccelX, stAccelY, stAccelZ, stGyroX, stGyroY, stGyroZ)

	deviation := func(aSTAvg, aAvg int32, stV byte) float64 {
		factoryTrim := 2620.0 * (math.Pow(1.01, float64(stV)-1.0))
		return math.Abs(100.0*float64(aSTAvg-aAvg)/factoryTrim) - 100.0
	}

	return &SelfTestResult{
		AccelDeviation: Deviation{
			X: deviation(avgAccelX, avgSTAccelX, stAccelX),
			Y: deviation(avgAccelY, avgSTAccelY, stAccelY),
			Z: deviation(avgAccelZ, avgSTAccelZ, stAccelZ),
		},
		GyroDeviation: Deviation{
			X: deviation(avgGyroX, avgSTGyroX, stGyroX),
			Y: deviation(avgGyroY, avgSTGyroY, stGyroY),
			Z: deviation(avgGyroZ, avgSTGyroZ, stGyroZ),
		},
	}, nil
}

// SetClockSource Set clock source setting.
//
// An internal 8MHz oscillator, gyroscope based clock, or external sources can
// be selected as the MPU-60X0 clock source. When the internal 8 MHz oscillator
// or an external source is chosen as the clock source, the MPU-60X0 can operate
// in low power modes with the gyroscopes disabled.
//
// Upon power up, the MPU-60X0 clock source defaults to the internal oscillator.
// However, it is highly recommended that the device be configured to use one of
// the gyroscopes (or an external clock source) as the clock reference for
// improved stability. The clock source can be selected according to the following table:
//
// CLK_SEL | Clock Source
// --------+--------------------------------------
// 0       | Internal oscillator
// 1       | PLL with X Gyro reference
// 2       | PLL with Y Gyro reference
// 3       | PLL with Z Gyro reference
// 4       | PLL with external 32.768kHz reference
// 5       | PLL with external 19.2MHz reference
// 6       | Reserved
// 7       | Stops the clock and keeps the timing generator in reset
func (m *MPU9250) SetClockSource(src byte) error {
	if src > 7 {
		return wrapf("clock should be in range 0 .. 7")
	}
	return m.transport.writeMaskedReg(reg.MPU9250_PWR_MGMT_1, reg.MPU9250_CLKSEL_MASK, src)
}

// SetGyroRange Set gyroscope range.
//
// range New gyroscope range value
// the mask is 11000
func (m *MPU9250) SetGyroRange(rangeVal byte) error {
	if rangeVal > 3 {
		return wrapf("accepted values are in the range 0 .. 3")
	}
	return m.transport.writeMaskedReg(reg.MPU9250_GYRO_CONFIG, reg.MPU9250_GYRO_FS_SEL_MASK, rangeVal<<3)
}

// GetGyroRange Get gyroscope range.
//
// The FS_SEL parameter allows setting the full-scale range of the gyro sensors,
// as described in the table below.
//
// 0 = +/- 250 degrees/sec
// 1 = +/- 500 degrees/sec
// 2 = +/- 1000 degrees/sec
// 3 = +/- 2000 degrees/sec
//
func (m *MPU9250) GetGyroRange() (byte, error) {
	b, err := m.transport.readMaskedReg(reg.MPU9250_GYRO_CONFIG, reg.MPU9250_GYRO_FS_SEL_MASK)
	return b >> 3, err // the mask is 11000
}

// SetAccelRange Set full-scale accelerometer range.
//
// The masked value is 11000
//
// range New full-scale accelerometer range setting
func (m *MPU9250) SetAccelRange(rangeVal byte) error {
	if (rangeVal >> 3) > 3 {
		return wrapf("accepted values are in the range 0 .. 3")
	}
	return m.transport.writeMaskedReg(reg.MPU9250_ACCEL_CONFIG, reg.MPU9250_ACCEL_FS_SEL_MASK, rangeVal)
}

// GetAccelRange Get full-scale accelerometer range.
//
// The FS_SEL parameter allows setting the full-scale range of the accelerometer
// sensors, as described in the table below.
//
// 0 = +/- 2g
// 1 = +/- 4g
// 2 = +/- 8g
// 3 = +/- 16g
func (m *MPU9250) GetAccelRange() (byte, error) {
	b, err := m.transport.readMaskedReg(reg.MPU9250_ACCEL_CONFIG, reg.MPU9250_ACCEL_FS_SEL_MASK)
	return b >> 3, err
}

// GetTempFIFOEnabled Get temperature FIFO enabled value.
//
// When set to 1, this bit enables TEMP_OUT_H and TEMP_OUT_L (Registers 65 and
// 66) to be written into the FIFO buffer.
func (m *MPU9250) GetTempFIFOEnabled() (bool, error) {
	response, err := m.transport.readMaskedReg(reg.MPU9250_FIFO_EN, reg.MPU9250_TEMP_FIFO_EN_MASK)
	if err != nil {
		return false, err
	}
	return response != 0, nil
}

// SetTempFIFOEnabled Set temperature FIFO enabled value.
//
// 	enabled New temperature FIFO enabled value
func (m *MPU9250) SetTempFIFOEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_FIFO_EN, reg.MPU9250_TEMP_FIFO_EN_MASK, boolByte(enabled, reg.MPU9250_TEMP_FIFO_EN_MASK))
}

// GetXGyroFIFOEnabled Get gyroscope X-axis FIFO enabled value.
//
// When set to 1, this bit enables GYRO_XOUT_H and GYRO_XOUT_L (Registers 67 and
// 68) to be written into the FIFO buffer.
func (m *MPU9250) GetXGyroFIFOEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_FIFO_EN, reg.MPU9250_GYRO_XOUT_MASK)
}

// SetXGyroFIFOEnabled Set gyroscope X-axis FIFO enabled value.
//
//	enabled New gyroscope X-axis FIFO enabled value
func (m *MPU9250) SetXGyroFIFOEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_FIFO_EN, reg.MPU9250_GYRO_XOUT_MASK, boolByte(enabled, reg.MPU9250_GYRO_XOUT_MASK))
}

// GetYGyroFIFOEnabled Get gyroscope Y-axis FIFO enabled value.
//
// When set to 1, this bit enables GYRO_YOUT_H and GYRO_YOUT_L (Registers 69 and
// 70) to be written into the FIFO buffer.
// @return Current gyroscope Y-axis FIFO enabled value
// @see MPU9250_RA_FIFO_EN
///
func (m *MPU9250) GetYGyroFIFOEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_FIFO_EN, reg.MPU9250_GYRO_YOUT_MASK)
}

// SetYGyroFIFOEnabled Set gyroscope Y-axis FIFO enabled value.
//
//	enabled New gyroscope Y-axis FIFO enabled value
func (m *MPU9250) SetYGyroFIFOEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_FIFO_EN, reg.MPU9250_GYRO_YOUT_MASK, boolByte(enabled, reg.MPU9250_GYRO_YOUT_MASK))
}

// GetZGyroFIFOEnabled Get gyroscope Z-axis FIFO enabled value.
//
// When set to 1, this bit enables GYRO_ZOUT_H and GYRO_ZOUT_L (Registers 71 and
// 72) to be written into the FIFO buffer.
func (m *MPU9250) GetZGyroFIFOEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_FIFO_EN, reg.MPU9250_GYRO_ZOUT_MASK)
}

// SetZGyroFIFOEnabled Set gyroscope Z-axis FIFO enabled value.
//
//	enabled New gyroscope Z-axis FIFO enabled value
func (m *MPU9250) SetZGyroFIFOEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_FIFO_EN, reg.MPU9250_GYRO_ZOUT_MASK, boolByte(enabled, reg.MPU9250_GYRO_ZOUT_MASK))
}

// GetAccelFIFOEnabled Get accelerometer FIFO enabled value.
//
// When set to 1, this bit enables ACCEL_XOUT_H, ACCEL_XOUT_L, ACCEL_YOUT_H,
// ACCEL_YOUT_L, ACCEL_ZOUT_H, and ACCEL_ZOUT_L (Registers 59 to 64) to be
// written into the FIFO buffer.
func (m *MPU9250) GetAccelFIFOEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_FIFO_EN, reg.MPU9250_ACCEL_MASK)
}

// SetAccelFIFOEnabled Set accelerometer FIFO enabled value.
//
//	enabled New accelerometer FIFO enabled value
func (m *MPU9250) SetAccelFIFOEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_FIFO_EN, reg.MPU9250_ACCEL_MASK, boolByte(enabled, reg.MPU9250_ACCEL_MASK))
}

// GetSlave2FIFOEnabled Get Slave 2 FIFO enabled value.
//
// When set to 1, this bit enables EXT_SENS_DATA registers (Registers 73 to 96)
// associated with Slave 2 to be written into the FIFO buffer.
func (m *MPU9250) GetSlave2FIFOEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_FIFO_EN, reg.MPU9250_SLV2_MASK)
}

// SetSlave2FIFOEnabled Set Slave 2 FIFO enabled value.
//
//	enabled New Slave 2 FIFO enabled value
func (m *MPU9250) SetSlave2FIFOEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_FIFO_EN, reg.MPU9250_SLV2_MASK, boolByte(enabled, reg.MPU9250_SLV2_MASK))
}

// GetSlave1FIFOEnabled Get Slave 1 FIFO enabled value.
//
// When set to 1, this bit enables EXT_SENS_DATA registers (Registers 73 to 96)
// associated with Slave 1 to be written into the FIFO buffer.
func (m *MPU9250) GetSlave1FIFOEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_FIFO_EN, reg.MPU9250_SLV1_MASK)
}

// SetSlave1FIFOEnabled Set Slave 1 FIFO enabled value.
//
//	enabled New Slave 1 FIFO enabled value
func (m *MPU9250) SetSlave1FIFOEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_FIFO_EN, reg.MPU9250_SLV1_MASK, boolByte(enabled, reg.MPU9250_SLV1_MASK))
}

// GetSlave0FIFOEnabled Get Slave 0 FIFO enabled value.
//
// When set to 1, this bit enables EXT_SENS_DATA registers (Registers 73 to 96)
// associated with Slave 0 to be written into the FIFO buffer.
func (m *MPU9250) GetSlave0FIFOEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_FIFO_EN, reg.MPU9250_SLV0_MASK)
}

// SetSlave0FIFOEnabled Set Slave 0 FIFO enabled value.
//
//	enabled New Slave 0 FIFO enabled value
func (m *MPU9250) SetSlave0FIFOEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_FIFO_EN, reg.MPU9250_SLV0_MASK, boolByte(enabled, reg.MPU9250_SLV0_MASK))
}

// I2C_MST_CTRL register

// GetMultiMasterEnabled Get multi-master enabled value.
//
// Multi-master capability allows multiple I2C masters to operate on the same
// bus. In circuits where multi-master capability is required, set MULT_MST_EN
// to 1. This will increase current drawn by approximately 30uA.
//
// In circuits where multi-master capability is required, the state of the I2C
// bus must always be monitored by each separate I2C Master. Before an I2C
// Master can assume arbitration of the bus, it must first confirm that no other
// I2C Master has arbitration of the bus. When MULT_MST_EN is set to 1, the
// MPU-60X0's bus arbitration detection logic is turned on, enabling it to
// detect when the bus is available.
func (m *MPU9250) GetMultiMasterEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_MST_CTRL, reg.MPU9250_MULT_MST_EN_MASK)
}

// SetMultiMasterEnabled Set multi-master enabled value.
//
//	enabled New multi-master enabled value
func (m *MPU9250) SetMultiMasterEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_I2C_MST_CTRL, reg.MPU9250_MULT_MST_EN_MASK, boolByte(enabled, reg.MPU9250_MULT_MST_EN_MASK))
}

// GetWaitForExternalSensorEnabled Get wait-for-external-sensor-data enabled value.
//
// When the WAIT_FOR_ES bit is set to 1, the Data Ready interrupt will be
// delayed until External Sensor data from the Slave Devices are loaded into the
// EXT_SENS_DATA registers. This is used to ensure that both the internal sensor
// data (i.e. from gyro and accel) and external sensor data have been loaded to
// their respective data registers (i.e. the data is synced) when the Data Ready
// interrupt is triggered.
func (m *MPU9250) GetWaitForExternalSensorEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_MST_CTRL, reg.MPU9250_WAIT_FOR_ES_MASK)
}

// SetWaitForExternalSensorEnabled Set wait-for-external-sensor-data enabled value.
//
//	enabled New wait-for-external-sensor-data enabled value
func (m *MPU9250) SetWaitForExternalSensorEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_I2C_MST_CTRL, reg.MPU9250_WAIT_FOR_ES_MASK, boolByte(enabled, reg.MPU9250_WAIT_FOR_ES_MASK))
}

// GetSlave3FIFOEnabled Get Slave 3 FIFO enabled value.
//
// When set to 1, this bit enables EXT_SENS_DATA registers (Registers 73 to 96)
// associated with Slave 3 to be written into the FIFO buffer.
func (m *MPU9250) GetSlave3FIFOEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_MST_CTRL, reg.MPU9250_SLV_3_FIFO_EN_MASK)
}

// SetSlave3FIFOEnabled Set Slave 3 FIFO enabled value.
//
//	enabled New Slave 3 FIFO enabled value
func (m *MPU9250) SetSlave3FIFOEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_I2C_MST_CTRL, reg.MPU9250_SLV_3_FIFO_EN_MASK, boolByte(enabled, reg.MPU9250_SLV_3_FIFO_EN_MASK))
}

// GetSlaveReadWriteTransitionEnabled Get slave read/write transition enabled value.
//
// The I2C_MST_P_NSR bit configures the I2C Master's transition from one slave
// read to the next slave read. If the bit equals 0, there will be a restart
// between reads. If the bit equals 1, there will be a stop followed by a start
// of the following read. When a write transaction follows a read transaction,
// the stop followed by a start of the successive write will be always used.
func (m *MPU9250) GetSlaveReadWriteTransitionEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_MST_CTRL, reg.MPU9250_I2C_MST_P_NSR_MASK)
}

// SetSlaveReadWriteTransitionEnabled Set slave read/write transition enabled value.
//
//	enabled New slave read/write transition enabled value
func (m *MPU9250) SetSlaveReadWriteTransitionEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_I2C_MST_CTRL, reg.MPU9250_I2C_MST_P_NSR_MASK, boolByte(enabled, reg.MPU9250_I2C_MST_P_NSR_MASK))
}

// GetMasterClockSpeed Get I2C master clock speed.
//
// I2C_MST_CLK is a 4 bit unsigned value which configures a divider on the
// MPU-60X0 internal 8MHz clock. It sets the I2C master clock speed according to
// the following table:
//
// I2C_MST_CLK | I2C Master Clock Speed | 8MHz Clock Divider
// ------------+------------------------+-------------------
// 0           | 348kHz                 | 23
// 1           | 333kHz                 | 24
// 2           | 320kHz                 | 25
// 3           | 308kHz                 | 26
// 4           | 296kHz                 | 27
// 5           | 286kHz                 | 28
// 6           | 276kHz                 | 29
// 7           | 267kHz                 | 30
// 8           | 258kHz                 | 31
// 9           | 500kHz                 | 16
// 10          | 471kHz                 | 17
// 11          | 444kHz                 | 18
// 12          | 421kHz                 | 19
// 13          | 400kHz                 | 20
// 14          | 381kHz                 | 21
// 15          | 364kHz                 | 22
func (m *MPU9250) GetMasterClockSpeed() (byte, error) {
	return m.transport.readMaskedReg(reg.MPU9250_I2C_MST_CTRL, reg.MPU9250_I2C_MST_CLK_MASK)
}

// SetMasterClockSpeed Set I2C master clock speed.
//
//	speed Current I2C master clock speed
func (m *MPU9250) SetMasterClockSpeed(speed byte) error {
	if speed > 15 {
		return wrapf("speed range 0 .. 15")
	}
	return m.transport.writeMaskedReg(reg.MPU9250_I2C_MST_CTRL, reg.MPU9250_I2C_MST_CLK_MASK, speed)
}

// I2C_SLV* registers (Slave 0-3)

// GetSlaveAddress Get the I2C address of the specified slave (0-3).
//
// Note that Bit 7 (MSB) controls read/write mode. If Bit 7 is set, it's a read
// operation, and if it is cleared, then it's a write operation. The remaining
// bits (6-0) are the 7-bit device address of the slave device.
//
// In read mode, the result of the read is placed in the lowest available
// EXT_SENS_DATA register. For further information regarding the allocation of
// read results, please refer to the EXT_SENS_DATA register description
// (Registers 73 - 96).
//
// The MPU-6050 supports a total of five slaves, but Slave 4 has unique
// characteristics, and so it has its own functions (getSlave4* and setSlave4*).
//
// I2C data transactions are performed at the Sample Rate, as defined in
// Register 25. The user is responsible for ensuring that I2C data transactions
// to and from each enabled Slave can be completed within a single period of the
// Sample Rate.
//
// The I2C slave access rate can be reduced relative to the Sample Rate. This
// reduced access rate is determined by I2C_MST_DLY (Register 52). Whether a
// slave's access rate is reduced relative to the Sample Rate is determined by
// I2C_MST_DELAY_CTRL (Register 103).
//
// The processing order for the slaves is fixed. The sequence followed for
// processing the slaves is Slave 0, Slave 1, Slave 2, Slave 3 and Slave 4. If a
// particular Slave is disabled it will be skipped.
//
// Each slave can either be accessed at the sample rate or at a reduced sample
// rate. In a case where some slaves are accessed at the Sample Rate and some
// slaves are accessed at the reduced rate, the sequence of accessing the slaves
// (Slave 0 to Slave 4) is still followed. However, the reduced rate slaves will
// be skipped if their access rate dictates that they should not be accessed
// during that particular cycle. For further information regarding the reduced
// access rate, please refer to Register 52. Whether a slave is accessed at the
// Sample Rate or at the reduced rate is determined by the Delay Enable bits in
// Register 103.
func (m *MPU9250) GetSlaveAddress(num byte) (byte, error) {
	if num > 3 {
		return 0, wrapf(slaveNumberError)
	}
	return m.transport.readByte(reg.MPU9250_I2C_SLV0_ADDR + num*3)
}

// SetSlaveAddress Set the I2C address of the specified slave (0-3).
//
//	num Slave number (0-3)
func (m *MPU9250) SetSlaveAddress(num, address byte) error {
	if num > 3 {
		return wrapf(slaveNumberError)
	}
	return m.transport.writeByte(reg.MPU9250_I2C_SLV0_ADDR+num*3, address)
}

// GetSlaveRegister Get the active internal register for the specified slave (0-3).
//
// Read/write operations for this slave will be done to whatever internal
// register address is stored in this MPU register.
//
// The MPU-6050 supports a total of five slaves, but Slave 4 has unique
// characteristics, and so it has its own functions.
//
//	num Slave number (0-3)
func (m *MPU9250) GetSlaveRegister(num byte) (byte, error) {
	if num > 3 {
		return 0, wrapf(slaveNumberError)
	}
	return m.transport.readByte(reg.MPU9250_I2C_SLV0_REG + num*3)
}

// SetSlaveRegister Set the active internal register for the specified slave (0-3).
//
//	num Slave number (0-3)
//	reg New active register for specified slave
func (m *MPU9250) SetSlaveRegister(num, r byte) error {
	if num > 3 {
		return wrapf(slaveNumberError)
	}
	return m.transport.writeByte(reg.MPU9250_I2C_SLV0_REG+num*3, r)
}

// GetSlaveEnabled Get the enabled value for the specified slave (0-3).
//
// When set to 1, this bit enables Slave 0 for data transfer operations. When
// cleared to 0, this bit disables Slave 0 from data transfer operations.
// @param num Slave number (0-3)
func (m *MPU9250) GetSlaveEnabled(num byte) (bool, error) {
	if num > 3 {
		return false, wrapf(slaveNumberError)
	}
	return m.readMaskedRegBool(reg.MPU9250_I2C_SLV0_CTRL+num*3, reg.MPU9250_I2C_SLV0_EN_MASK)
}

// SetSlaveEnabled Set the enabled value for the specified slave (0-3).
//
//	num Slave number (0-3)
//	enabled New enabled value for specified slave
func (m *MPU9250) SetSlaveEnabled(num byte, enabled bool) error {
	if num > 3 {
		return wrapf(slaveNumberError)
	}
	return m.transport.writeMaskedReg(reg.MPU9250_I2C_SLV0_CTRL+num*3, reg.MPU9250_I2C_SLV0_EN_MASK, boolByte(enabled, reg.MPU9250_I2C_SLV0_EN_MASK))
}

// GetSlaveWordByteSwap Get word pair byte-swapping enabled for the specified slave (0-3).
//
// When set to 1, this bit enables byte swapping. When byte swapping is enabled,
// the high and low bytes of a word pair are swapped. Please refer to
// I2C_SLV0_GRP for the pairing convention of the word pairs. When cleared to 0,
// bytes transferred to and from Slave 0 will be written to EXT_SENS_DATA
// registers in the order they were transferred.
//
//	num Slave number (0-3)
func (m *MPU9250) GetSlaveWordByteSwap(num byte) (bool, error) {
	if num > 3 {
		return false, wrapf(slaveNumberError)
	}
	return m.readMaskedRegBool(reg.MPU9250_I2C_SLV0_CTRL+num*3, reg.MPU9250_I2C_SLV0_BYTE_SW_MASK)
}

// SetSlaveWordByteSwap Set word pair byte-swapping enabled for the specified slave (0-3).
//
//	num Slave number (0-3)
//	enabled New word pair byte-swapping enabled value for specified slave
func (m *MPU9250) SetSlaveWordByteSwap(num byte, enabled bool) error {
	if num > 3 {
		return wrapf(slaveNumberError)
	}
	return m.transport.writeMaskedReg(reg.MPU9250_I2C_SLV0_CTRL+num*3, reg.MPU9250_I2C_SLV0_BYTE_SW_MASK, boolByte(enabled, reg.MPU9250_I2C_SLV0_BYTE_SW_MASK))
}

// GetSlaveWriteMode Get write mode for the specified slave (0-3).
//
// When set to 1, the transaction will read or write data only. When cleared to
// 0, the transaction will write a register address prior to reading or writing
// data. This should equal 0 when specifying the register address within the
// Slave device to/from which the ensuing data transaction will take place.
//
//	num Slave number (0-3)
func (m *MPU9250) GetSlaveWriteMode(num byte) (bool, error) {
	if num > 3 {
		return false, wrapf(slaveNumberError)
	}
	return m.readMaskedRegBool(reg.MPU9250_I2C_SLV0_CTRL+num*3, reg.MPU9250_I2C_SLV0_REG_DIS_MASK)
}

// SetSlaveWriteMode Set write mode for the specified slave (0-3).
//
//	num Slave number (0-3)
//	mode New write mode for specified slave (0 = register address + data, 1 = data only)
func (m *MPU9250) SetSlaveWriteMode(num byte, mode bool) error {
	if num > 3 {
		return wrapf(slaveNumberError)
	}
	return m.transport.writeMaskedReg(reg.MPU9250_I2C_SLV0_CTRL+num*3, reg.MPU9250_I2C_SLV0_REG_DIS_MASK, boolByte(mode, reg.MPU9250_I2C_SLV0_REG_DIS_MASK))
}

// GetSlaveWordGroupOffset Get word pair grouping order offset for the specified slave (0-3).
//
// This sets specifies the grouping order of word pairs received from registers.
// When cleared to 0, bytes from register addresses 0 and 1, 2 and 3, etc (even,
// then odd register addresses) are paired to form a word. When set to 1, bytes
// from register addresses are paired 1 and 2, 3 and 4, etc. (odd, then even
// register addresses) are paired to form a word.
//
//	num Slave number (0-3)
//	Current word pair grouping order offset for specified slave
func (m *MPU9250) GetSlaveWordGroupOffset(num byte) (bool, error) {
	if num > 3 {
		return false, wrapf(slaveNumberError)
	}
	return m.readMaskedRegBool(reg.MPU9250_I2C_SLV0_CTRL+num*3, reg.MPU9250_I2C_SLV0_GRP_MASK)
}

// SetSlaveWordGroupOffset Set word pair grouping order offset for the specified slave (0-3).
//
//	num Slave number (0-3)
//	enabled New word pair grouping order offset for specified slave
func (m *MPU9250) SetSlaveWordGroupOffset(num byte, enabled bool) error {
	if num > 3 {
		return wrapf(slaveNumberError)
	}
	return m.transport.writeMaskedReg(reg.MPU9250_I2C_SLV0_CTRL+num*3, reg.MPU9250_I2C_SLV0_GRP_MASK, boolByte(enabled, reg.MPU9250_I2C_SLV0_GRP_MASK))
}

// GetSlaveDataLength Get number of bytes to read for the specified slave (0-3).
//
// Specifies the number of bytes transferred to and from Slave 0. Clearing this
// bit to 0 is equivalent to disabling the register by writing 0 to I2C_SLV0_EN.
//
//	num Slave number (0-3)
func (m *MPU9250) GetSlaveDataLength(num byte) (byte, error) {
	if num > 3 {
		return 0, wrapf(slaveNumberError)
	}
	return m.transport.readMaskedReg(reg.MPU9250_I2C_SLV0_CTRL+num*3, reg.MPU9250_I2C_SLV0_LENG_MASK)
}

// SetSlaveDataLength Set number of bytes to read for the specified slave (0-3).
//
//	num Slave number (0-3)
//	length Number of bytes to read for specified slave
func (m *MPU9250) SetSlaveDataLength(num byte, length byte) error {
	if num > 3 {
		return wrapf(slaveNumberError)
	}
	return m.transport.writeMaskedReg(reg.MPU9250_I2C_SLV0_CTRL+num*3, reg.MPU9250_I2C_SLV0_LENG_MASK, length)
}

// I2C_SLV* registers (Slave 4)

// GetSlave4Address Get the I2C address of Slave 4.
//
// Note that Bit 7 (MSB) controls read/write mode. If Bit 7 is set, it's a read
// operation, and if it is cleared, then it's a write operation. The remaining
// bits (6-0) are the 7-bit device address of the slave device.
func (m *MPU9250) GetSlave4Address() (byte, error) {
	return m.transport.readByte(reg.MPU9250_I2C_SLV4_ADDR)
}

// SetSlave4Address Set the I2C address of Slave 4.
//
//	address New address for Slave 4
func (m *MPU9250) SetSlave4Address(address byte) error {
	return m.transport.writeByte(reg.MPU9250_I2C_SLV4_ADDR, address)
}

// GetSlave4Register Get the active internal register for the Slave 4.
//
// Read/write operations for this slave will be done to whatever internal
// register address is stored in this MPU register.
func (m *MPU9250) GetSlave4Register() (byte, error) {
	return m.transport.readByte(reg.MPU9250_I2C_SLV4_REG)
}

// SetSlave4Register Set the active internal register for Slave 4.
//	reg New active register for Slave 4
func (m *MPU9250) SetSlave4Register(r byte) error {
	return m.transport.writeByte(reg.MPU9250_I2C_SLV4_REG, r)
}

// SetSlave4OutputByte Set new byte to write to Slave 4.
//
// This register stores the data to be written into the Slave 4. If I2C_SLV4_RW
// is set 1 (set to read), this register has no effect.
//
//	data New byte to write to Slave 4
func (m *MPU9250) SetSlave4OutputByte(data byte) error {
	return m.transport.writeByte(reg.MPU9250_I2C_SLV4_DO, data)
}

// GetSlave4Enabled Get the enabled value for the Slave 4.
//
// When set to 1, this bit enables Slave 4 for data transfer operations. When
// cleared to 0, this bit disables Slave 4 from data transfer operations.
func (m *MPU9250) GetSlave4Enabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_SLV4_CTRL, reg.MPU9250_I2C_SLV4_EN_MASK)
}

// SetSlave4Enabled Set the enabled value for Slave 4.
//
//	enabled New enabled value for Slave 4
func (m *MPU9250) SetSlave4Enabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_I2C_SLV4_CTRL, reg.MPU9250_I2C_SLV4_EN_MASK, boolByte(enabled, reg.MPU9250_I2C_SLV4_EN_MASK))
}

// GetSlave4InterruptEnabled Get the enabled value for Slave 4 transaction interrupts.
//
// When set to 1, this bit enables the generation of an interrupt signal upon
// completion of a Slave 4 transaction. When cleared to 0, this bit disables the
// generation of an interrupt signal upon completion of a Slave 4 transaction.
// The interrupt status can be observed in Register 54.
func (m *MPU9250) GetSlave4InterruptEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_SLV4_CTRL, reg.MPU9250_SLV4_DONE_INT_EN_MASK)
}

// SetSlave4InterruptEnabled Set the enabled value for Slave 4 transaction interrupts.
//
//	enabled New enabled value for Slave 4 transaction interrupts.
func (m *MPU9250) SetSlave4InterruptEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_I2C_SLV4_CTRL, reg.MPU9250_SLV4_DONE_INT_EN_MASK, boolByte(enabled, reg.MPU9250_SLV4_DONE_INT_EN_MASK))
}

// GetSlave4WriteMode Get write mode for Slave 4.
//
// When set to 1, the transaction will read or write data only. When cleared to
// 0, the transaction will write a register address prior to reading or writing
// data. This should equal 0 when specifying the register address within the
// Slave device to/from which the ensuing data transaction will take place.
func (m *MPU9250) GetSlave4WriteMode() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_SLV4_CTRL, reg.MPU9250_I2C_SLV4_REG_DIS_MASK)
}

// SetSlave4WriteMode Set write mode for the Slave 4.
//	mode New write mode for Slave 4 (0 = register address + data, 1 = data only)
func (m *MPU9250) SetSlave4WriteMode(mode bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_I2C_SLV4_CTRL, reg.MPU9250_I2C_SLV4_REG_DIS_MASK, boolByte(mode, reg.MPU9250_I2C_SLV4_REG_DIS_MASK))
}

// GetSlave4MasterDelay Get Slave 4 master delay value.
//
// This configures the reduced access rate of I2C slaves relative to the Sample
// Rate. When a slave's access rate is decreased relative to the Sample Rate,
// the slave is accessed every:
//
//     1 / (1 + I2C_MST_DLY) samples
//
// This base Sample Rate in turn is determined by SMPLRT_DIV (register 25) and
// DLPF_CFG (register 26). Whether a slave's access rate is reduced relative to
// the Sample Rate is determined by I2C_MST_DELAY_CTRL (register 103). For
// further information regarding the Sample Rate, please refer to register 25.
func (m *MPU9250) GetSlave4MasterDelay() (byte, error) {
	return m.transport.readMaskedReg(reg.MPU9250_I2C_SLV4_CTRL, reg.MPU9250_I2C_MST_DLY_MASK)
}

// SetSlave4MasterDelay Set Slave 4 master delay value.
//
//	delay New Slave 4 master delay value
func (m *MPU9250) SetSlave4MasterDelay(delay byte) error {
	return m.transport.writeMaskedReg(reg.MPU9250_I2C_SLV4_CTRL, reg.MPU9250_I2C_MST_DLY_MASK, delay)
}

// GetSlave4InputByte Get last available byte read from Slave 4.
//
// This register stores the data read from Slave 4. This field is populated
// after a read transaction.
func (m *MPU9250) GetSlave4InputByte() (byte, error) {
	return m.transport.readByte(reg.MPU9250_I2C_SLV4_DI)
}

// I2C_MST_STATUS register

// GetPassthroughStatus Get FSYNC interrupt status.
//
// This bit reflects the status of the FSYNC interrupt from an external device
// into the MPU-60X0. This is used as a way to pass an external interrupt
// through the MPU-60X0 to the host application processor. When set to 1, this
// bit will cause an interrupt if FSYNC_INT_EN is asserted in INT_PIN_CFG
// (Register 55).
func (m *MPU9250) GetPassthroughStatus() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_MST_STATUS, reg.MPU9250_PASS_THROUGH_MASK)
}

// GetSlave4IsDone Get Slave 4 transaction done status.
//
// Automatically sets to 1 when a Slave 4 transaction has completed. This
// triggers an interrupt if the I2C_MST_INT_EN bit in the INT_ENABLE register
// (Register 56) is asserted and if the SLV_4_DONE_INT bit is asserted in the
// I2C_SLV4_CTRL register (Register 52).
func (m *MPU9250) GetSlave4IsDone() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_MST_STATUS, reg.MPU9250_I2C_SLV4_DONE_MASK)
}

// GetLostArbitration Get master arbitration lost status.
//
// This bit automatically sets to 1 when the I2C Master has lost arbitration of
// the auxiliary I2C bus (an error condition). This triggers an interrupt if the
// I2C_MST_INT_EN bit in the INT_ENABLE register (Register 56) is asserted.
func (m *MPU9250) GetLostArbitration() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_MST_STATUS, reg.MPU9250_I2C_LOST_ARB_MASK)
}

// GetSlave4Nack Get Slave 4 NACK status.
//
// This bit automatically sets to 1 when the I2C Master receives a NACK in a
// transaction with Slave 4. This triggers an interrupt if the I2C_MST_INT_EN
// bit in the INT_ENABLE register (Register 56) is asserted.
func (m *MPU9250) GetSlave4Nack() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_MST_STATUS, reg.MPU9250_I2C_SLV4_NACK_MASK)
}

// GetSlave3Nack Get Slave 3 NACK status.
//
// This bit automatically sets to 1 when the I2C Master receives a NACK in a
// transaction with Slave 3. This triggers an interrupt if the I2C_MST_INT_EN
// bit in the INT_ENABLE register (Register 56) is asserted.
func (m *MPU9250) GetSlave3Nack() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_MST_STATUS, reg.MPU9250_I2C_SLV3_NACK_MASK)
}

// GetSlave2Nack Get Slave 2 NACK status.
//
// This bit automatically sets to 1 when the I2C Master receives a NACK in a
// transaction with Slave 2. This triggers an interrupt if the I2C_MST_INT_EN
// bit in the INT_ENABLE register (Register 56) is asserted.
func (m *MPU9250) GetSlave2Nack() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_MST_STATUS, reg.MPU9250_I2C_SLV2_NACK_MASK)
}

// GetSlave1Nack Get Slave 1 NACK status.
//
// This bit automatically sets to 1 when the I2C Master receives a NACK in a
// transaction with Slave 1. This triggers an interrupt if the I2C_MST_INT_EN
// bit in the INT_ENABLE register (Register 56) is asserted.
func (m *MPU9250) GetSlave1Nack() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_MST_STATUS, reg.MPU9250_I2C_SLV1_NACK_MASK)
}

// GetSlave0Nack Get Slave 0 NACK status.
//
// This bit automatically sets to 1 when the I2C Master receives a NACK in a
// transaction with Slave 0. This triggers an interrupt if the I2C_MST_INT_EN
// bit in the INT_ENABLE register (Register 56) is asserted.
func (m *MPU9250) GetSlave0Nack() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_I2C_MST_STATUS, reg.MPU9250_I2C_SLV0_NACK_MASK)
}

// INT_PIN_CFG register

// GetInterruptMode Get interrupt logic level mode (false=active-high, true=active-low)
func (m *MPU9250) GetInterruptMode() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_INT_PIN_CFG, reg.MPU9250_ACTL_MASK)
}

// SetInterruptMode Set interrupt logic level mode.
//	mode New interrupt mode (false=active-high, true=active-low)
func (m *MPU9250) SetInterruptMode(mode bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_INT_PIN_CFG, reg.MPU9250_ACTL_MASK, boolByte(mode, reg.MPU9250_ACTL_MASK))
}

// GetInterruptDrive Get interrupt drive mode.
//
// Will be set false for push-pull, true for open-drain.
// @return Current interrupt drive mode (false=push-pull, true=open-drain)
func (m *MPU9250) GetInterruptDrive() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_INT_PIN_CFG, reg.MPU9250_OPEN_MASK)
}

// SetInterruptDrive Set interrupt drive mode.
//
//	drive New interrupt drive mode (0=push-pull, 1=open-drain)
func (m *MPU9250) SetInterruptDrive(drive bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_INT_PIN_CFG, reg.MPU9250_OPEN_MASK, boolByte(drive, reg.MPU9250_OPEN_MASK))
}

// GetInterruptLatch Get interrupt latch mode. (0=50us-pulse, 1=latch-until-int-cleared)
// Will be set 0 for 50us-pulse, 1 for latch-until-int-cleared.
func (m *MPU9250) GetInterruptLatch() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_INT_PIN_CFG, reg.MPU9250_LATCH_INT_EN_MASK)
}

// SetInterruptLatch Set interrupt latch mode.
//
//	latch New latch mode (0=50us-pulse, 1=latch-until-int-cleared)
func (m *MPU9250) SetInterruptLatch(latch bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_INT_PIN_CFG, reg.MPU9250_LATCH_INT_EN_MASK, boolByte(latch, reg.MPU9250_LATCH_INT_EN_MASK))
}

// GetInterruptLatchClear Get interrupt latch clear mode.
//
// Will be set 0 for status-read-only, 1 for any-register-read.
func (m *MPU9250) GetInterruptLatchClear() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_INT_PIN_CFG, reg.MPU9250_INT_ANYRD_2CLEAR_MASK)
}

// SetInterruptLatchClear Set interrupt latch clear mode.
//
//	clear New latch clear mode (0=status-read-only, 1=any-register-read)
func (m *MPU9250) SetInterruptLatchClear(clear bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_INT_PIN_CFG, reg.MPU9250_INT_ANYRD_2CLEAR_MASK, boolByte(clear, reg.MPU9250_INT_ANYRD_2CLEAR_MASK))
}

// GetFSyncInterruptLevel Get FSYNC interrupt logic level mode.
func (m *MPU9250) GetFSyncInterruptLevel() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_INT_PIN_CFG, reg.MPU9250_FSYNC_INT_MODE_EN_MASK)
}

// SetFSyncInterruptLevel Set FSYNC interrupt logic level mode.
//
//	mode New FSYNC interrupt mode (0=active-high, 1=active-low)
func (m *MPU9250) SetFSyncInterruptLevel(level bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_INT_PIN_CFG, reg.MPU9250_FSYNC_INT_MODE_EN_MASK, boolByte(level, reg.MPU9250_FSYNC_INT_MODE_EN_MASK))
}

// GetFSyncInterruptEnabled Get FSYNC pin interrupt enabled setting.
//
// Will be set 0 for disabled, 1 for enabled.
func (m *MPU9250) GetFSyncInterruptEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_INT_ENABLE, reg.MPU9250_FSYNC_INT_EN_MASK)
}

// SetFSyncInterruptEnabled Set FSYNC pin interrupt enabled setting.
//
//	enabled New FSYNC pin interrupt enabled setting
func (m *MPU9250) SetFSyncInterruptEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_INT_ENABLE, reg.MPU9250_FSYNC_INT_EN_MASK, boolByte(enabled, reg.MPU9250_FSYNC_INT_EN_MASK))
}

// GetI2CBypassEnabled Get I2C bypass enabled status.
//
// When this bit is equal to 1 and I2C_MST_EN (Register 106 bit[5]) is equal to
// 0, the host application processor will be able to directly access the
// auxiliary I2C bus of the MPU-60X0. When this bit is equal to 0, the host
// application processor will not be able to directly access the auxiliary I2C
// bus of the MPU-60X0 regardless of the state of I2C_MST_EN (Register 106
// bit[5]).
func (m *MPU9250) GetI2CBypassEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_INT_PIN_CFG, reg.MPU9250_BYPASS_EN_MASK)
}

// SetI2CBypassEnabled Set I2C bypass enabled status.
//
// When this bit is equal to 1 and I2C_MST_EN (Register 106 bit[5]) is equal to
// 0, the host application processor will be able to directly access the
// auxiliary I2C bus of the MPU-60X0. When this bit is equal to 0, the host
// application processor will not be able to directly access the auxiliary I2C
// bus of the MPU-60X0 regardless of the state of I2C_MST_EN (Register 106
// bit[5]).
//
//	enabled New I2C bypass enabled status
func (m *MPU9250) SetI2CBypassEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_INT_PIN_CFG, reg.MPU9250_BYPASS_EN_MASK, boolByte(enabled, reg.MPU9250_BYPASS_EN_MASK))
}

// INT_ENABLE register

// GetIntEnabled Get full interrupt enabled status.
//
// Full register byte for all interrupts, for quick reading. Each bit will be
// set 0 for disabled, 1 for enabled.
func (m *MPU9250) GetIntEnabled() (byte, error) {
	return m.transport.readByte(reg.MPU9250_INT_ENABLE)
}

// SetIntEnabled Set full interrupt enabled status.
//
// Full register byte for all interrupts, for quick reading. Each bit should be
// set 0 for disabled, 1 for enabled.
//
// enabled New interrupt enabled status
func (m *MPU9250) SetIntEnabled(enabled byte) error {
	return m.transport.writeByte(reg.MPU9250_INT_ENABLE, enabled)
}

// GetIntFIFOBufferOverflowEnabled Get FIFO Buffer Overflow interrupt enabled status.
//
// Will be set 0 for disabled, 1 for enabled.
func (m *MPU9250) GetIntFIFOBufferOverflowEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_INT_ENABLE, reg.MPU9250_FIFO_OFLOW_EN_MASK)
}

// SetIntFIFOBufferOverflowEnabled Set FIFO Buffer Overflow interrupt enabled status.
//
//	enabled New interrupt enabled status
func (m *MPU9250) SetIntFIFOBufferOverflowEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_INT_ENABLE, reg.MPU9250_FIFO_OFLOW_EN_MASK, boolByte(enabled, reg.MPU9250_FIFO_OFLOW_EN_MASK))
}

// GetIntI2CMasterEnabled Get I2C Master interrupt enabled status.
//
// This enables any of the I2C Master interrupt sources to generate an
// interrupt. Will be set 0 for disabled, 1 for enabled.
func (m *MPU9250) GetIntI2CMasterEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_USER_CTRL, reg.MPU9250_I2C_MST_EN_MASK)
}

// SetIntI2CMasterEnabled Set I2C Master interrupt enabled status.
//
//	enabled New interrupt enabled status
func (m *MPU9250) SetIntI2CMasterEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_USER_CTRL, reg.MPU9250_I2C_MST_EN_MASK, boolByte(enabled, reg.MPU9250_I2C_MST_EN_MASK))
}

// INT_STATUS register

// GetIntStatus Get full set of interrupt status bits.
//
// These bits clear to 0 after the register has been read. Very useful
// for getting multiple INT statuses, since each single bit read clears
// all of them because it has to read the whole byte.
func (m *MPU9250) GetIntStatus() (byte, error) {
	return m.transport.readByte(reg.MPU9250_INT_STATUS)
}

// SIGNAL_PATH_RESET register

// ResetGyroscopePath Reset gyroscope signal path.
//
// The reset will revert the signal path analog to digital converters and
// filters to their power up configurations.
func (m *MPU9250) ResetGyroscopePath() error {
	return m.transport.writeMaskedReg(reg.MPU9250_SIGNAL_PATH_RESET, reg.MPU9250_GYRO_RST_MASK, reg.MPU9250_GYRO_RST_MASK)
}

// ResetAccelerometerPath Reset accelerometer signal path.
//
// The reset will revert the signal path analog to digital converters and
// filters to their power up configurations.
func (m *MPU9250) ResetAccelerometerPath() error {
	return m.transport.writeMaskedReg(reg.MPU9250_SIGNAL_PATH_RESET, reg.MPU9250_ACCEL_RST_MASK, reg.MPU9250_ACCEL_RST_MASK)
}

// ResetTemperaturePath Reset temperature sensor signal path.
//
// The reset will revert the signal path analog to digital converters and
// filters to their power up configurations.
func (m *MPU9250) ResetTemperaturePath() error {
	return m.transport.writeMaskedReg(reg.MPU9250_SIGNAL_PATH_RESET, reg.MPU9250_TEMP_RST_MASK, reg.MPU9250_TEMP_RST_MASK)
}

// USER_CTRL register

// GetFIFOEnabled Get FIFO enabled status.
//
// When this bit is set to 0, the FIFO buffer is disabled. The FIFO buffer
// cannot be written to or read from while disabled. The FIFO buffer's state
// does not change unless the MPU-60X0 is power cycled.
func (m *MPU9250) GetFIFOEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_USER_CTRL, reg.MPU9250_FIFO_EN_MASK)
}

// SetFIFOEnabled Set FIFO enabled status.
//
//	enabled New FIFO enabled status
func (m *MPU9250) SetFIFOEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_USER_CTRL, reg.MPU9250_FIFO_EN_MASK, boolByte(enabled, reg.MPU9250_FIFO_EN_MASK))
}

// GetI2CMasterModeEnabled Get I2C Master Mode enabled status.
//
// When this mode is enabled, the MPU-60X0 acts as the I2C Master to the
// external sensor slave devices on the auxiliary I2C bus. When this bit is
// cleared to 0, the auxiliary I2C bus lines (AUX_DA and AUX_CL) are logically
// driven by the primary I2C bus (SDA and SCL). This is a precondition to
// enabling Bypass Mode. For further information regarding Bypass Mode, please
// refer to Register 55.
func (m *MPU9250) GetI2CMasterModeEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_USER_CTRL, reg.MPU9250_I2C_MST_EN_MASK)
}

// SetI2CMasterModeEnabled Set I2C Master Mode enabled status.
//
//	enabled New I2C Master Mode enabled status
func (m *MPU9250) SetI2CMasterModeEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_USER_CTRL, reg.MPU9250_I2C_MST_EN_MASK, boolByte(enabled, reg.MPU9250_I2C_MST_EN_MASK))
}

// SwitchSPIEnabled Switch from I2C to SPI mode (MPU-6000 only)
//
// If this is set, the primary SPI interface will be enabled in place of the
// disabled primary I2C interface.
func (m *MPU9250) SwitchSPIEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_USER_CTRL, reg.MPU9250_I2C_IF_DIS_MASK, boolByte(enabled, reg.MPU9250_I2C_IF_DIS_MASK))
}

// ResetFIFO Reset the FIFO.
//
// This bit resets the FIFO buffer when set to 1 while FIFO_EN equals 0. This
// bit automatically clears to 0 after the reset has been triggered.
func (m *MPU9250) ResetFIFO() error {
	return m.transport.writeMaskedReg(reg.MPU9250_USER_CTRL, reg.MPU9250_FIFO_RST_MASK, reg.MPU9250_FIFO_RST_MASK)
}

// ResetI2CMaster Reset the I2C Master.
//
// This bit resets the I2C Master when set to 1 while I2C_MST_EN equals 0.
// This bit automatically clears to 0 after the reset has been triggered.
func (m *MPU9250) ResetI2CMaster() error {
	return m.transport.writeMaskedReg(reg.MPU9250_USER_CTRL, reg.MPU9250_I2C_MST_RST_MASK, reg.MPU9250_I2C_MST_RST_MASK)
}

// ResetSensors Reset all sensor registers and signal paths.
//
// When set to 1, this bit resets the signal paths for all sensors (gyroscopes,
// accelerometers, and temperature sensor). This operation will also clear the
// sensor registers. This bit automatically clears to 0 after the reset has been
// triggered.
//
// When resetting only the signal path (and not the sensor registers), please
// use Register 104, SIGNAL_PATH_RESET.
func (m *MPU9250) ResetSensors() error {
	return m.transport.writeMaskedReg(reg.MPU9250_USER_CTRL, reg.MPU9250_SIG_COND_RST_MASK, reg.MPU9250_SIG_COND_RST_MASK)
}

// Reset Trigger a full device reset.
//
// A small delay of ~50ms may be desirable after triggering a reset.
func (m *MPU9250) Reset() error {
	return m.transport.writeMaskedReg(reg.MPU9250_PWR_MGMT_1, reg.MPU9250_H_RESET_MASK, reg.MPU9250_H_RESET_MASK)
}

// GetSleepEnabled Get sleep mode status.
//
// Setting the SLEEP bit in the register puts the device into very low power
// sleep mode. In this mode, only the serial interface and internal registers
// remain active, allowing for a very low standby current. Clearing this bit
// puts the device back into normal mode. To save power, the individual standby
// selections for each of the gyros should be used if any gyro axis is not used
// by the application.
func (m *MPU9250) GetSleepEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_PWR_MGMT_1, reg.MPU9250_SLEEP_MASK)
}

// SetSleepEnabled Set sleep mode status.
//
//	enabled New sleep mode enabled status
func (m *MPU9250) SetSleepEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_PWR_MGMT_1, reg.MPU9250_SLEEP_MASK, boolByte(enabled, reg.MPU9250_SLEEP_MASK))
}

// GetWakeCycleEnabled Get wake cycle enabled status.
//
// When this bit is set to 1 and SLEEP is disabled, the MPU-60X0 will cycle
// between sleep mode and waking up to take a single sample of data from active
// sensors at a rate determined by LP_WAKE_CTRL (register 108).
func (m *MPU9250) GetWakeCycleEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_PWR_MGMT_1, reg.MPU9250_CYCLE_MASK)
}

// SetWakeCycleEnabled Set wake cycle enabled status.
//
//	enabled New sleep mode enabled status
func (m *MPU9250) SetWakeCycleEnabled(enabled bool) error {
	return m.transport.writeMaskedReg(reg.MPU9250_PWR_MGMT_1, reg.MPU9250_CYCLE_MASK, boolByte(enabled, reg.MPU9250_CYCLE_MASK))
}

// GetClockSource Get clock source setting.
func (m *MPU9250) GetClockSource() (byte, error) {
	return m.transport.readMaskedReg(reg.MPU9250_PWR_MGMT_1, reg.MPU9250_CLKSEL_MASK)
}

// GetDeviceID Returns the devide ID
func (m *MPU9250) GetDeviceID() (byte, error) {
	return m.transport.readByte(reg.MPU9250_WHO_AM_I)
}

// FIFO_COUNT* registers

// GetFIFOCount Get current FIFO buffer size.
//
// This value indicates the number of bytes stored in the FIFO buffer. This
// number is in turn the number of bytes that can be read from the FIFO buffer
// and it is directly proportional to the number of samples available given the
// set of sensor data bound to be stored in the FIFO (register 35 and 36).
func (m *MPU9250) GetFIFOCount() (uint16, error) {
	return m.transport.readUint16(reg.MPU9250_FIFO_COUNTH, reg.MPU9250_FIFO_COUNTL)
}

// FIFO_R_W register

// GetFIFOByte Get byte from FIFO buffer.
//
// This register is used to read and write data from the FIFO buffer. Data is
// written to the FIFO in order of register number (from lowest to highest). If
// all the FIFO enable flags (see below) are enabled and all External Sensor
// Data registers (Registers 73 to 96) are associated with a Slave device, the
// contents of registers 59 through 96 will be written in order at the Sample
// Rate.
//
// The contents of the sensor data registers (Registers 59 to 96) are written
// into the FIFO buffer when their corresponding FIFO enable flags are set to 1
// in FIFO_EN (Register 35). An additional flag for the sensor data registers
// associated with I2C Slave 3 can be found in I2C_MST_CTRL (Register 36).
//
// If the FIFO buffer has overflowed, the status bit FIFO_OFLOW_INT is
// automatically set to 1. This bit is located in INT_STATUS (Register 58).
// When the FIFO buffer has overflowed, the oldest data will be lost and new
// data will be written to the FIFO.
//
// If the FIFO buffer is empty, reading this register will return the last byte
// that was previously read from the FIFO until new data is available. The user
// should check FIFO_COUNT to ensure that the FIFO buffer is not read when
// empty.
func (m *MPU9250) GetFIFOByte() (byte, error) {
	return m.transport.readByte(reg.MPU9250_FIFO_R_W)
}

// SetFIFOByte Write byte to FIFO buffer.
func (m *MPU9250) SetFIFOByte(data byte) error {
	return m.transport.writeByte(reg.MPU9250_FIFO_R_W, data)
}

func (m *MPU9250) getOffset(regs ...byte) (uint16, error) {
	return m.transport.readUint16(regs...)
}

func (m *MPU9250) setOffset(regHi, regLo byte, offset uint16) error {
	msbOffset := byte(offset >> 8)
	lsbOffset := byte(offset & 0x00FF)
	if err := m.transport.writeByte(regHi, msbOffset); err != nil {
		return err
	}
	return m.transport.writeByte(regLo, lsbOffset)
}

// GetXGyroOffset Returns the offset of X gyroscope.
func (m *MPU9250) GetXGyroOffset() (uint16, error) {
	return m.getOffset(reg.MPU9250_XG_OFFSET_H, reg.MPU9250_XG_OFFSET_L)
}

// SetXGyroOffset Sets the X gyroscope offset.
//
//	offset offset value.
func (m *MPU9250) SetXGyroOffset(offset uint16) error {
	return m.setOffset(reg.MPU9250_XG_OFFSET_H, reg.MPU9250_XG_OFFSET_L, offset)
}

// GetYGyroOffset Returns the offset of y gyroscope.
func (m *MPU9250) GetYGyroOffset() (uint16, error) {
	return m.getOffset(reg.MPU9250_YG_OFFSET_H, reg.MPU9250_YG_OFFSET_L)
}

// SetZGyroOffset Sets the Z gyroscope offset.
//
//	offset offset value.
func (m *MPU9250) SetZGyroOffset(offset uint16) error {
	return m.setOffset(reg.MPU9250_ZG_OFFSET_H, reg.MPU9250_ZG_OFFSET_L, offset)
}

// GetZGyroOffset Returns the offset of Z gyroscope.
func (m *MPU9250) GetZGyroOffset() (uint16, error) {
	return m.getOffset(reg.MPU9250_ZG_OFFSET_H, reg.MPU9250_ZG_OFFSET_L)
}

// SetYGyroOffset Sets the Y gyroscope offset.
//
//	offset offset value.
func (m *MPU9250) SetYGyroOffset(offset uint16) error {
	return m.setOffset(reg.MPU9250_YG_OFFSET_H, reg.MPU9250_YG_OFFSET_L, offset)
}

// EnableAccelerometerAxis Accelerometer functions
//
//	axis MPU9250_DISABLE_XA_MASK, MPU9250_DISABLE_YA_MASK, MPU9250_DISABLE_ZA_MASK
func (m *MPU9250) EnableAccelerometerAxis(axis byte) error {
	return m.transport.writeMaskedReg(reg.MPU9250_PWR_MGMT_2, axis, 0)
}

// DisableAccelerometerAxis Accelerometer functions
//
//	axis MPU9250_DISABLE_XA_MASK, MPU9250_DISABLE_YA_MASK, MPU9250_DISABLE_ZA_MASK
func (m *MPU9250) DisableAccelerometerAxis(axis byte) error {
	return m.transport.writeMaskedReg(reg.MPU9250_PWR_MGMT_2, axis, axis)
}

// EnableAccelerometer enables the accelerometer.
func (m *MPU9250) EnableAccelerometer() error {
	return m.transport.writeMaskedReg(reg.MPU9250_PWR_MGMT_2, reg.MPU9250_DISABLE_XYZA_MASK, 0)
}

// GetAccelerometerTestData Get the self-test data values from the registers.
func (m *MPU9250) GetAccelerometerTestData() (*AccelerometerData, error) {
	x, err := m.transport.readByte(reg.MPU9250_SELF_TEST_X_ACCEL)
	if err != nil {
		return nil, err
	}
	y, err := m.transport.readByte(reg.MPU9250_SELF_TEST_Y_ACCEL)
	if err != nil {
		return nil, err
	}
	z, err := m.transport.readByte(reg.MPU9250_SELF_TEST_Z_ACCEL)
	if err != nil {
		return nil, err
	}
	return &AccelerometerData{X: int16(x), Y: int16(y), Z: int16(z)}, nil
}

func negateBool(src bool, err error) (bool, error) {
	if err != nil {
		return src, err
	}
	return !src, err
}

// AccelerometerXIsEnabled Get the X accelerometer status.
func (m *MPU9250) AccelerometerXIsEnabled() (bool, error) {
	return negateBool(m.readMaskedRegBool(reg.MPU9250_PWR_MGMT_2, reg.MPU9250_DISABLE_XA_MASK))
}

// AccelerometerYIsEnabled Get the X accelerometer status.
func (m *MPU9250) AccelerometerYIsEnabled() (bool, error) {
	return negateBool(m.readMaskedRegBool(reg.MPU9250_PWR_MGMT_2, reg.MPU9250_DISABLE_YA_MASK))
}

// AccelerometerZIsEnabled Get the Z accelerometer status.
func (m *MPU9250) AccelerometerZIsEnabled() (bool, error) {
	return negateBool(m.readMaskedRegBool(reg.MPU9250_PWR_MGMT_2, reg.MPU9250_DISABLE_ZA_MASK))
}

func uintToInt(s uint16, err error) (int16, error) {
	return int16(s), err
}

// GetAccelerationX Get X-axis accelerometer reading.
func (m *MPU9250) GetAccelerationX() (int16, error) {
	if enabled, err := m.AccelerometerXIsEnabled(); err == nil && enabled {
		return uintToInt(m.ReadWord(reg.MPU9250_ACCEL_XOUT_H, reg.MPU9250_ACCEL_XOUT_L))
	} else if !enabled {
		return 0, wrapf("X acceleration disabled")
	} else {
		return 0, err
	}
}

// GetAccelerationY Get Y-axis accelerometer reading.
func (m *MPU9250) GetAccelerationY() (int16, error) {
	if enabled, err := m.AccelerometerYIsEnabled(); err == nil && enabled {
		return uintToInt(m.ReadWord(reg.MPU9250_ACCEL_YOUT_H, reg.MPU9250_ACCEL_YOUT_L))
	} else if !enabled {
		return 0, wrapf("Y acceleration disabled")
	} else {
		return 0, err
	}
}

// GetAccelerationZ Get Z-axis accelerometer reading.
func (m *MPU9250) GetAccelerationZ() (int16, error) {
	if enabled, err := m.AccelerometerZIsEnabled(); err == nil && enabled {
		return uintToInt(m.ReadWord(reg.MPU9250_ACCEL_ZOUT_H, reg.MPU9250_ACCEL_ZOUT_L))
	} else if !enabled {
		return 0, wrapf("Z acceleration disabled")
	} else {
		return 0, err
	}
}

// SetAccelerationOffsetZ Set the acceleration trim offset for Z axis.
func (m *MPU9250) SetAccelerationOffsetZ(offset uint16) error {
	return m.setOffset(reg.MPU9250_ZA_OFFSET_H, reg.MPU9250_ZA_OFFSET_L, offset)
}

// SetAccelerationOffsetY Set the acceleration trim offset for Y axis.
func (m *MPU9250) SetAccelerationOffsetY(offset uint16) error {
	return m.setOffset(reg.MPU9250_YA_OFFSET_H, reg.MPU9250_YA_OFFSET_L, offset)
}

// SetAccelerationOffsetX Set the acceleration trim offset for X axis.
func (m *MPU9250) SetAccelerationOffsetX(offset uint16) error {
	return m.setOffset(reg.MPU9250_XA_OFFSET_H, reg.MPU9250_XA_OFFSET_L, offset)
}

// GetAcceleration Get 3-axis accelerometer readings.
//
// These registers store the most recent accelerometer measurements.
// Accelerometer measurements are written to these registers at the Sample Rate
// as defined in Register 25.
//
// The accelerometer measurement registers, along with the temperature
// measurement registers, gyroscope measurement registers, and external sensor
// data registers, are composed of two sets of registers: an internal register
// set and a user-facing read register set.
//
// The data within the accelerometer sensors' internal register set is always
// updated at the Sample Rate. Meanwhile, the user-facing read register set
// duplicates the internal register set's data values whenever the serial
// interface is idle. This guarantees that a burst read of sensor registers will
// read measurements from the same sampling instant. Note that if burst reads
// are not used, the user is responsible for ensuring a set of single byte reads
// correspond to a single sampling instant by checking the Data Ready interrupt.
//
// Each 16-bit accelerometer measurement has a full scale defined in ACCEL_FS
// (Register 28). For each full scale setting, the accelerometers' sensitivity
// per LSB in ACCEL_xOUT is shown in the table below:
//
// AFS_SEL | Full Scale Range | LSB Sensitivity
// --------+------------------+----------------
// 0       | +/- 2g           | 8192 LSB/mg
// 1       | +/- 4g           | 4096 LSB/mg
// 2       | +/- 8g           | 2048 LSB/mg
// 3       | +/- 16g          | 1024 LSB/mg
func (m *MPU9250) GetAcceleration() (*AccelerometerData, error) {
	x, err := m.GetAccelerationX()
	if err != nil {
		return nil, err
	}
	y, err := m.GetAccelerationY()
	if err != nil {
		return nil, err
	}
	z, err := m.GetAccelerationZ()
	if err != nil {
		return nil, err
	}
	return &AccelerometerData{X: x, Y: y, Z: z}, nil
}

//Gyroscope functions

// GetGyroTestData Get the test data from self-test registers, factory settings.
func (m *MPU9250) GetGyroTestData() (*GyroscopeData, error) {
	x, err := m.transport.readByte(reg.MPU9250_SELF_TEST_X_GYRO)
	if err != nil {
		return nil, err
	}
	y, err := m.transport.readByte(reg.MPU9250_SELF_TEST_Y_GYRO)
	if err != nil {
		return nil, err
	}
	z, err := m.transport.readByte(reg.MPU9250_SELF_TEST_Z_GYRO)
	if err != nil {
		return nil, err
	}
	return &GyroscopeData{X: int16(x), Y: int16(y), Z: int16(z)}, nil
}

// EnableGyro Enables the gyroscope.
func (m *MPU9250) EnableGyro() error {
	return m.transport.writeMaskedReg(reg.MPU9250_PWR_MGMT_2, reg.MPU9250_DISABLE_XYZG_MASK, 0)
}

// DisableGyro Disables the gyroscope.
func (m *MPU9250) DisableGyro() error {
	return m.transport.writeMaskedReg(reg.MPU9250_PWR_MGMT_2, reg.MPU9250_DISABLE_XYZG_MASK, reg.MPU9250_DISABLE_XYZG_MASK)
}

// GyroXIsEnabled Get the X axis enabled flag.
func (m *MPU9250) GyroXIsEnabled() (bool, error) {
	//WARNING: this should be verified here
	return negateBool(m.readMaskedRegBool(reg.MPU9250_PWR_MGMT_2, reg.MPU9250_DISABLE_XG_MASK))
}

// GyroYIsEnabled Get the Y axis enabled flag.
func (m *MPU9250) GyroYIsEnabled() (bool, error) {
	//WARNING: this should be verified here
	return negateBool(m.readMaskedRegBool(reg.MPU9250_PWR_MGMT_2, reg.MPU9250_DISABLE_YG_MASK))
}

// GyroZIsEnabled Get the Z axis enabled flag.
func (m *MPU9250) GyroZIsEnabled() (bool, error) {
	//WARNING: this should be verified here
	return negateBool(m.readMaskedRegBool(reg.MPU9250_PWR_MGMT_2, reg.MPU9250_DISABLE_ZG_MASK))
}

// GetRotationX Get X-axis gyroscope reading.
func (m *MPU9250) GetRotationX() (int16, error) {
	if enabled, err := m.GyroXIsEnabled(); enabled && err == nil {
		return asInt16(m.ReadWord(reg.MPU9250_GYRO_XOUT_H, reg.MPU9250_GYRO_XOUT_L))
	} else if !enabled {
		return 0, wrapf("X rotation disabled")
	} else {
		return 0, err
	}
}

// GetRotationY Get Y-axis gyroscope reading.
func (m *MPU9250) GetRotationY() (int16, error) {
	if enabled, err := m.GyroYIsEnabled(); enabled && err == nil {
		return asInt16(m.ReadWord(reg.MPU9250_GYRO_YOUT_H, reg.MPU9250_GYRO_YOUT_L))
	} else if !enabled {
		return 0, wrapf("Y rotation disabled")
	} else {
		return 0, err
	}
}

// GetRotationZ Get Z-axis gyroscope reading.
func (m *MPU9250) GetRotationZ() (int16, error) {
	if enabled, err := m.GyroZIsEnabled(); enabled && err == nil {
		return asInt16(m.ReadWord(reg.MPU9250_GYRO_ZOUT_H, reg.MPU9250_GYRO_ZOUT_L))
	} else if !enabled {
		return 0, wrapf("Z rotation disabled")
	} else {
		return 0, err
	}
}

// GetRotation Get 3-axis gyroscope readings.
//
// These gyroscope measurement registers, along with the accelerometer
// measurement registers, temperature measurement registers, and external sensor
// data registers, are composed of two sets of registers: an internal register
// set and a user-facing read register set.
// The data within the gyroscope sensors' internal register set is always
// updated at the Sample Rate. Meanwhile, the user-facing read register set
// duplicates the internal register set's data values whenever the serial
// interface is idle. This guarantees that a burst read of sensor registers will
// read measurements from the same sampling instant. Note that if burst reads
// are not used, the user is responsible for ensuring a set of single byte reads
// correspond to a single sampling instant by checking the Data Ready interrupt.
//
// Each 16-bit gyroscope measurement has a full scale defined in FS_SEL
// (Register 27). For each full scale setting, the gyroscopes' sensitivity per
// LSB in GYRO_xOUT is shown in the table below:
//
// FS_SEL | Full Scale Range   | LSB Sensitivity
// -------+--------------------+----------------
// 0      | +/- 250 degrees/s  | 131 LSB/deg/s
// 1      | +/- 500 degrees/s  | 65.5 LSB/deg/s
// 2      | +/- 1000 degrees/s | 32.8 LSB/deg/s
// 3      | +/- 2000 degrees/s | 16.4 LSB/deg/s
func (m *MPU9250) GetRotation() (*RotationData, error) {
	x, err := m.GetRotationX()
	if err != nil {
		return nil, err
	}
	y, err := m.GetRotationY()
	if err != nil {
		return nil, err
	}
	z, err := m.GetRotationZ()
	if err != nil {
		return nil, err
	}
	return &RotationData{X: x, Y: y, Z: z}, nil
}

// GetMotion6 Get the motion data - accelerometer and rotation(gyroscope).
func (m *MPU9250) GetMotion6() (*AccelerometerData, *RotationData, error) {
	acc, err := m.GetAcceleration()
	if err != nil {
		return nil, nil, err
	}
	rot, err := m.GetRotation()
	if err != nil {
		return nil, nil, err
	}
	return acc, rot, nil
}

//Temperature functions

// EnableTemperature Enable internal temperature sensor.
func (m *MPU9250) EnableTemperature() error {
	return m.transport.writeMaskedReg(reg.MPU9250_FIFO_EN, reg.MPU9250_TEMP_FIFO_EN_MASK, boolByte(true, reg.MPU9250_TEMP_FIFO_EN_MASK))
}

// DisableTemperature Disables the internal temperature sensor.
func (m *MPU9250) DisableTemperature() error {
	return m.transport.writeMaskedReg(reg.MPU9250_FIFO_EN, reg.MPU9250_TEMP_FIFO_EN_MASK, boolByte(false, reg.MPU9250_TEMP_FIFO_EN_MASK))
}

// TemperatureIsEnabled Is the temperature sensor enabled.
func (m *MPU9250) TemperatureIsEnabled() (bool, error) {
	return m.readMaskedRegBool(reg.MPU9250_FIFO_EN, reg.MPU9250_TEMP_FIFO_EN_MASK)
}

// GetTemperature Get the current temperature.
//((Temp_out - room_temp_offset)/temp_sensitivity) + 21; //celcius
func (m *MPU9250) GetTemperature() (uint16, error) {
	if enabled, err := m.TemperatureIsEnabled(); enabled && err != nil {
		return 0, err
	}
	return m.ReadWord(reg.MPU9250_TEMP_OUT_H, reg.MPU9250_TEMP_OUT_L)
}

// WriteByteAddress Writes the byte address
func (m *MPU9250) WriteByteAddress(address, value byte) error {
	return m.transport.writeByte(address, value)
}

func (m *MPU9250) readMaskedRegBool(address, mask byte) (bool, error) {
	response, err := m.transport.readMaskedReg(address, mask)
	if err != nil {
		return false, err
	}
	return response != 0, nil
}

// ReadWord Reads unsigned int from the provided addresses.
//
//	hi high 8 bit address
//	lo low 8 bit address
func (m *MPU9250) ReadWord(hi, lo byte) (uint16, error) {
	hiByte, err := m.transport.readByte(hi)
	if err != nil {
		return 0, wrapf("can't read word %x(hi)=> %v", hi, lo, err)
	}
	loByte, err := m.transport.readByte(lo)
	if err != nil {
		return 0, wrapf("can't read word %x(lo)=> %v", hi, lo, err)
	}
	return uint16(hiByte)<<8 | uint16(loByte), nil
}

// ReadSignedWord Reads signed word from the provided high/low registers addresses.
func (m *MPU9250) ReadSignedWord(hi, lo byte) (int16, error) {
	res, err := m.ReadWord(hi, lo)
	return int16(res), err
}

func asInt16(src uint16, err error) (int16, error) {
	return int16(src), err
}

func boolByte(v bool, mask byte) byte {
	if v {
		return mask
	}
	return 0
}

func (m *MPU9250) transferBatch(seq [][]byte, msg string) error {
	for i, cmds := range seq {
		if len(cmds) == 2 {
			if err := m.transport.writeByte(cmds[0], cmds[1]); err != nil {
				return wrapf(msg, i, cmds[0], cmds[1], err)
			}
		} else {
			time.Sleep(time.Duration(cmds[0]) * time.Millisecond)
		}
	}
	return nil
}

func wrapf(format string, a ...interface{}) error {
	return fmt.Errorf("mpu9250 "+format, a...)
}

var (
	initSequence = [][]byte{
		{reg.MPU9250_PWR_MGMT_1, 0x0}, // Clear sleep mode bit (6), enable all sensors
		{100},
		{reg.MPU9250_PWR_MGMT_2, 0x01}, // Auto select clock source to be PLL gyroscope reference if ready else
		{100},
		// Configure Gyro and Thermometer
		// Disable FSYNC and set thermometer and gyro bandwidth to 41 and 42 Hz, respectively;
		// minimum delay time for this setting is 5.9 ms, which means sensor fusion update rates cannot
		// be higher than 1 / 0.0059 = 170 Hz
		// DLPF_CFG = bits 2:0 = 011; this limits the sample rate to 1000 Hz for both
		// With the MPU9250, it is possible to get gyro sample rates of 32 kHz (!), 8 kHz, or 1 kHz
		{reg.MPU9250_CONFIG, 0x03},
		{reg.MPU9250_SMPLRT_DIV, 0x04}, // Set sample rate = gyroscope output rate/(1 + SMPLRT_DIV)
	}
	calibrateSequence = [][]byte{
		{reg.MPU9250_PWR_MGMT_1, 0x80}, // reset device
		{100}, // sleep 100 ms
		{reg.MPU9250_PWR_MGMT_1, 1}, // get stable time source; Auto select clock source to be PLL gyroscope reference if ready else use the internal oscillator, bits 2:0 = 001
		{reg.MPU9250_PWR_MGMT_2, 0},
		{200}, // wait 200 ms
		{reg.MPU9250_INT_ENABLE, 0},   // Disable all interrupts
		{reg.MPU9250_FIFO_EN, 0},      // Disable FIFO
		{reg.MPU9250_PWR_MGMT_1, 0},   // Turn on internal clock source
		{reg.MPU9250_I2C_MST_CTRL, 0}, // Disable I2C master
		{reg.MPU9250_USER_CTRL, 0},    // Disable FIFO and I2C master modes
		{reg.MPU9250_USER_CTRL, 0x0C}, // Reset FIFO and DMP
		{15}, // wait 15 ms
		{reg.MPU9250_CONFIG, 0x01},    // Set low-pass filter to 188 Hz
		{reg.MPU9250_SMPLRT_DIV, 0},   // Set sample rate to 1 kHz
		{reg.MPU9250_GYRO_CONFIG, 0},  // Set gyro full-scale to 250 degrees per second, maximum sensitivity
		{reg.MPU9250_ACCEL_CONFIG, 0}, // Set accelerometer full-scale to 2 g, maximum sensitivity
		{reg.MPU9250_USER_CTRL, 0x40}, // Enable FIFO
		{reg.MPU9250_FIFO_EN, 0x78},   // Enable gyro and accelerometer sensors for FIFO  (max size 512 bytes in MPU-9150)
		{40}, // wait 40 ms
		{reg.MPU9250_FIFO_EN, 0x00}, // Disable gyro and accelerometer sensors for FIFO
	}
)
