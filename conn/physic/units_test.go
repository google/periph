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
		{"none", 0, none, 0},
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
		{"empty", &parseError{s: "", err: nil}, "parse error"},
		{"empty", &parseError{s: "", err: errors.New("test")}, "parse error: test: \"" + "\""},
		{"noUnits", noUnits("someunit"), "parse error: no units provided, need: \"someunit\""},
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

func TestAngle_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Angle
	}{
		{"1Degrees", "1Degrees", Degree},
		{"-1Degrees", "-1Degrees", -1 * Degree},
		{"180.00Degrees", "180.00Degrees", 180 * Degree},
		{"0.5Degrees", "0.5Degrees", 8726646 * NanoRadian},
		{"0.5°", "0.5°", 8726646 * NanoRadian},
		{"1nRadians", "1nRadians", NanoRadian},
		{"1uRadians", "1uRadians", MicroRadian},
		{"1mRadians", "1mRadians", MilliRadian},
		{"0.5uRadians", "0.5uRadians", 500 * NanoRadian},
		{"0.5mRadians", "0.5mRadians", 500 * MicroRadian},
		{"200uRadians", "200uRadians", 200 * MicroRadian},
		{"1Radians", "1Radians", Radian},
		{"1Pi", "1Pi", Pi},
		{"1π", "1π", Pi},
		{"2Pi", "2Pi", 2 * Pi},
		{"2Pi", "-2Pi", -2 * Pi},
		{"0.5Pi", "0.5Pi", 1570796326 * NanoRadian},
		{"200mRadians", "200mRadians", 200 * MilliRadian},
		{"20.0mRadians", "20.0mRadians", 20 * MilliRadian},
		{"1nRadian", "1nRadian", 1 * NanoRadian},
		{"1nradians", "1nradians", 1 * NanoRadian},
		{"1uradians", "1uradians", 1 * MicroRadian},
		{"1mradians", "1mradians", 1 * MilliRadian},
	}

	for _, tt := range tests {
		var got Angle
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "angle", "value of angle")
		fs.Parse([]string{"-angle", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v(%d) but got: %v(%d)", tt.name, tt.want, tt.want, got, got)
		}
	}
}

func TestFrequency_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Frequency
		err  bool
	}{
		{"1uHz", "1uHz", 1 * MicroHertz, false},
		{"10uHz", "10uHz", 10 * MicroHertz, false},
		{"100uHz", "100uHz", 100 * MicroHertz, false},
		{"1µHz", "1µHz", 1 * MicroHertz, false},
		{"10µHz", "10µHz", 10 * MicroHertz, false},
		{"100µHz", "100µHz", 100 * MicroHertz, false},
		{"1mHz", "1mHz", 1 * MilliHertz, false},
		{"10mHz", "10mHz", 10 * MilliHertz, false},
		{"100mHz", "100mHz", 100 * MilliHertz, false},
		{"1Hz", "1Hz", 1 * Hertz, false},
		{"10Hz", "10Hz", 10 * Hertz, false},
		{"100Hz", "100Hz", 100 * Hertz, false},
		{"1kHz", "1kHz", 1 * KiloHertz, false},
		{"10kHz", "10kHz", 10 * KiloHertz, false},
		{"100kHz", "100kHz", 100 * KiloHertz, false},
		{"1MHz", "1MHz", 1 * MegaHertz, false},
		{"10MHz", "10MHz", 10 * MegaHertz, false},
		{"100MHz", "100MHz", 100 * MegaHertz, false},
		{"1GHz", "1GHz", 1 * GigaHertz, false},
		{"10GHz", "10GHz", 10 * GigaHertz, false},
		{"100GHz", "100GHz", 100 * GigaHertz, false},
		{"1THz", "1THz", 1 * TeraHertz, false},
		{"12.345Hz", "12.345Hz", 12345 * MilliHertz, false},
		{"-12.345Hz", "-12.345Hz", -12345 * MilliHertz, false},
	}

	for _, tt := range tests {
		var got Frequency
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "f", "value of angle")
		fs.Parse([]string{"-f", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v but got: %v(%d)", tt.name, tt.want, got, got)
		}
	}
}

func TestDistance_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Distance
	}{
		{"1um", "1um", 1 * MicroMetre},
		{"10um", "10um", 10 * MicroMetre},
		{"100um", "100um", 100 * MicroMetre},
		{"1µm", "1µm", 1 * MicroMetre},
		{"10µm", "10µm", 10 * MicroMetre},
		{"100µm", "100µm", 100 * MicroMetre},
		{"1mm", "1mm", 1 * MilliMetre},
		{"10mm", "10mm", 10 * MilliMetre},
		{"100mm", "100mm", 100 * MilliMetre},
		{"1m", "1m", 1 * Metre},
		{"10m", "10m", 10 * Metre},
		{"100m", "100m", 100 * Metre},
		{"1km", "1km", 1 * KiloMetre},
		{"10km", "10km", 10 * KiloMetre},
		{"100km", "100km", 100 * KiloMetre},
		{"1Mm", "1Mm", 1 * MegaMetre},
		{"10Mm", "10Mm", 10 * MegaMetre},
		{"100Mm", "100Mm", 100 * MegaMetre},
		{"1Gm", "1Gm", 1 * GigaMetre},
		{"1metre", "1metre", 1 * Metre},
		{"1Metre", "1Metre", 1 * Metre},
		{"10metres", "10metres", 10 * Metre},
		{"10Metres", "10Metres", 10 * Metre},
		{"1in", "1in", 1 * Inch},
		{"1In", "1In", 1 * Inch},
		{"1inch", "1inch", 1 * Inch},
		{"1Inch", "1Inch", 1 * Inch},
		{"1inches", "1inches", 1 * Inch},
		{"1Inches", "1Inches", 1 * Inch},
		{"1foot", "1foot", 1 * Foot},
		{"1Foot", "1Foot", 1 * Foot},
		{"1ft", "1ft", 1 * Foot},
		{"1Ft", "1Ft", 1 * Foot},
		{"10Feet", "10Feet", 10 * Foot},
		{"10feet", "10feet", 10 * Foot},
		{"1Yard", "1Yard", 1 * Yard},
		{"1yard", "1yard", 1 * Yard},
		{"1Mile", "1Mile", 1 * Mile},
		{"1mile", "1mile", 1 * Mile},
		{"1Miles", "1Miles", 1 * Mile},
		{"1miles", "1miles", 1 * Mile},
	}

	for _, tt := range tests {
		var got Distance
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "d", "value of angle")
		fs.Parse([]string{"-d", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v(%d) but got: %v(%d)", tt.name, tt.want, tt.want, got, got)
		}
	}
}

