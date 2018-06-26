// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic

import (
	"strconv"
	"time"
)

// Distance is a measurement of length stored as an int64 nano metre.
//
// This is one of the base unit in the International System of Units.
//
// The highest representable value is 9.2Gm.
type Distance int64

// String returns the distance formatted as a string in metre.
func (d Distance) String() string {
	return nanoAsString(int64(d)) + "m"
}

// Distance constants.
const (
	NanoMetre  Distance = 1
	MicroMetre          = 1000 * NanoMetre
	MilliMetre          = 1000 * MicroMetre
	Metre               = 1000 * MilliMetre
	KiloMetre           = 1000 * Metre
	MegaMetre           = 1000 * KiloMetre

	// Conversion between Metre and imperial units.
	Thou = 25400 * NanoMetre
	Inch = 1000 * Thou
	Foot = 12 * Inch
	Yard = 3 * Foot
	Mile = 1760 * Yard
)

// ElectricCurrent is a measurement of a flow of electric charge stored as an
// int64 nano Ampere.
//
// This is one of the base unit in the International System of Units.
//
// The highest representable value is 9.2GA.
type ElectricCurrent int64

// String returns the current formatted as a string in Ampere.
func (e ElectricCurrent) String() string {
	return nanoAsString(int64(e)) + "A"
}

// ElectricCurrent constants.
const (
	NanoAmpere  ElectricCurrent = 1
	MicroAmpere                 = 1000 * NanoAmpere
	MilliAmpere                 = 1000 * MicroAmpere
	Ampere                      = 1000 * MilliAmpere
)

// ElectricPotential is a measurement of electric potential stored as an int64
// nano Volt.
//
// The highest representable value is 9.2GV.
type ElectricPotential int64

// String returns the tension formatted as a string in Volt.
func (e ElectricPotential) String() string {
	return nanoAsString(int64(e)) + "V"
}

// ElectricPotential constants.
const (
	// Volt is W/A, kg⋅m²/s³/A.
	NanoVolt  ElectricPotential = 1
	MicroVolt                   = 1000 * NanoVolt
	MilliVolt                   = 1000 * MicroVolt
	Volt                        = 1000 * MilliVolt
	KiloVolt                    = 1000 * Volt
)

// ElectricResistance is a measurement of the difficulty to pass an electric
// current through a conductor stored as an int64 nano Ohm.
//
// The highest representable value is 9.2GΩ.
type ElectricResistance int64

// String returns the resistance formatted as a string in Ohm.
func (e ElectricResistance) String() string {
	return nanoAsString(int64(e)) + "Ω"
}

// ElectricResistance constants.
const (
	// Ohm is V/A, kg⋅m²/s³/A².
	NanoOhm  ElectricResistance = 1
	MicroOhm                    = 1000 * NanoOhm
	MilliOhm                    = 1000 * MicroOhm
	Ohm                         = 1000 * MilliOhm
	KiloOhm                     = 1000 * Ohm
	MegaOhm                     = 1000 * KiloOhm
)

// Force is a measurement of interaction that will change the motion of an
// object stored as an int64 nano Newton.
//
// A measurement of Force is a vector and has a direction but this unit only
// represents the magnitude. The orientation needs to be stored as a Quaternion
// independently.
//
// The highest representable value is 9.2TN.
type Force int64

// String returns the force formatted as a string in Newton.
func (f Force) String() string {
	return nanoAsString(int64(f)) + "N"
}

// Force constants.
const (
	// Newton is kg⋅m/s².
	NanoNewton  Force = 1
	MicroNewton       = 1000 * NanoNewton
	MilliNewton       = 1000 * MicroNewton
	Newton            = 1000 * MilliNewton
	KiloNewton        = 1000 * Newton
	MegaNewton        = 1000 * KiloNewton

	EarthGravity = 9806650 * MicroNewton

	// Conversion between Newton and imperial units.
	// Pound is both a unit of mass and weight (force). The suffix Mass is added
	// to disambiguate the measurement it represents.
	PoundForce = 4448221615261 * NanoNewton
)

// Frequency is a measurement of cycle per second, stored as an int32 micro
// Hertz.
//
// The highest representable value is 9.2THz.
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

// Frequency constants.
const (
	// Hertz is 1/s.
	MicroHertz Frequency = 1
	MilliHertz           = 1000 * MicroHertz
	Hertz                = 1000 * MilliHertz
	KiloHertz            = 1000 * Hertz
	MegaHertz            = 1000 * KiloHertz
	GigaHertz            = 1000 * MegaHertz
)

// Mass is a measurement of mass stored as an int64 nano gram.
//
// This is one of the base unit in the International System of Units.
//
// The highest representable value is 9.2Gg.
type Mass int64

// String returns the mass formatted as a string in gram.
func (m Mass) String() string {
	return nanoAsString(int64(m)) + "g"
}

