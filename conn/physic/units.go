// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// Angle is the measurement of the difference in orientation between two vectors
// stored as an int64 nano radian.
//
// A negative angle is valid.
//
// The highest representable value is a bit over 500,000,000,000°.
type Angle int64

// String returns the angle formatted as a string in degree.
func (a Angle) String() string {
	// Angle is not a S.I. unit, so it must not be prefixed by S.I. prefixes.
	if a == 0 {
		return "0°"
	}
	// Round.
	prefix := ""
	if a < 0 {
		a = -a
		prefix = "-"
	}
	switch {
	case a < Degree:
		v := ((a * 1000) + Degree/2) / Degree
		return prefix + "0." + prefixZeros(3, int(v)) + "°"
	case a < 10*Degree:
		v := ((a * 1000) + Degree/2) / Degree
		i := v / 1000
		v = v - i*1000
		return prefix + strconv.FormatInt(int64(i), 10) + "." + prefixZeros(3, int(v)) + "°"
	case a < 100*Degree:
		v := ((a * 1000) + Degree/2) / Degree
		i := v / 1000
		v = v - i*1000
		return prefix + strconv.FormatInt(int64(i), 10) + "." + prefixZeros(2, int(v)) + "°"
	case a < 1000*Degree:
		v := ((a * 1000) + Degree/2) / Degree
		i := v / 1000
		v = v - i*1000
		return prefix + strconv.FormatInt(int64(i), 10) + "." + prefixZeros(1, int(v)) + "°"
	default:
		v := (a + Degree/2) / Degree
		return prefix + strconv.FormatInt(int64(v), 10) + "°"
	}
}

const (
	NanoRadian  Angle = 1
	MicroRadian Angle = 1000 * NanoRadian
	MilliRadian Angle = 1000 * MicroRadian
	Radian      Angle = 1000 * MilliRadian

	// Theta is 2π. This is equivalent to 360°.
	Theta  Angle = 6283185307 * NanoRadian
	Pi     Angle = 3141592653 * NanoRadian
	Degree Angle = 17453293 * NanoRadian
)

const (
	maxDegrees int64 = (maxInt64 - 500000000) / int64(Degree)
	maxPis     int64 = (maxInt64 - 500000000) / int64(Pi)
)

// Set sets the Angle to the value represented by s. Units can be provided in
// are Degree, Deg, Rad, Radians and pi with optional SI prefixes.
func (a *Angle) Set(s string) error {
	dec, n, err := atod(s)
	if err != nil {
		return err
	}
	prefix := prefix(none)
	if !(n == len(s)) {
		var n1 int
		prefix, n1 = parseSIPrefix([]rune(s[n:])[0])
		// Check for units that could be confused with prefixes.
		if prefix == pico {
			switch strings.ToLower(s[n:]) {
			case "pi", "π":
				prefix = none
			}
		}
		if prefix != none {
			n += n1
		}
	}
	v, err := dtoi(dec, int(prefix-nano))
	if err != nil {
		return err
	}
	switch strings.ToLower(s[n:]) {
	case "°", "deg", "degrees", "d":
		if v > maxDegrees || v < -maxDegrees {
			return &parseError{err: errors.New("value too large")}
		}
		if v >= 0 {
			*a = (Angle)((v*int64(Degree) + 500000000) / 1000000000)
		} else {
			*a = (Angle)((v*int64(Degree) - 500000000) / 1000000000)
		}
	case "pi", "π":
		if v > maxPis || v < -maxPis {
			return &parseError{err: errors.New("value too large")}
		}
		if v >= 0 {
			*a = (Angle)((v*int64(Pi) + 500000000) / 1000000000)
		} else {
			*a = (Angle)((v*int64(Pi) - 500000000) / 1000000000)
		}
	case "rad", "radians", "radian", "r":
		*a = (Angle)(v)
	default:
		return noUnits("Radian")
	}
	return nil
}

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

const (
	NanoMetre  Distance = 1
	MicroMetre Distance = 1000 * NanoMetre
	MilliMetre Distance = 1000 * MicroMetre
	Metre      Distance = 1000 * MilliMetre
	KiloMetre  Distance = 1000 * Metre
	MegaMetre  Distance = 1000 * KiloMetre
	GigaMetre  Distance = 1000 * MegaMetre

	// Conversion between Metre and imperial units.
	Thou Distance = 25400 * NanoMetre
	Inch Distance = 1000 * Thou
	Foot Distance = 12 * Inch
	Yard Distance = 3 * Foot
	Mile Distance = 1760 * Yard
)

const (
	maxMiles  int64 = (maxInt64 - 500) / (int64(Mile) / 1000000) // ~Max/1609344
	maxYards  int64 = (maxInt64 - 5000) / (int64(Yard) / 100000) // ~Max/9144
	maxFeet   int64 = (maxInt64 - 5000) / (int64(Foot) / 100000) // ~Max/3048
	maxInches int64 = (maxInt64 - 5000) / (int64(Inch) / 100000) // ~Max/254
)