func TestParseFrequency(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    Frequency
		wantErr bool
	}{
		{"100µHz", "100µHz", 100 * MicroHertz, false},
	}
	for _, tt := range tests {
		got, err := ParseFrequency(tt.s)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseFrequency() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if got != tt.want {
			t.Errorf("ParseFrequency() = %v, want %v", got, tt.want)
		}
	}
}

func TestElectricalCapacitance_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want ElectricalCapacitance
	}{

		{"1pF", "1pF", 1 * PicoFarad},
		{"10pF", "10pF", 10 * PicoFarad},
		{"100pF", "100pF", 100 * PicoFarad},
		{"1nF", "1nF", 1 * NanoFarad},
		{"10nF", "10nF", 10 * NanoFarad},
		{"100nF", "100nF", 100 * NanoFarad},
		{"1uF", "1uF", 1 * MicroFarad},
		{"10uF", "10uF", 10 * MicroFarad},
		{"100uF", "100uF", 100 * MicroFarad},
		{"1µF", "1µF", 1 * MicroFarad},
		{"10µF", "10µF", 10 * MicroFarad},
		{"100µF", "100µF", 100 * MicroFarad},
		{"1mF", "1mF", 1 * MilliFarad},
		{"10mF", "10mF", 10 * MilliFarad},
		{"100mF", "100mF", 100 * MilliFarad},
		{"1F", "1F", 1 * Farad},
		{"10F", "10F", 10 * Farad},
		{"100F", "100F", 100 * Farad},
		{"1kF", "1kF", 1 * KiloFarad},
		{"10kF", "10kF", 10 * KiloFarad},
		{"100kF", "100kF", 100 * KiloFarad},
		{"1f", "1f", 1 * Farad},
		{"1farad", "1farad", 1 * Farad},
		{"1Farad", "1Farad", 1 * Farad},
		{"10farads", "10farads", 10 * Farad},
		{"10Farads", "10Farads", 10 * Farad},
	}

	for _, tt := range tests {
		var got ElectricalCapacitance
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "farad", "value of capacitance")
		fs.Parse([]string{"-farad", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v but got: %v(%d)", tt.name, tt.want, got, got)
		}
	}
}

func TestElectricCurrent_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want ElectricCurrent
	}{

		{"1nA", "1nA", 1 * NanoAmpere},
		{"10nA", "10nA", 10 * NanoAmpere},
		{"100nA", "100nA", 100 * NanoAmpere},
		{"1uA", "1uA", 1 * MicroAmpere},
		{"10uA", "10uA", 10 * MicroAmpere},
		{"100uA", "100uA", 100 * MicroAmpere},
		{"1µA", "1µA", 1 * MicroAmpere},
		{"10µA", "10µA", 10 * MicroAmpere},
		{"100µA", "100µA", 100 * MicroAmpere},
		{"1mA", "1mA", 1 * MilliAmpere},
		{"10mA", "10mA", 10 * MilliAmpere},
		{"100mA", "100mA", 100 * MilliAmpere},
		{"1A", "1A", 1 * Ampere},
		{"10A", "10A", 10 * Ampere},
		{"100A", "100A", 100 * Ampere},
		{"1kA", "1kA", 1 * KiloAmpere},
		{"10kA", "10kA", 10 * KiloAmpere},
		{"100kA", "100kA", 100 * KiloAmpere},
		{"1MA", "1MA", 1 * MegaAmpere},
		{"10MA", "10MA", 10 * MegaAmpere},
		{"100MA", "100MA", 100 * MegaAmpere},
		{"1GA", "1GA", 1 * GigaAmpere},
		{"1a", "1a", 1 * Ampere},
		{"1Amp", "1Amp", 1 * Ampere},
		{"1amp", "1amp", 1 * Ampere},
		{"1amps", "1amps", 1 * Ampere},
		{"1Amps", "1Amps", 1 * Ampere},
	}

	for _, tt := range tests {
		var got ElectricCurrent
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "amps", "value of current")
		fs.Parse([]string{"-amps", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v but got: %v(%d)", tt.name, tt.want, got, got)
		}
	}
}

