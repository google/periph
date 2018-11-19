// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
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
	case a > (9223372036854775807 - 17453293):
		u := (uint64(a) + uint64(Degree)/2) / uint64(Degree)
		v := int64(u)
		return prefix + strconv.FormatInt(int64(v), 10) + "°"
	default:
		v := (a + Degree/2) / Degree
		return prefix + strconv.FormatInt(int64(v), 10) + "°"
	}
}

// Set sets the Angle to the value represented by s. Units are to be provided in
// "rad", "deg" or "°" with an optional SI prefix: "p", "n", "u", "µ", "m", "k",
// "M", "G" or "T".
func (f *Angle) Set(s string) error {
	d, n, err := atod(s)
	if err != nil {
		if e, ok := err.(*parseError); ok {
			switch e.err {
			case errNotANumber:
				if found, _ := containsUnitString(s[n:], "rad", "deg", "°"); found != "" {
					return errors.New("does not contain number")
				}
				return errors.New("does not contain number or unit \"Rad\"")
			case errOverflowsInt64:
				return errors.New("maximum value is " + maxAngle.String())
			case errOverflowsInt64Negative:
				return errors.New("minimum value is " + minAngle.String())
			}
		}
		return err
	}

	var si prefix
	if n != len(s) {
		r, rsize := utf8.DecodeRuneInString(s[n:])
		if r <= 1 || rsize == 0 {
			return &parseError{
				err: errors.New("unexpected end of string"),
			}
		}
		var siSize int
		si, siSize = parseSIPrefix(r)
		n += siSize
	}

	switch s[n:] {
	case "deg", "°", "Deg":
		degreePerRadian := decimal{
			base: 17453293,
			exp:  0,
			neg:  false,
		}
		lbf, _ := decimalMulScale(d, degreePerRadian)
		// Impossible for precision loss to exceed 9 since the number of
		// significant figures in degrees per radian is only 8.
		v, err := dtoi(lbf, int(si))
		if err != nil {
			if err != nil {
				if e, ok := err.(*parseError); ok {
					switch e.err {
					case errOverflowsInt64:
						return errors.New("maximum value is " + maxAngle.String())
					case errOverflowsInt64Negative:
						return errors.New("minimum value is " + minAngle.String())
					}
				}
				return err
			}
		}
		*f = (Angle)(v)
	case "rad", "Rad":
		v, err := dtoi(d, int(si-nano))
		if err != nil {
			if err != nil {
				if e, ok := err.(*parseError); ok {
					switch e.err {
					case errOverflowsInt64:
						return errors.New("maximum value is " + maxAngle.String())
					case errOverflowsInt64Negative:
						return errors.New("minimum value is " + minAngle.String())
					}
				}
				return err
			}
		}
		*f = (Angle)(v)
	case "":
		return noUnits("Rad")
	default:
		if found, extra := containsUnitString(s[n:], "Rad", "Deg", "°"); found != "" {
			return unknownUnitPrefix(found, extra, "p,n,u,µ,m,k,M,G or T")
		}
		return incorrectUnit(s[n:], "physic.Angle")
	}
	return nil
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

	maxAngle = 9223372036854775807 * NanoRadian
	minAngle = -9223372036854775807 * NanoRadian
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

// Set sets the Distance to the value represented by s. Units are to
// be provided in "m", "Mile", "Yard", "in", or "ft" with an optional SI
// prefix: "p", "n", "u", "µ", "m", "k", "M", "G" or "T".
func (d *Distance) Set(s string) error {
	decimal, n, err := atod(s)
	if err != nil {
		if e, ok := err.(*parseError); ok {
			switch e.err {
			case errNotANumber:
				if found, _ := containsUnitString(s[n:], "in", "ft", "Yard", "Mile", "m"); found != "" {
					return errors.New("does not contain number")
				}
				return errors.New("does not contain number or unit \"m\"")
			case errOverflowsInt64:
				return errors.New("maximum value is " + maxDistance.String())
			case errOverflowsInt64Negative:
				return errors.New("minimum value is " + minDistance.String())

			}
		}
		return err
	}
	si := prefix(unit)
	if n != len(s) {
		r, rsize := utf8.DecodeRuneInString(s[n:])
		if r <= 1 || rsize == 0 {
			return &parseError{
				err: errors.New("unexpected end of string"),
			}
		}
		var siSize int
		si, siSize = parseSIPrefix(r)
		if si == milli || si == mega {
			switch s[n:] {
			case "m", "Mile", "mile":
				si = unit
			}
		}
		if si != unit {
			n += siSize
		}
	}
	v, err := dtoi(decimal, int(si-nano))
	if err != nil {
		if err != nil {
			if e, ok := err.(*parseError); ok {
				switch e.err {
				case errOverflowsInt64:
					return errors.New("maximum value is " + maxDistance.String())
				case errOverflowsInt64Negative:
					return errors.New("minimum value is " + minDistance.String())
				}
			}

			return err
		}
	}
	switch s[n:] {
	case "m":
		*d = (Distance)(v)
	case "mile", "Mile":
		switch {
		case v > maxMiles:
			return errors.New("maximum value is 5731Mile")
		case v < minMiles:
			return errors.New("minimum value is -5731Mile")
		case v >= 0:
			*d = (Distance)((v*1609344 + 500) / 1000)
		default:
			*d = (Distance)((v*1609344 - 500) / 1000)
		}
	case "yard", "Yard":
		switch {
		case v > maxYards:
			return errors.New("maximum value is 1 Million Yard")
		case v < minYards:
			return errors.New("minimum value is -1 Million Yard")
		case v >= 0:
			*d = (Distance)((v*9144 + 5000) / 10000)
		default:
			*d = (Distance)((v*9144 - 5000) / 10000)
		}
	case "ft":
		switch {
		case v > maxFeet:
			return errors.New("maximum value is 3 Million ft")
		case v < minFeet:
			return errors.New("minimum value is 3 Million ft")
		case v >= 0:
			*d = (Distance)((v*3048 + 5000) / 10000)
		default:
			*d = (Distance)((v*3048 - 5000) / 10000)
		}
	case "in":
		switch {
		case v > maxInches:
			return errors.New("maximum value is 36 Million inch")
		case v < minInches:
			return errors.New("minimum value is 36 Million inch")
		case v >= 0:
			*d = (Distance)((v*254 + 5000) / 10000)
		default:
			*d = (Distance)((v*254 - 5000) / 10000)
		}
	case "":
		return noUnits("m, Mile, in, ft or Yard")
	default:
		if found, extra := containsUnitString(s[n:], "in", "ft", "Yard", "Mile", "m"); found != "" {
			return unknownUnitPrefix(found, extra, "p,n,u,µ,m,k,M,G or T")
		}
		return incorrectUnit(s[n:], "physic.Distance")
	}
	return nil
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

	maxDistance       = 9223372036854775807 * NanoMetre
	minDistance       = -9223372036854775807 * NanoMetre
	maxMiles    int64 = (int64(maxDistance) - 500) / int64((Mile)/1000000) // ~Max/1609344
	minMiles    int64 = (int64(minDistance) + 500) / int64((Mile)/1000000) // ~Min/1609344
	maxYards    int64 = (int64(maxDistance) - 5000) / int64((Yard)/100000) // ~Max/9144
	minYards    int64 = (int64(minDistance) + 5000) / int64((Yard)/100000) // ~Min/9144
	maxFeet     int64 = (int64(maxDistance) - 5000) / int64((Foot)/100000) // ~Max/3048
	minFeet     int64 = (int64(minDistance) + 5000) / int64((Foot)/100000) // ~Min/3048
	maxInches   int64 = (int64(maxDistance) - 5000) / int64((Inch)/100000) // ~Max/254
	minInches   int64 = (int64(minDistance) + 5000) / int64((Inch)/100000) // ~Min/254
)

// ElectricCurrent is a measurement of a flow of electric charge stored as an
// int64 nano Ampere.
//
// This is one of the base unit in the International System of Units.
//
// The highest representable value is 9.2GA.
type ElectricCurrent int64

// String returns the current formatted as a string in Ampere.
func (c ElectricCurrent) String() string {
	return nanoAsString(int64(c)) + "A"
}

// Set sets the ElectricCurrent to the value represented by s. Units are to
// be provided in "A" with an optional SI prefix: "p", "n", "u", "µ", "m", "k",
// "M", "G" or "T".
func (c *ElectricCurrent) Set(s string) error {
	v, n, err := valueOfUnitString(s, nano)
	if err != nil {
		if e, ok := err.(*parseError); ok {
			switch e.err {
			case errOverflowsInt64:
				return errors.New("maximum value is " + maxElectricCurrent.String())
			case errOverflowsInt64Negative:
				return errors.New("minimum value is " + minElectricCurrent.String())
			case errNotANumber:
				if found, _ := containsUnitString(s, "A"); found != "" {
					return errors.New("does not contain number")
				}
				return errors.New("does not contain number or unit \"A\"")
			}
		}
		return err
	}

	switch s[n:] {
	case "a", "A":
		*c = (ElectricCurrent)(v)
	case "":
		return noUnits("A")
	default:
		if found, extra := containsUnitString(s[n:], "A"); found != "" {
			return unknownUnitPrefix(found, extra, "p,n,u,µ,m,k,M,G or T")
		}
		return incorrectUnit(s[n:], "physic.ElectricCurrent")
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

	maxElectricCurrent = 9223372036854775807 * NanoAmpere
	minElectricCurrent = -9223372036854775807 * NanoAmpere
)

// ElectricPotential is a measurement of electric potential stored as an int64
// nano Volt.
//
// The highest representable value is 9.2GV.
type ElectricPotential int64

// String returns the tension formatted as a string in Volt.
func (p ElectricPotential) String() string {
	return nanoAsString(int64(p)) + "V"
}

// Set sets the ElectricPotential to the value represented by s. Units are to
// be provided in "V" with an optional SI prefix: "p", "n", "u", "µ", "m", "k",
// "M", "G" or "T".
func (p *ElectricPotential) Set(s string) error {
	v, n, err := valueOfUnitString(s, nano)
	if err != nil {
		if e, ok := err.(*parseError); ok {
			switch e.err {
			case errOverflowsInt64:
				return errors.New("maximum value is " + maxElectricPotential.String())
			case errOverflowsInt64Negative:
				return errors.New("minimum value is " + minElectricPotential.String())
			case errNotANumber:
				if found, _ := containsUnitString(s, "V"); found != "" {
					return errors.New("does not contain number")
				}
				return errors.New("does not contain number or unit \"V\"")
			}
		}
		return err
	}
	switch s[n:] {
	case "v", "V":
		*p = (ElectricPotential)(v)
	case "":
		return noUnits("V")
	default:
		if found, extra := containsUnitString(s[n:], "V"); found != "" {
			return unknownUnitPrefix(found, extra, "p,n,u,µ,m,k,M,G or T")
		}
		return incorrectUnit(s[n:], "physic.ElectricPotential")
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

	maxElectricPotential = 9223372036854775807 * NanoVolt
	minElectricPotential = -9223372036854775807 * NanoVolt
)

// ElectricResistance is a measurement of the difficulty to pass an electric
// current through a conductor stored as an int64 nano Ohm.
//
// The highest representable value is 9.2GΩ.
type ElectricResistance int64

// String returns the resistance formatted as a string in Ohm.
func (r ElectricResistance) String() string {
	return nanoAsString(int64(r)) + "Ω"
}

// Set sets the ElectricResistance to the value represented by s. Units are to
// be provided in "Ohm", or "Ω" with an optional SI prefix: "p", "n", "u", "µ",
// "m", "k", "M", "G" or "T".
func (r *ElectricResistance) Set(s string) error {
	v, n, err := valueOfUnitString(s, nano)
	if err != nil {
		if e, ok := err.(*parseError); ok {
			switch e.err {
			case errOverflowsInt64:
				return errors.New("maximum value is " + maxElectricResistance.String())
			case errOverflowsInt64Negative:
				return errors.New("minimum value is " + minElectricResistance.String())
			case errNotANumber:
				if found, _ := containsUnitString(s, "Ohm", "Ω"); found != "" {
					return errors.New("does not contain number")
				}
				return errors.New("does not contain number or unit \"Ohm\"")
			}
		}
		return err
	}

	switch s[n:] {
	case "ohm", "Ohm", "Ω":
		*r = (ElectricResistance)(v)
	case "":
		return noUnits("Ohm")
	default:
		if found, extra := containsUnitString(s[n:], "Ohm", "Ω"); found != "" {
			return unknownUnitPrefix(found, extra, "p,n,u,µ,m,k,M,G or T")
		}
		return incorrectUnit(s[n:], "physic.ElectricResistance")
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

	maxElectricResistance = 9223372036854775807 * NanoOhm
	minElectricResistance = -9223372036854775807 * NanoOhm
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

// Set sets the Force to the value represented by s. Units are to
// be provided in "N", or "lbf" (Pound force) with an optional SI prefix: "p",
// "n", "u", "µ", "m", "k", "M", "G" or "T".
func (f *Force) Set(s string) error {
	d, n, err := atod(s)
	if err != nil {
		if e, ok := err.(*parseError); ok {
			switch e.err {
			case errNotANumber:
				if found, _ := containsUnitString(s[n:], "N", "lbf"); found != "" {
					return errors.New("does not contain number")
				}
				return errors.New("does not contain number or unit \"N\" or \"lbf\"")
			case errOverflowsInt64:
				return errors.New("maximum value is " + maxForce.String())
			case errOverflowsInt64Negative:
				return errors.New("minimum value is " + minForce.String())

			}
		}
		return err
	}

	var si prefix
	if n != len(s) {
		r, rsize := utf8.DecodeRuneInString(s[n:])
		if r <= 1 || rsize == 0 {
			return &parseError{
				err: errors.New("unexpected end of string"),
			}
		}
		var siSize int
		si, siSize = parseSIPrefix(r)

		n += siSize
	}

	switch s[n:] {
	case "lbf":
		poundForce := decimal{
			base: 4448221615261,
			exp:  0,
			neg:  false,
		}
		lbf, loss := decimalMulScale(d, poundForce)
		if loss > 9 {
			return errors.New("converting to nano Newtons would overflow, consider using nN for maximum precision")
		}
		v, err := dtoi(lbf, int(si))
		if err != nil {
			if err != nil {
				if e, ok := err.(*parseError); ok {
					switch e.err {
					case errOverflowsInt64:
						return errors.New("maximum value is 2.073496Mlbf")
					case errOverflowsInt64Negative:
						return errors.New("minimum value is -2.073496Mlbf")
					}
				}
				return err
			}
		}
		*f = (Force)(v)
	case "N":
		v, err := dtoi(d, int(si-nano))
		if err != nil {
			if err != nil {
				if e, ok := err.(*parseError); ok {
					switch e.err {
					case errOverflowsInt64:
						return errors.New("maximum value is " + maxForce.String())
					case errOverflowsInt64Negative:
						return errors.New("minimum value is " + minForce.String())
					}
				}
				return err
			}
		}
		*f = (Force)(v)
	case "":
		return noUnits("N")
	default:
		if found, extra := containsUnitString(s[n:], "N"); found != "" {
			return unknownUnitPrefix(found, extra, "p,n,u,µ,m,k,M,G or T")
		}
		return incorrectUnit(s[n:], "physic.Force")
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

	maxForce = 9223372036854775807 * NanoNewton
	minForce = -9223372036854775807 * NanoNewton
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

// Set sets the Frequency to the value represented by s. Units are to
// be provided in "Hz" with an optional SI prefix: "p", "n", "u", "µ", "m", "k",
// "M", "G" or "T".
func (f *Frequency) Set(s string) error {
	v, n, err := valueOfUnitString(s, micro)
	if err != nil {
		if e, ok := err.(*parseError); ok {
			switch e.err {
			case errOverflowsInt64:
				return errors.New("maximum value is " + maxFrequency.String())
			case errOverflowsInt64Negative:
				return errors.New("minimum value is " + minFrequency.String())
			case errNotANumber:
				if found, _ := containsUnitString(s, "Hz"); found != "" {
					return errors.New("does not contain number")
				}
				return errors.New("does not contain number or unit \"Hz\"")
			}
		}
		return err
	}

	switch s[n:] {
	case "hz", "Hz":
		*f = (Frequency)(v)
	case "":
		return noUnits("Hz")
	default:
		if found, extra := containsUnitString(s[n:], "Hz"); found != "" {
			return unknownUnitPrefix(found, extra, "p,n,u,µ,m,k,M,G or T")
		}
		return incorrectUnit(s[n:], "physic.Frequency")
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

	// RPM is revolutions per minute. It is used to quantify angular frequency.
	RPM Frequency = 16667 * MicroHertz

	maxFrequency = 9223372036854775807 * MicroHertz
	minFrequency = -9223372036854775807 * MicroHertz
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

// Set sets the Pressure to the value represented by s. Units are to
// be provided in "Pa" with an optional SI prefix: "p", "n", "u", "µ", "m", "k",
// "M", "G" or "T".
func (p *Pressure) Set(s string) error {
	v, n, err := valueOfUnitString(s, nano)
	if err != nil {
		if e, ok := err.(*parseError); ok {
			switch e.err {
			case errOverflowsInt64:
				return errors.New("maximum value is " + maxPressure.String())
			case errOverflowsInt64Negative:
				return errors.New("minimum value is " + minPressure.String())
			case errNotANumber:
				if found, _ := containsUnitString(s, "Pa"); found != "" {
					return errors.New("does not contain number")
				}
				return errors.New("does not contain number or unit \"Pa\"")
			}
		}
		return err
	}

	switch s[n:] {
	case "pa", "Pa":
		*p = (Pressure)(v)
	case "":
		return noUnits("Pa")
	default:
		if found, extra := containsUnitString(s[n:], "Pa"); found != "" {
			return unknownUnitPrefix(found, extra, "p,n,u,µ,m,k,M,G or T")
		}
		return incorrectUnit(s[n:], "physic.Pressure")
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

	maxPressure = 9223372036854775807 * NanoPascal
	minPressure = -9223372036854775807 * NanoPascal
)

// RelativeHumidity is a humidity level measurement stored as an int32 fixed
// point integer at a precision of 0.00001%rH.
//
// Valid values are between 0% and 100%.
type RelativeHumidity int32

// String returns the humidity formatted as a string.
func (h RelativeHumidity) String() string {
	h /= MilliRH
	frac := int(h % 10)
	if frac == 0 {
		return strconv.Itoa(int(h)/10) + "%rH"
	}
	if frac < 0 {
		frac = -frac
	}
	return strconv.Itoa(int(h)/10) + "." + strconv.Itoa(frac) + "%rH"
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
	ZeroFahrenheit  Temperature = 255372 * MilliKelvin
	MilliFahrenheit Temperature = 555555 * NanoKelvin
	Fahrenheit      Temperature = 555555555 * NanoKelvin
)

// Power is a measurement of power stored as a nano watts.
//
// The highest representable value is 9.2GW.
type Power int64

// String returns the power formatted as a string in watts.
func (p Power) String() string {
	return nanoAsString(int64(p)) + "W"
}

// Set sets the Power to the value represented by s. Units are to
// be provided in "W" with an optional SI prefix: "p", "n", "u", "µ", "m", "k",
// "M", "G" or "T".
func (p *Power) Set(s string) error {
	v, n, err := valueOfUnitString(s, nano)
	if err != nil {
		if e, ok := err.(*parseError); ok {
			switch e.err {
			case errOverflowsInt64:
				return errors.New("maximum value is " + maxPower.String())
			case errOverflowsInt64Negative:
				return errors.New("minimum value is " + minPower.String())
			case errNotANumber:
				if found, _ := containsUnitString(s, "W"); found != "" {
					return errors.New("does not contain number")
				}
				return errors.New("does not contain number or unit \"W\"")
			}
		}
		return err
	}

	switch s[n:] {
	case "w", "W":
		*p = (Power)(v)
	case "":
		return noUnits("W")
	default:
		if found, extra := containsUnitString(s[n:], "W"); found != "" {
			return unknownUnitPrefix(found, extra, "p,n,u,µ,m,k,M,G or T")
		}
		return incorrectUnit(s[n:], "physic.Power")
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

	maxPower = 9223372036854775807 * NanoWatt
	minPower = -9223372036854775807 * NanoWatt
)

// Energy is a measurement of work stored as a nano joules.
//
// The highest representable value is 9.2GJ.
type Energy int64

// String returns the energy formatted as a string in Joules.
func (e Energy) String() string {
	return nanoAsString(int64(e)) + "J"
}

// Set sets the Energy to the value represented by s. Units are to
// be provided in "J" with an optional SI prefix: "p", "n", "u", "µ", "m", "k",
// "M", "G" or "T".
func (e *Energy) Set(s string) error {
	v, n, err := valueOfUnitString(s, nano)
	if err != nil {
		if e, ok := err.(*parseError); ok {
			switch e.err {
			case errOverflowsInt64:
				return errors.New("maximum value is " + maxEnergy.String())
			case errOverflowsInt64Negative:
				return errors.New("minimum value is " + minEnergy.String())
			case errNotANumber:
				if found, _ := containsUnitString(s, "J"); found != "" {
					return errors.New("does not contain number")
				}
				return errors.New("does not contain number or unit \"J\"")
			}
		}
		return err
	}

	switch s[n:] {
	case "j", "J":
		*e = (Energy)(v)
	case "":
		return noUnits("J")
	default:
		if found, extra := containsUnitString(s[n:], "J"); found != "" {
			return unknownUnitPrefix(found, extra, "p,n,u,µ,m,k,M,G or T")
		}
		return incorrectUnit(s[n:], "physic.Energy")
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

	// BTU (British thermal unit) is the heat required to raise the temperature
	// of one pound of water by one degree Fahrenheit. This is the ISO value.
	BTU Energy = 1055060 * MilliJoule

	WattSecond   Energy = Joule
	WattHour     Energy = 3600 * Joule
	KiloWattHour Energy = 3600 * KiloJoule

	maxEnergy = 9223372036854775807 * NanoJoule
	minEnergy = -9223372036854775807 * NanoJoule
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
// to be provided in "F" with an optional SI prefix: "p", "n", "u", "µ", "m",
// "k", "M", "G" or "T".
func (c *ElectricalCapacitance) Set(s string) error {
	v, n, err := valueOfUnitString(s, pico)
	if err != nil {
		if e, ok := err.(*parseError); ok {
			switch e.err {
			case errOverflowsInt64:
				return errors.New("maximum value is " + maxElectricalCapacitance.String())
			case errOverflowsInt64Negative:
				return errors.New("minimum value is " + minElectricalCapacitance.String())
			case errNotANumber:
				if found, _ := containsUnitString(s, "F"); found != "" {
					return errors.New("does not contain number")
				}
				return errors.New("does not contain number or unit \"F\"")
			}
		}
		return err
	}

	switch s[n:] {
	case "f", "F":
		*c = (ElectricalCapacitance)(v)
	case "":
		return noUnits("F")
	default:
		if found, extra := containsUnitString(s[n:], "F"); found != "" {
			return unknownUnitPrefix(found, extra, "p,n,u,µ,m,k,M,G or T")
		}
		return incorrectUnit(s[n:], "physic.ElectricalCapacitance")
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

	maxElectricalCapacitance = 9223372036854775807 * PicoFarad
	minElectricalCapacitance = -9223372036854775807 * PicoFarad
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
func (i LuminousIntensity) String() string {
	return nanoAsString(int64(i)) + "cd"
}

// Set sets the LuminousIntensity to the value represented by s. Units are to
// be provided in "cd" with an optional SI prefix: "p", "n", "u", "µ", "m", "k",
// "M", "G" or "T".
func (i *LuminousIntensity) Set(s string) error {
	v, n, err := valueOfUnitString(s, nano)
	if err != nil {
		if e, ok := err.(*parseError); ok {
			switch e.err {
			case errOverflowsInt64:
				return errors.New("maximum value is " + maxLuminousIntensity.String())
			case errOverflowsInt64Negative:
				return errors.New("minimum value is " + minLuminousIntensity.String())
			case errNotANumber:
				if found, _ := containsUnitString(s, "cd"); found != "" {
					return errors.New("does not contain number")
				}
				return errors.New("does not contain number or unit \"cd\"")
			}
		}
		return err
	}

	switch s[n:] {
	case "cd":
		*i = (LuminousIntensity)(v)
	case "":
		return noUnits("cd")
	default:
		if found, extra := containsUnitString(s[n:], "cd"); found != "" {
			return unknownUnitPrefix(found, extra, "p,n,u,µ,m,k,M,G or T")
		}
		return incorrectUnit(s[n:], "physic.LuminousIntensity")
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

	maxLuminousIntensity = 9223372036854775807 * NanoCandela
	minLuminousIntensity = -9223372036854775807 * NanoCandela
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

// Set sets the LuminousFlux to the value represented by s. Units are to
// be provided in "lm" with an optional SI prefix: "p", "n", "u", "µ", "m", "k",
// "M", "G" or "T".
func (f *LuminousFlux) Set(s string) error {
	v, n, err := valueOfUnitString(s, nano)
	if err != nil {
		if e, ok := err.(*parseError); ok {
			switch e.err {
			case errOverflowsInt64:
				return errors.New("maximum value is " + maxLuminousFlux.String())
			case errOverflowsInt64Negative:
				return errors.New("minimum value is " + minLuminousFlux.String())
			case errNotANumber:
				if found, _ := containsUnitString(s, "lm"); found != "" {
					return errors.New("does not contain number")
				}
				return errors.New("does not contain number or unit \"lm\"")
			}
		}
		return err
	}

	switch s[n:] {
	case "lm":
		*f = (LuminousFlux)(v)
	case "":
		return noUnits("lm")
	default:
		if found, extra := containsUnitString(s[n:], "lm"); found != "" {
			return unknownUnitPrefix(found, extra, "p,n,u,µ,m,k,M,G or T")
		}
		return incorrectUnit(s[n:], "physic.LuminousFlux")
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

	maxLuminousFlux = 9223372036854775807 * NanoLumen
	minLuminousFlux = -9223372036854775807 * NanoLumen
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

// Decimal is the representation of decimal number.
type decimal struct {
	// base hold the significant digits.
	base uint64
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

var (
	errOverflowsInt64         = errors.New("exceeds maximum")
	errOverflowsInt64Negative = errors.New("exceeds minimum")
	errNotANumber             = errors.New("not a number")
	errExponentOverflow       = errors.New("exponent exceeds int64")
)

// Converts from decimal to int64.
// Scale is combined with the decimal exponent to maximise the resolution and is
// in powers of ten.
func dtoi(d decimal, scale int) (int64, error) {
	// Get the total magnitude of the number.
	// a^x * b^y = a*b^(x+y) since scale is of the order unity this becomes
	// 1^x * b^y = b^(x+y).
	// mag must be positive to use as index in to powerOf10 array.
	u := d.base
	mag := d.exp + scale
	if mag < 0 {
		mag = -mag
	}
	if mag > 18 {
		return 0, errExponentOverflow
	}
	// Divide is = 10^(-mag)
	if d.exp+scale < 0 {
		u = (u + powerOf10[mag]/2) / powerOf10[mag]
	} else {
		check := u * powerOf10[mag]
		if check/powerOf10[mag] != u || check > maxInt64 {
			if d.neg {
				return -maxInt64, &parseError{
					msg: "-" + maxUint64Str,
					err: errOverflowsInt64Negative,
				}
			}
			return maxInt64, &parseError{
				msg: maxUint64Str,
				err: errOverflowsInt64,
			}
		}
		u *= powerOf10[mag]
	}

	n := int64(u)
	if d.neg {
		n = -n
	}
	return n, nil
}

// Converts a string to a decimal form. The return int is how many bytes of the
// string are considered numeric. The string may contain +-0 prefixes and
// arbitrary suffixes as trailing non number characters are ignored.
// Significant digits are stored without leading or trailing zeros, rather a
// base and exponent is used. Significant digits are stored as uint64, max size
// of significant digits is int64
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
					msg: s,
					err: errors.New("can't contain both plus and minus symbols"),
				}
			}
			if d.neg {
				return decimal{}, 0, &parseError{
					msg: s,
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
					msg: s,
					err: errors.New("can't contain both plus and minus symbols"),
				}
			}
			if seenPlus {
				return decimal{}, 0, &parseError{
					msg: s,
					err: errors.New("multiple plus symbols"),
				}
			}
			seenPlus = true
			start++
		case s[i] == '.':
			if isPoint {
				return decimal{}, 0, &parseError{
					msg: s,
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
					msg: s,
					err: errNotANumber,
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

	for i := start; i < end; i++ {
		c := s[i]
		// Check that is is a digit.
		if c >= '0' && c <= '9' {
			// *10 is decimal shift left.
			d.base *= 10
			// Convert ascii digit into number.
			check := d.base + uint64(c-'0')
			// Check should always be larger than u unless we have overflowed.
			// Similarly if check > max it will overflow when converted to int64.
			if check < d.base || check > maxInt64 {
				if d.neg {
					return decimal{}, 0, &parseError{
						msg: "-" + maxUint64Str,
						err: errOverflowsInt64Negative,
					}
				}
				return decimal{}, 0, &parseError{
					msg: maxUint64Str,
					err: errOverflowsInt64,
				}
			}
			d.base = check
		} else if c != '.' {
			return decimal{}, 0, &parseError{err: errNotANumber}
		}
	}
	if !isPoint {
		d.exp = exp
	} else {
		if dp > start && dp < end {
			// Decimal Point is in the middle of a number.
			end--
		}
		// Find the exponent based on decimal point distance from left and the
		// length of the number.
		d.exp = (dp - start) - (end - start)
		if dp <= start {
			// Account for numbers of the form 1 > n < -1 eg 0.0001.
			d.exp++
		}
	}
	return d, last, nil
}

// valueOfUnitString is a helper for converting a string and a prefix in to a
// physic unit. It can be used when characters of the units do not conflict with
// any of the SI prefixes.
func valueOfUnitString(s string, base prefix) (int64, int, error) {
	d, n, err := atod(s)
	if err != nil {
		return 0, n, err
	}
	si := prefix(unit)
	if n != len(s) {
		r, rsize := utf8.DecodeRuneInString(s[n:])
		if r <= 1 || rsize == 0 {
			return 0, 0, &parseError{
				err: errors.New("unexpected end of string"),
				msg: s,
			}
		}
		var siSize int
		si, siSize = parseSIPrefix(r)
		n += siSize

	}
	v, err := dtoi(d, int(si-base))
	if err != nil {
		return v, 0, err
	}
	return v, n, nil
}

// Calcululates the product of two decimals; a and b, keeping the base less than
// maxInt64. Degrade controls the maximum limit of the precision loss in the
// product. Returns an error if the product base exceeds an int64 after losing a
// degrade number of least significant figures. This function is to aid in the
// multiplication of numbers that combined have more than 18 significant figures
// each. The minimum limit of significant figures is 9 figures.
func decimalMulScale(a, b decimal) (decimal, uint) {
	if a.base > 18446744073709551609 || b.base > 18446744073709551609 {
		return decimal{}, 21
	}
	exp := a.exp + b.exp
	neg := a.neg != b.neg
	ab := a.base
	bb := b.base
	for i := uint(0); i < 21; i++ {
		if ab <= 1 || bb <= 1 {
			// This will always fit inside uint64.
			return decimal{ab * bb, exp, neg}, i
		}
		if base := ab * bb; (base/ab == bb) && base < maxInt64 {
			// Return if product did not overflow or exceed int64.
			return decimal{base, exp, neg}, i
		}
		// Truncate least significant digit in product.
		if bb > ab {
			bb = (bb + 5) / 10
			// Compact trailing zeros if any.
			for bb > 0 && bb%10 == 0 {
				bb /= 10
				exp++
			}
		} else {
			ab = (ab + 5) / 10
			// Compact trailing zeros if any.
			for ab > 0 && ab%10 == 0 {
				ab /= 10
				exp++
			}
		}
		exp++
	}
	return decimal{}, 21
}

// For units with short and or plural variations order units with longest first.
// eg Degrees, Degree, Deg.
func containsUnitString(s string, units ...string) (string, string) {
	sub := strings.ToLower(s)
	for _, unit := range units {
		unitLow := strings.ToLower(unit)
		if strings.Contains(sub, unitLow) {
			index := strings.Index(sub, unitLow)
			if index >= 0 {
				// prefix
				return unit, s[:index]
			}
		}
	}
	return "", ""
}

type parseError struct {
	msg string
	err error
}

func (p *parseError) Error() string {
	if p.err == nil {
		return "parse error"
	}
	if p.msg == "" {
		return p.err.Error()
	}
	return p.err.Error() + " " + p.msg
}

func noUnits(s string) error {
	return &parseError{msg: s, err: errors.New("no units provided, need")}
}

func incorrectUnit(inputString, want string) error {
	return &parseError{err: errors.New("\"" + inputString + "\"" + " is not a valid unit for " + want)}
}

func unknownUnitPrefix(unit string, prefix string, valid string) error {
	return &parseError{
		msg: "valid prefixes for \"" + unit + "\" are " + valid,
		err: errors.New("contains unknown unit prefix \"" + prefix + "\"."),
	}
}

type prefix int

const (
	pico  prefix = -12
	nano  prefix = -9
	micro prefix = -6
	milli prefix = -3
	unit  prefix = 0
	deca  prefix = 1
	hecto prefix = 2
	kilo  prefix = 3
	mega  prefix = 6
	giga  prefix = 9
	tera  prefix = 12
)

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
		return unit, 0
	}
}
