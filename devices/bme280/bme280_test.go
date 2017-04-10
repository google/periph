// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bme280

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"periph.io/x/periph/conn/conntest"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/i2c/i2ctest"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/conn/spi/spitest"
	"periph.io/x/periph/devices"
)

// Real data extracted from a device.
var calib = calibration{
	t1: 28176,
	t2: 26220,
	t3: 350,
	p1: 38237,
	p2: -10824,
	p3: 3024,
	p4: 7799,
	p5: -99,
	p6: -7,
	p7: 9900,
	p8: -10230,
	p9: 4285,
	h2: 366, // Note they are inversed for bit packing.
	h1: 75,
	h3: 0,
	h4: 309,
	h5: 0,
	h6: 30,
}

func TestSPISense_success(t *testing.T) {
	bus := spitest.Playback{
		Playback: conntest.Playback{
			Ops: []conntest.IO{
				// Chipd ID detection.
				{
					W: []byte{0xD0, 0x00},
					R: []byte{0x00, 0x60},
				},
				// Calibration data.
				{
					W: []byte{0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					R: []byte{0x00, 0xC9, 0x6C, 0x63, 0x65, 0x32, 0x00, 0x77, 0x93, 0x98, 0xD5, 0xD0, 0x0B, 0x67, 0x23, 0xBA, 0x00, 0xF9, 0xFF, 0xAC, 0x26, 0x0A, 0xD8, 0xBD, 0x10, 0x00, 0x4B},
				},
				// Calibration data.
				{
					W: []byte{0xE1, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					R: []byte{0x00, 0x5C, 0x01, 0x00, 0x15, 0x0F, 0x00, 0x1E},
				},
				{W: []byte{0x74, 0xB4, 0x72, 0x05, 0x75, 0xA0, 0x74, 0xB7}},
				// R.
				{
					W: []byte{0xF7, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					R: []byte{0x00, 0x51, 0x9F, 0xC0, 0x9E, 0x3A, 0x50, 0x5E, 0x5B},
				},
			},
		},
	}
	opts := Opts{
		Temperature: O16x,
		Pressure:    O16x,
		Humidity:    O16x,
		Standby:     S1s,
		Filter:      FOff,
	}
	dev, err := NewSPI(&bus, &opts)
	if err != nil {
		t.Fatal(err)
	}
	if s := dev.String(); s != "BME280{playback}" {
		t.Fatal(s)
	}
	env := devices.Environment{}
	if err := dev.Sense(&env); err != nil {
		t.Fatal(err)
	}
	// TODO(maruel): The values do not make sense but I think I burned my SPI
	// BME280 by misconnecting it in reverse for a few minutes. It still "work"
	// but fail to read data. It could also be a bug in the driver. :(
	if env.Temperature != 62680 {
		t.Fatalf("temp %d", env.Temperature)
	}
	if env.Pressure != 99576 {
		t.Fatalf("pressure %d", env.Pressure)
	}
	if env.Humidity != 995 {
		t.Fatalf("humidity %d", env.Humidity)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewSPI_fail(t *testing.T) {
	if d, err := NewSPI(&spiFail{}, nil); d != nil || err == nil {
		t.Fatal("DevParams() have failed")
	}
}

func TestNewSPI_fail_len(t *testing.T) {
	bus := spitest.Playback{
		Playback: conntest.Playback{
			Ops: []conntest.IO{
				{
					// Chipd ID detection.
					W: []byte{0xD0, 0x00},
					R: []byte{0x00},
				},
			},
			DontPanic: true,
		},
	}
	if dev, err := NewSPI(&bus, nil); dev != nil || err == nil {
		t.Fatal("read failed")
	}
	// The I/O didn't occur.
	bus.Count++
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewSPI_fail_chipid(t *testing.T) {
	bus := spitest.Playback{
		Playback: conntest.Playback{
			Ops: []conntest.IO{
				{
					// Chipd ID detection.
					W: []byte{0xD0, 0x00},
					R: []byte{0x00, 0xFF},
				},
			},
		},
	}
	if dev, err := NewSPI(&bus, nil); dev != nil || err == nil {
		t.Fatal("read failed")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewI2C_fail_io(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chipd ID detection.
			{Addr: 0x76, W: []byte{0xd0}},
		},
		DontPanic: true,
	}
	if dev, err := NewI2C(&bus, nil); dev != nil || err == nil {
		t.Fatal("read failed")
	}
	// The I/O didn't occur.
	bus.Count++
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewI2C_fail_chipid(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chipd ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
		},
		DontPanic: true,
	}
	if dev, err := NewI2C(&bus, nil); dev != nil || err == nil {
		t.Fatal("invalid chip id")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewI2C_calib1(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chipd ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
			// Calibration data.
			{
				Addr: 0x76,
				W:    []byte{0x88},
				R:    []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
		},
		DontPanic: true,
	}
	opts := Opts{Address: 0}
	if dev, err := NewI2C(&bus, &opts); dev != nil || err == nil {
		t.Fatal("2nd calib read failed")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewI2C_calib2(t *testing.T) {
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chipd ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
			// Calibration data.
			{
				Addr: 0x76,
				W:    []byte{0x88},
				R:    []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
			// Calibration data.
			{Addr: 0x76, W: []byte{0xe1}, R: []byte{0x6e, 0x1, 0x0, 0x13, 0x5, 0x0, 0x1e}},
		},
		DontPanic: true,
	}
	if dev, err := NewI2C(&bus, nil); dev != nil || err == nil {
		t.Fatal("3rd calib read failed")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2COpts_bad_addr(t *testing.T) {
	bus := i2ctest.Playback{}
	opts := Opts{Address: 1}
	if dev, err := NewI2C(&bus, &opts); dev != nil || err == nil {
		t.Fatal("bad addr")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2COpts(t *testing.T) {
	bus := i2ctest.Playback{DontPanic: true}
	opts := Opts{Address: 0x76}
	if dev, err := NewI2C(&bus, &opts); dev != nil || err == nil {
		t.Fatal("write fails")
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2CSense_fail(t *testing.T) {
	// This data was generated with "bme280 -r"
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chipd ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
			// Calibration data.
			{
				Addr: 0x76,
				W:    []byte{0x88},
				R:    []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
			// Calibration data.
			{Addr: 0x76, W: []byte{0xe1}, R: []byte{0x6e, 0x1, 0x0, 0x13, 0x5, 0x0, 0x1e}},
			// Configuration.
			{Addr: 0x76, W: []byte{0xf4, 0x6c, 0xf2, 0x3, 0xf5, 0xe0, 0xf4, 0x6f}, R: nil},
			// Read.
			{Addr: 0x76, W: []byte{0xf7}},
		},
		DontPanic: true,
	}
	dev, err := NewI2C(&bus, nil)
	if err != nil {
		t.Fatal(err)
	}
	if dev.Sense(&devices.Environment{}) == nil {
		t.Fatal("sense fail read")
	}
	// The I/O didn't occur.
	bus.Count++
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestI2CSense_success(t *testing.T) {
	// This data was generated with "bme280 -r"
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chipd ID detection.
			{Addr: 0x76, W: []byte{0xd0}, R: []byte{0x60}},
			// Calibration data.
			{
				Addr: 0x76,
				W:    []byte{0x88},
				R:    []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
			// Calibration data.
			{Addr: 0x76, W: []byte{0xe1}, R: []byte{0x6e, 0x1, 0x0, 0x13, 0x5, 0x0, 0x1e}},
			// Configuration.
			{Addr: 0x76, W: []byte{0xf4, 0x6c, 0xf2, 0x3, 0xf5, 0xe0, 0xf4, 0x6f}, R: nil},
			// Read.
			{Addr: 0x76, W: []byte{0xf7}, R: []byte{0x4a, 0x52, 0xc0, 0x80, 0x96, 0xc0, 0x7a, 0x76}},
			// Halt.
			{Addr: 0x76, W: []byte{0xf4, 0x0}},
		},
	}
	dev, err := NewI2C(&bus, nil)
	if err != nil {
		t.Fatal(err)
	}
	if s := dev.String(); s != "BME280{playback(118)}" {
		t.Fatal(s)
	}
	env := devices.Environment{}
	if err := dev.Sense(&env); err != nil {
		t.Fatal(err)
	}
	if env.Temperature != 23720 {
		t.Fatalf("temp %d", env.Temperature)
	}
	if env.Pressure != 100943 {
		t.Fatalf("pressure %d", env.Pressure)
	}
	if env.Humidity != 6531 {
		t.Fatalf("humidity %d", env.Humidity)
	}
	if err := dev.Halt(); err != nil {
		t.Fatal(err)
	}
	if err := bus.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCalibrationFloat(t *testing.T) {
	// Real data extracted from measurements from this device.
	tRaw := int32(524112)
	pRaw := int32(309104)
	hRaw := int32(30987)

	// Compare the values with the 3 algorithms.
	temp, tFine := calib.compensateTempFloat(tRaw)
	pres := calib.compensatePressureFloat(pRaw, tFine)
	humi := calib.compensateHumidityFloat(hRaw, tFine)
	if tFine != 117494 {
		t.Fatalf("tFine %d", tFine)
	}
	if !floatEqual(temp, 22.948120) {
		// 22.95°C
		t.Fatalf("temp %f", temp)
	}
	if !floatEqual(pres, 100.046074) {
		// 100.046kPa
		t.Fatalf("pressure %f", pres)
	}
	if !floatEqual(humi, 63.167889) {
		// 63.17%
		t.Fatalf("humidity %f", humi)
	}
}

func TestCalibrationInt(t *testing.T) {
	// Real data extracted from measurements from this device.
	tRaw := int32(524112)
	pRaw := int32(309104)
	hRaw := int32(30987)

	temp, tFine := calib.compensateTempInt(tRaw)
	pres64 := calib.compensatePressureInt64(pRaw, tFine)
	pres32 := calib.compensatePressureInt32(pRaw, tFine)
	humi := calib.compensateHumidityInt(hRaw, tFine)
	if tFine != 117407 {
		t.Fatalf("tFine %d", tFine)
	}
	if temp != 2293 {
		// 2293/100 = 22.93°C
		// Delta is <0.02°C which is pretty good.
		t.Fatalf("temp %d", temp)
	}
	if pres64 != 25611063 {
		// 25611063/256/1000 = 100.043214844
		// Delta is 3Pa which is ok.
		t.Fatalf("pressure64 %d", pres64)
	}
	if pres32 != 100045 {
		// 100045/1000 = 100.045kPa
		// Delta is 1Pa which is pretty good.
		t.Fatalf("pressure32 %d", pres32)
	}
	if humi != 64686 {
		// 64686/1024 = 63.17%
		// Delta is <0.01% which is pretty good.
		t.Fatalf("humidity %d", humi)
	}
}

func TestCalibration_limits_0(t *testing.T) {
	c := calibration{h1: 0xFF, h2: 1, h3: 1, h6: 1}
	if v := c.compensateHumidityInt(0x7FFFFFFF>>14, 0xFFFFFFF); v != 0 {
		t.Fatal(v)
	}
}

func TestCalibration_limits_419430400(t *testing.T) {
	// TODO(maruel): Reverse the equation to overflow  419430400
}

//

func Example() {
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatalf("failed to open I²C: %v", err)
	}
	defer bus.Close()
	dev, err := NewI2C(bus, nil)
	if err != nil {
		log.Fatalf("failed to initialize bme280: %v", err)
	}
	env := devices.Environment{}
	if err := dev.Sense(&env); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%8s %10s %9s\n", env.Temperature, env.Pressure, env.Humidity)
}

func TestCalibration_compensatePressureInt64(t *testing.T) {
	c := calibration{}
	if x := c.compensatePressureInt64(0, 0); x != 0 {
		t.Fatal(x)
	}
}

func TestCalibration_compensateHumidityInt(t *testing.T) {
	c := calibration{
		h1: 0xFF,
	}
	if x := c.compensateHumidityInt(0, 0); x != 0 {
		t.Fatal(x)
	}
}

//

var epsilon float32 = 0.00000001

func floatEqual(a, b float32) bool {
	return (a-b) < epsilon && (b-a) < epsilon
}

// Page 50

// compensatePressureInt32 returns pressure in Pa. Output value of "96386"
// equals 96386 Pa = 963.86 hPa
//
// "Compensating the pressure value with 32 bit integer has an accuracy of
// typically 1 Pa"
//
// raw has 20 bits of resolution.
//
// BUG(maruel): Output is incorrect.
func (c *calibration) compensatePressureInt32(raw, tFine int32) uint32 {
	x := tFine>>1 - 64000
	y := (((x >> 2) * (x >> 2)) >> 11) * int32(c.p6)
	y += (x * int32(c.p5)) << 1
	y = y>>2 + int32(c.p4)<<16
	x = (((int32(c.p3) * (((x >> 2) * (x >> 2)) >> 13)) >> 3) + ((int32(c.p2) * x) >> 1)) >> 18
	x = ((32768 + x) * int32(c.p1)) >> 15
	if x == 0 {
		return 0
	}
	p := ((uint32(int32(1048576)-raw) - uint32(y>>12)) * 3125)
	if p < 0x80000000 {
		p = (p << 1) / uint32(x)
	} else {
		p = (p / uint32(x)) * 2
	}
	x = (int32(c.p9) * int32(((p>>3)*(p>>3))>>13)) >> 12
	y = (int32(p>>2) * int32(c.p8)) >> 13
	return uint32(int32(p) + ((x + y + int32(c.p7)) >> 4))
}

var _ devices.Environmental = &Dev{}

// Page 49

// compensateTempFloat returns temperature in °C. Output value of "51.23"
// equals 51.23 °C.
//
// raw has 20 bits of resolution.
func (c *calibration) compensateTempFloat(raw int32) (float32, int32) {
	x := (float64(raw)/16384. - float64(c.t1)/1024.) * float64(c.t2)
	y := (float64(raw)/131072. - float64(c.t1)/8192.) * float64(c.t3)
	tFine := int32(x + y)
	return float32((x + y) / 5120.), tFine
}

// compensateHumidityFloat returns pressure in Pa. Output value of "96386.2"
// equals 96386.2 Pa = 963.862 hPa.
//
// raw has 20 bits of resolution.
func (c *calibration) compensatePressureFloat(raw, tFine int32) float32 {
	x := float64(tFine)*0.5 - 64000.
	y := x * x * float64(c.p6) / 32768.
	y += x * float64(c.p5) * 2.
	y = y*0.25 + float64(c.p4)*65536.
	x = (float64(c.p3)*x*x/524288. + float64(c.p2)*x) / 524288.
	x = (1. + x/32768.) * float64(c.p1)
	if x <= 0 {
		return 0
	}
	p := float64(1048576 - raw)
	p = (p - y/4096.) * 6250. / x
	x = float64(c.p9) * p * p / 2147483648.
	y = p * float64(c.p8) / 32768.
	return float32(p+(x+y+float64(c.p7))/16.) / 1000.
}

// compensateHumidityFloat returns humidity in %rH. Output value of "46.332"
// represents 46.332 %rH.
//
// raw has 16 bits of resolution.
func (c *calibration) compensateHumidityFloat(raw, tFine int32) float32 {
	h := float64(tFine - 76800)
	h = (float64(raw) - float64(c.h4)*64. + float64(c.h5)/16384.*h) * float64(c.h2) / 65536. * (1. + float64(c.h6)/67108864.*h*(1.+float64(c.h3)/67108864.*h))
	h *= 1. - float64(c.h1)*h/524288.
	if h > 100. {
		return 100.
	}
	if h < 0. {
		return 0.
	}
	return float32(h)
}

type spiFail struct {
	spitest.Playback
}

func (s *spiFail) DevParams(maxHz int64, mode spi.Mode, bits int) error {
	return errors.New("failing")
}