func TestTemperature_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Temperature
	}{

		{"1nK", "1nK", 1 * NanoKelvin},
		{"10nK", "10nK", 10 * NanoKelvin},
		{"100nK", "100nK", 100 * NanoKelvin},
		{"1uK", "1uK", 1 * MicroKelvin},
		{"10uK", "10uK", 10 * MicroKelvin},
		{"100uK", "100uK", 100 * MicroKelvin},
		{"1µK", "1µK", 1 * MicroKelvin},
		{"10µK", "10µK", 10 * MicroKelvin},
		{"100µK", "100µK", 100 * MicroKelvin},
		{"1mK", "1mK", 1 * MilliKelvin},
		{"10mK", "10mK", 10 * MilliKelvin},
		{"100mK", "100mK", 100 * MilliKelvin},
		{"1K", "1K", 1 * Kelvin},
		{"10K", "10K", 10 * Kelvin},
		{"100K", "100K", 100 * Kelvin},
		{"1kK", "1kK", 1 * KiloKelvin},
		{"10kK", "10kK", 10 * KiloKelvin},
		{"100kK", "100kK", 100 * KiloKelvin},
		{"1MK", "1MK", 1 * MegaKelvin},
		{"10MK", "10MK", 10 * MegaKelvin},
		{"100MK", "100MK", 100 * MegaKelvin},
		{"1GK", "1GK", 1 * GigaKelvin},
		{"0C", "0C", ZeroCelsius},
		{"0°C", "0°C", ZeroCelsius},
		{"20C", "20C", ZeroCelsius + 20*Kelvin},
		{"-20C", "-20C", ZeroCelsius - 20*Kelvin},
	}

	for _, tt := range tests {
		var got Temperature
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "t", "value of temperature")
		fs.Parse([]string{"-t", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v but got: %v(%d)", tt.name, tt.want, got, got)
		}
	}
}

func TestElectricPotential_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want ElectricPotential
	}{

		{"1nV", "1nV", 1 * NanoVolt},
		{"10nV", "10nV", 10 * NanoVolt},
		{"100nV", "100nV", 100 * NanoVolt},
		{"1uV", "1uV", 1 * MicroVolt},
		{"10uV", "10uV", 10 * MicroVolt},
		{"100uV", "100uV", 100 * MicroVolt},
		{"1µV", "1µV", 1 * MicroVolt},
		{"10µV", "10µV", 10 * MicroVolt},
		{"100µV", "100µV", 100 * MicroVolt},
		{"1mV", "1mV", 1 * MilliVolt},
		{"10mV", "10mV", 10 * MilliVolt},
		{"100mV", "100mV", 100 * MilliVolt},
		{"1V", "1V", 1 * Volt},
		{"10V", "10V", 10 * Volt},
		{"100V", "100V", 100 * Volt},
		{"1kV", "1kV", 1 * KiloVolt},
		{"10kV", "10kV", 10 * KiloVolt},
		{"100kV", "100kV", 100 * KiloVolt},
		{"1MV", "1MV", 1 * MegaVolt},
		{"10MV", "10MV", 10 * MegaVolt},
		{"100MV", "100MV", 100 * MegaVolt},
		{"1GV", "1GV", 1 * GigaVolt},
		{"10volt", "10volt", 10 * Volt},
		{"10volts", "10volts", 10 * Volt},
		{"10Volt", "10Volt", 10 * Volt},
		{"10Volts", "10Volts", 10 * Volt},
	}

	for _, tt := range tests {
		var got ElectricPotential
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "v", "value of voltage")
		fs.Parse([]string{"-v", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v but got: %v(%d)", tt.name, tt.want, got, got)
		}
	}
}

func TestElectricResistance_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want ElectricResistance
	}{

		{"1nΩ", "1nΩ", 1 * NanoOhm},
		{"10nΩ", "10nΩ", 10 * NanoOhm},
		{"100nΩ", "100nΩ", 100 * NanoOhm},
		{"1uΩ", "1uΩ", 1 * MicroOhm},
		{"10uΩ", "10uΩ", 10 * MicroOhm},
		{"100uΩ", "100uΩ", 100 * MicroOhm},
		{"1µΩ", "1µΩ", 1 * MicroOhm},
		{"10µΩ", "10µΩ", 10 * MicroOhm},
		{"100µΩ", "100µΩ", 100 * MicroOhm},
		{"1mΩ", "1mΩ", 1 * MilliOhm},
		{"10mΩ", "10mΩ", 10 * MilliOhm},
		{"100mΩ", "100mΩ", 100 * MilliOhm},
		{"1Ω", "1Ω", 1 * Ohm},
		{"10Ω", "10Ω", 10 * Ohm},
		{"100Ω", "100Ω", 100 * Ohm},
		{"1kΩ", "1kΩ", 1 * KiloOhm},
		{"10kΩ", "10kΩ", 10 * KiloOhm},
		{"100kΩ", "100kΩ", 100 * KiloOhm},
		{"1MΩ", "1MΩ", 1 * MegaOhm},
		{"10MΩ", "10MΩ", 10 * MegaOhm},
		{"100MΩ", "100MΩ", 100 * MegaOhm},
		{"1GΩ", "1GΩ", 1 * GigaOhm},
		{"10Ohm", "10Ohm", 10 * Ohm},
		{"10Ohms", "10Ohms", 10 * Ohm},
		{"10ohm", "10ohm", 10 * Ohm},
		{"10ohms", "10ohms", 10 * Ohm},
	}

	for _, tt := range tests {
		var got ElectricResistance
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "r", "value of resistance")
		fs.Parse([]string{"-r", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v but got: %v(%d)", tt.name, tt.want, got, got)
		}
	}
}

