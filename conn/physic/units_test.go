// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestAngle_String(t *testing.T) {
	data := []struct {
		in       Angle
		expected string
	}{
		{0, "0°"},
		{Degree/10000 + Degree/2000, "0.001°"},
		{-Degree/10000 - Degree/2000, "-0.001°"},
		{Degree / 1000, "0.001°"},
		{-Degree / 1000, "-0.001°"},
		{Degree / 2, "0.500°"},
		{-Degree / 2, "-0.500°"},
		{Degree, "1.000°"},
		{-Degree, "-1.000°"},
		{10 * Degree, "10.00°"},
		{-10 * Degree, "-10.00°"},
		{100 * Degree, "100.0°"},
		{-100 * Degree, "-100.0°"},
		{1000 * Degree, "1000°"},
		{-1000 * Degree, "-1000°"},
		{100000000000 * Degree, "100000000000°"},
		{-100000000000 * Degree, "-100000000000°"},
		{maxAngle, "528460276055°"},
		{minAngle, "-528460276055°"},
		{Pi, "180.0°"},
		{Theta, "360.0°"},
		{Radian, "57.296°"},
	}
	for i, line := range data {
		if s := line.in.String(); s != line.expected {
			t.Fatalf("%d: Degree(%d).String() = %s != %s", i, int64(line.in), s, line.expected)
		}
	}
}

func TestDistance_String(t *testing.T) {
	if s := Mile.String(); s != "1.609km" {
		t.Fatalf("%#v", s)
	}
}

func TestElectricCurrent_String(t *testing.T) {
	if s := Ampere.String(); s != "1A" {
		t.Fatalf("%#v", s)
	}
}

func TestElectricPotential_String(t *testing.T) {
	if s := Volt.String(); s != "1V" {
		t.Fatalf("%#v", s)
	}
}

func TestElectricResistance_String(t *testing.T) {
	if s := Ohm.String(); s != "1Ω" {
		t.Fatalf("%#v", s)
	}
}

func TestForce_String(t *testing.T) {
	if s := Newton.String(); s != "1N" {
		t.Fatalf("%#v", s)
	}
}

func TestFrequency_String(t *testing.T) {
	if s := Hertz.String(); s != "1Hz" {
		t.Fatalf("%#v", s)
	}
}

func TestFrequency_Duration(t *testing.T) {
	if v := MegaHertz.Duration(); v != time.Microsecond {
		t.Fatalf("%#v", v)
	}
}

func TestFrequency_PeriodToFrequency(t *testing.T) {
	if v := PeriodToFrequency(time.Millisecond); v != KiloHertz {
		t.Fatalf("%#v", v)
	}
}

func TestMass_String(t *testing.T) {
	if s := PoundMass.String(); s != "453.592g" {
		t.Fatalf("%#v", s)
	}
}

func TestPressure_String(t *testing.T) {
	if s := NanoPascal.String(); s != "1nPa" {
		t.Fatalf("%v", s)
	}
	if s := MicroPascal.String(); s != "1µPa" {
		t.Fatalf("%v", s)
	}
	if s := MilliPascal.String(); s != "1mPa" {
		t.Fatalf("%v", s)
	}
	if s := Pascal.String(); s != "1Pa" {
		t.Fatalf("%v", s)
	}
	if s := KiloPascal.String(); s != "1kPa" {
		t.Fatalf("%v", s)
	}
	if s := MegaPascal.String(); s != "1MPa" {
		t.Fatalf("%v", s)
	}
	if s := GigaPascal.String(); s != "1GPa" {
		t.Fatalf("%v", s)
	}

}

func TestRelativeHumidity_String(t *testing.T) {
	data := []struct {
		in       RelativeHumidity
		expected string
	}{
		{TenthMicroRH, "0%rH"},
		{MicroRH, "0%rH"},
		{10 * MicroRH, "0%rH"},
		{100 * MicroRH, "0%rH"},
		{1000 * MicroRH, "0.1%rH"},
		{506000 * MicroRH, "50.6%rH"},
		{90 * PercentRH, "90%rH"},
		{100 * PercentRH, "100%rH"},
		// That's a lot of humidity. This is to test the value doesn't overflow
		// int32 too quickly.
		{1000 * PercentRH, "1000%rH"},
		// That's really dry.
		{-501000 * MicroRH, "-50.1%rH"},
	}
	for i, line := range data {
		if s := line.in.String(); s != line.expected {
			t.Fatalf("%d: RelativeHumidity(%d).String() = %s != %s", i, int64(line.in), s, line.expected)
		}
	}
}

func TestSpeed_String(t *testing.T) {
	if s := MilePerHour.String(); s != "447.040mm/s" {
		t.Fatalf("%#v", s)
	}
}

func TestTemperature_String(t *testing.T) {
	if s := ZeroCelsius.String(); s != "0°C" {
		t.Fatalf("%#v", s)
	}
	if s := Temperature(0).String(); s != "-273.150°C" {
		t.Fatalf("%#v", s)
	}
}

func TestPower_String(t *testing.T) {
	if s := NanoWatt.String(); s != "1nW" {
		t.Fatalf("%v", s)
	}
	if s := MicroWatt.String(); s != "1µW" {
		t.Fatalf("%v", s)
	}
	if s := MilliWatt.String(); s != "1mW" {
		t.Fatalf("%v", s)
	}
	if s := Watt.String(); s != "1W" {
		t.Fatalf("%v", s)
	}
	if s := KiloWatt.String(); s != "1kW" {
		t.Fatalf("%v", s)
	}
	if s := MegaWatt.String(); s != "1MW" {
		t.Fatalf("%v", s)
	}
	if s := GigaWatt.String(); s != "1GW" {
		t.Fatalf("%v", s)
	}
}
func TestEnergy_String(t *testing.T) {
	if s := NanoJoule.String(); s != "1nJ" {
		t.Fatalf("%v", s)
	}
	if s := MicroJoule.String(); s != "1µJ" {
		t.Fatalf("%v", s)
	}
	if s := MilliJoule.String(); s != "1mJ" {
		t.Fatalf("%v", s)
	}
	if s := Joule.String(); s != "1J" {
		t.Fatalf("%v", s)
	}
	if s := KiloJoule.String(); s != "1kJ" {
		t.Fatalf("%v", s)
	}
	if s := MegaJoule.String(); s != "1MJ" {
		t.Fatalf("%v", s)
	}
	if s := GigaJoule.String(); s != "1GJ" {
		t.Fatalf("%v", s)
	}
}

func TestCapacitance_String(t *testing.T) {
	if s := PicoFarad.String(); s != "1pF" {
		t.Fatalf("%v", s)
	}
	if s := NanoFarad.String(); s != "1nF" {
		t.Fatalf("%v", s)
	}
	if s := MicroFarad.String(); s != "1µF" {
		t.Fatalf("%v", s)
	}
	if s := MilliFarad.String(); s != "1mF" {
		t.Fatalf("%v", s)
	}
	if s := Farad.String(); s != "1F" {
		t.Fatalf("%v", s)
	}
	if s := KiloFarad.String(); s != "1kF" {
		t.Fatalf("%v", s)
	}
	if s := MegaFarad.String(); s != "1MF" {
		t.Fatalf("%v", s)
	}
}

func TestLuminousIntensity_String(t *testing.T) {
	if s := NanoCandela.String(); s != "1ncd" {
		t.Fatalf("%v", s)
	}
	if s := MicroCandela.String(); s != "1µcd" {
		t.Fatalf("%v", s)
	}
	if s := MilliCandela.String(); s != "1mcd" {
		t.Fatalf("%v", s)
	}
	if s := Candela.String(); s != "1cd" {
		t.Fatalf("%v", s)
	}
	if s := KiloCandela.String(); s != "1kcd" {
		t.Fatalf("%v", s)
	}
	if s := MegaCandela.String(); s != "1Mcd" {
		t.Fatalf("%v", s)
	}
	if s := GigaCandela.String(); s != "1Gcd" {
		t.Fatalf("%v", s)
	}
}

func TestFlux_String(t *testing.T) {
	if s := NanoLumen.String(); s != "1nlm" {
		t.Fatalf("%v", s)
	}
	if s := MicroLumen.String(); s != "1µlm" {
		t.Fatalf("%v", s)
	}
	if s := MilliLumen.String(); s != "1mlm" {
		t.Fatalf("%v", s)
	}
	if s := Lumen.String(); s != "1lm" {
		t.Fatalf("%v", s)
	}
	if s := KiloLumen.String(); s != "1klm" {
		t.Fatalf("%v", s)
	}
	if s := MegaLumen.String(); s != "1Mlm" {
		t.Fatalf("%v", s)
	}
	if s := GigaLumen.String(); s != "1Glm" {
		t.Fatalf("%v", s)
	}
}

func TestPicoAsString(t *testing.T) {
	data := []struct {
		in       int64
		expected string
	}{
		{0, "0"}, // 0
		{1, "1p"},
		{-1, "-1p"},
		{900, "900p"},
		{-900, "-900p"},
		{999, "999p"},
		{-999, "-999p"},
		{1000, "1n"},
		{-1000, "-1n"},
		{1100, "1.100n"},
		{-1100, "-1.100n"}, // 10
		{999999, "999.999n"},
		{-999999, "-999.999n"},
		{1000000, "1µ"},
		{-1000000, "-1µ"},
		{1000501, "1.001µ"},
		{-1000501, "-1.001µ"},
		{1100000, "1.100µ"},
		{-1100000, "-1.100µ"},
		{999999501, "1m"},
		{-999999501, "-1m"},
		{999999999, "1m"},
		{-999999999, "-1m"},
		{1000000000, "1m"},
		{-1000000000, "-1m"}, // 20
		{1100000000, "1.100m"},
		{-1100000000, "-1.100m"},
		{999999499999, "999.999m"},
		{-999999499999, "-999.999m"},
		{999999500001, "1"},
		{-999999500001, "-1"},
		{1000000000000, "1"},
		{-1000000000000, "-1"},
		{1100000000000, "1.100"},
		{-1100000000000, "-1.100"},
		{999999499999999, "999.999"},
		{-999999499999999, "-999.999"},
		{999999500000001, "1k"},
		{-999999500000001, "-1k"},
		{1000000000000000, "1k"}, //30
		{-1000000000000000, "-1k"},
		{1100000000000000, "1.100k"},
		{-1100000000000000, "-1.100k"},
		{999999499999999999, "999.999k"},
		{-999999499999999999, "-999.999k"},
		{999999500000000001, "1M"},
		{-999999500000000001, "-1M"},
		{1000000000000000000, "1M"},
		{-1000000000000000000, "-1M"},
		{1100000000000000000, "1.100M"},
		{-1100000000000000000, "-1.100M"},
		{-1999499999999999999, "-1.999M"},
		{1999499999999999999, "1.999M"},
		{-1999500000000000001, "-2M"},
		{1999500000000000001, "2M"},
		{9223372036854775807, "9.223M"},
		{-9223372036854775807, "-9.223M"},
		{-9223372036854775808, "-9.223M"},
	}
	for i, line := range data {
		if s := picoAsString(line.in); s != line.expected {
			t.Fatalf("%d: picoAsString(%d).String() = %s != %s", i, line.in, s, line.expected)
		}
	}
}

