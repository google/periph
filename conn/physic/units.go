// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic

import (
	"strconv"
)

// Centi is a fixed point value with a 0.01 precision.
type Centi int32

// Float64 returns the value as float64.
func (c Centi) Float64() float64 {
	return float64(c) * .01
}

// String returns the value formatted as a string.
func (c Centi) String() string {
	d := c % 100
	if d < 0 {
		d = -d
	}
	return strconv.Itoa(int(c)/100) + "." + prefixZeros(2, int(d))
}

// Milli is a fixed point value with a 0.001 precision.
type Milli int32

// Float64 returns the value as float64.
func (m Milli) Float64() float64 {
	return float64(m) * .001
}

// String returns the value formatted as a string.
func (m Milli) String() string {
	d := m % 1000
	if d < 0 {
		d = -d
	}
	return strconv.Itoa(int(m)/1000) + "." + prefixZeros(3, int(d))
}

//

// Ampere is a current measurement stored as a fixed point integer at a
// precision of 1mA.
type Ampere Milli

// Float64 returns the value as float64.
func (a Ampere) Float64() float64 {
	return Milli(a).Float64()
}

// String returns the current formatted as a string.
//
// For small values, it is printed in mA unit.
func (a Ampere) String() string {
	if a < 1000 && a > -1000 {
		return strconv.Itoa(int(a)) + "mA"
	}
	return Milli(a).String() + "A"
}

// Celsius is a temperature measurement stored as a fixed point integer at a
// precision of 0.001°C.
type Celsius Milli

// Float64 returns the value as float64.
func (c Celsius) Float64() float64 {
	return Milli(c).Float64()
}

// String returns the temperature formatted as a string.
func (c Celsius) String() string {
	return Milli(c).String() + "°C"
}

// ToF returns the temperature in Fahrenheit, an unsound unit used in the
// United States.
func (c Celsius) ToF() Fahrenheit {
	return Fahrenheit((c*9+2)/5 + 32000)
}

// Fahrenheit is a temperature measurement stored as a fixed point integer at a
// precision of 0.001°F.
type Fahrenheit Milli

// Float64 returns the value as float64.
func (f Fahrenheit) Float64() float64 {
	return Milli(f).Float64()
}

// String returns the temperature formatted as a string.
func (f Fahrenheit) String() string {
	return Milli(f).String() + "°F"
}

// KPascal is a pressure measurement stored as a fixed point integer at a
// precision of 1Pa.
type KPascal Milli

// Float64 returns the value as float64.
func (k KPascal) Float64() float64 {
	return Milli(k).Float64()
}

// String returns the pressure formatted as a string.
func (k KPascal) String() string {
	return Milli(k).String() + "KPa"
}

// RelativeHumidity is a humidity level measurement stored as a fixed point
// integer at a precision of 0.01%rH.
type RelativeHumidity Centi

// Float64 returns the value as float64.
func (r RelativeHumidity) Float64() float64 {
	return Centi(r).Float64()
}

// String returns the humidity formatted as a string.
func (r RelativeHumidity) String() string {
	return Centi(r).String() + "%rH"
}

// Volt is a tension measurement stored as a fixed point integer at a
// precision of 1mV.
type Volt Milli

// Float64 returns the value as float64.
func (v Volt) Float64() float64 {
	return Milli(v).Float64()
}

// String returns the tension formatted as a string.
//
// For small values, it is printed in mV unit.
func (v Volt) String() string {
	if v < 1000 && v > -1000 {
		return strconv.Itoa(int(v)) + "mV"
	}
	return Milli(v).String() + "V"
}

//

func prefixZeros(digits, v int) string {
	// digits is expected to be around 2~3.
	s := strconv.Itoa(v)
	for len(s) < digits {
		// O(n²) but since digits is expected to run 2~3 times at most, it doesn't matter.
		s = "0" + s
	}
	return s
}