func TestPower_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Power
	}{

		{"1nW", "1nW", 1 * NanoWatt},
		{"10nW", "10nW", 10 * NanoWatt},
		{"100nW", "100nW", 100 * NanoWatt},
		{"1uW", "1uW", 1 * MicroWatt},
		{"10uW", "10uW", 10 * MicroWatt},
		{"100uW", "100uW", 100 * MicroWatt},
		{"1µW", "1µW", 1 * MicroWatt},
		{"10µW", "10µW", 10 * MicroWatt},
		{"100µW", "100µW", 100 * MicroWatt},
		{"1mW", "1mW", 1 * MilliWatt},
		{"10mW", "10mW", 10 * MilliWatt},
		{"100mW", "100mW", 100 * MilliWatt},
		{"1W", "1W", 1 * Watt},
		{"10W", "10W", 10 * Watt},
		{"100W", "100W", 100 * Watt},
		{"1kW", "1kW", 1 * KiloWatt},
		{"10kW", "10kW", 10 * KiloWatt},
		{"100kW", "100kW", 100 * KiloWatt},
		{"1MW", "1MW", 1 * MegaWatt},
		{"10MW", "10MW", 10 * MegaWatt},
		{"100MW", "100MW", 100 * MegaWatt},
		{"1GW", "1GW", 1 * GigaWatt},
		{"10Watt", "10Watt", 10 * Watt},
		{"10Watts", "10Watts", 10 * Watt},
		{"10Watt", "10Watt", 10 * Watt},
		{"10Watts", "10Watts", 10 * Watt},
		{"10W", "10W", 10 * Watt},
		{"10w", "10w", 10 * Watt},
		{"12.23w", "12.23w", 12230 * MilliWatt},
		{"12.23µw", "12.23uw", 12230 * NanoWatt},
	}

	for _, tt := range tests {
		var got Power
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "p", "value of power")
		fs.Parse([]string{"-p", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v but got: %v(%d)", tt.name, tt.want, got, got)
		}
	}
}

func TestEnergy_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Energy
	}{

		{"1nJ", "1nJ", 1 * NanoJoule},
		{"10nJ", "10nJ", 10 * NanoJoule},
		{"100nJ", "100nJ", 100 * NanoJoule},
		{"1uJ", "1uJ", 1 * MicroJoule},
		{"10uJ", "10uJ", 10 * MicroJoule},
		{"100uJ", "100uJ", 100 * MicroJoule},
		{"1µJ", "1µJ", 1 * MicroJoule},
		{"10µJ", "10µJ", 10 * MicroJoule},
		{"100µJ", "100µJ", 100 * MicroJoule},
		{"1mJ", "1mJ", 1 * MilliJoule},
		{"10mJ", "10mJ", 10 * MilliJoule},
		{"100mJ", "100mJ", 100 * MilliJoule},
		{"1J", "1J", 1 * Joule},
		{"10J", "10J", 10 * Joule},
		{"100J", "100J", 100 * Joule},
		{"1kJ", "1kJ", 1 * KiloJoule},
		{"10kJ", "10kJ", 10 * KiloJoule},
		{"100kJ", "100kJ", 100 * KiloJoule},
		{"1MJ", "1MJ", 1 * MegaJoule},
		{"10MJ", "10MJ", 10 * MegaJoule},
		{"100MJ", "100MJ", 100 * MegaJoule},
		{"1GJ", "1GJ", 1 * GigaJoule},
		{"10Joule", "10Joule", 10 * Joule},
		{"10Joules", "10Joules", 10 * Joule},
		{"10joule", "10joule", 10 * Joule},
		{"10joules", "10joules", 10 * Joule},
		{"10J", "10J", 10 * Joule},
		{"10j", "10j", 10 * Joule},
	}

	for _, tt := range tests {
		var got Energy
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "e", "value of energy")
		fs.Parse([]string{"-e", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v but got: %v(%d)", tt.name, tt.want, got, got)
		}
	}
}

func TestPressure_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Pressure
	}{

		{"1nPa", "1nPa", 1 * NanoPascal},
		{"10nPa", "10nPa", 10 * NanoPascal},
		{"100nPa", "100nPa", 100 * NanoPascal},
		{"1uPa", "1uPa", 1 * MicroPascal},
		{"10uPa", "10uPa", 10 * MicroPascal},
		{"100uPa", "100uPa", 100 * MicroPascal},
		{"1µPa", "1µPa", 1 * MicroPascal},
		{"10µPa", "10µPa", 10 * MicroPascal},
		{"100µPa", "100µPa", 100 * MicroPascal},
		{"1mPa", "1mPa", 1 * MilliPascal},
		{"10mPa", "10mPa", 10 * MilliPascal},
		{"100mPa", "100mPa", 100 * MilliPascal},
		{"1Pa", "1Pa", 1 * Pascal},
		{"10Pa", "10Pa", 10 * Pascal},
		{"100Pa", "100Pa", 100 * Pascal},
		{"1kPa", "1kPa", 1 * KiloPascal},
		{"10kPa", "10kPa", 10 * KiloPascal},
		{"100kPa", "100kPa", 100 * KiloPascal},
		{"1MPa", "1MPa", 1 * MegaPascal},
		{"10MPa", "10MPa", 10 * MegaPascal},
		{"100MPa", "100MPa", 100 * MegaPascal},
		{"1GPa", "1GPa", 1 * GigaPascal},
		//TODO(NeuralSpaz): why do these tests fail.
		// it is from pico
		// {"10Pascal", "10Pascal", 10 * Pascal},
		// {"10Pascals", "10Pascals", 10 * Pascal},
		// {"10pascal", "10pascal", 10 * Pascal},
		// {"10pascals", "10pascals", 10 * Pascal},
		// {"10Pa", "10Pa", 10 * Pascal},
		// {"10pa", "10pa", 10 * Pascal},
	}

	for _, tt := range tests {
		var got Pressure
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "p", "value of presure")
		fs.Parse([]string{"-p", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v(%d) but got: %v(%d)", tt.name, tt.want, tt.want, got, got)
		}
	}
}