func TestNanoAsString(t *testing.T) {
	data := []struct {
		in       int64
		expected string
	}{
		{0, "0"}, // 0
		{1, "1n"},
		{-1, "-1n"},
		{900, "900n"},
		{-900, "-900n"},
		{999, "999n"},
		{-999, "-999n"},
		{1000, "1µ"},
		{-1000, "-1µ"},
		{1100, "1.100µ"},
		{-1100, "-1.100µ"}, // 10
		{999999, "999.999µ"},
		{-999999, "-999.999µ"},
		{1000000, "1m"},
		{-1000000, "-1m"},
		{1100000, "1.100m"},
		{1100100, "1.100m"},
		{1101000, "1.101m"},
		{-1100000, "-1.100m"},
		{1100499, "1.100m"},
		{1199999, "1.200m"},
		{4999501, "5m"},
		{1999501, "2m"},
		{-1100501, "-1.101m"},
		{111100501, "111.101m"},
		{999999499, "999.999m"},
		{999999501, "1"},
		{999999999, "1"},
		{1000000000, "1"},
		{-1000000000, "-1"}, // 20
		{1100000000, "1.100"},
		{-1100000000, "-1.100"},
		{1100499000, "1.100"},
		{-1100501000, "-1.101"},
		{999999499000, "999.999"},
		{999999501000, "1k"},
		{999999999999, "1k"},
		{-999999999999, "-1k"},
		{1000000000000, "1k"},
		{-1000000000000, "-1k"},
		{1100000000000, "1.100k"},
		{-1100000000000, "-1.100k"},
		{1100499000000, "1.100k"},
		{1199999000000, "1.200k"},
		{-1100501000000, "-1.101k"},
		{999999499000000, "999.999k"},
		{999999501000000, "1M"},
		{999999999999999, "1M"},
		{-999999999999999, "-1M"}, // 30
		{1000000000000000, "1M"},
		{-1000000000000000, "-1M"},
		{1100000000000000, "1.100M"},
		{-1100000000000000, "-1.100M"},
		{1100499000000000, "1.100M"},
		{-1100501000000000, "-1.101M"},
		{999999499000000000, "999.999M"},
		{999999501100000000, "1G"},
		{999999999999999999, "1G"},
		{-999999999999999999, "-1G"},
		{1000000000000000000, "1G"},
		{-1000000000000000000, "-1G"},
		{1100000000000000000, "1.100G"},
		{-1100000000000000000, "-1.100G"},
		{1999999999999999999, "2G"},
		{-1999999999999999999, "-2G"},
		{1100499000000000000, "1.100G"},
		{-1100501000000000000, "-1.101G"},
		{9223372036854775807, "9.223G"},
		{-9223372036854775807, "-9.223G"},
		{-9223372036854775808, "-9.223G"},
	}
	for i, line := range data {
		if s := nanoAsString(line.in); s != line.expected {
			t.Fatalf("%d: nanoAsString(%d).String() = %s != %s", i, line.in, s, line.expected)
		}
	}
}

func TestMicroAsString(t *testing.T) {
	data := []struct {
		in       int64
		expected string
	}{
		{0, "0"}, // 0
		{1, "1µ"},
		{-1, "-1µ"},
		{900, "900µ"},
		{-900, "-900µ"},
		{999, "999µ"},
		{-999, "-999µ"},
		{1000, "1m"},
		{-1000, "-1m"},
		{1100, "1.100m"},
		{-1100, "-1.100m"}, // 10
		{999999, "999.999m"},
		{-999999, "-999.999m"},
		{1000000, "1"},
		{-1000000, "-1"},
		{1000501, "1.001"},
		{-1000501, "-1.001"},
		{1100000, "1.100"},
		{-1100000, "-1.100"},
		{999999501, "1k"},
		{-999999501, "-1k"},
		{999999999, "1k"},
		{-999999999, "-1k"},
		{1000000000, "1k"},
		{-1000000000, "-1k"}, // 20
		{1100000000, "1.100k"},
		{-1100000000, "-1.100k"},
		{999999499999, "999.999k"},
		{-999999499999, "-999.999k"},
		{999999500001, "1M"},
		{-999999500001, "-1M"},
		{1000000000000, "1M"},
		{-1000000000000, "-1M"},
		{1100000000000, "1.100M"},
		{-1100000000000, "-1.100M"},
		{999999499999999, "999.999M"},
		{-999999499999999, "-999.999M"},
		{999999500000001, "1G"},
		{-999999500000001, "-1G"},
		{1000000000000000, "1G"}, //30
		{-1000000000000000, "-1G"},
		{1100000000000000, "1.100G"},
		{-1100000000000000, "-1.100G"},
		{999999499999999999, "999.999G"},
		{-999999499999999999, "-999.999G"},
		{999999500000000001, "1T"},
		{-999999500000000001, "-1T"},
		{1000000000000000000, "1T"},
		{-1000000000000000000, "-1T"},
		{1100000000000000000, "1.100T"},
		{-1100000000000000000, "-1.100T"},
		{-1999499999999999999, "-1.999T"},
		{1999499999999999999, "1.999T"},
		{-1999500000000000001, "-2T"},
		{1999500000000000001, "2T"},
		{9223372036854775807, "9.223T"},
		{-9223372036854775807, "-9.223T"},
		{-9223372036854775808, "-9.223T"},
	}
	for i, line := range data {
		if s := microAsString(line.in); s != line.expected {
			t.Fatalf("%d: microAsString(%d).String() = %s != %s", i, line.in, s, line.expected)
		}
	}
}

func BenchmarkCelsiusString(b *testing.B) {
	v := 10*Celsius + ZeroCelsius
	buf := bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		buf.WriteString(v.String())
		buf.Reset()
	}
}

func BenchmarkCelsiusFloatf(b *testing.B) {
	v := float64(10)
	buf := bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		buf.WriteString(fmt.Sprintf("%.1f°C", v))
		buf.Reset()
	}
}

func BenchmarkCelsiusFloatg(b *testing.B) {
	v := float64(10)
	buf := bytes.Buffer{}
	for i := 0; i < b.N; i++ {
		buf.WriteString(fmt.Sprintf("%g°C", v))
		buf.Reset()
	}
}

func TestAtod(t *testing.T) {
	const (
		negative = true
		positive = false
	)
	succeeds := []struct {
		in       string
		expected decimal
		n        int
	}{
		{"123456789", decimal{123456789, 0, positive}, 9},
		{"1nM", decimal{1, 0, positive}, 1},
		{"2.2", decimal{22, -1, positive}, 3},
		{"12.5mA", decimal{125, -1, positive}, 4},
		{"-12.5mA", decimal{125, -1, negative}, 5},
		{"1ma1", decimal{1, 0, positive}, 1},
		{"+1ma1", decimal{1, 0, positive}, 2},
		{"-1ma1", decimal{1, 0, negative}, 2},
		{"-0.00001%rH", decimal{1, -5, negative}, 8},
		{"0.00001%rH", decimal{1, -5, positive}, 7},
		{"1.0", decimal{1, 0, positive}, 3},
		{"0.10001", decimal{10001, -5, positive}, 7},
		{"+0.10001", decimal{10001, -5, positive}, 8},
		{"-0.10001", decimal{10001, -5, negative}, 8},
		{"1n", decimal{1, 0, positive}, 1},
		{"1.n", decimal{1, 0, positive}, 2},
		{"-1.n", decimal{1, 0, negative}, 3},
		{"200n", decimal{2, 2, positive}, 3},
		{".01", decimal{1, -2, positive}, 3},
		{"+.01", decimal{1, -2, positive}, 4},
		{"-.01", decimal{1, -2, negative}, 4},
		{"1-2", decimal{1, 0, positive}, 1},
		{"1+2", decimal{1, 0, positive}, 1},
		{"-1-2", decimal{1, 0, negative}, 2},
		{"-1+2", decimal{1, 0, negative}, 2},
		{"+1-2", decimal{1, 0, positive}, 2},
		{"+1+2", decimal{1, 0, positive}, 2},
		{"010", decimal{1, 1, positive}, 3},
		{"001", decimal{1, 0, positive}, 3},
	}

	fails := []struct {
		in       string
		expected decimal
		n        int
	}{
		{"1.1.1", decimal{}, 0},
		{"1a2b3a", decimal{}, 0},
		{"aba", decimal{}, 0},
		{"%-0.10001", decimal{}, 0},
		{"--100ma", decimal{}, 0},
		{"++100ma", decimal{}, 0},
		{"+-100ma", decimal{}, 0},
		{"-+100ma", decimal{}, 0},
	}

	for _, tt := range succeeds {
		got, n, err := atod(tt.in)

		if got != tt.expected {
			t.Errorf("case atod(\"%s\") got %v expected %v", tt.in, got, tt.expected)
		}
		if err != nil {
			t.Errorf("case atod(\"%s\") unexpected expected error %v", tt.in, err)
		}
		if n != tt.n {
			t.Errorf("case atod(\"%s\") expected to consume %d char but used %d", tt.in, tt.n, n)
		}
	}

	for _, tt := range fails {
		got, n, err := atod(tt.in)

		if got != tt.expected {
			t.Errorf("case atod(\"%s\") got %v expected %v", tt.in, got, tt.expected)
		}
		if err == nil {
			t.Errorf("case atod(\"%s\") expected error %v", tt.in, err)
		}
		if n != tt.n {
			t.Errorf("case atod(\"%s\") expected to consume %d char but used %d", tt.in, tt.n, n)
		}
	}
}

func TestDoti(t *testing.T) {
	const (
		negative = true
		positive = false
	)
	succeeds := []struct {
		name     string
		in       decimal
		expected int64
	}{
		{"123", decimal{123, 0, positive}, 123},
		{"-123", decimal{123, 0, negative}, -123},
		{"1230", decimal{123, 1, positive}, 1230},
		{"-1230", decimal{123, 1, negative}, -1230},
		{"12.3", decimal{123, -1, positive}, 12},
		{"-12.3", decimal{123, -1, negative}, -12},
		{"123n", decimal{123, 0, positive}, 123},
		{"max", decimal{9223372036854775807, 0, positive}, 9223372036854775807},
		{"rounding(5.6)", decimal{56, -1, positive}, 6},
		{"rounding(5.5)", decimal{55, -1, positive}, 6},
		{"rounding(5.4)", decimal{54, -1, positive}, 5},
		{"rounding(-5.6)", decimal{56, -1, negative}, -6},
		{"rounding(-5.5)", decimal{55, -1, negative}, -6},
		{"rounding(-5.4)", decimal{54, -1, negative}, -5},
		{"rounding(0.6)", decimal{6, -1, positive}, 1},
		{"rounding(0.5)", decimal{5, -1, positive}, 1},
		{"rounding(0.4)", decimal{4, -1, positive}, 0},
		{"rounding(-0.6)", decimal{6, -1, negative}, -1},
		{"rounding(-0.5)", decimal{5, -1, negative}, -1},
		{"rounding(-0.4)", decimal{4, -1, negative}, -0},
	}

	fails := []struct {
		name     string
		in       decimal
		expected int64
	}{
		{"max+1", decimal{9223372036854775808, 0, positive}, 9223372036854775807},
		{"-max-1", decimal{9223372036854775808, 0, negative}, -9223372036854775807},
		{"exponet too large for int64", decimal{123, 20, positive}, 0},
		{"exponet too small for int64", decimal{123, -20, positive}, 0},
		{"max*10^1", decimal{9223372036854775807, 1, positive}, 9223372036854775807},
		{"-max*10^1", decimal{9223372036854775807, 1, negative}, -9223372036854775807},
		{"overflow", decimal{7588728005190, 9, positive}, 9223372036854775807},
	}

	for _, tt := range succeeds {
		got, err := dtoi(tt.in, 0)

		if got != tt.expected {
			t.Errorf("case dtoi() %s got %v expected %v", tt.name, got, tt.expected)
		}
		if err != nil {
			t.Errorf("case dtoi() %s got an unexpected error %v", tt.name, err)
		}
	}

	for _, tt := range fails {
		got, err := dtoi(tt.in, 0)

		if got != tt.expected {
			t.Errorf("case dtoi() %s got %v expected %v", tt.name, got, tt.expected)
		}
		if err == nil {
			t.Errorf("case dtoi() %s expected %v but got nil", tt.name, err)
		}
	}
}

