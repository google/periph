// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package devices

import (
	"strconv"
)

// Milli is a fixed point value with 0.001 precision.
//
// Deprecated: This interface will be removed in v3.
type Milli int32

// Float64 returns the value as float64 with 0.001 precision.
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

// Celsius is a temperature at a precision of 0.001°C.
//
// Deprecated: This interface will be removed in v3. Use physic.Temperature
// instead.
type Celsius Milli

// Float64 returns the value as float64 with 0.001 precision.
func (c Celsius) Float64() float64 {
	return Milli(c).Float64()
}

// String returns the temperature formatted as a string.
func (c Celsius) String() string {
	return Milli(c).String() + "°C"
}

// ToF returns the temperature as Fahrenheit, a unit used in the United States.
func (c Celsius) ToF() Fahrenheit {
	return Fahrenheit((c*9+2)/5 + 32000)
}

// Fahrenheit is an unsound unit used in the United States.
//
// Deprecated: This interface will be removed in v3. Use physic.Temperature
// instead.
type Fahrenheit Milli

// Float64 returns the value as float64 with 0.001 precision.
func (f Fahrenheit) Float64() float64 {
	return Milli(f).Float64()
}

// String returns the temperature formatted as a string.
func (f Fahrenheit) String() string {
	return Milli(f).String() + "°F"
}

// KPascal is pressure at precision of 1Pa.
//
// Deprecated: This interface will be removed in v3. Use physic.Pressure
// instead.
type KPascal Milli

// Float64 returns the value as float64 with 0.001 precision.
func (k KPascal) Float64() float64 {
	return Milli(k).Float64()
}

// String returns the pressure formatted as a string.
func (k KPascal) String() string {
	return Milli(k).String() + "KPa"
}

// RelativeHumidity is humidity level in %rH with 0.01%rH precision.
//
// Deprecated: This interface will be removed in v3. Use
// physic.RelativeHumidity instead.
type RelativeHumidity int32

// Float64 returns the value in %.
func (r RelativeHumidity) Float64() float64 {
	return float64(r) * .01
}

// String returns the humidity formatted as a string.
func (r RelativeHumidity) String() string {
	m := r % 100
	if m < 0 {
		m = -m
	}
	return strconv.Itoa(int(r)/100) + "." + prefixZeros(2, int(m)) + "%rH"
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
