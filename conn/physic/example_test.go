// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic_test

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn/physic"
)

func ExampleDistance() {
	fmt.Printf("%s\n", physic.Inch)
	fmt.Printf("%s\n", physic.Foot)
	fmt.Printf("%s\n", physic.Mile)
	// Output:
	// 25.400mm
	// 304.800mm
	// 1.609km
}

func ExampleElectricCurrent() {
	fmt.Printf("%s\n", 10010*physic.MilliAmpere)
	fmt.Printf("%s\n", 10*physic.Ampere)
	fmt.Printf("%s\n", -10*physic.MilliAmpere)
	// Output:
	// 10.010A
	// 10A
	// -10mA
}

func ExampleElectricPotential() {
	fmt.Printf("%s\n", 10010*physic.MilliVolt)
	fmt.Printf("%s\n", 10*physic.Volt)
	fmt.Printf("%s\n", -10*physic.MilliVolt)
	// Output:
	// 10.010V
	// 10V
	// -10mV
}

func ExampleElectricResistance() {
	fmt.Printf("%s\n", 10010*physic.MilliOhm)
	fmt.Printf("%s\n", 10*physic.Ohm)
	fmt.Printf("%s\n", 24*physic.MegaOhm)
	// Output:
	// 10.010Ω
	// 10Ω
	// 24MΩ
}

func ExampleForce() {
	fmt.Printf("%s\n", 10*physic.MilliNewton)
	fmt.Printf("%s\n", 101010*physic.EarthGravity)
	fmt.Printf("%s\n", physic.PoundForce)
	// Output:
	// 10mN
	// 990.569kN
	// 4.448kN
}

func ExampleFrequency() {
	fmt.Printf("%s\n", 10*physic.MilliHertz)
	fmt.Printf("%s\n", 101010*physic.MilliHertz)
	fmt.Printf("%s\n", 10*physic.MegaHertz)
	// Output:
	// 10mHz
	// 101.010Hz
	// 10MHz
}

func ExampleFrequency_Duration() {
	fmt.Printf("%s\n", physic.MilliHertz.Duration())
	fmt.Printf("%s\n", physic.MegaHertz.Duration())
	// Output:
	// 16m40s
	// 1µs
}

func ExamplePeriodToFrequency() {
	fmt.Printf("%s\n", physic.PeriodToFrequency(time.Microsecond))
	fmt.Printf("%s\n", physic.PeriodToFrequency(time.Minute))
	// Output:
	// 1MHz
	// 16.666mHz
}

func ExampleMass() {
	fmt.Printf("%s\n", 10*physic.MilliGram)
	fmt.Printf("%s\n", physic.OunceMass)
	fmt.Printf("%s\n", physic.PoundMass)
	fmt.Printf("%s\n", physic.Slug)
	// Output:
	// 10mg
	// 28.349g
	// 453.592g
	// 14.593kg
}

func ExamplePressure() {
	fmt.Printf("%s\n", 101010*physic.Pascal)
	fmt.Printf("%s\n", 101*physic.KiloPascal)
	// Output:
	// 101.010kPa
	// 101kPa
}

func ExampleRelativeHumidity() {
	fmt.Printf("%s\n", 506*physic.MilliRH)
	fmt.Printf("%s\n", 20*physic.PercentRH)
	// Output:
	// 50.6%rH
	// 20%rH
}

func ExampleSpeed() {
	fmt.Printf("%s\n", 10*physic.MilliMetrePerSecond)
	fmt.Printf("%s\n", physic.LightSpeed)
	fmt.Printf("%s\n", physic.KilometrePerHour)
	fmt.Printf("%s\n", physic.MilePerHour)
	fmt.Printf("%s\n", physic.FootPerSecond)
	// Output:
	// 10mm/s
	// 299.792Mm/s
	// 3.600m/s
	// 447.040mm/s
	// 304.800mm/s
}

func ExampleTemperature() {
	fmt.Printf("%s\n", 0*physic.Kelvin)
	fmt.Printf("%s\n", 23010*physic.MilliCelsius+physic.ZeroCelsius)
	fmt.Printf("%s\n", 80*physic.Fahrenheit+physic.ZeroFahrenheit)
	// Output:
	// -273.150°C
	// 23.010°C
	// 26.666°C
}
