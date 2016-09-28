// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package bme280

import (
	"fmt"
	"log"
	"testing"

	"github.com/google/pio/conn/i2c"
	"github.com/google/pio/conn/i2c/i2ctest"
	"github.com/google/pio/devices"
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

func TestRead(t *testing.T) {
	// This data was generated with "bme280 -r"
	bus := i2ctest.Playback{
		Ops: []i2ctest.IO{
			// Chipd ID detection.
			{Addr: 0x76, Write: []byte{0xd0}, Read: []byte{0x60}},
			// Calibration data.
			{
				Addr:  0x76,
				Write: []byte{0x88},
				Read:  []byte{0x10, 0x6e, 0x6c, 0x66, 0x32, 0x0, 0x5d, 0x95, 0xb8, 0xd5, 0xd0, 0xb, 0x77, 0x1e, 0x9d, 0xff, 0xf9, 0xff, 0xac, 0x26, 0xa, 0xd8, 0xbd, 0x10, 0x0, 0x4b},
			},
			// Calibration data.
			{Addr: 0x76, Write: []byte{0xe1}, Read: []byte{0x6e, 0x1, 0x0, 0x13, 0x5, 0x0, 0x1e}},
			// Configuration.
			{Addr: 0x76, Write: []byte{0xf4, 0x6c, 0xf2, 0x3, 0xf5, 0xe0, 0xf4, 0x6f}, Read: nil},
			// Read.
			{Addr: 0x76, Write: []byte{0xf7}, Read: []byte{0x4a, 0x52, 0xc0, 0x80, 0x96, 0xc0, 0x7a, 0x76}},
		},
	}
	dev, err := NewI2C(&bus, nil)
	if err != nil {
		t.Fatal(err)
	}
	env := devices.Environment{}
	if err := dev.Sense(&env); err != nil {
		t.Fatalf("Sense(): %v", err)
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

//

func Example() {
	bus, err := i2c.New(-1)
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
