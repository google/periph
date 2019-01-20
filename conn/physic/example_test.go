// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic_test

import (
	"flag"
	"fmt"
	"log"
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

func ExampleAngle_Set() {
	var a physic.Angle

	if err := a.Set("2°"); err != nil {
		log.Fatal(a)
	}
	fmt.Println(a)

	if err := a.Set("90deg"); err != nil {
		log.Fatal(a)
	}
	fmt.Println(a)

	if err := a.Set("1rad"); err != nil {
		log.Fatal(a)
	}
	fmt.Println(a)
	// Output:
	// 2.000°
	// 90.00°
	// 57.296°
}

func ExampleAngle_flag() {
	var a physic.Angle

	flag.Var(&a, "angle", "angle to set the servo to")
	flag.Parse()
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

func ExampleDistance_Set() {
	var d physic.Distance

	if err := d.Set("1ft"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(d)

	if err := d.Set("1m"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(d)

	if err := d.Set("9Mile"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(d)
	// Output:
	// 304.800mm
	// 1m
	// 14.484km

}

func ExampleDistance_flag() {
	var d physic.Distance

	flag.Var(&d, "distance", "x axis travel length")
	flag.Parse()
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

func ExampleElectricCurrent_Set() {
	var e physic.ElectricCurrent

	if err := e.Set("12.5mA"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(e)

	if err := e.Set("2.4kA"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(e)

	if err := e.Set("2A"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(e)
	// Output:
	// 12.500mA
	// 2.400kA
	// 2A
}

func ExampleElectricCurrent_flag() {
	var m physic.ElectricCurrent

	flag.Var(&m, "motor", "rated motor current")
	flag.Parse()
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

func ExampleElectricPotential_flag() {
	var v physic.ElectricPotential
	flag.Var(&v, "cutout", "battery full charge voltage")
	flag.Parse()
}

func ExampleElectricPotential_Set() {
	var v physic.ElectricPotential
	if err := v.Set("250uV"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(v)

	if err := v.Set("100kV"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(v)

	if err := v.Set("12V"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(v)
	// Output:
	// 250µV
	// 100kV
	// 12V
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

func ExampleElectricResistance_Set() {
	var r physic.ElectricResistance

	if err := r.Set("33.3kOhm"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(r)

	if err := r.Set("1Ohm"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(r)

	if err := r.Set("5MOhm"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(r)
	// Output:
	// 33.300kΩ
	// 1Ω
	// 5MΩ
}

func ExampleElectricResistance_flag() {
	var r physic.ElectricResistance

	flag.Var(&r, "shunt", "shunt resistor value")
	flag.Parse()
}

func ExampleForce() {
	fmt.Println(10 * physic.MilliNewton)
	fmt.Println(physic.EarthGravity)
	fmt.Println(physic.PoundForce)
	// Output:
	// 10mN
	// 9.807N
	// 4.448N
}

func ExampleForce_Set() {
	var f physic.Force

	if err := f.Set("9.8N"); err != nil {
		log.Fatal(f)
	}
	fmt.Println(f)

	if err := f.Set("20lbf"); err != nil {
		log.Fatal(f)
	}
	fmt.Println(f)

	// Output:
	// 9.800N
	// 88.964N
}

func ExampleForce_flag() {
	var f physic.Force

	flag.Var(&f, "force", "load cell wakeup force")
	flag.Parse()
}

func ExampleFrequency() {
	fmt.Println(10 * physic.MilliHertz)
	fmt.Println(101010 * physic.MilliHertz)
	fmt.Println(10 * physic.MegaHertz)
	fmt.Println(60 * physic.RPM)
	// Output:
	// 10mHz
	// 101.010Hz
	// 10MHz
	// 1Hz
}

func ExampleFrequency_Set() {
	var f physic.Frequency

	if err := f.Set("10MHz"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(f)

	if err := f.Set("10mHz"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(f)

	if err := f.Set("1kHz"); err != nil {
		log.Fatal(err)
	}
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

func ExampleFrequency_Period() {
	fmt.Println(physic.MilliHertz.Period())
	fmt.Println(physic.MegaHertz.Period())
	// Output:
	// 16m40s
	// 1µs
}

func ExamplePeriodToFrequency() {
	fmt.Println(physic.PeriodToFrequency(time.Microsecond))
	fmt.Println(physic.PeriodToFrequency(time.Minute))
	// Output:
	// 1MHz
	// 16.667mHz
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

func ExampleMass_Set() {
	var m physic.Mass

	if err := m.Set("10mg"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(m)

	if err := m.Set("16.5kg"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(m)

	if err := m.Set("2.2oz"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(m)

	if err := m.Set("16lb"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(m)
	// Output:
	// 10mg
	// 16.500kg
	// 62.369g
	// 7.257kg
}

func ExampleMass_flag() {
	var m physic.Mass

	flag.Var(&m, "weight", "amount of cat food to dispense")
	flag.Parse()
}

func ExamplePressure() {
	fmt.Println(101010 * physic.Pascal)
	fmt.Println(101 * physic.KiloPascal)
	// Output:
	// 101.010kPa
	// 101kPa
}

func ExamplePressure_Set() {
	var p physic.Pressure

	if err := p.Set("300kPa"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(p)

	if err := p.Set("16MPa"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(p)
	// Output:
	// 300kPa
	// 16MPa
}

func ExamplePressure_flag() {
	var p physic.Pressure

	flag.Var(&p, "setpoint", "pressure for pump to maintain")
	flag.Parse()
}

func ExampleRelativeHumidity() {
	fmt.Println(506 * physic.MilliRH)
	fmt.Println(20 * physic.PercentRH)
	// Output:
	// 50.6%rH
	// 20%rH
}

func ExampleRelativeHumidity_Set() {
	var r physic.RelativeHumidity

	if err := r.Set("50.6%rH"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(r)

	if err := r.Set("20%"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(r)
	// Output:
	// 50.6%rH
	// 20%rH
}

func ExampleRelativeHumidity_flag() {
	var h physic.RelativeHumidity

	flag.Var(&h, "humidity", "green house humidity high alarm level")
	flag.Parse()
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

func ExampleSpeed_Set() {
	var s physic.Speed

	if err := s.Set("10m/s"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)

	if err := s.Set("100kph"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)

	if err := s.Set("2067fps"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)

	if err := s.Set("55mph"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
	// Output:
	// 10m/s
	// 27.778m/s
	// 630.022m/s
	// 24.587m/s
}

func ExampleSpeed_flag() {
	var s physic.Speed

	flag.Var(&s, "speed", "window shutter closing speed")
	flag.Parse()
}

func ExampleTemperature() {
	fmt.Println(0 * physic.Kelvin)
	fmt.Println(23010*physic.MilliCelsius + physic.ZeroCelsius)
	fmt.Println(80*physic.Fahrenheit + physic.ZeroFahrenheit)
	// Output:
	// -273.150°C
	// 23.010°C
	// 26.667°C
}

func ExampleTemperature_Set() {
	var t physic.Temperature

	if err := t.Set("0°C"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(t)

	if err := t.Set("1C"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(t)

	if err := t.Set("5MK"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(t)

	if err := t.Set("0°F"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(t)

	if err := t.Set("32F"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(t)

	// Output:
	// 0°C
	// 1°C
	// 5M°C
	// -17.778°C
	// 0°C
}

func ExampleTemperature_flag() {
	var t physic.Temperature

	flag.Var(&t, "temp", "thermostat setpoint")
	flag.Parse()
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

func ExamplePower_Set() {
	var p physic.Power

	if err := p.Set("25mW"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(p)

	if err := p.Set("1W"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(p)

	if err := p.Set("1.21GW"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(p)

	// Output:
	// 25mW
	// 1W
	// 1.210GW
}

func ExamplePower_flag() {
	var p physic.Power

	flag.Var(&p, "power", "heater maximum power")
	flag.Parse()
}

func ExampleElectricalCapacitance() {
	fmt.Println(1 * physic.Farad)
	fmt.Println(22 * physic.PicoFarad)
	// Output:
	// 1F
	// 22pF
}

func ExampleElectricalCapacitance_Set() {
	var c physic.ElectricalCapacitance

	if err := c.Set("1F"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(c)

	if err := c.Set("22pF"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(c)
	// Output:
	// 1F
	// 22pF
}

func ExampleElectricalCapacitance_flag() {
	var c physic.ElectricalCapacitance

	flag.Var(&c, "mintouch", "minimum touch sensitivity")
	flag.Parse()
}

func ExampleLuminousFlux() {
	fmt.Println(18282 * physic.Lumen)
	// Output:
	// 18.282klm
}

func ExampleLuminousFlux_Set() {
	var l physic.LuminousFlux

	if err := l.Set("25mlm"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(l)

	if err := l.Set("2.5Mlm"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(l)

	// Output:
	// 25mlm
	// 2.500Mlm
}
func ExampleLuminousFlux_flag() {
	var l physic.LuminousFlux

	flag.Var(&l, "low", "mood light level")
	flag.Parse()
}

func ExampleLuminousIntensity() {
	fmt.Println(12 * physic.Candela)
	// Output:
	// 12cd
}

func ExampleLuminousIntensity_Set() {
	var l physic.LuminousIntensity

	if err := l.Set("16cd"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(l)

	// Output:
	// 16cd
}

func ExampleLuminousIntensity_flag() {
	var l physic.LuminousIntensity

	flag.Var(&l, "dusk", "light level to turn on light")
	flag.Parse()
}

func ExampleEnergy() {
	fmt.Println(1 * physic.Joule)
	fmt.Println(1 * physic.WattSecond)
	fmt.Println(1 * physic.KiloWattHour)
	// Output:
	// 1J
	// 1J
	// 3.600MJ
}

func ExampleEnergy_Set() {
	var e physic.Energy

	if err := e.Set("2.6kJ"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(e)

	if err := e.Set("45mJ"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(e)

	// Output:
	// 2.600kJ
	// 45mJ
}

func ExampleEnergy_flag() {
	var e physic.Energy

	flag.Var(&e, "capacity", "capacity of battery")
	flag.Parse()
}
