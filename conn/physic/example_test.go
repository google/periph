// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic_test

import (
	"fmt"
	"time"

	"periph.io/x/periph/conn/physic"
)

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

func ExampleTemperature() {
	fmt.Printf("%s\n", 0*physic.Kelvin)
	fmt.Printf("%s\n", 23010*physic.MilliCelsius+physic.ZeroCelsius)
	fmt.Printf("%s\n", 80*physic.Fahrenheit+physic.ZeroFahrenheit)
	// Output:
	// -273.150°C
	// 23.010°C
	// 26.666°C
}
