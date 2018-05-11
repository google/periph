// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic

import "testing"

func TestCenti(t *testing.T) {
	o := Centi(10010)
	if s := o.String(); s != "100.10" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f > 100.11 || f < 100.09 {
		t.Fatalf("%f", f)
	}
}

func TestCenti_neg(t *testing.T) {
	o := Centi(-10010)
	if s := o.String(); s != "-100.10" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f > -100.09 || f < -100.11 {
		t.Fatalf("%f", f)
	}
}

func TestMilli(t *testing.T) {
	o := Milli(10010)
	if s := o.String(); s != "10.010" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f > 10.011 || f < 10.009 {
		t.Fatalf("%f", f)
	}
}

func TestMilli_neg(t *testing.T) {
	o := Milli(-10010)
	if s := o.String(); s != "-10.010" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f > -10.009 || f < -10.011 {
		t.Fatalf("%f", f)
	}
}

//

func TestAmpere(t *testing.T) {
	o := Ampere(10010)
	if s := o.String(); s != "10.010A" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f > 10.011 || f < 10.009 {
		t.Fatalf("%f", f)
	}

	o = Ampere(10)
	if s := o.String(); s != "10mA" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f > 0.011 || f < 0.009 {
		t.Fatalf("%f", f)
	}
	o = Ampere(-10)
	if s := o.String(); s != "-10mA" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f < -0.011 || f > -0.009 {
		t.Fatalf("%f", f)
	}
}

func TestCelsius(t *testing.T) {
	o := Celsius(10010)
	if s := o.String(); s != "10.010°C" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f > 10.011 || f < 10.009 {
		t.Fatalf("%f", f)
	}
	if f := o.ToF(); f != 50018 {
		t.Fatalf("%d", f)
	}
}

func TestFahrenheit(t *testing.T) {
	o := Fahrenheit(10010)
	if s := o.String(); s != "10.010°F" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f > 10.011 || f < 10.009 {
		t.Fatalf("%f", f)
	}
}

func TestKPascal(t *testing.T) {
	o := KPascal(10010)
	if s := o.String(); s != "10.010KPa" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f > 10.011 || f < 10.009 {
		t.Fatalf("%f", f)
	}
}

func TestRelativeHumidity(t *testing.T) {
	o := RelativeHumidity(5006)
	if s := o.String(); s != "50.06%rH" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f >= 50.07 || f <= 50.05 {
		t.Fatalf("%f", f)
	}
}

func TestRelativeHumidity_neg(t *testing.T) {
	o := RelativeHumidity(-5010)
	if s := o.String(); s != "-50.10%rH" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f <= -50.11 || f >= -50.09 {
		t.Fatalf("%f", f)
	}
}

func TestVolt(t *testing.T) {
	o := Volt(10010)
	if s := o.String(); s != "10.010V" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f > 10.011 || f < 10.009 {
		t.Fatalf("%f", f)
	}

	o = Volt(10)
	if s := o.String(); s != "10mV" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f > 0.011 || f < 0.009 {
		t.Fatalf("%f", f)
	}
	o = Volt(-10)
	if s := o.String(); s != "-10mV" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f < -0.011 || f > -0.009 {
		t.Fatalf("%f", f)
	}
}
