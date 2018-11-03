// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic_test

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn/physic"
)

func ExampleAngle() {
	fmt.Println(physic.Degree)
	fmt.Println(physic.Pi)
	fmt.Println(physic.Theta)
	// Output:
	// 1.000°
	// 180.0°
	// 360.0°
}

func ExampleDistance() {
	fmt.Println(physic.Inch)
	fmt.Println(physic.Foot)
	fmt.Println(physic.Mile)
	// Output:
	// 25.400mm
	// 304.800mm
	// 1.609km
}

func ExampleElectricCurrent() {
	fmt.Println(10010 * physic.MilliAmpere)
	fmt.Println(10 * physic.Ampere)
	fmt.Println(-10 * physic.MilliAmpere)
	// Output:
	// 10.010A
	// 10A
	// -10mA
}

func ExampleElectricPotential() {
	fmt.Println(10010 * physic.MilliVolt)
	fmt.Println(10 * physic.Volt)
	fmt.Println(-10 * physic.MilliVolt)
	// Output:
	// 10.010V
	// 10V
	// -10mV
}

func ExampleElectricResistance() {
	fmt.Println(10010 * physic.MilliOhm)
	fmt.Println(10 * physic.Ohm)
	fmt.Println(24 * physic.MegaOhm)
	// Output:
	// 10.010Ω
	// 10Ω
	// 24MΩ
}

func ExampleForce() {
	fmt.Println(10 * physic.MilliNewton)
	fmt.Println(physic.EarthGravity)
	fmt.Println(physic.PoundForce)
	// Output:
	// 10mN
	// 9.807N
	// 4.448kN
}

func ExampleFrequency() {
	fmt.Println(10 * physic.MilliHertz)
	fmt.Println(101010 * physic.MilliHertz)
	fmt.Println(10 * physic.MegaHertz)
	// Output:
	// 10mHz
	// 101.010Hz
	// 10MHz
}

func ExampleFrequency_Duration() {
	fmt.Println(physic.MilliHertz.Duration())
	fmt.Println(physic.MegaHertz.Duration())
	// Output:
	// 16m40s
	// 1µs
}

func ExamplePeriodToFrequency() {
	fmt.Println(physic.PeriodToFrequency(time.Microsecond))
	fmt.Println(physic.PeriodToFrequency(time.Minute))
	// Output:
	// 1MHz
	// 16.666mHz
}

func ExampleMass() {
	fmt.Println(10 * physic.MilliGram)
	fmt.Println(physic.OunceMass)
	fmt.Println(physic.PoundMass)
	fmt.Println(physic.Slug)
	// Output:
	// 10mg
	// 28.350g
	// 453.592g
	// 14.594kg
}

func ExamplePressure() {
	fmt.Println(101010 * physic.Pascal)
	fmt.Println(101 * physic.KiloPascal)
	// Output:
	// 101.010kPa
	// 101kPa
}

func ExampleRelativeHumidity() {
	fmt.Println(506 * physic.MilliRH)
	fmt.Println(20 * physic.PercentRH)
	// Output:
	// 50.6%rH
	// 20%rH
}

func ExampleSpeed() {
	fmt.Println(10 * physic.MilliMetrePerSecond)
	fmt.Println(physic.LightSpeed)
	fmt.Println(physic.KilometrePerHour)
	fmt.Println(physic.MilePerHour)
	fmt.Println(physic.FootPerSecond)
	// Output:
	// 10mm/s
	// 299.792Mm/s
	// 277.778mm/s
	// 447.040mm/s
	// 304.800mm/s
}

func ExampleTemperature() {
	fmt.Println(0 * physic.Kelvin)
	fmt.Println(23010*physic.MilliCelsius + physic.ZeroCelsius)
	fmt.Println(80*physic.Fahrenheit + physic.ZeroFahrenheit)
	// Output:
	// -273.150°C
	// 23.010°C
	// 26.666°C
}

func ExamplePower() {
	fmt.Println(1 * physic.Watt)
	fmt.Println(16 * physic.MilliWatt)
	fmt.Println(1210 * physic.MegaWatt)
	// Output:
	// 1W
	// 16mW
	// 1.210GW
}

func ExampleElectricalCapacitance() {
	fmt.Println(1 * physic.Farad)
	fmt.Println(22 * physic.PicoFarad)
	// Output:
	// 1F
	// 22pF
}

func ExampleLuminousFlux() {
	fmt.Println(18282 * physic.Lumen)
	// Output:
	// 18.282klm
}

func ExampleLuminousIntensity() {
	fmt.Println(12 * physic.Candela)
	// Output:
	// 12cd
}

func ExampleEnergy() {
	fmt.Println(1 * physic.Joule)
	// Output:
	// 1J
}