func Test_decimalMulScale(t *testing.T) {
	const (
		negative = true
		positive = false
	)
	succeeds := []struct {
		loss   uint
		a, b   decimal
		expect decimal
	}{
		{
			0,
			decimal{123, 0, positive},
			decimal{123, 0, positive},
			decimal{15129, 0, positive},
		},
		{
			0,
			decimal{123, 0, negative},
			decimal{123, 0, positive},
			decimal{15129, 0, negative},
		},
		{
			0,
			decimal{123, 0, positive},
			decimal{123, 0, negative},
			decimal{15129, 0, negative},
		},
		{
			0,
			decimal{123, 0, negative},
			decimal{123, 0, negative},
			decimal{15129, 0, positive},
		},
		{
			0,
			decimal{1000000001, 0, positive},
			decimal{1000000001, 0, positive},
			decimal{1000000002000000001, 0, positive},
		},
		{
			1,
			decimal{10000000001, 0, positive},
			decimal{10000000001, 0, positive},
			decimal{10000000001, 10, positive},
		},
		{
			2,
			decimal{10000000011, 0, positive},
			decimal{10000000001, 0, positive},
			decimal{1000000001, 11, positive},
		},
		{
			2,
			decimal{10000000011, 0, positive},
			decimal{10000000011, 0, positive},
			decimal{1000000002000000001, 2, positive},
		},
		{
			4,
			decimal{100000000111, 0, positive},
			decimal{100000000111, 0, positive},
			decimal{1000000002000000001, 4, positive},
		},
		{
			6,
			decimal{1000000001111, 0, positive},
			decimal{1000000001111, 0, positive},
			decimal{1000000002000000001, 6, positive},
		},
		{
			8,
			decimal{10000000011111, 0, positive},
			decimal{10000000011111, 0, positive},
			decimal{1000000002000000001, 8, positive},
		},
		{
			10,
			decimal{100000000111111, 0, positive},
			decimal{100000000111111, 0, positive},
			decimal{1000000002000000001, 10, positive},
		},
		{
			12,
			decimal{1000000001111111, 0, positive},
			decimal{1000000001111111, 0, positive},
			decimal{1000000002000000001, 12, positive},
		},
		{
			14,
			decimal{10000000011111111, 0, positive},
			decimal{10000000011111111, 0, positive},
			decimal{1000000002000000001, 14, positive},
		},
		{
			16,
			decimal{100000000111111111, 0, positive},
			decimal{100000000111111111, 0, positive},
			decimal{1000000002000000001, 16, positive},
		},
		{
			18,
			decimal{1000000001111111111, 0, positive},
			decimal{1000000001111111111, 0, positive},
			decimal{1000000002000000001, 18, positive},
		},
		{
			20,
			decimal{10000000011111111111, 0, positive},
			decimal{10000000011111111111, 0, positive},
			decimal{1000000002000000001, 20, positive},
		},
		{
			19,
			decimal{maxInt64, 0, positive},
			decimal{maxInt64, 0, positive},
			decimal{8507059176058364548, 19, positive},
		},
		{
			18,
			decimal{(1 << 64) - 6, 0, positive},
			decimal{(1 << 64) - 6, 0, positive},
			decimal{3402823667840801649, 20, positive},
		},
		{
			0,
			decimal{(1 << 64) - 6, 100, positive},
			decimal{0, 0, positive},
			decimal{0, 0, positive},
		},
	}

	fails := []struct {
		loss   uint
		a, b   decimal
		expect decimal
	}{
		{
			21,
			decimal{(1 << 64) - 5, 0, positive},
			decimal{(1 << 64) - 5, 0, positive},
			decimal{},
		},
	}

	for _, tt := range succeeds {
		got, loss := decimalMul(tt.a, tt.b)
		if loss != tt.loss {
			t.Errorf("decimalMulScale(%v,%v) expected %d loss but got %d", tt.a, tt.b, tt.loss, loss)
		}
		if got != tt.expect {
			t.Errorf("decimalMulScale(%v,%v) got: %v expected: %v", tt.a, tt.b, got, tt.expect)
		}
	}

	for _, tt := range fails {
		got, loss := decimalMul(tt.a, tt.b)
		if loss != tt.loss {
			t.Errorf("decimalMulScale(%v,%v) expected %d loss but got %d", tt.a, tt.b, tt.loss, loss)
		}
		if got != tt.expect {
			t.Errorf("decimalMulScale(%v,%v) got: %v expected: %v", tt.a, tt.b, got, tt.expect)
		}
	}
}

func TestPrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix rune
		want   prefix
		n      int
	}{
		{"pico", 'p', pico, 1},
		{"nano", 'n', nano, 1},
		{"micro", 'u', micro, 1},
		{"mu", 'µ', micro, 2},
		{"milli", 'm', milli, 1},
		{"unit", 0, unit, 0},
		{"kilo", 'k', kilo, 1},
		{"mega", 'M', mega, 1},
		{"giga", 'G', giga, 1},
		{"tera", 'T', tera, 1},
	}
	for _, tt := range tests {
		got, n := parseSIPrefix(tt.prefix)

		if got != tt.want || n != tt.n {
			t.Errorf("wanted prefix %d, and len %d, but got prefix %d, and len %d", tt.want, tt.n, got, n)
		}
	}
}

func TestParseError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"empty", &parseError{msg: "", err: nil}, "parse error"},
		{"empty", &parseError{msg: "", err: errors.New("test")}, "test"},
		{"noUnits", noUnits("someunit"), "no units provided, need someunit"},
	}
	for _, tt := range tests {
		got := tt.err.Error()

		if got != tt.want {
			t.Errorf("wanted err string:\n%s but got:\n%s", tt.want, got)
		}
	}
}

func TestMaxInt64(t *testing.T) {
	if strconv.FormatUint(maxInt64, 10) != maxInt64Str {
		t.Fatal("unexpected text representation of max")
	}
}

func TestValueOfUnitString(t *testing.T) {
	succeeds := []struct {
		in        string
		uintbase  prefix
		expected  int64
		usedChars int
	}{
		{"1p", pico, 1, 2},
		{"1n", pico, 1000, 2},
		{"1u", pico, 1000000, 2},
		{"1µ", pico, 1000000, 3},
		{"1m", pico, 1000000000, 2},
		{"1k", pico, 1000000000000000, 2},
		{"1M", pico, 1000000000000000000, 2},
		{"9.223372036854775807M", pico, 9223372036854775807, 21},
		{"9223372036854775807p", pico, 9223372036854775807, 20},
		{"-1p", pico, -1, 3},
		{"-1n", pico, -1000, 3},
		{"-1u", pico, -1000000, 3},
		{"-1µ", pico, -1000000, 4},
		{"-1m", pico, -1000000000, 3},
		{"-1k", pico, -1000000000000000, 3},
		{"-1M", pico, -1000000000000000000, 3},
		{"-9.223372036854775807M", pico, -9223372036854775807, 22},
		{"-9223372036854775807p", pico, -9223372036854775807, 21},
		{"1p", nano, 0, 2},
		{"1n", nano, 1, 2},
		{"1u", nano, 1000, 2},
		{"1µ", nano, 1000, 3},
		{"1m", nano, 1000000, 2},
		{"1k", nano, 1000000000000, 2},
		{"1M", nano, 1000000000000000, 2},
		{"1G", nano, 1000000000000000000, 2},
		{"9.223372036854775807G", nano, 9223372036854775807, 21},
		{"9223372036854775807n", nano, 9223372036854775807, 20},
		{"-1p", nano, -0, 3},
		{"-1n", nano, -1, 3},
		{"-1u", nano, -1000, 3},
		{"-1µ", nano, -1000, 4},
		{"-1m", nano, -1000000, 3},
		{"-1k", nano, -1000000000000, 3},
		{"-1M", nano, -1000000000000000, 3},
		{"-1G", nano, -1000000000000000000, 3},
		{"-9.223372036854775807G", nano, -9223372036854775807, 22},
		{"-9223372036854775807n", nano, -9223372036854775807, 21},
		{"1p", micro, 0, 2},
		{"1n", micro, 0, 2},
		{"1u", micro, 1, 2},
		{"1µ", micro, 1, 3},
		{"1m", micro, 1000, 2},
		{"1k", micro, 1000000000, 2},
		{"1M", micro, 1000000000000, 2},
		{"1G", micro, 1000000000000000, 2},
		{"1T", micro, 1000000000000000000, 2},
		{"9.223372036854775807T", micro, 9223372036854775807, 21},
		{"9223372036854775807u", micro, 9223372036854775807, 20},
		{"-1p", micro, -0, 3},
		{"-1n", micro, -0, 3},
		{"-1u", micro, -1, 3},
		{"-1µ", micro, -1, 4},
		{"-1m", micro, -1000, 3},
		{"-1k", micro, -1000000000, 3},
		{"-1M", micro, -1000000000000, 3},
		{"-1G", micro, -1000000000000000, 3},
		{"-1T", micro, -1000000000000000000, 3},
		{"-9.223372036854775807T", micro, -9223372036854775807, 22},
		{"-9223372036854775807u", micro, -9223372036854775807, 21},
	}

	fails := []struct {
		in     string
		prefix prefix
	}{
		{"9.223372036854775808M", pico},
		{"9.223372036854775808G", nano},
		{"9.223372036854775808T", micro},
		{"9223372036854775808p", pico},
		{"9223372036854775808n", nano},
		{"9223372036854775808u", micro},
		{"-9.223372036854775808M", pico},
		{"-9.223372036854775808G", nano},
		{"-9.223372036854775808T", micro},
		{"-9223372036854775808p", pico},
		{"-9223372036854775808n", nano},
		{"-9223372036854775808u", micro},
		{"not a number", nano},
		{string([]byte{0x31, 0x01}), nano}, // 0x01 is a invalid utf8 start byte.
	}

	for _, tt := range succeeds {
		got, used, err := valueOfUnitString(tt.in, tt.uintbase)

		if got != tt.expected {
			t.Errorf("valueOfUnitString(%s,%d) wanted: %v(%d) but got: %v(%d)", tt.in, tt.uintbase, tt.expected, tt.expected, got, got)
		}
		if used != tt.usedChars {
			t.Errorf("valueOfUnitString(%s,%d) used %d chars but should use: %d chars", tt.in, tt.uintbase, used, tt.usedChars)
		}
		if err != nil {
			t.Errorf("valueOfUnitString(%s,%d) unexpected error: %v", tt.in, tt.uintbase, err)
		}
	}

	for _, tt := range fails {
		_, _, err := valueOfUnitString(tt.in, tt.prefix)

		if err == nil {
			t.Errorf("valueOfUnitString(%s,%d) expected an error", tt.in, tt.prefix)
		}
	}
}