// Mass constants.
const (
	NanoGram  Mass = 1
	MicroGram      = 1000 * NanoGram
	MilliGram      = 1000 * MicroGram
	Gram           = 1000 * MilliGram
	KiloGram       = 1000 * Gram
	MegaGram       = 1000 * KiloGram
	Tonne          = MegaGram

	// Conversion between Gram and imperial units.
	// Ounce is both a unit of mass, weight (force) or volume depending on
	// context. The suffix Mass is added to disambiguate the measurement it
	// represents.
	OunceMass = 28349523125 * NanoGram
	// Pound is both a unit of mass and weight (force). The suffix Mass is added
	// to disambiguate the measurement it represents.
	PoundMass = 16 * OunceMass
	Slug      = 14593903 * MilliGram
)

// Pressure is a measurement of force applied to a surface per unit
// area (stress) stored as an int64 nano Pascal.
//
// The highest representable value is 9.2GPa.
type Pressure int64

// String returns the pressure formatted as a string in Pascal.
func (p Pressure) String() string {
	return nanoAsString(int64(p)) + "Pa"
}

// Pressure constants.
const (
	// Pascal is N/m², kg/m/s².
	NanoPascal  Pressure = 1
	MicroPascal          = 1000 * NanoPascal
	MilliPascal          = 1000 * MicroPascal
	Pascal               = 1000 * MilliPascal
	KiloPascal           = 1000 * Pascal
)

// RelativeHumidity is a humidity level measurement stored as an int32 fixed
// point integer at a precision of 0.00001%rH.
//
// Valid values are between 0 and 10000000.
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

// RelativeHumidity constants.
const (
	TenthMicroRH RelativeHumidity = 1                 // 0.00001%rH
	MicroRH                       = 10 * TenthMicroRH // 0.0001%rH
	MilliRH                       = 1000 * MicroRH    // 0.1%rH
	PercentRH                     = 10 * MilliRH      // 1%rH
)

// Speed is a measurement of magnitude of velocity stored as an int64 nano
// Metre per Second.
//
// The highest representable value is 9.2Gm/s.
type Speed int64

// String returns the speed formatted as a string in m/s.
func (s Speed) String() string {
	return nanoAsString(int64(s)) + "m/s"
}

// Speed constants.
const (
	// MetrePerSecond is m/s.
	NanoMetrePerSecond  Speed = 1
	MicroMetrePerSecond       = 1000 * NanoMetrePerSecond
	MilliMetrePerSecond       = 1000 * MicroMetrePerSecond
	MetrePerSecond            = 1000 * MilliMetrePerSecond
	KiloMetrePerSecond        = 1000 * MetrePerSecond
	MegaMetrePerSecond        = 1000 * KiloMetrePerSecond

	LightSpeed = 299792458 * MetrePerSecond

	KilometrePerHour = 3600 * MilliMetrePerSecond
	MilePerHour      = 447040 * MicroMetrePerSecond
	FootPerSecond    = 304800 * MicroMetrePerSecond
)

// Temperature is a measurement of hotness stored as a nano kelvin.
//
// Negative values are invalid.
//
// The highest representable value is 9.2GK.
type Temperature int64

// String returns the temperature formatted as a string in °Celsius.
func (t Temperature) String() string {
	return nanoAsString(int64(t-ZeroCelsius)) + "°C"
}

// Temperature constants.
const (
	NanoKelvin  Temperature = 1
	MicroKelvin             = 1000 * NanoKelvin
	MilliKelvin             = 1000 * MicroKelvin
	Kelvin                  = 1000 * MilliKelvin

	// Conversion between Kelvin and Celsius.
	ZeroCelsius  = 273150 * MilliKelvin
	MilliCelsius = MilliKelvin
	Celsius      = Kelvin

	// Conversion between Kelvin and Fahrenheit.
	ZeroFahrenheit  = 255372 * MilliKelvin
	MilliFahrenheit = 555555 * NanoKelvin
	Fahrenheit      = 555555555 * NanoKelvin
)

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

// nanoAsString converts a value in S.I. unit in a string with the predefined
// prefix.
func nanoAsString(v int64) string {
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
		unit = "G"
	case v >= 1000000000000000:
		frac = int(v % 1000000000000000 / 1000000000000)
		base = int(v / 1000000000000000)
		unit = "M"
	case v >= 1000000000000:
		frac = int(v % 1000000000000 / 1000000000)
		base = int(v / 1000000000000)
		unit = "k"
	case v >= 1000000000:
		frac = int(v % 1000000000 / 1000000)
		base = int(v / 1000000000)
		unit = ""
	case v >= 1000000:
		frac = int(v % 1000000 / 1000)
		base = int(v / 1000000)
		unit = "m"
	case v >= 1000:
		frac = int(v) % 1000
		base = int(v) / 1000
		unit = "µ"
	default:
		if v == 0 {
			return "0"
		}
		base = int(v)
		unit = "n"
	}
	if frac == 0 {
		return sign + strconv.Itoa(base) + unit
	}
	return sign + strconv.Itoa(base) + "." + prefixZeros(3, frac) + unit
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