func ExampleAngle_Set() {
	var a physic.Angle

	a.Set("2Pi")
	fmt.Println(a)

	a.Set("90Degrees")
	fmt.Println(a)

	a.Set("1Radian")
	fmt.Println(a)
	// Output:
	// 360.0°
	// 90.00°
	// 57.296°
}

func ExampleDistance_Set() {
	var d physic.Distance

	d.Set("1Foot")
	fmt.Println(d)

	d.Set("1Metre")
	fmt.Println(d)

	d.Set("9Mile")
	fmt.Println(d)
	// Output:
	// 304.800mm
	// 1m
	// 14.484km

}

func ExampleElectricCurrent_Set() {
	var e physic.ElectricCurrent

	e.Set("12.5mA")
	fmt.Println(e)

	e.Set("2.4kA")
	fmt.Println(e)

	e.Set("2Amps")
	fmt.Println(e)
	// Output:
	// 12.500mA
	// 2.400kA
	// 2A
}

func ExampleElectricPotential_Set() {
	var e physic.ElectricPotential

	e.Set("250uV")
	fmt.Println(e)

	e.Set("100kV")
	fmt.Println(e)

	e.Set("12Volts")
	fmt.Println(e)
	// Output:
	// 250µV
	// 100kV
	// 12V
}

func ExampleElectricResistance_Set() {
	var e physic.ElectricResistance

	e.Set("33.3kOhms")
	fmt.Println(e)

	e.Set("1Ohm")
	fmt.Println(e)

	e.Set("5MOhm")
	fmt.Println(e)
	// Output:
	// 33.300kΩ
	// 1Ω
	// 5MΩ
}

func ExampleForce_Set() {
	var f physic.Force

	f.Set("9.8N")
	fmt.Println(f)

	// Output:
	// 9.800N
}

func ExampleFrequency_Set() {
	var f physic.Frequency

	f.Set("10MHz")
	fmt.Println(f)

	f.Set("10mHz")
	fmt.Println(f)

	f.Set("1kHz")
	fmt.Println(f)
	// Output:
	// 10MHz
	// 10mHz
	// 1kHz
}

func ExampleMass_Set() {
	var m physic.Mass

	m.Set("10mg")
	fmt.Println(m)

	m.Set("16.5kg")
	fmt.Println(m)

	m.Set("2.2oz")
	fmt.Println(m)

	m.Set("16Tonne")
	fmt.Println(m)
	// Output:
	// 10mg
	// 16.500kg
	// 56.699g
	// 16Mg
}

func ExamplePressure_Set() {
	var p physic.Pressure

	p.Set("300kPa")
	fmt.Println(p)

	p.Set("16MPascal")
	fmt.Println(p)
	// Output:
	// 300kPa
	// 16MPa
}

func ExampleRelativeHumidity_Set() {
	var r physic.RelativeHumidity

	r.Set("50.6%rH")
	fmt.Println(r)

	r.Set("20%rH")
	fmt.Println(r)
	// Output:
	// 50.6%rH
	// 20%rH
}

func ExampleSpeed_Set() {
	var s physic.Speed

	s.Set("10m/s")
	fmt.Println(s)

	s.Set("100km/h")
	fmt.Println(s)

	s.Set("2067fps")
	fmt.Println(s)

	s.Set("55mph")
	fmt.Println(s)
	// Output:
	// 10m/s
	// 27.778m/s
	// 630.022m/s
	// 24.587m/s
}

func ExampleTemperature_Set() {
	var t physic.Temperature

	t.Set("0C")
	fmt.Println(t)

	t.Set("1C")
	fmt.Println(t)

	t.Set("5MK")
	fmt.Println(t)

	// Output:
	// 0°C
	// 1°C
	// 5M°C
}

func ExamplePower_Set() {
	var p physic.Power

	p.Set("25mW")
	fmt.Println(p)

	p.Set("1W")
	fmt.Println(p)

	p.Set("1.21GW")
	fmt.Println(p)

	// Output:
	// 25mW
	// 1W
	// 1.210GW
}

func ExampleElectricalCapacitance_Set() {
	var c physic.ElectricalCapacitance

	c.Set("1F")
	fmt.Println(c)

	c.Set("22pF")
	fmt.Println(c)
	// Output:
	// 1F
	// 22pF
}

func ExampleLuminousFlux_Set() {
	var l physic.LuminousFlux

	l.Set("25mlm")
	fmt.Println(l)

	l.Set("2.5Mlm")
	fmt.Println(l)

	// Output:
	// 25mlm
	// 2.500Mlm
}

func ExampleLuminousIntensity_Set() {
	var l physic.LuminousIntensity

	l.Set("16cd")
	fmt.Println(l)

	// Output:
	// 16cd
}

func ExampleEnergy_Set() {
	var e physic.Energy

	e.Set("2.6kJ")
	fmt.Println(e)

	e.Set("45mJ")
	fmt.Println(e)

	// Output:
	// 2.600kJ
	// 45mJ
}
