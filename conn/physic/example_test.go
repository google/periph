// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic_test

import (
	"fmt"

	"periph.io/x/periph/conn/physic"
)

func ExampleAmpere() {
	fmt.Printf("%s\n", physic.Ampere(10010))
	fmt.Printf("%s\n", physic.Ampere(10))
	fmt.Printf("%s\n", physic.Ampere(-10))
	// Output:
	// 10.010A
	// 10mA
	// -10mA
}

func ExampleCelsius() {
	fmt.Printf("%s\n", physic.Celsius(23010))
	fmt.Printf("%s\n", physic.Celsius(10))
	// Output:
	// 23.010°C
	// 0.010°C
}

func ExampleFahrenheit() {
	fmt.Printf("%s\n", physic.Fahrenheit(80010))
	fmt.Printf("%s\n", physic.Fahrenheit(10))
	// Output:
	// 80.010°F
	// 0.010°F
}

func ExampleKPascal() {
	fmt.Printf("%s\n", physic.KPascal(101010))
	fmt.Printf("%s\n", physic.KPascal(10))
	// Output:
	// 101.010KPa
	// 0.010KPa
}

func ExampleRelativeHumidity() {
	fmt.Printf("%s\n", physic.RelativeHumidity(5006))
	fmt.Printf("%s\n", physic.RelativeHumidity(10))
	// Output:
	// 50.06%rH
	// 0.10%rH
}

func ExampleVolt() {
	fmt.Printf("%s\n", physic.Volt(10010))
	fmt.Printf("%s\n", physic.Volt(10))
	fmt.Printf("%s\n", physic.Volt(-10))
	// Output:
	// 10.010V
	// 10mV
	// -10mV
}