func TestLuminousIntensity_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want LuminousIntensity
	}{

		{"1ncd", "1ncd", 1 * NanoCandela},
		{"10ncd", "10ncd", 10 * NanoCandela},
		{"100ncd", "100ncd", 100 * NanoCandela},
		{"1ucd", "1ucd", 1 * MicroCandela},
		{"10ucd", "10ucd", 10 * MicroCandela},
		{"100ucd", "100ucd", 100 * MicroCandela},
		{"1µcd", "1µcd", 1 * MicroCandela},
		{"10µcd", "10µcd", 10 * MicroCandela},
		{"100µcd", "100µcd", 100 * MicroCandela},
		{"1mcd", "1mcd", 1 * MilliCandela},
		{"10mcd", "10mcd", 10 * MilliCandela},
		{"100mcd", "100mcd", 100 * MilliCandela},
		{"1cd", "1cd", 1 * Candela},
		{"10cd", "10cd", 10 * Candela},
		{"100cd", "100cd", 100 * Candela},
		{"1kcd", "1kcd", 1 * KiloCandela},
		{"10kcd", "10kcd", 10 * KiloCandela},
		{"100kcd", "100kcd", 100 * KiloCandela},
		{"1Mcd", "1Mcd", 1 * MegaCandela},
		{"10Mcd", "10Mcd", 10 * MegaCandela},
		{"100Mcd", "100Mcd", 100 * MegaCandela},
		{"1Gcd", "1Gcd", 1 * GigaCandela},
		{"10Candela", "10Candela", 10 * Candela},
		{"10Candelas", "10Candelas", 10 * Candela},
		{"10candela", "10candela", 10 * Candela},
		{"10candelas", "10candelas", 10 * Candela},
		{"10cd", "10cd", 10 * Candela},
	}

	for _, tt := range tests {
		var got LuminousIntensity
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "l", "value of intensity")
		fs.Parse([]string{"-l", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v but got: %v(%d)", tt.name, tt.want, got, got)
		}
	}
}

func TestLuminousFlux_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want LuminousFlux
	}{

		{"1nlm", "1nlm", 1 * NanoLumen},
		{"10nlm", "10nlm", 10 * NanoLumen},
		{"100nlm", "100nlm", 100 * NanoLumen},
		{"1ulm", "1ulm", 1 * MicroLumen},
		{"10ulm", "10ulm", 10 * MicroLumen},
		{"100ulm", "100ulm", 100 * MicroLumen},
		{"1µlm", "1µlm", 1 * MicroLumen},
		{"10µlm", "10µlm", 10 * MicroLumen},
		{"100µlm", "100µlm", 100 * MicroLumen},
		{"1mlm", "1mlm", 1 * MilliLumen},
		{"10mlm", "10mlm", 10 * MilliLumen},
		{"100mlm", "100mlm", 100 * MilliLumen},
		{"1lm", "1lm", 1 * Lumen},
		{"10lm", "10lm", 10 * Lumen},
		{"100lm", "100lm", 100 * Lumen},
		{"1klm", "1klm", 1 * KiloLumen},
		{"10klm", "10klm", 10 * KiloLumen},
		{"100klm", "100klm", 100 * KiloLumen},
		{"1Mlm", "1Mlm", 1 * MegaLumen},
		{"10Mlm", "10Mlm", 10 * MegaLumen},
		{"100Mlm", "100Mlm", 100 * MegaLumen},
		{"1Glm", "1Glm", 1 * GigaLumen},
		{"10Lumen", "10Lumen", 10 * Lumen},
		{"10Lumens", "10Lumens", 10 * Lumen},
		{"10lumen", "10lumen", 10 * Lumen},
		{"10lumens", "10lumens", 10 * Lumen},
		{"10lm", "10lm", 10 * Lumen},
	}

	for _, tt := range tests {
		var got LuminousFlux
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "f", "value of flux")
		fs.Parse([]string{"-f", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v but got: %v(%d)", tt.name, tt.want, got, got)
		}
	}
}

