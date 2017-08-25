// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package devices

import "testing"

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
	o := RelativeHumidity(5010)
	if s := o.String(); s != "50.10%rH" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f > 50.11 || f < 50.09 {
		t.Fatalf("%f", f)
	}
}

func TestRelativeHumidity_neg(t *testing.T) {
	o := RelativeHumidity(-5010)
	if s := o.String(); s != "-50.10%rH" {
		t.Fatalf("%#v", s)
	}
	if f := o.Float64(); f < -50.11 || f > -50.09 {
		t.Fatalf("%f", f)
	}
}
