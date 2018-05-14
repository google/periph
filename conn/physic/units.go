// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic

import (
	"strconv"
	"time"
)

const (
	MicroAmpere ElectricCurrent = 1
	MilliAmpere                 = 1000 * MicroAmpere
	Ampere                      = 1000 * MilliAmpere

	MicroVolt ElectricPotential = 1
	MilliVolt                   = 1000 * MicroVolt
	Volt                        = 1000 * MilliVolt
	KiloVolt                    = 1000 * Volt

	MicroHertz Frequency = 1
	MilliHertz           = 1000 * MicroHertz
	Hertz                = 1000 * MilliHertz
	KiloHertz            = 1000 * Hertz
	MegaHertz            = 1000 * KiloHertz
	GigaHertz            = 1000 * MegaHertz

	MicroPascal Pressure = 1
	MilliPascal          = 1000 * MicroPascal
	Pascal               = 1000 * MilliPascal
	KiloPascal           = 1000 * Pascal

	MicroRH   RelativeHumidity = 1
	MilliRH                    = 1000 * MicroRH
	PercentRH                  = 10 * MilliRH

	MicroKelvin Temperature = 1
	MilliKelvin             = 1000 * MicroKelvin
	Kelvin                  = 1000 * MilliKelvin

	// Conversion between Kelvin and Celsius.
	ZeroCelsius  = 273150 * MilliKelvin
	MilliCelsius = MilliKelvin
	Celsius      = Kelvin

	// Conversion between Kelvin and Fahrenheit.
	ZeroFahrenheit  = 255372 * MilliKelvin
	MilliFahrenheit = 555 * MicroKelvin
	Fahrenheit      = 555555 * MicroKelvin
)

// ElectricCurrent is a measurement of a flow of electric charge as an int64
// micro Ampere.
type ElectricCurrent int64

// String returns the current formatted as a string in Ampere.
func (e ElectricCurrent) String() string {
	return microAsString(int64(e)) + "A"
}

// ElectricPotential is a measurement of electric potential stored as micro
// Volt.
type ElectricPotential int64

// String returns the tension formatted as a string in Volt.
func (e ElectricPotential) String() string {
	return microAsString(int64(e)) + "V"
}

// Frequency is a measurement of cycle per second, stored as micro Hertz.
type Frequency int64

// String returns the frequency formatted as a string in Hertz.
func (f Frequency) String() string {
	return microAsString(int64(f)) + "Hz"
}

// Duration returns the duration of one cycle at this frequency.
func (f Frequency) Duration() time.Duration {
	return time.Second * time.Duration(Hertz) / time.Duration(f)
}

// PeriodToFrequency returns the frequency for a period of this interval.
func PeriodToFrequency(t time.Duration) Frequency {
	return Frequency(time.Second) * Hertz / Frequency(t)
}

// Pressure is a measurement of stress stored as micro Pascal.
type Pressure int64

// String returns the pressure formatted as a string in Pascal.
func (p Pressure) String() string {
	return microAsString(int64(p)) + "Pa"
}

// RelativeHumidity is a humidity level measurement stored as a fixed point
// integer at a precision of 0.0001%rH.
//
// Valid values are between 0 and 1000000.
type RelativeHumidity int32

// String returns the humidity formatted as a string.
func (r RelativeHumidity) String() string {
	r /= MilliRH
	frac := int(r % 10)
	if frac == 0 {
		return strconv.Itoa(int(r)/10) + "%rH"
	}
	if frac < 0 {
		frac = -frac
	}
	return strconv.Itoa(int(r)/10) + "." + strconv.Itoa(frac) + "%rH"
}

// Temperature is a measurement of hotness stored as a micro kelvin.
//
// Negative values are invalid.
type Temperature int64

// String returns the temperature formatted as a string in °Celsius.
func (t Temperature) String() string {
	return microAsString(int64(t-ZeroCelsius)) + "°C"
}

//

func prefixZeros(digits, v int) string {
	// digits is expected to be around 2~3.
	s := strconv.Itoa(v)
	for len(s) < digits {
		// O(n²) but since digits is expected to run 2~3 times at most, it doesn't
		// matter.
		s = "0" + s
	}
	return s
}

// microAsString converts a value in S.I. unit in a string with the predefined
// prefix.
func microAsString(v int64) string {
	sign := ""
	if v < 0 {
		if v == -9223372036854775808 {
			v++
		}
		sign = "-"
		v = -v
	}
	// TODO(maruel): Round a bit.
	var frac int
	var base int
	unit := ""
	switch {
	case v >= 1000000000000000000:
		frac = int(v % 1000000000000000000 / 1000000000000000)
		base = int(v / 1000000000000000000)
		unit = "T"
	case v >= 1000000000000000:
		frac = int(v % 1000000000000000 / 1000000000000)
		base = int(v / 1000000000000000)
		unit = "G"
	case v >= 1000000000000:
		frac = int(v % 1000000000000 / 1000000000)
		base = int(v / 1000000000000)
		unit = "M"
	case v >= 1000000000:
		frac = int(v % 1000000000 / 1000000)
		base = int(v / 1000000000)
		unit = "k"
	case v >= 1000000:
		frac = int(v % 1000000 / 1000)
		base = int(v / 1000000)
		unit = ""
	case v >= 1000:
		frac = int(v) % 1000
		base = int(v) / 1000
		unit = "m"
	default:
		if v == 0 {
			return "0"
		}
		base = int(v)
		unit = "µ"
	}
	if frac == 0 {
		return sign + strconv.Itoa(base) + unit
	}
	return sign + strconv.Itoa(base) + "." + prefixZeros(3, frac) + unit
}