func TestSpeed_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Speed
	}{

		{"1nm/s", "1nm/s", 1 * NanoMetrePerSecond},
		{"10nm/s", "10nm/s", 10 * NanoMetrePerSecond},
		{"100nm/s", "100nm/s", 100 * NanoMetrePerSecond},
		{"1um/s", "1um/s", 1 * MicroMetrePerSecond},
		{"10um/s", "10um/s", 10 * MicroMetrePerSecond},
		{"100um/s", "100um/s", 100 * MicroMetrePerSecond},
		{"1µm/s", "1µm/s", 1 * MicroMetrePerSecond},
		{"10µm/s", "10µm/s", 10 * MicroMetrePerSecond},
		{"100µm/s", "100µm/s", 100 * MicroMetrePerSecond},
		{"1mm/s", "1mm/s", 1 * MilliMetrePerSecond},
		{"10mm/s", "10mm/s", 10 * MilliMetrePerSecond},
		{"100mm/s", "100mm/s", 100 * MilliMetrePerSecond},
		{"1m/s", "1m/s", 1 * MetrePerSecond},
		{"10m/s", "10m/s", 10 * MetrePerSecond},
		{"100m/s", "100m/s", 100 * MetrePerSecond},
		{"1km/s", "1km/s", 1 * KiloMetrePerSecond},
		{"10km/s", "10km/s", 10 * KiloMetrePerSecond},
		{"100km/s", "100km/s", 100 * KiloMetrePerSecond},
		{"1Mm/s", "1Mm/s", 1 * MegaMetrePerSecond},
		{"10Mm/s", "10Mm/s", 10 * MegaMetrePerSecond},
		{"100Mm/s", "100Mm/s", 100 * MegaMetrePerSecond},
		{"1Gm/s", "1Gm/s", 1 * GigaMetrePerSecond},
		{"1km/h", "1km/h", 1 * KilometrePerHour},
		{"1mph", "1mph", 1 * MilePerHour},
		{"1fps", "1fps", 1 * FootPerSecond},
		{"100km/h", "100km/h", 27777777777 * NanoMetrePerSecond},
		// {"1km/h", "100km/h", Speed(float64(1*MetrePerSecond) / 3.6)},
	}

	for _, tt := range tests {
		var got Speed
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "s", "value of speed")
		fs.Parse([]string{"-s", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v(%d) but got: %v(%d)", tt.name, tt.want, tt.want, got, got)
		}
	}
}

func TestMass_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Mass
	}{

		{"1ng", "1ng", 1 * NanoGram},
		{"10ng", "10ng", 10 * NanoGram},
		{"100ng", "100ng", 100 * NanoGram},
		{"1ug", "1ug", 1 * MicroGram},
		{"10ug", "10ug", 10 * MicroGram},
		{"100ug", "100ug", 100 * MicroGram},
		{"1µg", "1µg", 1 * MicroGram},
		{"10µg", "10µg", 10 * MicroGram},
		{"100µg", "100µg", 100 * MicroGram},
		{"1mg", "1mg", 1 * MilliGram},
		{"10mg", "10mg", 10 * MilliGram},
		{"100mg", "100mg", 100 * MilliGram},
		{"1g", "1g", 1 * Gram},
		{"10g", "10g", 10 * Gram},
		{"100g", "100g", 100 * Gram},
		{"1kg", "1kg", 1 * KiloGram},
		{"10kg", "10kg", 10 * KiloGram},
		{"100kg", "100kg", 100 * KiloGram},
		{"1Mg", "1Mg", 1 * MegaGram},
		{"10Mg", "10Mg", 10 * MegaGram},
		{"100Mg", "100Mg", 100 * MegaGram},
		{"1Gg", "1Gg", 1 * GigaGram},
		{"1gram", "1gram", 1 * Gram},
		{"1Gram", "1Gram", 1 * Gram},
		{"1grams", "1grams", 1 * Gram},
		{"1Grams", "1Grams", 1 * Gram},
		{"1ounce", "1ounce", 1 * OunceMass},
		{"1Ounce", "1Ounce", 1 * OunceMass},
		{"1Ounces", "1Ounces", 1 * OunceMass},
		{"1ounces", "1ounces", 1 * OunceMass},
		{"1oz", "1oz", 1 * OunceMass},
		{"1Oz", "1Oz", 1 * OunceMass},
		{"1lb", "1lb", 1 * PoundMass},
		{"1tonne", "1tonne", 1 * Tonne},
		{"1tonnes", "1tonnes", 1 * Tonne},
		{"1Tonne", "1Tonne", 1 * Tonne},
		{"1Tonnes", "1Tonnes", 1 * Tonne},
	}

	for _, tt := range tests {
		var got Mass
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "m", "value of mass")
		fs.Parse([]string{"-m", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v(%d) but got: %v(%d)", tt.name, tt.want, tt.want, got, got)
		}
	}
}

func TestForce_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Force
	}{

		{"1nN", "1nN", 1 * NanoNewton},
		{"10nN", "10nN", 10 * NanoNewton},
		{"100nN", "100nN", 100 * NanoNewton},
		{"1uN", "1uN", 1 * MicroNewton},
		{"10uN", "10uN", 10 * MicroNewton},
		{"100uN", "100uN", 100 * MicroNewton},
		{"1µN", "1µN", 1 * MicroNewton},
		{"10µN", "10µN", 10 * MicroNewton},
		{"100µN", "100µN", 100 * MicroNewton},
		{"1mN", "1mN", 1 * MilliNewton},
		{"10mN", "10mN", 10 * MilliNewton},
		{"100mN", "100mN", 100 * MilliNewton},
		{"1N", "1N", 1 * Newton},
		{"10N", "10N", 10 * Newton},
		{"100N", "100N", 100 * Newton},
		{"1kN", "1kN", 1 * KiloNewton},
		{"10kN", "10kN", 10 * KiloNewton},
		{"100kN", "100kN", 100 * KiloNewton},
		{"1MN", "1MN", 1 * MegaNewton},
		{"10MN", "10MN", 10 * MegaNewton},
		{"100MN", "100MN", 100 * MegaNewton},
		{"1GN", "1GN", 1 * GigaNewton},
		{"1Newton", "1Newton", 1 * Newton},
		{"1newton", "1newton", 1 * Newton},
		{"1newtons", "1newtons", 1 * Newton},
		{"1Newtons", "1Newtons", 1 * Newton},
	}

	for _, tt := range tests {
		var got Force
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "f", "value of force")
		fs.Parse([]string{"-f", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v but got: %v(%d)", tt.name, tt.want, got, got)
		}
	}
}