// Set sets the Distance to the value represented by s. Units can be provided in
// are miles, metres, feet, inches and yards with optional SI prefixes.
func (d *Distance) Set(s string) error {
	// Only default base metric unit has full range all other units are limited,
	// as any conversion calculations must fit inside a int64. Conversion math
	// should be ordered so as to maintain as much fidelity as possible.
	dec, n, err := atod(s)
	if err != nil {
		return err
	}
	prefix := prefix(none)
	if !(n == len(s)) {
		var n1 int
		prefix, n1 = parseSIPrefix([]rune(s[n:])[0])
		// Check for units that could be confused with prefixes.
		if prefix == milli {
			switch strings.ToLower(s[n:]) {
			case "mile", "metre", "miles", "metres", "m":
				prefix = none
			}
		}
		if prefix == mega {
			switch strings.ToLower(s[n:]) {
			case "mile", "metre", "miles", "metres", "m":
				prefix = none
			}
		}
		if prefix != none {
			n += n1
		}
	}
	v, err := dtoi(dec, int(prefix-nano))
	if err != nil {
		return err
	}
	switch strings.ToLower(s[n:]) {
	case "mile", "miles":
		if v > maxMiles || v < -maxMiles {
			return &parseError{err: errors.New("too big use metric")}
		}
		if v >= 0 {
			*d = (Distance)((v*1609344 + 500) / 1000)
		} else {
			*d = (Distance)((v*1609344 - 500) / 1000)
		}
	case "yard", "yards":

		if v > maxYards || v < -maxYards {
			return &parseError{err: errors.New("too big use metric")}
		}
		if v >= 0 {
			*d = (Distance)((v*9144 + 5000) / 10000)
		} else {
			*d = (Distance)((v*9144 - 5000) / 10000)
		}
	case "foot", "feet", "ft":
		if v > maxFeet || v < -maxFeet {
			return &parseError{err: errors.New("too big use metric")}
		}
		if v >= 0 {
			*d = (Distance)((v*3048 + 5000) / 10000)
		} else {
			*d = (Distance)((v*3048 - 5000) / 10000)
		}
	case "in", "inch", "inches":
		if v > maxInches || v < -maxInches {
			return &parseError{err: errors.New("too big use metric")}
		}
		if v >= 0 {
			*d = (Distance)((v*254 + 5000) / 10000)
		} else {
			*d = (Distance)((v*254 - 5000) / 10000)
		}
	case "m", "metre", "metres":
		*d = (Distance)(v)
	default:
		return noUnits("m, Metre, Mile, Inch, Foot or Yard")
	}
	return nil
}

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

// Set sets the ElectricCurrent to the value represented by s. Units can be
// provided in amps with optional SI prefixes.
func (e *ElectricCurrent) Set(s string) error {
	v, n, err := set(s, nano)
	if err != nil {
		return err
	}

	switch strings.ToLower(s[n:]) {
	case "a", "amp", "amps", "ampere":
		*e = (ElectricCurrent)(v)
	default:
		return noUnits("A")
	}
	return nil
}

