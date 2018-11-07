// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic_test

import (
	"flag"
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

func ExampleFrequency_flag() {
	var pwm physic.Frequency

	flag.Var(&pwm, "pwm", "pwm frequency")
	flag.Parse()
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