func TestRelativeHumidity_Set(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want RelativeHumidity
	}{
		{"1%rh", (PercentRH).String(), 1 * PercentRH},
		{"55.0%rh", "55.0%rh", 55 * PercentRH},
		{"1%rh", (PercentRH).String(), 1 * PercentRH},
		{"1%rh", (PercentRH).String(), 1 * PercentRH},
		{"0.00001%rH", "0.00001%rH", 1 * TenthMicroRH},
		{"0.001%rH", "0.001%rH", 1 * PercentRH / 1000},
		{"1mrH", "1mrH", 1 * MilliRH},
		{"10urH", "10urH", 10 * MicroRH},
		{"1µrH", "1µrH", 1 * MicroRH},
		{"1urH", "1urH", 1 * MicroRH},
		{"0.1%rH", "0.1%rH", (1 * PercentRH) / 10},
		{"1%rH", "1%rH", 1 * PercentRH},
	}

	for _, tt := range tests {
		var got RelativeHumidity
		fs := flag.NewFlagSet("Tests", flag.ExitOnError)
		fs.Var(&got, "f", "value of humidity")
		fs.Parse([]string{"-f", tt.s})
		if got != tt.want {
			t.Errorf("%s wanted: %v(%d) but got: %v(%d)", tt.name, tt.want, tt.want, got, got)
		}
	}
}