func TestAngleSet(t *testing.T) {
	succeeds := []struct {
		in       string
		expected Angle
	}{
		{"1nrad", NanoRadian},
		{"10nrad", 10 * NanoRadian},
		{"100nrad", 100 * NanoRadian},
		{"1urad", 1 * MicroRadian},
		{"10urad", 10 * MicroRadian},
		{"100urad", 100 * MicroRadian},
		{"1µrad", 1 * MicroRadian},
		{"10µrad", 10 * MicroRadian},
		{"100µrad", 100 * MicroRadian},
		{"1mrad", 1 * MilliRadian},
		{"10mrad", 10 * MilliRadian},
		{"100mrad", 100 * MilliRadian},
		{"1rad", 1 * Radian},
		{"10rad", 10 * Radian},
		{"100rad", 100 * Radian},
		{"1krad", 1000 * Radian},
		{"10krad", 10000 * Radian},
		{"100krad", 100000 * Radian},
		{"1Mrad", 1000000 * Radian},
		{"10Mrad", 10000000 * Radian},
		{"100Mrad", 100000000 * Radian},
		{"1Grad", 1000000000 * Radian},
		{"12.345rad", 12345 * MilliRadian},
		{"-12.345rad", -12345 * MilliRadian},
		{fmt.Sprintf("%dnrad", maxAngle), maxAngle},
		{"1deg", 1 * Degree},
		{"1Mdeg", 1000000 * Degree},
		{"100Gdeg", 100000000000 * Degree},
		{"500Gdeg", 500000000000 * Degree},
		{maxAngle.String(), 528460276055 * Degree},
		{minAngle.String(), -528460276055 * Degree},
		{"1mdeg", Degree / 1000},
		{"1udeg", Degree / 1000000},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10000000000Tdeg",
			"exponent exceeds int64",
		},
		{
			"10Trad",
			"exponent exceeds int64",
		},
		{
			"10Erad",
			"contains unknown unit prefix \"E\". valid prefixes for \"Rad\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10Exarad",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"Rad\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eRadianE",
			"contains unknown unit prefix \"e\". valid prefixes for \"Rad\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need Rad",
		},
		{
			fmt.Sprintf("%dnrad", uint64(maxAngle)+1),
			"maximum value is 528460276055°",
		},
		{
			fmt.Sprintf("-%dnrad", uint64(maxAngle)+1),
			"minimum value is -528460276055°",
		},
		{
			"528460276056deg",
			"maximum value is 528460276055°",
		},
		{
			"-528460276056deg",
			"minimum value is -528460276055°",
		},
		{
			"-9.223372036854775808Grad",
			"minimum value is -528460276055°",
		},
		{
			"9.223372036854775808Grad",
			"maximum value is 528460276055°",
		},
		{
			"9.224Grad",
			"maximum value is 528460276055°",
		},
		{
			"-9.224Grad",
			"minimum value is -528460276055°",
		},
		{
			"1cup",
			"\"cup\" is not a valid unit for physic.Angle",
		},
		{
			"rad",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"Rad\"",
		},
		{
			"++1rad",
			"multiple plus symbols ++1rad",
		},
		{
			"--1rad",
			"multiple minus symbols --1rad",
		},
		{
			"+-1rad",
			"can't contain both plus and minus symbols +-1rad",
		},
		{
			"1.1.1.1rad",
			"multiple decimal points 1.1.1.1rad",
		},
		{
			string([]byte{0x33, 0x01}),
			"unexpected end of string",
		},
	}

	for _, tt := range succeeds {
		var got Angle
		if err := got.Set(tt.in); err != nil {
			t.Errorf("Angle.Set(%s) got unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("Angle.Set(%s) expected: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got Angle
		if err := got.Set(tt.in); err != nil {
			if err.Error() != tt.err {
				t.Errorf("Angle.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
			}
		} else {
			t.Errorf("Angle.Set(%s) expected error: %s but got none", tt.in, tt.err)
		}
	}
}

func TestDistance_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected Distance
	}{
		{"1nm", 1 * NanoMetre},
		{"10nm", 10 * NanoMetre},
		{"100nm", 100 * NanoMetre},
		{"1um", 1 * MicroMetre},
		{"10um", 10 * MicroMetre},
		{"100um", 100 * MicroMetre},
		{"1µm", 1 * MicroMetre},
		{"10µm", 10 * MicroMetre},
		{"100µm", 100 * MicroMetre},
		{"1mm", 1 * MilliMetre},
		{"1mm", 1 * MilliMetre},
		{"10mm", 10 * MilliMetre},
		{"100mm", 100 * MilliMetre},
		{"1m", 1 * Metre},
		{"10m", 10 * Metre},
		{"100m", 100 * Metre},
		{"1km", 1 * KiloMetre},
		{"10km", 10 * KiloMetre},
		{"100km", 100 * KiloMetre},
		{"1Mm", 1 * MegaMetre},
		{"1Mm", 1 * MegaMetre},
		{"10Mm", 10 * MegaMetre},
		{"100Mm", 100 * MegaMetre},
		{"1Gm", 1 * GigaMetre},
		{"12.345m", 12345 * MilliMetre},
		{"-12.345m", -12345 * MilliMetre},
		{"9.223372036854775807Gm", 9223372036854775807 * NanoMetre},
		{"-9.223372036854775807Gm", -9223372036854775807 * NanoMetre},
		{"1Mm", 1 * MegaMetre},
		{"5Mile", 8046720000000 * NanoMetre},
		{"3ft", 914400000 * NanoMetre},
		{"10Yard", 9144000000 * NanoMetre},
		{"5731.137678988Mile", 9223372036853264 * NanoMetre},
		{"-5731.137678988Mile", -9223372036853264 * NanoMetre},
		{"1.008680231502051MYard", 922337203685475 * NanoMetre},
		{"-1008680.231502051Yard", -922337203685475 * NanoMetre},
		{"3026040.694506158ft", 922337203685477 * NanoMetre},
		{"-3.026040694506158Mft", -922337203685477 * NanoMetre},
		{"36.312488334073900Min", 922337203685477 * NanoMetre},
		{"-36312488.334073900in", -922337203685477 * NanoMetre},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10Tm",
			"exponent exceeds int64",
		},
		{
			"10Em",
			"contains unknown unit prefix \"E\". valid prefixes for \"m\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10Exam",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"m\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eMetreE",
			"contains unknown unit prefix \"e\". valid prefixes for \"m\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need m, Mile, in, ft or Yard",
		},
		{
			"9.3Gm",
			"maximum value is 9.223Gm",
		},
		{
			"-9.3Gm",
			"minimum value is -9.223Gm",
		},
		{
			"9223372036854775808",
			"maximum value is 9.223Gm",
		},
		{
			"-9223372036854775808",
			"minimum value is -9.223Gm",
		},
		{
			"9.223372036854775808Gm",
			"maximum value is 9.223Gm",
		},
		{
			"-9.223372036854775808Gm",
			"minimum value is -9.223Gm",
		},
		{
			"9.223372036854775808Gm",
			"maximum value is 9.223Gm",
		},
		{
			"-9.223372036854775808Gm",
			"minimum value is -9.223Gm",
		},
		{
			"5731.137678989Mile",
			"maximum value is 5731Mile",
		},
		{
			"-5731.1376789889Mile",
			"minimum value is -5731Mile",
		},
		{
			"1.008680231502053MYard",
			"maximum value is 1 Million Yard",
		},
		{
			"-1008680.231502053Yard",
			"minimum value is -1 Million Yard",
		},
		{
			"3026040.694506159ft",
			"maximum value is 3 Million ft",
		},
		{
			"-3.026040694506159Mft",
			"minimum value is 3 Million ft",
		},
		{
			"36.312488334073901Min",
			"maximum value is 36 Million inch",
		},
		{
			"-36312488.334073901in",
			"minimum value is 36 Million inch",
		},
		{
			"1random",
			"contains unknown unit prefix \"rando\". valid prefixes for \"m\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"m",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number",
		},
		{
			"cd",
			"does not contain number or unit \"m\"",
		},
		{
			"1Jaunt",
			"\"Jaunt\" is not a valid unit for physic.Distance",
		},
		{
			"++1m",
			"multiple plus symbols ++1m",
		},
		{
			"--1m",
			"multiple minus symbols --1m",
		},
		{
			"+-1m",
			"can't contain both plus and minus symbols +-1m",
		},
		{
			"1.1.1.1m",
			"multiple decimal points 1.1.1.1m",
		},
		{
			string([]byte{0x31, 0x01}),
			"unexpected end of string",
		},
	}

	for _, tt := range succeeds {
		var got Distance
		if err := got.Set(tt.in); err != nil {
			t.Errorf("Distance.Set(%s) got unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("Distance.Set(%s) expected: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got Distance
		if err := got.Set(tt.in); err != nil {
			if err.Error() != tt.err {
				t.Errorf("Distance.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
			}
		} else {
			t.Errorf("Distance.Set(%s) expected error: %s but got none", tt.in, tt.err)
		}
	}
}

func TestElectricPotential_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected ElectricPotential
	}{
		{"1nV", 1 * NanoVolt},
		{"10nV", 10 * NanoVolt},
		{"100nV", 100 * NanoVolt},
		{"1uV", 1 * MicroVolt},
		{"10uV", 10 * MicroVolt},
		{"100uV", 100 * MicroVolt},
		{"1µV", 1 * MicroVolt},
		{"10µV", 10 * MicroVolt},
		{"100µV", 100 * MicroVolt},
		{"1mV", 1 * MilliVolt},
		{"10mV", 10 * MilliVolt},
		{"100mV", 100 * MilliVolt},
		{"1V", 1 * Volt},
		{"10V", 10 * Volt},
		{"100V", 100 * Volt},
		{"1kV", 1 * KiloVolt},
		{"10kV", 10 * KiloVolt},
		{"100kV", 100 * KiloVolt},
		{"1MV", 1 * MegaVolt},
		{"10MV", 10 * MegaVolt},
		{"100MV", 100 * MegaVolt},
		{"1GV", 1 * GigaVolt},
		{"12.345V", 12345 * MilliVolt},
		{"-12.345V", -12345 * MilliVolt},
		{"9.223372036854775807GV", 9223372036854775807 * NanoVolt},
		{"-9.223372036854775807GV", -9223372036854775807 * NanoVolt},
		{"1MV", 1 * MegaVolt},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TV",
			"exponent exceeds int64",
		},
		{
			"10EV",
			"contains unknown unit prefix \"E\". valid prefixes for \"V\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eVoltE",
			"contains unknown unit prefix \"e\". valid prefixes for \"V\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need V",
		},
		{
			"9223372036854775808",
			"maximum value is 9.223GV",
		},
		{
			"-9223372036854775808",
			"minimum value is -9.223GV",
		},
		{
			"9.223372036854775808TV",
			"maximum value is 9.223GV",
		},
		{
			"-9.223372036854775808GV",
			"minimum value is -9.223GV",
		},
		{
			"9.223372036854775808GV",
			"maximum value is 9.223GV",
		},
		{
			"-9.223372036854775808GOhm",
			"minimum value is -9.223GV",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.ElectricPotential",
		},
		{
			"V",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"V\"",
		},
		{
			"++1V",
			"multiple plus symbols ++1V",
		},
		{
			"--1V",
			"multiple minus symbols --1V",
		},
		{
			"+-1V",
			"can't contain both plus and minus symbols +-1V",
		},
		{
			"1.1.1.1V",
			"multiple decimal points 1.1.1.1V",
		},
	}

	for _, tt := range succeeds {
		var got ElectricPotential
		if err := got.Set(tt.in); err != nil {
			t.Errorf("ElectricPotential.Set(%s) got unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("ElectricPotential.Set(%s) expected: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got ElectricPotential
		if err := got.Set(tt.in); err.Error() != tt.err {
			t.Errorf("ElectricPotential.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
		}
	}
}

func TestElectricCurrent_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected ElectricCurrent
	}{
		{"1nA", 1 * NanoAmpere},
		{"10nA", 10 * NanoAmpere},
		{"100nA", 100 * NanoAmpere},
		{"1uA", 1 * MicroAmpere},
		{"10uA", 10 * MicroAmpere},
		{"100uA", 100 * MicroAmpere},
		{"1µA", 1 * MicroAmpere},
		{"10µA", 10 * MicroAmpere},
		{"100µA", 100 * MicroAmpere},
		{"1mA", 1 * MilliAmpere},
		{"10mA", 10 * MilliAmpere},
		{"100mA", 100 * MilliAmpere},
		{"1A", 1 * Ampere},
		{"10A", 10 * Ampere},
		{"100A", 100 * Ampere},
		{"1kA", 1 * KiloAmpere},
		{"10kA", 10 * KiloAmpere},
		{"100kA", 100 * KiloAmpere},
		{"1MA", 1 * MegaAmpere},
		{"10MA", 10 * MegaAmpere},
		{"100MA", 100 * MegaAmpere},
		{"1GA", 1 * GigaAmpere},
		{"12.345A", 12345 * MilliAmpere},
		{"-12.345A", -12345 * MilliAmpere},
		{"9.223372036854775807GA", 9223372036854775807 * NanoAmpere},
		{"-9.223372036854775807GA", -9223372036854775807 * NanoAmpere},
		{"1A", 1 * Ampere},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TA",
			"exponent exceeds int64",
		},
		{
			"10EA",
			"contains unknown unit prefix \"E\". valid prefixes for \"A\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eAmpE",
			"contains unknown unit prefix \"e\". valid prefixes for \"A\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need A",
		},
		{
			"922337203685477580",
			"maximum value is 9.223GA",
		},
		{
			"-922337203685477580",
			"minimum value is -9.223GA",
		},
		{
			"9.223372036854775808GA",
			"maximum value is 9.223GA",
		},
		{
			"-9.223372036854775808GA",
			"minimum value is -9.223GA",
		},
		{
			"9.223372036854775808GA",
			"maximum value is 9.223GA",
		},
		{
			"-9.223372036854775808GA",
			"minimum value is -9.223GA",
		},
		{
			"1junk",
			"\"junk\" is not a valid unit for physic.ElectricCurrent",
		},
		{
			"A",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"A\"",
		},
		{
			"++1A",
			"multiple plus symbols ++1A",
		},
		{
			"--1A",
			"multiple minus symbols --1A",
		},
		{
			"+-1A",
			"can't contain both plus and minus symbols +-1A",
		},
		{
			"1.1.1.1A",
			"multiple decimal points 1.1.1.1A",
		},
	}

	for _, tt := range succeeds {
		var got ElectricCurrent
		if err := got.Set(tt.in); err != nil {
			t.Errorf("ElectricCurrent.Set(%s) unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("ElectricCurrent.Set(%s) wanted: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got ElectricCurrent
		if err := got.Set(tt.in); err.Error() != tt.err {
			t.Errorf("ElectricCurrent.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
		}
	}
}

func TestElectricResistance_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected ElectricResistance
	}{
		{"1nOhm", 1 * NanoOhm},
		{"10nOhm", 10 * NanoOhm},
		{"100nOhm", 100 * NanoOhm},
		{"1uOhm", 1 * MicroOhm},
		{"10uOhm", 10 * MicroOhm},
		{"100uOhm", 100 * MicroOhm},
		{"1µOhm", 1 * MicroOhm},
		{"10µOhm", 10 * MicroOhm},
		{"100µOhm", 100 * MicroOhm},
		{"1mOhm", 1 * MilliOhm},
		{"10mOhm", 10 * MilliOhm},
		{"100mOhm", 100 * MilliOhm},
		{"1Ohm", 1 * Ohm},
		{"10Ohm", 10 * Ohm},
		{"100Ohm", 100 * Ohm},
		{"1kOhm", 1 * KiloOhm},
		{"10kOhm", 10 * KiloOhm},
		{"100kOhm", 100 * KiloOhm},
		{"1MOhm", 1 * MegaOhm},
		{"10MOhm", 10 * MegaOhm},
		{"100MOhm", 100 * MegaOhm},
		{"1GOhm", 1 * GigaOhm},
		{"12.345Ohm", 12345 * MilliOhm},
		{"-12.345Ohm", -12345 * MilliOhm},
		{"9.223372036854775807GOhm", 9223372036854775807 * NanoOhm},
		{"-9.223372036854775807GOhm", -9223372036854775807 * NanoOhm},
		{"1MΩ", 1 * MegaOhm},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TOhm",
			"exponent exceeds int64",
		},
		{
			"10EOhm",
			"contains unknown unit prefix \"E\". valid prefixes for \"Ohm\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaOhm",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"Ohm\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eOhmE",
			"contains unknown unit prefix \"e\". valid prefixes for \"Ohm\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need Ohm",
		},
		{
			"9223372036854775808",
			"maximum value is 9.223GΩ",
		},
		{
			"-9223372036854775808",
			"minimum value is -9.223GΩ",
		},
		{
			"9.223372036854775808GOhm",
			"maximum value is 9.223GΩ",
		},
		{
			"-9.223372036854775808GOhm",
			"minimum value is -9.223GΩ",
		},
		{
			"9.223372036854775808GOhm",
			"maximum value is 9.223GΩ",
		},
		{
			"-9.223372036854775808GOhm",
			"minimum value is -9.223GΩ",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.ElectricResistance",
		},
		{
			"Ohm",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"Ohm\"",
		},
		{
			"++1Ohm",
			"multiple plus symbols ++1Ohm",
		},
		{
			"--1Ohm",
			"multiple minus symbols --1Ohm",
		},
		{
			"+-1Ohm",
			"can't contain both plus and minus symbols +-1Ohm",
		},
		{
			"1.1.1.1Ohm",
			"multiple decimal points 1.1.1.1Ohm",
		},
	}

	for _, tt := range succeeds {
		var got ElectricResistance
		if err := got.Set(tt.in); err != nil {
			t.Errorf("ElectricResistance.Set(%s) got unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("ElectricResistance.Set(%s) expected: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got ElectricResistance
		if err := got.Set(tt.in); err.Error() != tt.err {
			t.Errorf("ElectricResistance.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
		}
	}
}

func TestForceSet(t *testing.T) {
	succeeds := []struct {
		in       string
		expected Force
	}{
		{"1nN", 1 * NanoNewton},
		{"10nN", 10 * NanoNewton},
		{"100nN", 100 * NanoNewton},
		{"1uN", 1 * MicroNewton},
		{"10uN", 10 * MicroNewton},
		{"100uN", 100 * MicroNewton},
		{"1µN", 1 * MicroNewton},
		{"10µN", 10 * MicroNewton},
		{"100µN", 100 * MicroNewton},
		{"1mN", 1 * MilliNewton},
		{"10mN", 10 * MilliNewton},
		{"100mN", 100 * MilliNewton},
		{"1N", 1 * Newton},
		{"10N", 10 * Newton},
		{"100N", 100 * Newton},
		{"1kN", 1 * KiloNewton},
		{"10kN", 10 * KiloNewton},
		{"100kN", 100 * KiloNewton},
		{"1MN", 1 * MegaNewton},
		{"10MN", 10 * MegaNewton},
		{"100MN", 100 * MegaNewton},
		{"1GN", 1 * GigaNewton},
		{"12.345N", 12345 * MilliNewton},
		{"-12.345N", -12345 * MilliNewton},
		{"9.223372036854775807GN", 9223372036854775807 * NanoNewton},
		{"-9.223372036854775807GN", -9223372036854775807 * NanoNewton},
		{"1MN", 1 * MegaNewton},
		{"1nN", 1 * NanoNewton},
		{"1mlbf", 4448222 * NanoNewton},
		{"1lbf", 1 * PoundForce},
		{"1lbf", 4448221615 * NanoNewton},
		{"20lbf", 88964432305 * NanoNewton},
		{"1klbf", 4448221615261 * NanoNewton},
		{"1Mlbf", 4448221615261000 * NanoNewton},
		{"2Mlbf", 8896443230522000 * NanoNewton},
		{"2073496519lbf", 9223372034443058185 * NanoNewton},
		{"1.0000000000101lbf", 4448221615 * NanoNewton},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"2073496520lbf",
			"maximum value is 2.073496519Glbf",
		},
		{
			"-2073496520lbf",
			"minimum value is -2.073496519Glbf",
		},
		{
			"1234567.890123456789lbf",
			"converting to nano Newtons would overflow, consider using nN for maximum precision",
		},
		{
			"100000000000Tlbf",
			"exponent exceeds int64",
		},
		{
			"10TN",
			"exponent exceeds int64",
		},
		{
			"10EN",
			"contains unknown unit prefix \"E\". valid prefixes for \"N\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaN",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"N\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eNewtonE",
			"contains unknown unit prefix \"e\". valid prefixes for \"N\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need N",
		},
		{
			"9223372036854775808",
			"maximum value is 9.223GN",
		},
		{
			"-9223372036854775808",
			"minimum value is -9.223GN",
		},
		{
			"9.223372036854775808GN",
			"maximum value is 9.223GN",
		},
		{
			"-9.223372036854775808GN",
			"minimum value is -9.223GN",
		},
		{
			"9.223372036854775808GN",
			"maximum value is 9.223GN",
		},
		{
			"-9.223372036854775808GN",
			"minimum value is -9.223GN",
		},
		{
			"9.3GN",
			"maximum value is 9.223GN",
		},
		{
			"-9.3GN",
			"minimum value is -9.223GN",
		},
		{
			"1random",
			"contains unknown unit prefix \"ra\". valid prefixes for \"N\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"1cup",
			"\"cup\" is not a valid unit for physic.Force",
		},
		{
			"N",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"N\" or \"lbf\"",
		},
		{
			"++1N",
			"multiple plus symbols ++1N",
		},
		{
			"--1N",
			"multiple minus symbols --1N",
		},
		{
			"+-1N",
			"can't contain both plus and minus symbols +-1N",
		},
		{
			"1.1.1.1N",
			"multiple decimal points 1.1.1.1N",
		},
		{
			string([]byte{0x33, 0x01}),
			"unexpected end of string",
		},
	}

	for _, tt := range succeeds {
		var got Force
		if err := got.Set(tt.in); err != nil {
			t.Errorf("Force.Set(%s) got unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("Force.Set(%s) expected: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got Force
		if err := got.Set(tt.in); err != nil {
			if err.Error() != tt.err {
				t.Errorf("Force.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
			}
		} else {
			t.Errorf("Force.Set(%s) expected error: %s but got none", tt.in, tt.err)
		}
	}
}

func TestFrequency_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected Frequency
	}{
		{"1uHz", 1 * MicroHertz},
		{"10uHz", 10 * MicroHertz},
		{"100uHz", 100 * MicroHertz},
		{"1µHz", 1 * MicroHertz},
		{"10µHz", 10 * MicroHertz},
		{"100µHz", 100 * MicroHertz},
		{"1mHz", 1 * MilliHertz},
		{"10mHz", 10 * MilliHertz},
		{"100mHz", 100 * MilliHertz},
		{"1Hz", 1 * Hertz},
		{"10Hz", 10 * Hertz},
		{"100Hz", 100 * Hertz},
		{"1kHz", 1 * KiloHertz},
		{"10kHz", 10 * KiloHertz},
		{"100kHz", 100 * KiloHertz},
		{"1MHz", 1 * MegaHertz},
		{"10MHz", 10 * MegaHertz},
		{"100MHz", 100 * MegaHertz},
		{"1GHz", 1 * GigaHertz},
		{"10GHz", 10 * GigaHertz},
		{"100GHz", 100 * GigaHertz},
		{"1THz", 1 * TeraHertz},
		{"12.345Hz", 12345 * MilliHertz},
		{"-12.345Hz", -12345 * MilliHertz},
		{"9.223372036854775807THz", 9223372036854775807 * MicroHertz},
		{"-9.223372036854775807THz", -9223372036854775807 * MicroHertz},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10THz",
			"exponent exceeds int64",
		},
		{
			"10EHz",
			"contains unknown unit prefix \"E\". valid prefixes for \"Hz\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaHz",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"Hz\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eHzE",
			"contains unknown unit prefix \"e\". valid prefixes for \"Hz\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need Hz",
		},
		{
			"922337203685477580",
			"maximum value is 9.223THz",
		},
		{
			"-922337203685477580",
			"minimum value is -9.223THz",
		},
		{
			"9.223372036854775808THz",
			"maximum value is 9.223THz",
		},
		{
			"-9.223372036854775808THz",
			"minimum value is -9.223THz",
		},
		{
			"9.223372036854775808THertz",
			"maximum value is 9.223THz",
		},
		{
			"-9.223372036854775808THertz",
			"minimum value is -9.223THz",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.Frequency",
		},
		{
			"Hz",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"Hz\"",
		},
		{
			"++1Hz",
			"multiple plus symbols ++1Hz",
		},
		{
			"--1Hz",
			"multiple minus symbols --1Hz",
		},
		{
			"+-1Hz",
			"can't contain both plus and minus symbols +-1Hz",
		},
		{
			"1.1.1.1Hz",
			"multiple decimal points 1.1.1.1Hz",
		},
	}

	for _, tt := range succeeds {
		var got Frequency
		if err := got.Set(tt.in); err != nil {
			t.Errorf("Frequency.Set(%s) unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("Frequency.Set(%s) wanted: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got Frequency
		if err := got.Set(tt.in); err.Error() != tt.err {
			t.Errorf("Frequency.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
		}
	}
}

func TestMass_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected Mass
	}{
		{"1ng", NanoGram},
		{"1ug", MicroGram},
		{"1µg", MicroGram},
		{"1mg", MilliGram},
		{"1g", Gram},
		{"1kg", KiloGram},
		{"1Mg", MegaGram},
		{"1Gg", GigaGram},
		{"1oz", OunceMass},
		{"1lb", PoundMass},
		// Maximum and minimum values that are allowed.
		{"9.223372036854775807Gg", 9223372036854775807},
		{"-9.223372036854775807Gg", -9223372036854775807},
		{"20334054lb", maxPoundMass * PoundMass},
		{"-20334054lb", minPoundMass * PoundMass},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10Gg",
			"exponent exceeds int64",
		},
		{
			"10Eg",
			"contains unknown unit prefix \"E\". valid prefixes for \"g\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need g",
		},
		{
			fmt.Sprintf("%dlb", maxPoundMass+1),
			fmt.Sprintf("maximum value is %dlb", maxPoundMass),
		},
		{
			fmt.Sprintf("%dlb", minPoundMass-1),
			fmt.Sprintf("minimum value is %dlb", minPoundMass),
		},
		{
			fmt.Sprintf("%doz", maxOunceMass+1),
			fmt.Sprintf("maximum value is %doz", maxOunceMass),
		},
		{
			fmt.Sprintf("%doz", minOunceMass-1),
			fmt.Sprintf("minimum value is %doz", minOunceMass),
		},
		{
			fmt.Sprintf("%dlb", maxPoundMass+1),
			fmt.Sprintf("maximum value is %dlb", maxPoundMass),
		},
		{
			fmt.Sprintf("%dlb", minPoundMass-1),
			fmt.Sprintf("minimum value is %dlb", minPoundMass),
		},
		{
			"9.224Gg",
			"maximum value is 9.223Gg",
		},
		{
			"-9.224Gg",
			"minimum value is -9.223Gg",
		},
		{
			"9223372036854775808ng",
			"maximum value is 9.223Gg",
		},
		{
			"-9223372036854775808ng",
			"minimum value is -9.223Gg",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.Mass",
		},
		{
			"g",
			"does not contain number",
		},
		{
			"oz",
			"does not contain number",
		},
		{
			"lb",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"g\"",
		},
		{
			"++1g",
			"multiple plus symbols ++1g",
		},
		{
			"--1g",
			"multiple minus symbols --1g",
		},
		{
			"+-1g",
			"can't contain both plus and minus symbols +-1g",
		},
		{
			"1.1.1.1g",
			"multiple decimal points 1.1.1.1g",
		},
		{
			string([]byte{0x33, 0x01}),
			"unexpected end of string",
		},
		{
			"10000000Tlb",
			errExponentOverflow.Error(),
		},
		{
			"10000000Toz",
			errExponentOverflow.Error(),
		},
	}

	for _, tt := range succeeds {
		var got Mass
		if err := got.Set(tt.in); err != nil {
			t.Errorf("Mass.Set(%s) unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("Mass.Set(%s) wanted: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got Mass
		if err := got.Set(tt.in); err.Error() != tt.err {
			t.Errorf("Mass.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
		}
	}
}

func TestPressure_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected Pressure
	}{
		{"1nPa", 1 * NanoPascal},
		{"10nPa", 10 * NanoPascal},
		{"100nPa", 100 * NanoPascal},
		{"1uPa", 1 * MicroPascal},
		{"10uPa", 10 * MicroPascal},
		{"100uPa", 100 * MicroPascal},
		{"1µPa", 1 * MicroPascal},
		{"10µPa", 10 * MicroPascal},
		{"100µPa", 100 * MicroPascal},
		{"1mPa", 1 * MilliPascal},
		{"10mPa", 10 * MilliPascal},
		{"100mPa", 100 * MilliPascal},
		{"1Pa", 1 * Pascal},
		{"10Pa", 10 * Pascal},
		{"100Pa", 100 * Pascal},
		{"1kPa", 1 * KiloPascal},
		{"10kPa", 10 * KiloPascal},
		{"100kPa", 100 * KiloPascal},
		{"1MPa", 1 * MegaPascal},
		{"10MPa", 10 * MegaPascal},
		{"100MPa", 100 * MegaPascal},
		{"1GPa", 1 * GigaPascal},
		{"12.345Pa", 12345 * MilliPascal},
		{"-12.345Pa", -12345 * MilliPascal},
		{"9.223372036854775807GPa", 9223372036854775807 * NanoPascal},
		{"-9.223372036854775807GPa", -9223372036854775807 * NanoPascal},
		{"1MPa", 1 * MegaPascal},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TPa",
			"exponent exceeds int64",
		},
		{
			"10EPa",
			"contains unknown unit prefix \"E\". valid prefixes for \"Pa\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaPa",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"Pa\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ePascalE",
			"contains unknown unit prefix \"e\". valid prefixes for \"Pa\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need Pa",
		},
		{
			"9223372036854775808",
			"maximum value is 9.223GPa",
		},
		{
			"-9223372036854775808",
			"minimum value is -9.223GPa",
		},
		{
			"9.223372036854775808GPa",
			"maximum value is 9.223GPa",
		},
		{
			"-9.223372036854775808GPa",
			"minimum value is -9.223GPa",
		},
		{
			"9.223372036854775808GPa",
			"maximum value is 9.223GPa",
		},
		{
			"-9.223372036854775808GPa",
			"minimum value is -9.223GPa",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.Pressure",
		},
		{
			"Pa",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"Pa\"",
		},
		{
			"++1Pa",
			"multiple plus symbols ++1Pa",
		},
		{
			"--1Pa",
			"multiple minus symbols --1Pa",
		},
		{
			"+-1Pa",
			"can't contain both plus and minus symbols +-1Pa",
		},
		{
			"1.1.1.1Pa",
			"multiple decimal points 1.1.1.1Pa",
		},
	}

	for _, tt := range succeeds {
		var got Pressure
		if err := got.Set(tt.in); err != nil {
			t.Errorf("Pressure.Set(%s) got unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("Pressure.Set(%s) expected: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got Pressure
		if err := got.Set(tt.in); err.Error() != tt.err {
			t.Errorf("Pressure.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
		}
	}
}

func TestRelativeHumidity_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected RelativeHumidity
	}{
		{"10u%rH", PercentRH / 100000},
		{"1m%rH", PercentRH / 1000},
		{"1%rH", PercentRH},
		{"10%rH", 10 * PercentRH},
		{"100%rH", 100 * PercentRH},
		{"10u%", PercentRH / 100000},
		{"1m%", PercentRH / 1000},
		{"1%", PercentRH},
		{"10%", 10 * PercentRH},
		{"100%", 100 * PercentRH},
		{fmt.Sprintf("%du%%rH", int64(maxRelativeHumidity)*10), maxRelativeHumidity},
		{fmt.Sprintf("%du%%rH", int64(minRelativeHumidity)*10), minRelativeHumidity},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"1000T%rH",
			"exponent exceeds int64",
		},
		{
			"10E%rH",
			"contains unknown unit prefix \"E\". valid prefixes for \"%rH\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need %rH",
		},
		{
			"21474836.48m%rH",
			"maximum value is 100%rH",
		},
		{
			"-21474836.48m%rH",
			"minimum value is 0%rH",
		},
		{
			"90224T%rH",
			"maximum value is 100%rH",
		},
		{
			"-90224T%rH",
			"minimum value is 0%rH",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.RelativeHumidity",
		},
		{
			"%rH",
			"does not contain number",
		},
		{
			"%",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"%rH\"",
		},
		{
			"++1%rH",
			"multiple plus symbols ++1%rH",
		},
		{
			"--1%rH",
			"multiple minus symbols --1%rH",
		},
		{
			"+-1%rH",
			"can't contain both plus and minus symbols +-1%rH",
		},
		{
			"1.1.1.1%rH",
			"multiple decimal points 1.1.1.1%rH",
		},
	}

	for _, tt := range succeeds {
		var got RelativeHumidity
		if err := got.Set(tt.in); err != nil {
			t.Errorf("RelativeHumidity.Set(%s) unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("RelativeHumidity.Set(%s) wanted: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got RelativeHumidity
		if err := got.Set(tt.in); err == nil {
			t.Errorf("RelativeHumidity.Set(%s) \nexpected: error %v but got none", tt.in, tt.err)
		} else if err.Error() != tt.err {
			t.Errorf("RelativeHumidity.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
		}
	}
}

func TestSpeed_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected Speed
	}{
		{"1nmps", NanoMetrePerSecond},
		{"1umps", MicroMetrePerSecond},
		{"1µmps", MicroMetrePerSecond},
		{"1mmps", MilliMetrePerSecond},
		{"1mps", MetrePerSecond},
		{"1kmps", KiloMetrePerSecond},
		{"1Mmps", MegaMetrePerSecond},
		{"1Gmps", GigaMetrePerSecond},
		{"1nm/s", NanoMetrePerSecond},
		{"1um/s", MicroMetrePerSecond},
		{"1µm/s", MicroMetrePerSecond},
		{"1mm/s", MilliMetrePerSecond},
		{"1m/s", MetrePerSecond},
		{"1km/s", KiloMetrePerSecond},
		{"1Mm/s", MegaMetrePerSecond},
		{"1Gm/s", GigaMetrePerSecond},
		{"1mph", MilePerHour},
		{"1fps", FootPerSecond},
		{"1kph", KilometrePerHour},
		// Maximum and minimum values that are allowed.
		{fmt.Sprintf("%dnmps", minSpeed), minSpeed},
		{fmt.Sprintf("%dnmps", maxSpeed), maxSpeed},
		{fmt.Sprintf("%dkph", minKilometrePerHour), minKilometrePerHour * KilometrePerHour},
		{fmt.Sprintf("%dkph", maxKilometrePerHour), maxKilometrePerHour * KilometrePerHour},
		{fmt.Sprintf("%dmph", minMilePerHour), minMilePerHour * MilePerHour},
		{fmt.Sprintf("%dmph", maxMilePerHour), maxMilePerHour * MilePerHour},
		{fmt.Sprintf("%dfps", minFootPerSecond), minFootPerSecond * FootPerSecond},
		{fmt.Sprintf("%dfps", maxFootPerSecond), maxFootPerSecond * FootPerSecond},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10Gm/s",
			"exponent exceeds int64",
		},
		{
			"10Em/s",
			"contains unknown unit prefix \"E\". valid prefixes for \"m/s\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need m/s",
		},
		{
			fmt.Sprintf("%dkph", maxKilometrePerHour+1),
			fmt.Sprintf("maximum value is %dkph", maxKilometrePerHour),
		},
		{
			fmt.Sprintf("%dkph", minKilometrePerHour-1),
			fmt.Sprintf("minimum value is %dkph", minKilometrePerHour),
		},
		{
			fmt.Sprintf("%dmph", maxMilePerHour+1),
			fmt.Sprintf("maximum value is %dmph", maxMilePerHour),
		},
		{
			fmt.Sprintf("%dmph", minMilePerHour-1),
			fmt.Sprintf("minimum value is %dmph", minMilePerHour),
		},
		{
			fmt.Sprintf("%dfps", maxFootPerSecond+1),
			fmt.Sprintf("maximum value is %dfps", maxFootPerSecond),
		},
		{
			fmt.Sprintf("%dfps", minFootPerSecond-1),
			fmt.Sprintf("minimum value is %dfps", minFootPerSecond),
		},
		{
			"9.224Gm/s",
			"maximum value is 9.223Gm/s",
		},
		{
			"-9.224Gm/s",
			"minimum value is -9.223Gm/s",
		},
		{
			"9223372036854775808nm/s",
			"maximum value is 9.223Gm/s",
		},
		{
			"-9223372036854775808nm/s",
			"minimum value is -9.223Gm/s",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.Speed",
		},
		{
			"m/s",
			"does not contain number",
		},
		{
			"fps",
			"does not contain number",
		},
		{
			"mph",
			"does not contain number",
		},
		{
			"kph",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"m/s\"",
		},
		{
			"++1m/s",
			"multiple plus symbols ++1m/s",
		},
		{
			"--1m/s",
			"multiple minus symbols --1m/s",
		},
		{
			"+-1m/s",
			"can't contain both plus and minus symbols +-1m/s",
		},
		{
			"1.1.1.1m/s",
			"multiple decimal points 1.1.1.1m/s",
		},
		{
			string([]byte{0x33, 0x01}),
			"unexpected end of string",
		},
		{
			"10000000Tmph",
			errExponentOverflow.Error(),
		},
		{
			"10000000Tfps",
			errExponentOverflow.Error(),
		},
		{
			"10000000Tkph",
			errExponentOverflow.Error(),
		},
	}

	for _, tt := range succeeds {
		var got Speed
		if err := got.Set(tt.in); err != nil {
			t.Errorf("Speed.Set(%s) unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("Speed.Set(%s) wanted: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got Speed
		if err := got.Set(tt.in); err.Error() != tt.err {
			t.Errorf("Speed.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
		}
	}
}

func TestPower_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected Power
	}{
		{"1nW", 1 * NanoWatt},
		{"10nW", 10 * NanoWatt},
		{"100nW", 100 * NanoWatt},
		{"1uW", 1 * MicroWatt},
		{"10uW", 10 * MicroWatt},
		{"100uW", 100 * MicroWatt},
		{"1µW", 1 * MicroWatt},
		{"10µW", 10 * MicroWatt},
		{"100µW", 100 * MicroWatt},
		{"1mW", 1 * MilliWatt},
		{"10mW", 10 * MilliWatt},
		{"100mW", 100 * MilliWatt},
		{"1W", 1 * Watt},
		{"10W", 10 * Watt},
		{"100W", 100 * Watt},
		{"1kW", 1 * KiloWatt},
		{"10kW", 10 * KiloWatt},
		{"100kW", 100 * KiloWatt},
		{"1MW", 1 * MegaWatt},
		{"10MW", 10 * MegaWatt},
		{"100MW", 100 * MegaWatt},
		{"1GW", 1 * GigaWatt},
		{"12.345W", 12345 * MilliWatt},
		{"-12.345W", -12345 * MilliWatt},
		{"9.223372036854775807GW", 9223372036854775807 * NanoWatt},
		{"-9.223372036854775807GW", -9223372036854775807 * NanoWatt},
		{"1MW", 1 * MegaWatt},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TW",
			"exponent exceeds int64",
		},
		{
			"10EW",
			"contains unknown unit prefix \"E\". valid prefixes for \"W\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaW",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"W\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eWattE",
			"contains unknown unit prefix \"e\". valid prefixes for \"W\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need W",
		},
		{
			"9223372036854775808",
			"maximum value is 9.223GW",
		},
		{
			"-9223372036854775808",
			"minimum value is -9.223GW",
		},
		{
			"9.223372036854775808GW",
			"maximum value is 9.223GW",
		},
		{
			"-9.223372036854775808GW",
			"minimum value is -9.223GW",
		},
		{
			"9.223372036854775808GW",
			"maximum value is 9.223GW",
		},
		{
			"-9.223372036854775808GW",
			"minimum value is -9.223GW",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.Power",
		},
		{
			"W",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"W\"",
		},
		{
			"++1W",
			"multiple plus symbols ++1W",
		},
		{
			"--1W",
			"multiple minus symbols --1W",
		},
		{
			"+-1W",
			"can't contain both plus and minus symbols +-1W",
		},
		{
			"1.1.1.1W",
			"multiple decimal points 1.1.1.1W",
		},
	}

	for _, tt := range succeeds {
		var got Power
		if err := got.Set(tt.in); err != nil {
			t.Errorf("Power.Set(%s) got unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("Power.Set(%s) expected: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got Power
		if err := got.Set(tt.in); err.Error() != tt.err {
			t.Errorf("Power.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
		}
	}
}

func TestEnergy_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected Energy
	}{
		{"1nJ", 1 * NanoJoule},
		{"10nJ", 10 * NanoJoule},
		{"100nJ", 100 * NanoJoule},
		{"1uJ", 1 * MicroJoule},
		{"10uJ", 10 * MicroJoule},
		{"100uJ", 100 * MicroJoule},
		{"1µJ", 1 * MicroJoule},
		{"10µJ", 10 * MicroJoule},
		{"100µJ", 100 * MicroJoule},
		{"1mJ", 1 * MilliJoule},
		{"10mJ", 10 * MilliJoule},
		{"100mJ", 100 * MilliJoule},
		{"1J", 1 * Joule},
		{"10J", 10 * Joule},
		{"100J", 100 * Joule},
		{"1kJ", 1 * KiloJoule},
		{"10kJ", 10 * KiloJoule},
		{"100kJ", 100 * KiloJoule},
		{"1MJ", 1 * MegaJoule},
		{"10MJ", 10 * MegaJoule},
		{"100MJ", 100 * MegaJoule},
		{"1GJ", 1 * GigaJoule},
		{"12.345J", 12345 * MilliJoule},
		{"-12.345J", -12345 * MilliJoule},
		{"9.223372036854775807GJ", 9223372036854775807 * NanoJoule},
		{"-9.223372036854775807GJ", -9223372036854775807 * NanoJoule},
		{"1MJ", 1 * MegaJoule},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TJ",
			"exponent exceeds int64",
		},
		{
			"10EJ",
			"contains unknown unit prefix \"E\". valid prefixes for \"J\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaJ",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"J\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eJouleE",
			"contains unknown unit prefix \"e\". valid prefixes for \"J\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need J",
		},
		{
			"9223372036854775808",
			"maximum value is 9.223GJ",
		},
		{
			"-9223372036854775808",
			"minimum value is -9.223GJ",
		},
		{
			"9.223372036854775808GJ",
			"maximum value is 9.223GJ",
		},
		{
			"-9.223372036854775808GJ",
			"minimum value is -9.223GJ",
		},
		{
			"9.223372036854775808GJ",
			"maximum value is 9.223GJ",
		},
		{
			"-9.223372036854775808GJ",
			"minimum value is -9.223GJ",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.Energy",
		},
		{
			"J",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"J\"",
		},
		{
			"++1J",
			"multiple plus symbols ++1J",
		},
		{
			"--1J",
			"multiple minus symbols --1J",
		},
		{
			"+-1J",
			"can't contain both plus and minus symbols +-1J",
		},
		{
			"1.1.1.1J",
			"multiple decimal points 1.1.1.1J",
		},
	}

	for _, tt := range succeeds {
		var got Energy
		if err := got.Set(tt.in); err != nil {
			t.Errorf("Energy.Set(%s) got unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("Energy.Set(%s) expected: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got Energy
		if err := got.Set(tt.in); err.Error() != tt.err {
			t.Errorf("Energy.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
		}
	}
}

func TestElectricalCapacitance_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected ElectricalCapacitance
	}{
		{"1pF", 1 * PicoFarad},
		{"10pF", 10 * PicoFarad},
		{"100pF", 100 * PicoFarad},
		{"1nF", 1 * NanoFarad},
		{"10nF", 10 * NanoFarad},
		{"100nF", 100 * NanoFarad},
		{"1uF", 1 * MicroFarad},
		{"10uF", 10 * MicroFarad},
		{"100uF", 100 * MicroFarad},
		{"1µF", 1 * MicroFarad},
		{"10µF", 10 * MicroFarad},
		{"100µF", 100 * MicroFarad},
		{"1mF", 1 * MilliFarad},
		{"10mF", 10 * MilliFarad},
		{"100mF", 100 * MilliFarad},
		{"1F", 1 * Farad},
		{"10F", 10 * Farad},
		{"100F", 100 * Farad},
		{"1kF", 1 * KiloFarad},
		{"10kF", 10 * KiloFarad},
		{"100kF", 100 * KiloFarad},
		{"1MF", 1 * MegaFarad},
		{"12.345F", 12345 * MilliFarad},
		{"-12.345F", -12345 * MilliFarad},
		{"9.223372036854775807MF", 9223372036854775807 * PicoFarad},
		{"-9.223372036854775807MF", -9223372036854775807 * PicoFarad},
		{"1MF", 1 * MegaFarad},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TF",
			"exponent exceeds int64",
		},
		{
			"10EF",
			"contains unknown unit prefix \"E\". valid prefixes for \"F\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaF",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"F\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eFaradE",
			"contains unknown unit prefix \"e\". valid prefixes for \"F\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need F",
		},
		{
			"9223372036854775808",
			"maximum value is 9.223MF",
		},
		{
			"-9223372036854775808",
			"minimum value is -9.223MF",
		},
		{
			"9.223372036854775808MF",
			"maximum value is 9.223MF",
		},
		{
			"-9.223372036854775808MF",
			"minimum value is -9.223MF",
		},
		{
			"9.223372036854775808MF",
			"maximum value is 9.223MF",
		},
		{
			"-9.223372036854775808MF",
			"minimum value is -9.223MF",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.ElectricalCapacitance",
		},
		{
			"F",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"F\"",
		},
		{
			"++1F",
			"multiple plus symbols ++1F",
		},
		{
			"--1F",
			"multiple minus symbols --1F",
		},
		{
			"+-1F",
			"can't contain both plus and minus symbols +-1F",
		},
		{
			"1.1.1.1F",
			"multiple decimal points 1.1.1.1F",
		},
	}

	for _, tt := range succeeds {
		var got ElectricalCapacitance
		if err := got.Set(tt.in); err != nil {
			t.Errorf("ElectricalCapacitance.Set(%s) got unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("ElectricalCapacitance.Set(%s) expected: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got ElectricalCapacitance
		if err := got.Set(tt.in); err.Error() != tt.err {
			t.Errorf("ElectricalCapacitance.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
		}
	}
}

func TestLuminousIntensity_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected LuminousIntensity
	}{
		{"1ncd", 1 * NanoCandela},
		{"10ncd", 10 * NanoCandela},
		{"100ncd", 100 * NanoCandela},
		{"1ucd", 1 * MicroCandela},
		{"10ucd", 10 * MicroCandela},
		{"100ucd", 100 * MicroCandela},
		{"1µcd", 1 * MicroCandela},
		{"10µcd", 10 * MicroCandela},
		{"100µcd", 100 * MicroCandela},
		{"1mcd", 1 * MilliCandela},
		{"10mcd", 10 * MilliCandela},
		{"100mcd", 100 * MilliCandela},
		{"1cd", 1 * Candela},
		{"10cd", 10 * Candela},
		{"100cd", 100 * Candela},
		{"1kcd", 1 * KiloCandela},
		{"10kcd", 10 * KiloCandela},
		{"100kcd", 100 * KiloCandela},
		{"1Mcd", 1 * MegaCandela},
		{"10Mcd", 10 * MegaCandela},
		{"100Mcd", 100 * MegaCandela},
		{"1Gcd", 1 * GigaCandela},
		{"12.345cd", 12345 * MilliCandela},
		{"-12.345cd", -12345 * MilliCandela},
		{"9.223372036854775807Gcd", 9223372036854775807 * NanoCandela},
		{"-9.223372036854775807Gcd", -9223372036854775807 * NanoCandela},
		{"1Mcd", 1 * MegaCandela},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10Tcd",
			"exponent exceeds int64",
		},
		{
			"10Ecd",
			"contains unknown unit prefix \"E\". valid prefixes for \"cd\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10Exacd",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"cd\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ecdE",
			"contains unknown unit prefix \"e\". valid prefixes for \"cd\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need cd",
		},
		{
			"9223372036854775808",
			"maximum value is 9.223Gcd",
		},
		{
			"-9223372036854775808",
			"minimum value is -9.223Gcd",
		},
		{
			"9.223372036854775808Gcd",
			"maximum value is 9.223Gcd",
		},
		{
			"-9.223372036854775808Gcd",
			"minimum value is -9.223Gcd",
		},
		{
			"9.223372036854775808Gcd",
			"maximum value is 9.223Gcd",
		},
		{
			"-9.223372036854775808Gcd",
			"minimum value is -9.223Gcd",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.LuminousIntensity",
		},
		{
			"cd",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"cd\"",
		},
		{
			"++1cd",
			"multiple plus symbols ++1cd",
		},
		{
			"--1cd",
			"multiple minus symbols --1cd",
		},
		{
			"+-1cd",
			"can't contain both plus and minus symbols +-1cd",
		},
		{
			"1.1.1.1cd",
			"multiple decimal points 1.1.1.1cd",
		},
	}

	for _, tt := range succeeds {
		var got LuminousIntensity
		if err := got.Set(tt.in); err != nil {
			t.Errorf("LuminousIntensity.Set(%s) got unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("LuminousIntensity.Set(%s) expected: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got LuminousIntensity
		if err := got.Set(tt.in); err.Error() != tt.err {
			t.Errorf("LuminousIntensity.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
		}
	}
}

func TestLuminousFlux_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected LuminousFlux
	}{
		{"1nlm", 1 * NanoLumen},
		{"10nlm", 10 * NanoLumen},
		{"100nlm", 100 * NanoLumen},
		{"1ulm", 1 * MicroLumen},
		{"10ulm", 10 * MicroLumen},
		{"100ulm", 100 * MicroLumen},
		{"1µlm", 1 * MicroLumen},
		{"10µlm", 10 * MicroLumen},
		{"100µlm", 100 * MicroLumen},
		{"1mlm", 1 * MilliLumen},
		{"10mlm", 10 * MilliLumen},
		{"100mlm", 100 * MilliLumen},
		{"1lm", 1 * Lumen},
		{"10lm", 10 * Lumen},
		{"100lm", 100 * Lumen},
		{"1klm", 1 * KiloLumen},
		{"10klm", 10 * KiloLumen},
		{"100klm", 100 * KiloLumen},
		{"1Mlm", 1 * MegaLumen},
		{"10Mlm", 10 * MegaLumen},
		{"100Mlm", 100 * MegaLumen},
		{"1Glm", 1 * GigaLumen},
		{"12.345lm", 12345 * MilliLumen},
		{"-12.345lm", -12345 * MilliLumen},
		{"9.223372036854775807Glm", 9223372036854775807 * NanoLumen},
		{"-9.223372036854775807Glm", -9223372036854775807 * NanoLumen},
		{"1Mlm", 1 * MegaLumen},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10Tlm",
			"exponent exceeds int64",
		},
		{
			"10Elm",
			"contains unknown unit prefix \"E\". valid prefixes for \"lm\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10Exalm",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"lm\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10elmE",
			"contains unknown unit prefix \"e\". valid prefixes for \"lm\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need lm",
		},
		{
			"9223372036854775808",
			"maximum value is 9.223Glm",
		},
		{
			"-9223372036854775808",
			"minimum value is -9.223Glm",
		},
		{
			"9.223372036854775808Glm",
			"maximum value is 9.223Glm",
		},
		{
			"-9.223372036854775808Glm",
			"minimum value is -9.223Glm",
		},
		{
			"9.223372036854775808Glm",
			"maximum value is 9.223Glm",
		},
		{
			"-9.223372036854775808Glm",
			"minimum value is -9.223Glm",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.LuminousFlux",
		},
		{
			"lm",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"lm\"",
		},
		{
			"++1lm",
			"multiple plus symbols ++1lm",
		},
		{
			"--1lm",
			"multiple minus symbols --1lm",
		},
		{
			"+-1lm",
			"can't contain both plus and minus symbols +-1lm",
		},
		{
			"1.1.1.1lm",
			"multiple decimal points 1.1.1.1lm",
		},
	}

	for _, tt := range succeeds {
		var got LuminousFlux
		if err := got.Set(tt.in); err != nil {
			t.Errorf("LuminousFlux.Set(%s) got unexpected error: %v", tt.in, err)
		}
		if got != tt.expected {
			t.Errorf("LuminousFlux.Set(%s) expected: %v(%d) but got: %v(%d)", tt.in, tt.expected, tt.expected, got, got)
		}
	}

	for _, tt := range fails {
		var got LuminousFlux
		if err := got.Set(tt.in); err.Error() != tt.err {
			t.Errorf("LuminousFlux.Set(%s) \nexpected: %s\ngot: %s", tt.in, tt.err, err)
		}
	}
}

func BenchmarkDecimal(b *testing.B) {
	var d decimal
	var n int
	var err error
	for i := 0; i < b.N; i++ {
		if d, n, err = atod("337.2m"); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%v %d", d, n)
}

func BenchmarkDecimal2Int(b *testing.B) {
	d := decimal{1234, 5, false}
	var err error
	var v int64
	for i := 0; i < b.N; i++ {
		if v, err = dtoi(d, 0); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%d", v)
}

func BenchmarkString2Decimal2Int(b *testing.B) {
	var d decimal
	var n int
	var err error
	var v int64
	for i := 0; i < b.N; i++ {
		if d, n, err = atod("337.2m"); err != nil {
			b.Fatal(err)
		}
		if v, err = dtoi(d, 0); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%d %d", v, n)
}

func BenchmarkDecimalNeg(b *testing.B) {
	var d decimal
	var n int
	var err error
	for i := 0; i < b.N; i++ {
		if d, n, err = atod("-337.2m"); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%v %d", d, n)
}

func BenchmarkString2Decimal2IntNeg(b *testing.B) {
	var d decimal
	var n int
	var err error
	var v int64
	for i := 0; i < b.N; i++ {
		if d, n, err = atod("-337.2m"); err != nil {
			b.Fatal(err)
		}
		if v, err = dtoi(d, 0); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%d %d", v, n)
}

func BenchmarkDistanceSet(b *testing.B) {
	var err error
	var d Distance
	for i := 0; i < b.N; i++ {
		if err = d.Set("1ft"); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%d", d)
}

func BenchmarkElectricCurrentSet(b *testing.B) {
	var err error
	var e ElectricCurrent
	for i := 0; i < b.N; i++ {
		if err = e.Set("1A"); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%d", e)
}

func BenchmarkForceSetMetric(b *testing.B) {
	var err error
	var f Force
	for i := 0; i < b.N; i++ {
		if err = f.Set("123N"); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%d", f)
}

func BenchmarkForceSetImperial(b *testing.B) {
	var err error
	var f Force
	for i := 0; i < b.N; i++ {
		if err = f.Set("1.23Mlbf"); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%d", f)
}

func BenchmarkForceSetImperialWorstCase(b *testing.B) {
	var err error
	var f Force
	for i := 0; i < b.N; i++ {
		if err = f.Set("1.0000000000101lbf"); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%d", f)
}

func BenchmarkAngleSetRadian(b *testing.B) {
	var err error
	var a Angle
	for i := 0; i < b.N; i++ {
		if err = a.Set("1rad"); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%d", a)
}

func BenchmarkAngleSet1Degree(b *testing.B) {
	var err error
	var a Angle
	for i := 0; i < b.N; i++ {
		if err = a.Set("1deg"); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%d", a)
}

func BenchmarkAngleSet2Degree(b *testing.B) {
	var err error
	var a Angle
	for i := 0; i < b.N; i++ {
		if err = a.Set("2deg"); err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%d", a)
}
