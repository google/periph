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
		{(9223372036854775807 - 17453293) * NanoRadian, "528460276054°"},
		{(-9223372036854775807 + 17453293) * NanoRadian, "-528460276054°"},
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
		{"123456789", decimal{"123456789", 0, positive}, 9},
		{"1nM", decimal{"1", 0, positive}, 1},
		{"2.2nM", decimal{"22", -1, positive}, 3},
		{"12.5mA", decimal{"125", -1, positive}, 4},
		{"-12.5mA", decimal{"125", -1, negative}, 5},
		{"1ma1", decimal{"1", 0, positive}, 1},
		{"+1ma1", decimal{"1", 0, positive}, 2},
		{"-1ma1", decimal{"1", 0, negative}, 2},
		{"-0.00001%rH", decimal{"1", -5, negative}, 8},
		{"0.00001%rH", decimal{"1", -5, positive}, 7},
		{"1.0", decimal{"1", 0, positive}, 3},
		{"0.10001", decimal{"10001", -5, positive}, 7},
		{"+0.10001", decimal{"10001", -5, positive}, 8},
		{"-0.10001", decimal{"10001", -5, negative}, 8},
		{"1n", decimal{"1", 0, positive}, 1},
		{"1.n", decimal{"1", 0, positive}, 2},
		{"-1.n", decimal{"1", 0, negative}, 3},
		{"200n", decimal{"2", 2, positive}, 3},
		{".01", decimal{"1", -2, positive}, 3},
		{"+.01", decimal{"1", -2, positive}, 4},
		{"-.01", decimal{"1", -2, negative}, 4},
		{"1-2", decimal{"1", 0, positive}, 1},
		{"1+2", decimal{"1", 0, positive}, 1},
		{"-1-2", decimal{"1", 0, negative}, 2},
		{"-1+2", decimal{"1", 0, negative}, 2},
		{"+1-2", decimal{"1", 0, positive}, 2},
		{"+1+2", decimal{"1", 0, positive}, 2},
		{"010", decimal{"1", 1, positive}, 3},
		{"001", decimal{"1", 0, positive}, 3},
	}

	fails := []struct {
		in       string
		expected decimal
		n        int
	}{
		{"1.1.1", decimal{}, 0},
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
		{"123", decimal{"123", 0, positive}, 123},
		{"-123", decimal{"123", 0, negative}, -123},
		{"1230", decimal{"123", 1, positive}, 1230},
		{"-1230", decimal{"123", 1, negative}, -1230},
		{"12.3", decimal{"123", -1, positive}, 12},
		{"-12.3", decimal{"123", -1, negative}, -12},
		{"123n", decimal{"123", 0, positive}, 123},
		{"max", decimal{"9223372036854775807", 0, positive}, 9223372036854775807},
		{"rounding(5.6)", decimal{"56", -1, positive}, 6},
		{"rounding(5.5)", decimal{"55", -1, positive}, 6},
		{"rounding(5.4)", decimal{"54", -1, positive}, 5},
		{"rounding(-5.6)", decimal{"56", -1, negative}, -6},
		{"rounding(-5.5)", decimal{"55", -1, negative}, -6},
		{"rounding(-5.4)", decimal{"54", -1, negative}, -5},
		{"rounding(0.6)", decimal{"6", -1, positive}, 1},
		{"rounding(0.5)", decimal{"5", -1, positive}, 1},
		{"rounding(0.4)", decimal{"4", -1, positive}, 0},
		{"rounding(-0.6)", decimal{"6", -1, negative}, -1},
		{"rounding(-0.5)", decimal{"5", -1, negative}, -1},
		{"rounding(-0.4)", decimal{"4", -1, negative}, -0},
	}

	fails := []struct {
		name     string
		in       decimal
		expected int64
	}{
		{"max+1", decimal{"9223372036854775808", 0, positive}, 9223372036854775807},
		{"-max-1", decimal{"9223372036854775808", 0, negative}, -9223372036854775807},
		{"non digit in decimal.digit)", decimal{"1a", 0, positive}, 0},
		{"non digit in decimal.digit)", decimal{"2.7b", 0, negative}, 0},
		{"exponet too large for int64", decimal{"123", 20, positive}, 0},
		{"exponet too small for int64", decimal{"123", -20, positive}, 0},
		{"max*10^1", decimal{"9223372036854775807", 1, positive}, 9223372036854775807},
		{"-max*10^1", decimal{"9223372036854775807", 1, negative}, -9223372036854775807},
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
	if strconv.FormatUint(maxInt64, 10) != maxUint64Str {
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

func TestElectricPotential_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected ElectricPotential
	}{
		{"1nVolt", 1 * NanoVolt},
		{"10nVolts", 10 * NanoVolt},
		{"100nVolts", 100 * NanoVolt},
		{"1uVolt", 1 * MicroVolt},
		{"10uVolts", 10 * MicroVolt},
		{"100uVolts", 100 * MicroVolt},
		{"1µVolt", 1 * MicroVolt},
		{"10µVolts", 10 * MicroVolt},
		{"100µVolts", 100 * MicroVolt},
		{"1mVolt", 1 * MilliVolt},
		{"10mVolts", 10 * MilliVolt},
		{"100mVolts", 100 * MilliVolt},
		{"1Volt", 1 * Volt},
		{"10Volts", 10 * Volt},
		{"100Volts", 100 * Volt},
		{"1kVolt", 1 * KiloVolt},
		{"10kVolts", 10 * KiloVolt},
		{"100kVolts", 100 * KiloVolt},
		{"1MVolt", 1 * MegaVolt},
		{"10MVolts", 10 * MegaVolt},
		{"100MVolts", 100 * MegaVolt},
		{"1GVolt", 1 * GigaVolt},
		{"12.345Volts", 12345 * MilliVolt},
		{"-12.345Volts", -12345 * MilliVolt},
		{"9.223372036854775807GVolts", 9223372036854775807 * NanoVolt},
		{"-9.223372036854775807GVolts", -9223372036854775807 * NanoVolt},
		{"1MV", 1 * MegaVolt},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TVolt",
			"exponent exceeds int64",
		},
		{
			"10EVolt",
			"contains unknown unit prefix \"E\". valid prefixes for \"Volt\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaVolt",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"Volt\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eVoltE",
			"contains unknown unit prefix \"e\". valid prefixes for \"Volt\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need Volt",
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
			"Volt",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"Volt\"",
		},
		{
			"++1Volt",
			"multiple plus symbols ++1Volt",
		},
		{
			"--1Volt",
			"multiple minus symbols --1Volt",
		},
		{
			"+-1Volt",
			"can't contain both plus and minus symbols +-1Volt",
		},
		{
			"1.1.1.1Volt",
			"multiple decimal points 1.1.1.1Volt",
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
			"9.223372036854775808TOhm",
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