const (
	NanoAmpere  ElectricCurrent = 1
	MicroAmpere ElectricCurrent = 1000 * NanoAmpere
	MilliAmpere ElectricCurrent = 1000 * MicroAmpere
	Ampere      ElectricCurrent = 1000 * MilliAmpere
	KiloAmpere  ElectricCurrent = 1000 * Ampere
	MegaAmpere  ElectricCurrent = 1000 * KiloAmpere
	GigaAmpere  ElectricCurrent = 1000 * MegaAmpere
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

// Set sets the ElectricPotential to the value represented by s. Units are to be
// provided in volts with optional SI prefixes.
func (e *ElectricPotential) Set(s string) error {
	v, n, err := set(s, nano)
	if err != nil {
		return err
	}

	switch strings.ToLower(s[n:]) {
	case "v", "volt", "volts":
		*e = (ElectricPotential)(v)
	default:
		return noUnits("V")
	}
	return nil
}

const (
	// Volt is W/A, kg⋅m²/s³/A.
	NanoVolt  ElectricPotential = 1
	MicroVolt ElectricPotential = 1000 * NanoVolt
	MilliVolt ElectricPotential = 1000 * MicroVolt
	Volt      ElectricPotential = 1000 * MilliVolt
	KiloVolt  ElectricPotential = 1000 * Volt
	MegaVolt  ElectricPotential = 1000 * KiloVolt
	GigaVolt  ElectricPotential = 1000 * MegaVolt
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

// Set sets the ElectricResistance to the value represented by s. Units are to be
// provided in ohms with optional SI prefixes.
func (e *ElectricResistance) Set(s string) error {
	v, n, err := set(s, nano)
	if err != nil {
		return err
	}

	switch strings.ToLower(s[n:]) {
	case "ohm", "ohms", "ω": // ω is lowercase Ω.
		*e = (ElectricResistance)(v)
	default:
		return noUnits("Ohm")
	}
	return nil
}

const (
	// Ohm is V/A, kg⋅m²/s³/A².
	NanoOhm  ElectricResistance = 1
	MicroOhm ElectricResistance = 1000 * NanoOhm
	MilliOhm ElectricResistance = 1000 * MicroOhm
	Ohm      ElectricResistance = 1000 * MilliOhm
	KiloOhm  ElectricResistance = 1000 * Ohm
	MegaOhm  ElectricResistance = 1000 * KiloOhm
	GigaOhm  ElectricResistance = 1000 * MegaOhm
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

// Set sets the Force to the value represented by s. Units are to be provided in
// newtons with optional SI prefixes.
func (f *Force) Set(s string) error {
	d, n, err := atod(s)
	if err != nil {
		return err
	}
	prefix := prefix(none)
	if !(n == len(s)) {
		var n1 int
		prefix, n1 = parseSIPrefix([]rune(s[n:])[0])
		if prefix == nano && !(s[n:] == "nN") {
			prefix = none
		}
		if prefix != none {
			n += n1
		}
	}
	v, err := dtoi(d, int(prefix-nano))
	if err != nil {
		return err
	}

	switch strings.ToLower(s[n:]) {
	case "n", "newton", "newtons":
		*f = (Force)(v)
	default:
		return noUnits("N")
	}
	return nil
}

const (
	// Newton is kg⋅m/s².
	NanoNewton  Force = 1
	MicroNewton Force = 1000 * NanoNewton
	MilliNewton Force = 1000 * MicroNewton
	Newton      Force = 1000 * MilliNewton
	KiloNewton  Force = 1000 * Newton
	MegaNewton  Force = 1000 * KiloNewton
	GigaNewton  Force = 1000 * MegaNewton

	EarthGravity Force = 9806650 * MicroNewton

	// Conversion between Newton and imperial units.
	// Pound is both a unit of mass and weight (force). The suffix Force is added
	// to disambiguate the measurement it represents.
	PoundForce Force = 4448221615261 * NanoNewton
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

// Set sets the Frequency to the value represented by s. Units are to be
// provided in hertz with optional SI prefixes.
func (f *Frequency) Set(s string) error {
	v, n, err := set(s, micro)
	if err != nil {
		return err
	}

	switch strings.ToLower(s[n:]) {
	case "hz", "hertz":
		*f = (Frequency)(v)
	default:
		return noUnits("Hz")
	}
	return nil
}

// Duration returns the duration of one cycle at this frequency.
func (f Frequency) Duration() time.Duration {
	// Note: Duration() should have been named Period().
	// TODO(maruel): Rounding should be fine-tuned.
	return time.Second * time.Duration(Hertz) / time.Duration(f)
}

// PeriodToFrequency returns the frequency for a period of this interval.
func PeriodToFrequency(t time.Duration) Frequency {
	return Frequency(time.Second) * Hertz / Frequency(t)
}

const (
	// Hertz is 1/s.
	MicroHertz Frequency = 1
	MilliHertz Frequency = 1000 * MicroHertz
	Hertz      Frequency = 1000 * MilliHertz
	KiloHertz  Frequency = 1000 * Hertz
	MegaHertz  Frequency = 1000 * KiloHertz
	GigaHertz  Frequency = 1000 * MegaHertz
	TeraHertz  Frequency = 1000 * GigaHertz
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

// Set sets the Mass to the value represented by s. Units are to be provided
// in grams, ounces, pounds or tonnes with optional SI prefixes.
func (m *Mass) Set(s string) error {
	d, n, err := atod(s)
	if err != nil {
		return err
	}
	prefix := prefix(none)
	if !(n == len(s)) {
		var n1 int
		prefix, n1 = parseSIPrefix([]rune(s[n:])[0])
		if prefix == giga {
			switch strings.ToLower(s[n:]) {
			case "gram", "grams", "g":
				prefix = none
			}
		}
		if prefix == tera {
			switch strings.ToLower(s[n:]) {
			case "tonne", "tonnes":
				prefix = none
			}
		}
		if prefix != none {
			n += n1
		}
	}
	v, err := dtoi(d, int(prefix-nano))
	if err != nil {
		return err
	}
	// TODO(NeuralSpaz): Rounding
	switch strings.ToLower(s[n:]) {
	case "tonne", "tonnes":
		//TODO(NeuralSpaz): Max Tonne
		*m = (Mass)(v * 1000000)
	case "pound", "pounds", "lb":
		//TODO(NeuralSpaz): Max lb
		if v >= 0 {
			*m = (Mass)((v*45359237 + 50000) / 100000)
		} else {
			*m = (Mass)((v*45359237 - 50000) / 100000)
		}
	case "ounce", "ounces", "oz":
		// Scales v up to maintain precision of oz/gram.
		// Re-ordering of operations to stay within the range of int64.
		//TODO(NeuralSpaz): Max oz
		if v >= 0 {
			if v < 325344874 {
				*m = (Mass)(((v*28349523125 + 500000000) / 1000000000))
			} else {
				*m = (Mass)(((v + 500000000) / 1000000000) * 28349523125)
			}
		} else {
			if v > -325344874 {
				*m = (Mass)(((v*28349523125 - 500000000) / 1000000000))
			} else {
				*m = (Mass)(((v - 500000000) / 1000000000) * 28349523125)
			}
		}
	case "g", "gram", "grams":
		*m = (Mass)(v)
	default:
		return noUnits("g")
	}
	return nil
}

const (
	NanoGram  Mass = 1
	MicroGram Mass = 1000 * NanoGram
	MilliGram Mass = 1000 * MicroGram
	Gram      Mass = 1000 * MilliGram
	KiloGram  Mass = 1000 * Gram
	MegaGram  Mass = 1000 * KiloGram
	GigaGram  Mass = 1000 * MegaGram
	Tonne     Mass = MegaGram

	// Conversion between Gram and imperial units.
	// Ounce is both a unit of mass, weight (force) or volume depending on
	// context. The suffix Mass is added to disambiguate the measurement it
	// represents.
	OunceMass Mass = 28349523125 * NanoGram
	// Pound is both a unit of mass and weight (force). The suffix Mass is added
	// to disambiguate the measurement it represents.
	PoundMass Mass = 16 * OunceMass
	Slug      Mass = 14593903 * MilliGram
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

// Set sets the Pressure to the value represented by s. Units are to be provided
// in pascals with optional SI prefixes.
func (p *Pressure) Set(s string) error {
	d, n, err := atod(s)
	if err != nil {
		return err
	}
	prefix := prefix(none)
	if !(n == len(s)) {
		var n1 int
		prefix, n1 = parseSIPrefix([]rune(s[n:])[0])
		if prefix == pico {
			prefix = none
		}
		if prefix != none {
			n += n1
		}
	}
	v, err := dtoi(d, int(prefix-nano))
	if err != nil {
		return err
	}
	// TODO(NeuralSpaz): Add support for hg, hPa and psi
	switch strings.ToLower(s[n:]) {
	case "pa", "pascal", "pascals":
		*p = (Pressure)(v)
	default:
		return noUnits("Pa")
	}
	return nil
}

const (
	// Pascal is N/m², kg/m/s².
	NanoPascal  Pressure = 1
	MicroPascal Pressure = 1000 * NanoPascal
	MilliPascal Pressure = 1000 * MicroPascal
	Pascal      Pressure = 1000 * MilliPascal
	KiloPascal  Pressure = 1000 * Pascal
	MegaPascal  Pressure = 1000 * KiloPascal
	GigaPascal  Pressure = 1000 * MegaPascal
)

// RelativeHumidity is a humidity level measurement stored as an int32 fixed
// point integer at a precision of 0.00001%rH.
//
// Valid values are between 0% and 100%.
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

// Set sets the RelativeHumidity to the value represented by s. Units are to be
// provided in relative humidity or percent relative humidity with optional
// SI prefixes.
func (r *RelativeHumidity) Set(s string) error {
	v, n, err := set(s, prefix(-7))
	if err != nil {
		return err
	}
	switch strings.ToLower(s[n:]) {
	case "rh":
		*r = (RelativeHumidity)(v)
	case "%", "%rh":
		if v >= 0 {
			*r = (RelativeHumidity)((v + 50) / 100)
		} else {
			*r = (RelativeHumidity)((v - 50) / 100)
		}
	default:
		return noUnits("%rH")
	}
	return nil
}

const (
	TenthMicroRH RelativeHumidity = 1                 // 0.00001%rH
	MicroRH      RelativeHumidity = 10 * TenthMicroRH // 0.0001%rH
	MilliRH      RelativeHumidity = 1000 * MicroRH    // 0.1%rH
	PercentRH    RelativeHumidity = 10 * MilliRH      // 1%rH
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

// Set sets the Speed to the value represented by s. Units are to be provided in
// fps, mph, km/h or m/s with optional SI prefixes.
func (s *Speed) Set(str string) error {
	d, n, err := atod(str)
	if err != nil {
		return err
	}
	prefix := prefix(none)
	if !(n == len(str)) {
		var n1 int
		prefix, n1 = parseSIPrefix([]rune(str[n:])[0])
		if prefix == milli && !(str[n:] == "mm/s") && !(str[n:] == "mfps") {
			prefix = none
		}
		if prefix == kilo && !(str[n:] == "km/s") {
			prefix = none
		}
		if prefix != none {
			n += n1
		}
	}
	v, err := dtoi(d, int(prefix-nano))
	if err != nil {
		return err
	}

	switch strings.ToLower(str[n:]) {
	case "fps":
		//TODO(NeuralSpaz): Max
		if v >= 0 {
			*s = (Speed)(((v*3048 + 5000) / 10000))
		} else {
			*s = (Speed)(((v*3048 - 5000) / 10000))
		}
	case "mph":
		//TODO(NeuralSpaz): Max
		if v >= 0 {
			*s = (Speed)(((v*44704 + 50000) / 100000))
		} else {
			*s = (Speed)(((v*44704 - 50000) / 100000))
		}
	case "km/h":
		//TODO(NeuralSpaz): Max
		if v >= 0 {
			*s = (Speed)((v*10 + 18) / 36)
		} else {
			*s = (Speed)((v*10 - 18) / 36)
		}
	case "m/s":
		*s = (Speed)(v)
	default:
		return noUnits("m/s")
	}
	return nil
}

const (
	// MetrePerSecond is m/s.
	NanoMetrePerSecond  Speed = 1
	MicroMetrePerSecond Speed = 1000 * NanoMetrePerSecond
	MilliMetrePerSecond Speed = 1000 * MicroMetrePerSecond
	MetrePerSecond      Speed = 1000 * MilliMetrePerSecond
	KiloMetrePerSecond  Speed = 1000 * MetrePerSecond
	MegaMetrePerSecond  Speed = 1000 * KiloMetrePerSecond
	GigaMetrePerSecond  Speed = 1000 * MegaMetrePerSecond

	LightSpeed Speed = 299792458 * MetrePerSecond

	KilometrePerHour Speed = 277777778 * NanoMetrePerSecond
	MilePerHour      Speed = 447040 * MicroMetrePerSecond
	FootPerSecond    Speed = 304800 * MicroMetrePerSecond
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

// Set sets the Temperature to the value represented by s. Units are to be
// provided in kelvin or celsius with optional SI prefixes.
func (t *Temperature) Set(s string) error {
	v, n, err := set(s, nano)
	if err != nil {
		return err
	}
	switch strings.ToLower(s[n:]) {
	case "k", "kelvin":
		*t = (Temperature)(v)
	case "c", "°c", "celsius":
		*t = (Temperature)(v + int64(ZeroCelsius))
	case "f", "fahrenheit":
		if v > 18446744000000 {
			// Maximum F is ~18446.744F
			return &parseError{s: s, err: errors.New("maximum is 18446.744F")}
		}
		*t = (Temperature)(((5*v + 2298350000000) + 4) / 9)
	default:
		return noUnits("K, C or F")
	}
	return nil
}

const (
	NanoKelvin  Temperature = 1
	MicroKelvin Temperature = 1000 * NanoKelvin
	MilliKelvin Temperature = 1000 * MicroKelvin
	Kelvin      Temperature = 1000 * MilliKelvin
	KiloKelvin  Temperature = 1000 * Kelvin
	MegaKelvin  Temperature = 1000 * KiloKelvin
	GigaKelvin  Temperature = 1000 * MegaKelvin

	// Conversion between Kelvin and Celsius.
	ZeroCelsius  Temperature = 273150 * MilliKelvin
	MilliCelsius Temperature = MilliKelvin
	Celsius      Temperature = Kelvin

	// Conversion between Kelvin and Fahrenheit.
	ZeroFahrenheit  Temperature = 255372222222 * NanoKelvin
	MilliFahrenheit Temperature = 555555 * NanoKelvin
	Fahrenheit      Temperature = 555555556 * NanoKelvin
)

// Power is a measurement of power stored as a nano watts.
//
// The highest representable value is 9.2GW.
type Power int64

// String returns the power formatted as a string in watts.
func (p Power) String() string {
	return nanoAsString(int64(p)) + "W"
}

// Set sets the Power to the value represented by s. Units are to be provided in
// watts with optional SI prefixes.
func (p *Power) Set(s string) error {
	v, n, err := set(s, nano)
	if err != nil {
		return err
	}

	switch strings.ToLower(s[n:]) {
	case "watt", "watts", "w":
		*p = (Power)(v)
	default:
		return noUnits("W")
	}
	return nil
}

const (
	// Watt is unit of power J/s, kg⋅m²⋅s⁻³
	NanoWatt  Power = 1
	MicroWatt Power = 1000 * NanoWatt
	MilliWatt Power = 1000 * MicroWatt
	Watt      Power = 1000 * MilliWatt
	KiloWatt  Power = 1000 * Watt
	MegaWatt  Power = 1000 * KiloWatt
	GigaWatt  Power = 1000 * MegaWatt
)

// Energy is a measurement of work stored as a nano joules.
//
// The highest representable value is 9.2GJ.
type Energy int64

// String returns the energy formatted as a string in Joules.
func (e Energy) String() string {
	return nanoAsString(int64(e)) + "J"
}

// Set sets the Energy to the value represented by s. Units are to be provided
// in joules with optional SI prefixes.
func (e *Energy) Set(s string) error {
	v, n, err := set(s, nano)
	if err != nil {
		return err
	}

	switch strings.ToLower(s[n:]) {
	case "joule", "joules", "j":
		*e = (Energy)(v)
	default:
		return noUnits("J")
	}
	return nil
}

const (
	// Joule is a unit of work. kg⋅m²⋅s⁻²
	NanoJoule  Energy = 1
	MicroJoule Energy = 1000 * NanoJoule
	MilliJoule Energy = 1000 * MicroJoule
	Joule      Energy = 1000 * MilliJoule
	KiloJoule  Energy = 1000 * Joule
	MegaJoule  Energy = 1000 * KiloJoule
	GigaJoule  Energy = 1000 * MegaJoule
	// TODO(NeuralSpaz): Add Calorie support.
)

// ElectricalCapacitance is a measurement of capacitance stored as a pico farad.
//
// The highest representable value is 9.2MF.
type ElectricalCapacitance int64

// String returns the energy formatted as a string in Farad.
func (c ElectricalCapacitance) String() string {
	return picoAsString(int64(c)) + "F"
}

// Set sets the ElectricalCapacitance to the value represented by s. Units are
// to be provided in farads with optional SI prefixes.
func (e *ElectricalCapacitance) Set(s string) error {
	v, n, err := set(s, pico)
	if err != nil {
		return err
	}

	switch strings.ToLower(s[n:]) {
	case "f", "farad", "farads":
		*e = (ElectricalCapacitance)(v)
	default:
		return noUnits("F")
	}
	return nil
}

const (
	// Farad is a unit of capacitance. kg⁻¹⋅m⁻²⋅s⁴A²
	PicoFarad  ElectricalCapacitance = 1
	NanoFarad  ElectricalCapacitance = 1000 * PicoFarad
	MicroFarad ElectricalCapacitance = 1000 * NanoFarad
	MilliFarad ElectricalCapacitance = 1000 * MicroFarad
	Farad      ElectricalCapacitance = 1000 * MilliFarad
	KiloFarad  ElectricalCapacitance = 1000 * Farad
	MegaFarad  ElectricalCapacitance = 1000 * KiloFarad
)

// LuminousIntensity is a measurement of the quantity of visible light energy
// emitted per unit solid angle with wavelength power weighted by a luminosity
// function which represents the human eye's response to different wavelengths.
// The CIE 1931 luminosity function is the SI standard for candela.
//
// LuminousIntensity is stored as nano candela.
//
// This is one of the base unit in the International System of Units.
//
// The highest representable value is 9.2Gcd.
type LuminousIntensity int64

// String returns the energy formatted as a string in Candela.
func (l LuminousIntensity) String() string {
	return nanoAsString(int64(l)) + "cd"
}

// Set sets the LuminousIntensity to the value represented by s. Units are to be
// provided in candelas with optional SI prefixes.
func (l *LuminousIntensity) Set(s string) error {
	v, n, err := set(s, nano)
	if err != nil {
		return err
	}

	switch strings.ToLower(s[n:]) {
	case "cd", "candela", "candelas":
		*l = (LuminousIntensity)(v)
	default:
		return noUnits("cd")
	}
	return nil
}

const (
	// Candela is a unit of luminous intensity. cd
	NanoCandela  LuminousIntensity = 1
	MicroCandela LuminousIntensity = 1000 * NanoCandela
	MilliCandela LuminousIntensity = 1000 * MicroCandela
	Candela      LuminousIntensity = 1000 * MilliCandela
	KiloCandela  LuminousIntensity = 1000 * Candela
	MegaCandela  LuminousIntensity = 1000 * KiloCandela
	GigaCandela  LuminousIntensity = 1000 * MegaCandela
)

// LuminousFlux is a measurement of total quantity of visible light energy
// emitted with wavelength power weighted by a luminosity function which
// represents a model of the human eye's response to different wavelengths.
// The CIE 1931 luminosity function is the standard for lumens.
//
// LuminousFlux is stored as nano lumens.
//
// The highest representable value is 9.2Glm.
type LuminousFlux int64

// String returns the energy formatted as a string in Lumens.
func (f LuminousFlux) String() string {
	return nanoAsString(int64(f)) + "lm"
}

// Set sets the LuminousFlux to the value represented by s. Units are to be
// provided in lumens with optional SI prefixes.
func (l *LuminousFlux) Set(s string) error {
	v, n, err := set(s, nano)
	if err != nil {
		return err
	}

	switch strings.ToLower(s[n:]) {
	case "lm", "lumen", "lumens":
		*l = (LuminousFlux)(v)
	default:
		return noUnits("lm")
	}
	return nil
}

const (
	// Lumen is a unit of luminous flux. cd⋅sr
	NanoLumen  LuminousFlux = 1
	MicroLumen LuminousFlux = 1000 * NanoLumen
	MilliLumen LuminousFlux = 1000 * MicroLumen
	Lumen      LuminousFlux = 1000 * MilliLumen
	KiloLumen  LuminousFlux = 1000 * Lumen
	MegaLumen  LuminousFlux = 1000 * KiloLumen
	GigaLumen  LuminousFlux = 1000 * MegaLumen
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
	var frac int
	var base int
	var precision int64
	unit := ""
	switch {
	case v >= 999999500000000001:
		precision = v % 1000000000000000
		base = int(v / 1000000000000000)
		if precision > 500000000000000 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
		unit = "G"
	case v >= 999999500000001:
		precision = v % 1000000000000
		base = int(v / 1000000000000)
		if precision > 500000000000 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
		unit = "M"
	case v >= 999999500001:
		precision = v % 1000000000
		base = int(v / 1000000000)
		if precision > 500000000 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
		unit = "k"
	case v >= 999999501:
		precision = v % 1000000
		base = int(v / 1000000)
		if precision > 500000 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
		unit = ""
	case v >= 1000000:
		precision = v % 1000
		base = int(v / 1000)
		if precision > 500 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
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
	var frac int
	var base int
	var precision int64
	unit := ""
	switch {
	case v >= 999999500000000001:
		precision = v % 1000000000000000
		base = int(v / 1000000000000000)
		if precision > 500000000000000 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
		unit = "T"
	case v >= 999999500000001:
		precision = v % 1000000000000
		base = int(v / 1000000000000)
		if precision > 500000000000 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
		unit = "G"
	case v >= 999999500001:
		precision = v % 1000000000
		base = int(v / 1000000000)
		if precision > 500000000 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
		unit = "M"
	case v >= 999999501:
		precision = v % 1000000
		base = int(v / 1000000)
		if precision > 500000 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
		unit = "k"
	case v >= 1000000:
		precision = v % 1000
		base = int(v / 1000)
		if precision > 500 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
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

// picoAsString converts a value in S.I. unit in a string with the predefined
// prefix.
func picoAsString(v int64) string {
	sign := ""
	if v < 0 {
		if v == -9223372036854775808 {
			v++
		}
		sign = "-"
		v = -v
	}
	var frac int
	var base int
	var precision int64
	unit := ""
	switch {
	case v >= 999999500000000001:
		precision = v % 1000000000000000
		base = int(v / 1000000000000000)
		if precision > 500000000000000 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
		unit = "M"
	case v >= 999999500000001:
		precision = v % 1000000000000
		base = int(v / 1000000000000)
		if precision > 500000000000 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
		unit = "k"
	case v >= 999999500001:
		precision = v % 1000000000
		base = int(v / 1000000000)
		if precision > 500000000 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
		unit = ""
	case v >= 999999501:
		precision = v % 1000000
		base = int(v / 1000000)
		if precision > 500000 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
		unit = "m"
	case v >= 1000000:
		precision = v % 1000
		base = int(v / 1000)
		if precision > 500 {
			base++
		}
		frac = (base % 1000)
		base = base / 1000
		unit = "µ"
	case v >= 1000:
		frac = int(v) % 1000
		base = int(v) / 1000
		unit = "n"
	default:
		if v == 0 {
			return "0"
		}
		base = int(v)
		unit = "p"
	}
	if frac == 0 {
		return sign + strconv.Itoa(base) + unit
	}
	return sign + strconv.Itoa(base) + "." + prefixZeros(3, frac) + unit
}

// Decimal is the exact representation of decimal number.
type decimal struct {
	// digits hold the string representation of the significant decimal digits.
	digits string
	// exponent is the left or right decimal shift. (powers of ten).
	exp int
	// neg it true if the number is negative.
	neg bool
}

// Positive powers of 10 in the form such that powerOF10[index] = 10^index.
var powerOf10 = [...]uint64{
	1,
	10,
	100,
	1000,
	10000,
	100000,
	1000000,
	10000000,
	100000000,
	1000000000,
	10000000000,
	100000000000,
	1000000000000,
	10000000000000,
	100000000000000,
	1000000000000000,
	10000000000000000,
	100000000000000000,
	1000000000000000000,
}

// Maximum value for a int64.
const maxInt64 = (1<<63 - 1)

var maxUint64Str = "9223372036854775807"

// Converts from decimal to int64, using the decimal.digits character values and
// converting to a intermediate unit64.
// Scale is combined with the decimal exponent to maximise the resolution and is
// in powers of ten.
func dtoi(d decimal, scale int) (int64, error) {
	// Use uint till the last as it allows checks for overflows.
	var u uint64
	for i := 0; i < len(d.digits); i++ {
		// Check that is is a digit.
		if d.digits[i] >= '0' && d.digits[i] <= '9' {
			// '0' = 0x30 '1' = 0x31 ...etc.
			digit := d.digits[i] - '0'
			// *10 is decimal shift left.
			u *= 10
			check := u + uint64(digit)
			// Check should always be larger than u unless we have overflowed.
			// Similarly if check > max it will overflow when converted to int64.
			if check < u || check > maxInt64 {
				if d.neg {
					return -maxInt64, &parseError{
						s:   "-" + maxUint64Str,
						err: errors.New("overflows minimum is"),
					}
				}
				return maxInt64, &parseError{
					s:   maxUint64Str,
					err: errors.New("overflows maximum is"),
				}
			}
			u = check
		} else {
			// Should not get here if used atod to generate the decimal.
			return 0, &parseError{err: errors.New("not a number")}
		}
	}

	// Get the total magnitude of the number.
	// a^x * b^y = a*b^(x+y) since scale is of the order unity this becomes
	// 1^x * b^y = b^(x+y).
	// mag must be positive to use as index in to powerOf10 array.
	mag := d.exp + scale
	if mag < 0 {
		mag *= -1
	}
	if mag > 18 {
		return 0, errors.New("exponent exceeds int64")
	}
	// Divide is = 10^(-mag)
	if d.exp+scale < 0 {
		u = (u + powerOf10[mag]/2) / powerOf10[mag]
	} else {
		check := u * powerOf10[mag]
		if check < u || check > maxInt64 {
			if d.neg {
				return -maxInt64, &parseError{
					s:   "-" + maxUint64Str,
					err: errors.New("overflows minimum is"),
				}
			}
			return maxInt64, &parseError{
				s:   maxUint64Str,
				err: errors.New("overflows maximum is"),
			}
		}
		u *= powerOf10[mag]
	}

	n := int64(u)
	if d.neg {
		n *= -1
	}
	return n, nil
}

// Converts a string to a decimal form. The return int is how many bytes of the
// string are numeric. The string may contain +-0 prefixes and arbitrary
// suffixes as trailing non number characters are ignored.
// Significant digits are stored without leading or trailing zeros, rather an
// exponent is used.
func atod(s string) (decimal, int, error) {
	var d decimal
	start := 0
	dp := 0
	end := len(s)
	seenDigit := false
	seenZero := false
	isPoint := false
	seenPlus := false

	// Strip leading zeros, +/- and mark DP.
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '-':
			if seenDigit {
				end = i
				break
			}
			if seenPlus {
				return decimal{}, 0, &parseError{
					s:   s,
					err: errors.New("can't contain both plus and minus symbols"),
				}
			}
			if d.neg {
				return decimal{}, 0, &parseError{
					s:   s,
					err: errors.New("multiple minus symbols"),
				}
			}
			d.neg = true
			start++
		case s[i] == '+':
			if seenDigit {
				end = i
				break
			}
			if d.neg {
				return decimal{}, 0, &parseError{
					s:   s,
					err: errors.New("can't contain both plus and minus symbols"),
				}
			}
			if seenPlus {
				return decimal{}, 0, &parseError{
					s:   s,
					err: errors.New("multiple plus symbols"),
				}
			}
			seenPlus = true
			start++
		case s[i] == '.':
			if isPoint {
				return decimal{}, 0, &parseError{
					s:   s,
					err: errors.New("multiple decimal points"),
				}
			}
			isPoint = true
			dp = i
			if !seenDigit {
				start++
			}
		case s[i] == '0':
			if !seenDigit {
				start++
			}
			seenZero = true
		case s[i] >= '1' && s[i] <= '9':
			seenDigit = true
		default:
			if !seenDigit && !seenZero {
				return decimal{}, 0, &parseError{
					s:   s,
					err: errors.New("is not a number"),
				}
			}
			end = i
			break
		}
	}

	last := end
	seenDigit = false
	exp := 0
	// Strip non significant zeros to find base exponent.
	for i := end - 1; i > start-1; i-- {
		switch {
		case s[i] >= '1' && s[i] <= '9':
			seenDigit = true
		case s[i] == '.':
			if !seenDigit {
				end--
			}
		case s[i] == '0':
			if !seenDigit {
				if i > dp {
					end--
				}
				if i <= dp || dp == 0 {
					exp++
				}
			}
		default:
			last--
			end--
		}
	}

	if dp > start && dp < end {
		// Concatenate with out decimal point.
		d.digits = s[start:dp] + s[dp+1:end]
	} else {
		d.digits = s[start:end]
	}
	if !isPoint {
		d.exp = exp
	} else {
		ttl := dp - start
		length := len(d.digits)
		if ttl > 0 {
			d.exp = ttl - length
		} else {
			d.exp = ttl - length + 1
		}
	}
	return d, last, nil
}

// Set is a helper for converting a string and a prefix in to a physic unit. It
// can be used when characters of the units do not conflict with any of the SI
// prefixes.
func set(s string, base prefix) (int64, int, error) {
	d, n, err := atod(s)
	if err != nil {
		return 0, n, err
	}
	si := prefix(none)
	if !(n == len(s)) {
		var n1 int
		si, n1 = parseSIPrefix([]rune(s[n:])[0])
		if si != none {
			n += n1
		}
	}
	v, err := dtoi(d, int(si-base))
	if err != nil {
		return v, 0, err
	}
	return v, n, nil
}

type parseError struct {
	s        string
	position int
	err      error
}

func (p *parseError) Error() string {
	if p.err == nil {
		return "parse error"
	}
	if p.s == "" {
		return "parse error: " + p.err.Error()
	}
	return "parse error: " + p.err.Error() + ": \"" + p.s + "\""
}

func noUnits(s string) error {
	return &parseError{s: s, err: errors.New("no units provided, need")}
}

type prefix int

const (
	pico  prefix = -12
	nano  prefix = -9
	micro prefix = -6
	milli prefix = -3
	none  prefix = 0
	deca  prefix = 1
	hecto prefix = 2
	kilo  prefix = 3
	mega  prefix = 6
	giga  prefix = 9
	tera  prefix = 12
)

// take a prefix rune and converts it to an int. For use with indexing directly
// into a string length of the prefix rune is also returned.
func parseSIPrefix(r rune) (prefix, int) {
	switch r {
	case 'p':
		return pico, len("p")
	case 'n':
		return nano, len("n")
	case 'u':
		return micro, len("u")
	case 'µ':
		return micro, len("µ")
	case 'm':
		return milli, len("m")
	case 'k':
		return kilo, len("k")
	case 'M':
		return mega, len("M")
	case 'G':
		return giga, len("G")
	case 'T':
		return tera, len("T")
	default:
		return none, 0
	}
}