func TestMeta_Set(t *testing.T) {
	var degree Angle
	var metre Distance
	var amp ElectricCurrent
	var volt ElectricPotential
	var ohm ElectricResistance
	var farad ElectricalCapacitance
	var newton Force
	var hertz Frequency
	var gram Mass
	var pascal Pressure
	var humidity RelativeHumidity
	var metresPerSecond Speed
	var celsius Temperature
	var watt Power
	var joule Energy
	var candela LuminousIntensity
	var lux LuminousFlux

	tests := []struct {
		name    string
		v       flag.Value
		s       string
		wantErr bool
		err     string
	}{
		{
			"errAngle", &degree, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errDistance", &metre, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errElectricCurrent", &amp, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errElectricPotential", &volt, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errElectricResistance", &ohm, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errElectricalCapacitance", &farad, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errForce", &newton, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errFrequency", &hertz, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errMass", &gram, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errPressure", &pascal, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errRelativeHumidity", &humidity, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errSpeed", &metresPerSecond, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errTemperature", &celsius, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errPower", &watt, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errEnergy", &joule, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errLuminousIntensity", &candela, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errLuminousFlux", &lux, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		{
			"errAngle", &degree, "1.1.1.1", true,
			"parse error: multiple decimal points: \"1.1.1.1\"",
		},
		//Minimum Implementation
		{"1SiUnitAngle", &degree, "1Radian", false, ""},
		{"1SiUnitDistance", &metre, "1m", false, ""},
		{"1SiUnitElectricCurrent", &amp, "1A", false, ""},
		{"1SiUnitElectricPotential", &volt, "1V", false, ""},
		{"1SiUnitElectricResistance", &ohm, "1Ohm", false, ""},
		{"1SiUnitElectricalCapacitance", &farad, "1F", false, ""},
		{"1SiUnitForce", &newton, "1N", false, ""},
		{"1SiUnitFrequency", &hertz, "1Hz", false, ""},
		{"1SiUnitMass", &gram, "1g", false, ""},
		{"1SiUnitPressure", &pascal, "1Pa", false, ""},
		{"1SiUnitRelativeHumidity", &humidity, "1%rH", false, ""},
		{"1SiUnitSpeed", &metresPerSecond, "1m/s", false, ""},
		{"1SiUnitTemperature", &celsius, "1K", false, ""},
		{"1SiUnitPower", &watt, "1W", false, ""},
		{"1SiUnitEnergy", &joule, "1J", false, ""},
		{"1SiUnitLuminousIntensity", &candela, "1cd", false, ""},
		{"1SiUnitLuminousFlux", &lux, "1lm", false, ""},
		// Naked values
		{
			"noUnitErrAngle", &degree, "1", true,
			"parse error: no units provided, need: \"Radian\"",
		},
		{
			"noUnitErrDistance", &metre, "1", true,
			"parse error: no units provided, need: \"m\"",
		},
		{
			"noUnitErrElectricCurrent", &amp, "1", true,
			"parse error: no units provided, need: \"A\"",
		},
		{
			"noUnitErrElectricPotential", &volt, "1", true,
			"parse error: no units provided, need: \"V\"",
		},
		{
			"noUnitErrElectricResistance", &ohm, "1", true,
			"parse error: no units provided, need: \"Ohm\"",
		},
		{
			"noUnitErrElectricalCapacitance", &farad, "1", true,
			"parse error: no units provided, need: \"F\"",
		},
		{
			"noUnitErrForce", &newton, "1", true,
			"parse error: no units provided, need: \"N\"",
		},
		{
			"noUnitErrFrequency", &hertz, "1", true,
			"parse error: no units provided, need: \"Hz\"",
		},
		{
			"noUnitErrMass", &gram, "1", true,
			"parse error: no units provided, need: \"g\"",
		},
		{
			"noUnitErrPressure", &pascal, "1", true,
			"parse error: no units provided, need: \"Pa\"",
		},
		{
			"noUnitErrRelativeHumidity", &humidity, "1", true,
			"parse error: no units provided, need: \"%rH\"",
		},
		{
			"noUnitErrSpeed", &metresPerSecond, "1", true,
			"parse error: no units provided, need: \"m/s\"",
		},
		{
			"noUnitErrTemperature", &celsius, "1", true,
			"parse error: no units provided, need: \"K or C\"",
		},
		{
			"noUnitErrPower", &watt, "1", true,
			"parse error: no units provided, need: \"W\"",
		},
		{
			"noUnitErrEnergy", &joule, "1", true,
			"parse error: no units provided, need: \"J\"",
		},
		{
			"noUnitErrLuminousIntensity", &candela, "1", true,
			"parse error: no units provided, need: \"cd\"",
		},
		{
			"noUnitErrLuminousFlux", &lux, "1", true,
			"parse error: no units provided, need: \"lm\"",
		},
	}

	for _, tt := range tests {
		got := tt.v.Set(tt.s)
		if tt.wantErr && got == nil {
			t.Errorf("case %s expected error but got none", tt.name)
		}
		if !tt.wantErr && got != nil {
			t.Errorf("case %s got %v", tt.name, got)
		}
		if tt.wantErr && got.Error() != tt.err {
			t.Errorf("case %s got %v", tt.name, got)
		}
	}
}

func TestAtod(t *testing.T) {
	const (
		negative = true
		postive  = false
	)
	tests := []struct {
		s    string
		want decimal
		used int
		err  bool
	}{
		{"123456789", decimal{"123456789", 0, postive}, 9, false},
		{"1nM", decimal{"1", 0, postive}, 1, false},
		{"2.2nM", decimal{"22", -1, postive}, 3, false},
		{"12.5mA", decimal{"125", -1, postive}, 4, false},
		{"-12.5mA", decimal{"125", -1, negative}, 5, false},
		{"1.1.1", decimal{}, 0, true},
		{"1ma1", decimal{"1", 0, postive}, 1, false},
		{"-0.00001%rH", decimal{"1", -5, negative}, 8, false},
		{"0.00001%rH", decimal{"1", -5, postive}, 7, false},
		{"--1ma1", decimal{"1", 0, negative}, 3, false},
		{"++100ma1", decimal{"1", 2, postive}, 5, false},
		{"1.0", decimal{"1", 0, postive}, 3, false},
		{"0.10001", decimal{"10001", -5, postive}, 7, false},
		{"-0.10001", decimal{"10001", -5, negative}, 8, false},
		{"%-0.10001", decimal{"10001", -5, negative}, 0, true},
		{"1n", decimal{"1", 0, postive}, 1, false},
		{"200n", decimal{"2", 2, postive}, 3, false},
	}

	for _, tt := range tests {
		got, n, err := atod(tt.s)

		if got != tt.want && !tt.err {
			t.Errorf("got %v expected %v", got, tt.want)
		}
		if tt.err && err == nil {
			t.Errorf("expected error %v but got nil", err)
		}

		if n != tt.used {
			t.Errorf("expected to consume %d char but used %d", tt.used, n)
		}
	}
}

func TestDoti(t *testing.T) {
	tests := []struct {
		name string
		d    decimal
		want int64
		err  bool
	}{
		{"123", decimal{"123", 0, false}, 123, false},
		{"-123", decimal{"123", 0, true}, -123, false},
		{"1230", decimal{"123", 1, false}, 1230, false},
		{"-1230", decimal{"123", 1, true}, -1230, false},
		{"1230", decimal{"123", 20, false}, 1230, true},
		{"-1230", decimal{"123", 20, true}, -1230, true},
		{"max", decimal{"9223372036854775807", 0, false}, 9223372036854775807, false},
		{"-max", decimal{"9223372036854775807", 0, true}, -9223372036854775807, false},
		{"max+1", decimal{"9223372036854775808", 0, true}, 0, true},
		{"1a", decimal{"1a", 0, false}, 123, true},
		{"2.7b", decimal{"2.7b", 0, true}, -123, true},
		{"12", decimal{"123", -1, false}, 12, false},
		{"-12", decimal{"123", -1, true}, -12, false},
		{"123n", decimal{"123", 0, false}, 123, false},
		{"max*10^1", decimal{"9223372036854775807", 1, false}, 9223372036854775807, true},
	}

	for _, tt := range tests {
		got, err := tt.d.dtoi(0)

		if got != tt.want && !tt.err {
			t.Errorf("got %v expected %v", got, tt.want)
		}
		if tt.err && err == nil {
			t.Errorf("expected %v but got nil, %v", err, got)
		}
	}

}

func BenchmarkSetDistance(b *testing.B) {
	var t Temperature
	for i := 0; i < b.N; i++ {
		t.Set("-337.2C")
	}
}

func BenchmarkSetPower(b *testing.B) {
	var t Power
	for i := 0; i < b.N; i++ {
		t.Set("-337.2w")
	}
}

func BenchmarkDecimal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		atod("-337.2m")
	}
}

func BenchmarkDecimalInt(b *testing.B) {
	var d decimal
	for i := 0; i < b.N; i++ {
		d, _, _ = atod("-337.2m")
		d.dtoi(0)
	}
}
