// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package physic

import (
	"bytes"
	"errors"
	"fmt"
	"log"
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

func TestDistance_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected Distance
	}{
		{"1nMetre", 1 * NanoMetre},
		{"10nMetres", 10 * NanoMetre},
		{"100nMetres", 100 * NanoMetre},
		{"1uMetre", 1 * MicroMetre},
		{"10uMetres", 10 * MicroMetre},
		{"100uMetres", 100 * MicroMetre},
		{"1µMetre", 1 * MicroMetre},
		{"10µMetres", 10 * MicroMetre},
		{"100µMetres", 100 * MicroMetre},
		{"1mm", 1 * MilliMetre},
		{"1mMetre", 1 * MilliMetre},
		{"10mMetres", 10 * MilliMetre},
		{"100mMetres", 100 * MilliMetre},
		{"1Metre", 1 * Metre},
		{"10Metres", 10 * Metre},
		{"100Metres", 100 * Metre},
		{"1kMetre", 1 * KiloMetre},
		{"10kMetres", 10 * KiloMetre},
		{"100kMetres", 100 * KiloMetre},
		{"1MMetre", 1 * MegaMetre},
		{"1Mm", 1 * MegaMetre},
		{"10MMetres", 10 * MegaMetre},
		{"100MMetres", 100 * MegaMetre},
		{"1GMetre", 1 * GigaMetre},
		{"12.345Metres", 12345 * MilliMetre},
		{"-12.345Metres", -12345 * MilliMetre},
		{"9.223372036854775807GMetres", 9223372036854775807 * NanoMetre},
		{"-9.223372036854775807GMetres", -9223372036854775807 * NanoMetre},
		{"1Mm", 1 * MegaMetre},
		{"5Miles", 8046720000000 * NanoMetre},
		{"3Feet", 914400000 * NanoMetre},
		{"10Yards", 9144000000 * NanoMetre},
		{"5731.137678988Mile", 9223372036853264 * NanoMetre},
		{"-5731.137678988Mile", -9223372036853264 * NanoMetre},
		{"1.008680231502051MYards", 922337203685475 * NanoMetre},
		{"-1008680.231502051Yards", -922337203685475 * NanoMetre},
		{"3026040.694506158Feet", 922337203685477 * NanoMetre},
		{"-3.026040694506158MFeet", -922337203685477 * NanoMetre},
		{"36.312488334073900MInch", 922337203685477 * NanoMetre},
		{"-36312488.334073900Inch", -922337203685477 * NanoMetre},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TMetre",
			"exponent exceeds int64",
		},
		{
			"10EMetre",
			"contains unknown unit prefix \"E\". valid prefixes for \"Metre\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaMetre",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"Metre\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eMetreE",
			"contains unknown unit prefix \"e\". valid prefixes for \"Metre\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need m, Metre, Mile, Inch, Foot or Yard",
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
			"9.223372036854775808GMetre",
			"maximum value is 9.223Gm",
		},
		{
			"-9.223372036854775808GMetre",
			"minimum value is -9.223Gm",
		},
		{
			"9.223372036854775808GMetre",
			"maximum value is 9.223Gm",
		},
		{
			"-9.223372036854775808GMetre",
			"minimum value is -9.223Gm",
		},
		{
			"5731.137678989Mile",
			"maximum value is 5731Miles",
		},
		{
			"-5731.1376789889Mile",
			"minimum value is -5731Miles",
		},
		{
			"1.008680231502053MYards",
			"maximum value is 1 Million Yards",
		},
		{
			"-1008680.231502053Yards",
			"minimum value is -1 Million Yards",
		},
		{
			"3026040.694506159Feet",
			"maximum value is 3 Million Feet",
		},
		{
			"-3.026040694506159MFeet",
			"minimum value is 3 Million Feet",
		},
		{
			"36.312488334073901MInch",
			"maximum value is 36 Million Inches",
		},
		{
			"-36312488.334073901Inch",
			"minimum value is 36 Million Inches",
		},
		{
			"1random",
			"contains unknown unit prefix \"rando\". valid prefixes for \"m\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"Metre",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number",
		},
		{
			"CANDELA",
			"does not contain number or unit \"Metre\"",
		},
		{
			"1Jaunt",
			"\"Jaunt\" is not a valid unit for physic.Distance",
		},
		{
			"++1Metre",
			"multiple plus symbols ++1Metre",
		},
		{
			"--1Metre",
			"multiple minus symbols --1Metre",
		},
		{
			"+-1Metre",
			"can't contain both plus and minus symbols +-1Metre",
		},
		{
			"1.1.1.1Metre",
			"multiple decimal points 1.1.1.1Metre",
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

func TestElectricCurrent_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected ElectricCurrent
	}{
		{"1nAmp", 1 * NanoAmpere},
		{"10nAmps", 10 * NanoAmpere},
		{"100nAmps", 100 * NanoAmpere},
		{"1uAmp", 1 * MicroAmpere},
		{"10uAmps", 10 * MicroAmpere},
		{"100uAmps", 100 * MicroAmpere},
		{"1µAmp", 1 * MicroAmpere},
		{"10µAmps", 10 * MicroAmpere},
		{"100µAmps", 100 * MicroAmpere},
		{"1mAmp", 1 * MilliAmpere},
		{"10mAmps", 10 * MilliAmpere},
		{"100mAmps", 100 * MilliAmpere},
		{"1Amp", 1 * Ampere},
		{"10Amps", 10 * Ampere},
		{"100Amps", 100 * Ampere},
		{"1kAmp", 1 * KiloAmpere},
		{"10kAmps", 10 * KiloAmpere},
		{"100kAmps", 100 * KiloAmpere},
		{"1MAmp", 1 * MegaAmpere},
		{"10MAmps", 10 * MegaAmpere},
		{"100MAmps", 100 * MegaAmpere},
		{"1GAmp", 1 * GigaAmpere},
		{"12.345Amps", 12345 * MilliAmpere},
		{"-12.345Amps", -12345 * MilliAmpere},
		{"9.223372036854775807GAmps", 9223372036854775807 * NanoAmpere},
		{"-9.223372036854775807GAmps", -9223372036854775807 * NanoAmpere},
		{"1A", 1 * Ampere},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TAmp",
			"exponent exceeds int64",
		},
		{
			"10EAmp",
			"contains unknown unit prefix \"E\". valid prefixes for \"Amp\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaAmp",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"Amp\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eAmpE",
			"contains unknown unit prefix \"e\". valid prefixes for \"Amp\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need Amp",
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
			"9.223372036854775808GAmp",
			"maximum value is 9.223GA",
		},
		{
			"-9.223372036854775808GAmp",
			"minimum value is -9.223GA",
		},
		{
			"9.223372036854775808GAmp",
			"maximum value is 9.223GA",
		},
		{
			"-9.223372036854775808GAmp",
			"minimum value is -9.223GA",
		},
		{
			"1junk",
			"\"junk\" is not a valid unit for physic.ElectricCurrent",
		},
		{
			"Amp",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"Amp\"",
		},
		{
			"++1Amp",
			"multiple plus symbols ++1Amp",
		},
		{
			"--1Amp",
			"multiple minus symbols --1Amp",
		},
		{
			"+-1Amp",
			"can't contain both plus and minus symbols +-1Amp",
		},
		{
			"1.1.1.1Amp",
			"multiple decimal points 1.1.1.1Amp",
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

func TestPressure_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected Pressure
	}{
		{"1nPascal", 1 * NanoPascal},
		{"10nPascals", 10 * NanoPascal},
		{"100nPascals", 100 * NanoPascal},
		{"1uPascal", 1 * MicroPascal},
		{"10uPascals", 10 * MicroPascal},
		{"100uPascals", 100 * MicroPascal},
		{"1µPascal", 1 * MicroPascal},
		{"10µPascals", 10 * MicroPascal},
		{"100µPascals", 100 * MicroPascal},
		{"1mPascal", 1 * MilliPascal},
		{"10mPascals", 10 * MilliPascal},
		{"100mPascals", 100 * MilliPascal},
		{"1Pascal", 1 * Pascal},
		{"10Pascals", 10 * Pascal},
		{"100Pascals", 100 * Pascal},
		{"1kPascal", 1 * KiloPascal},
		{"10kPascals", 10 * KiloPascal},
		{"100kPascals", 100 * KiloPascal},
		{"1MPascal", 1 * MegaPascal},
		{"10MPascals", 10 * MegaPascal},
		{"100MPascals", 100 * MegaPascal},
		{"1GPascal", 1 * GigaPascal},
		{"12.345Pascals", 12345 * MilliPascal},
		{"-12.345Pascals", -12345 * MilliPascal},
		{"9.223372036854775807GPascals", 9223372036854775807 * NanoPascal},
		{"-9.223372036854775807GPascals", -9223372036854775807 * NanoPascal},
		{"1MPa", 1 * MegaPascal},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TPascal",
			"exponent exceeds int64",
		},
		{
			"10EPascal",
			"contains unknown unit prefix \"E\". valid prefixes for \"Pascal\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaPascal",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"Pascal\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ePascalE",
			"contains unknown unit prefix \"e\". valid prefixes for \"Pascal\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need Pascal",
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
			"9.223372036854775808GPascal",
			"maximum value is 9.223GPa",
		},
		{
			"-9.223372036854775808GPascal",
			"minimum value is -9.223GPa",
		},
		{
			"9.223372036854775808GPascal",
			"maximum value is 9.223GPa",
		},
		{
			"-9.223372036854775808GPascal",
			"minimum value is -9.223GPa",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.Pressure",
		},
		{
			"Pascal",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"Pascal\"",
		},
		{
			"++1Pascal",
			"multiple plus symbols ++1Pascal",
		},
		{
			"--1Pascal",
			"multiple minus symbols --1Pascal",
		},
		{
			"+-1Pascal",
			"can't contain both plus and minus symbols +-1Pascal",
		},
		{
			"1.1.1.1Pascal",
			"multiple decimal points 1.1.1.1Pascal",
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

func TestPower_Set(t *testing.T) {
	succeeds := []struct {
		in       string
		expected Power
	}{
		{"1nWatt", 1 * NanoWatt},
		{"10nWatts", 10 * NanoWatt},
		{"100nWatts", 100 * NanoWatt},
		{"1uWatt", 1 * MicroWatt},
		{"10uWatts", 10 * MicroWatt},
		{"100uWatts", 100 * MicroWatt},
		{"1µWatt", 1 * MicroWatt},
		{"10µWatts", 10 * MicroWatt},
		{"100µWatts", 100 * MicroWatt},
		{"1mWatt", 1 * MilliWatt},
		{"10mWatts", 10 * MilliWatt},
		{"100mWatts", 100 * MilliWatt},
		{"1Watt", 1 * Watt},
		{"10Watts", 10 * Watt},
		{"100Watts", 100 * Watt},
		{"1kWatt", 1 * KiloWatt},
		{"10kWatts", 10 * KiloWatt},
		{"100kWatts", 100 * KiloWatt},
		{"1MWatt", 1 * MegaWatt},
		{"10MWatts", 10 * MegaWatt},
		{"100MWatts", 100 * MegaWatt},
		{"1GWatt", 1 * GigaWatt},
		{"12.345Watts", 12345 * MilliWatt},
		{"-12.345Watts", -12345 * MilliWatt},
		{"9.223372036854775807GWatts", 9223372036854775807 * NanoWatt},
		{"-9.223372036854775807GWatts", -9223372036854775807 * NanoWatt},
		{"1MW", 1 * MegaWatt},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TWatt",
			"exponent exceeds int64",
		},
		{
			"10EWatt",
			"contains unknown unit prefix \"E\". valid prefixes for \"Watt\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaWatt",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"Watt\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eWattE",
			"contains unknown unit prefix \"e\". valid prefixes for \"Watt\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need Watt",
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
			"9.223372036854775808GWatt",
			"maximum value is 9.223GW",
		},
		{
			"-9.223372036854775808GWatt",
			"minimum value is -9.223GW",
		},
		{
			"9.223372036854775808GWatt",
			"maximum value is 9.223GW",
		},
		{
			"-9.223372036854775808GWatt",
			"minimum value is -9.223GW",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.Power",
		},
		{
			"Watt",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"Watt\"",
		},
		{
			"++1Watt",
			"multiple plus symbols ++1Watt",
		},
		{
			"--1Watt",
			"multiple minus symbols --1Watt",
		},
		{
			"+-1Watt",
			"can't contain both plus and minus symbols +-1Watt",
		},
		{
			"1.1.1.1Watt",
			"multiple decimal points 1.1.1.1Watt",
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
		{"1nJoule", 1 * NanoJoule},
		{"10nJoules", 10 * NanoJoule},
		{"100nJoules", 100 * NanoJoule},
		{"1uJoule", 1 * MicroJoule},
		{"10uJoules", 10 * MicroJoule},
		{"100uJoules", 100 * MicroJoule},
		{"1µJoule", 1 * MicroJoule},
		{"10µJoules", 10 * MicroJoule},
		{"100µJoules", 100 * MicroJoule},
		{"1mJoule", 1 * MilliJoule},
		{"10mJoules", 10 * MilliJoule},
		{"100mJoules", 100 * MilliJoule},
		{"1Joule", 1 * Joule},
		{"10Joules", 10 * Joule},
		{"100Joules", 100 * Joule},
		{"1kJoule", 1 * KiloJoule},
		{"10kJoules", 10 * KiloJoule},
		{"100kJoules", 100 * KiloJoule},
		{"1MJoule", 1 * MegaJoule},
		{"10MJoules", 10 * MegaJoule},
		{"100MJoules", 100 * MegaJoule},
		{"1GJoule", 1 * GigaJoule},
		{"12.345Joules", 12345 * MilliJoule},
		{"-12.345Joules", -12345 * MilliJoule},
		{"9.223372036854775807GJoules", 9223372036854775807 * NanoJoule},
		{"-9.223372036854775807GJoules", -9223372036854775807 * NanoJoule},
		{"1MJ", 1 * MegaJoule},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TJoule",
			"exponent exceeds int64",
		},
		{
			"10EJoule",
			"contains unknown unit prefix \"E\". valid prefixes for \"Joule\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaJoule",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"Joule\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eJouleE",
			"contains unknown unit prefix \"e\". valid prefixes for \"Joule\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need Joule",
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
			"9.223372036854775808GJoule",
			"maximum value is 9.223GJ",
		},
		{
			"-9.223372036854775808GJoule",
			"minimum value is -9.223GJ",
		},
		{
			"9.223372036854775808GJoule",
			"maximum value is 9.223GJ",
		},
		{
			"-9.223372036854775808GJoule",
			"minimum value is -9.223GJ",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.Energy",
		},
		{
			"Joule",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"Joule\"",
		},
		{
			"++1Joule",
			"multiple plus symbols ++1Joule",
		},
		{
			"--1Joule",
			"multiple minus symbols --1Joule",
		},
		{
			"+-1Joule",
			"can't contain both plus and minus symbols +-1Joule",
		},
		{
			"1.1.1.1Joule",
			"multiple decimal points 1.1.1.1Joule",
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
		{"1pFarad", 1 * PicoFarad},
		{"10pFarads", 10 * PicoFarad},
		{"100pFarads", 100 * PicoFarad},
		{"1nFarad", 1 * NanoFarad},
		{"10nFarads", 10 * NanoFarad},
		{"100nFarads", 100 * NanoFarad},
		{"1uFarad", 1 * MicroFarad},
		{"10uFarads", 10 * MicroFarad},
		{"100uFarads", 100 * MicroFarad},
		{"1µFarad", 1 * MicroFarad},
		{"10µFarads", 10 * MicroFarad},
		{"100µFarads", 100 * MicroFarad},
		{"1mFarad", 1 * MilliFarad},
		{"10mFarads", 10 * MilliFarad},
		{"100mFarads", 100 * MilliFarad},
		{"1Farad", 1 * Farad},
		{"10Farads", 10 * Farad},
		{"100Farads", 100 * Farad},
		{"1kFarad", 1 * KiloFarad},
		{"10kFarads", 10 * KiloFarad},
		{"100kFarads", 100 * KiloFarad},
		{"1MFarad", 1 * MegaFarad},
		{"12.345Farads", 12345 * MilliFarad},
		{"-12.345Farads", -12345 * MilliFarad},
		{"9.223372036854775807MFarads", 9223372036854775807 * PicoFarad},
		{"-9.223372036854775807MFarads", -9223372036854775807 * PicoFarad},
		{"1MF", 1 * MegaFarad},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TFarad",
			"exponent exceeds int64",
		},
		{
			"10EFarad",
			"contains unknown unit prefix \"E\". valid prefixes for \"Farad\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaFarad",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"Farad\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eFaradE",
			"contains unknown unit prefix \"e\". valid prefixes for \"Farad\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need Farad",
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
			"9.223372036854775808MFarad",
			"maximum value is 9.223MF",
		},
		{
			"-9.223372036854775808MFarad",
			"minimum value is -9.223MF",
		},
		{
			"9.223372036854775808MFarad",
			"maximum value is 9.223MF",
		},
		{
			"-9.223372036854775808MFarad",
			"minimum value is -9.223MF",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.ElectricalCapacitance",
		},
		{
			"Farad",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"Farad\"",
		},
		{
			"++1Farad",
			"multiple plus symbols ++1Farad",
		},
		{
			"--1Farad",
			"multiple minus symbols --1Farad",
		},
		{
			"+-1Farad",
			"can't contain both plus and minus symbols +-1Farad",
		},
		{
			"1.1.1.1Farad",
			"multiple decimal points 1.1.1.1Farad",
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
		{"1nCandela", 1 * NanoCandela},
		{"10nCandelas", 10 * NanoCandela},
		{"100nCandelas", 100 * NanoCandela},
		{"1uCandela", 1 * MicroCandela},
		{"10uCandelas", 10 * MicroCandela},
		{"100uCandelas", 100 * MicroCandela},
		{"1µCandela", 1 * MicroCandela},
		{"10µCandelas", 10 * MicroCandela},
		{"100µCandelas", 100 * MicroCandela},
		{"1mCandela", 1 * MilliCandela},
		{"10mCandelas", 10 * MilliCandela},
		{"100mCandelas", 100 * MilliCandela},
		{"1Candela", 1 * Candela},
		{"10Candelas", 10 * Candela},
		{"100Candelas", 100 * Candela},
		{"1kCandela", 1 * KiloCandela},
		{"10kCandelas", 10 * KiloCandela},
		{"100kCandelas", 100 * KiloCandela},
		{"1MCandela", 1 * MegaCandela},
		{"10MCandelas", 10 * MegaCandela},
		{"100MCandelas", 100 * MegaCandela},
		{"1GCandela", 1 * GigaCandela},
		{"12.345Candelas", 12345 * MilliCandela},
		{"-12.345Candelas", -12345 * MilliCandela},
		{"9.223372036854775807GCandelas", 9223372036854775807 * NanoCandela},
		{"-9.223372036854775807GCandelas", -9223372036854775807 * NanoCandela},
		{"1Mcd", 1 * MegaCandela},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TCandela",
			"exponent exceeds int64",
		},
		{
			"10ECandela",
			"contains unknown unit prefix \"E\". valid prefixes for \"Candela\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaCandela",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"Candela\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eCandelaE",
			"contains unknown unit prefix \"e\". valid prefixes for \"Candela\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need Candela",
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
			"9.223372036854775808GCandela",
			"maximum value is 9.223Gcd",
		},
		{
			"-9.223372036854775808GCandela",
			"minimum value is -9.223Gcd",
		},
		{
			"9.223372036854775808GCandela",
			"maximum value is 9.223Gcd",
		},
		{
			"-9.223372036854775808GCandela",
			"minimum value is -9.223Gcd",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.LuminousIntensity",
		},
		{
			"Candela",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"Candela\"",
		},
		{
			"++1Candela",
			"multiple plus symbols ++1Candela",
		},
		{
			"--1Candela",
			"multiple minus symbols --1Candela",
		},
		{
			"+-1Candela",
			"can't contain both plus and minus symbols +-1Candela",
		},
		{
			"1.1.1.1Candela",
			"multiple decimal points 1.1.1.1Candela",
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
		{"1nLumen", 1 * NanoLumen},
		{"10nLumens", 10 * NanoLumen},
		{"100nLumens", 100 * NanoLumen},
		{"1uLumen", 1 * MicroLumen},
		{"10uLumens", 10 * MicroLumen},
		{"100uLumens", 100 * MicroLumen},
		{"1µLumen", 1 * MicroLumen},
		{"10µLumens", 10 * MicroLumen},
		{"100µLumens", 100 * MicroLumen},
		{"1mLumen", 1 * MilliLumen},
		{"10mLumens", 10 * MilliLumen},
		{"100mLumens", 100 * MilliLumen},
		{"1Lumen", 1 * Lumen},
		{"10Lumens", 10 * Lumen},
		{"100Lumens", 100 * Lumen},
		{"1kLumen", 1 * KiloLumen},
		{"10kLumens", 10 * KiloLumen},
		{"100kLumens", 100 * KiloLumen},
		{"1MLumen", 1 * MegaLumen},
		{"10MLumens", 10 * MegaLumen},
		{"100MLumens", 100 * MegaLumen},
		{"1GLumen", 1 * GigaLumen},
		{"12.345Lumens", 12345 * MilliLumen},
		{"-12.345Lumens", -12345 * MilliLumen},
		{"9.223372036854775807GLumens", 9223372036854775807 * NanoLumen},
		{"-9.223372036854775807GLumens", -9223372036854775807 * NanoLumen},
		{"1Mlm", 1 * MegaLumen},
	}

	fails := []struct {
		in  string
		err string
	}{
		{
			"10TLumen",
			"exponent exceeds int64",
		},
		{
			"10ELumen",
			"contains unknown unit prefix \"E\". valid prefixes for \"Lumen\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10ExaLumen",
			"contains unknown unit prefix \"Exa\". valid prefixes for \"Lumen\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10eLumenE",
			"contains unknown unit prefix \"e\". valid prefixes for \"Lumen\" are p,n,u,µ,m,k,M,G or T",
		},
		{
			"10",
			"no units provided, need Lumen",
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
			"9.223372036854775808GLumen",
			"maximum value is 9.223Glm",
		},
		{
			"-9.223372036854775808GLumen",
			"minimum value is -9.223Glm",
		},
		{
			"9.223372036854775808GLumen",
			"maximum value is 9.223Glm",
		},
		{
			"-9.223372036854775808GLumen",
			"minimum value is -9.223Glm",
		},
		{
			"1random",
			"\"random\" is not a valid unit for physic.LuminousFlux",
		},
		{
			"Lumen",
			"does not contain number",
		},
		{
			"RPM",
			"does not contain number or unit \"Lumen\"",
		},
		{
			"++1Lumen",
			"multiple plus symbols ++1Lumen",
		},
		{
			"--1Lumen",
			"multiple minus symbols --1Lumen",
		},
		{
			"+-1Lumen",
			"can't contain both plus and minus symbols +-1Lumen",
		},
		{
			"1.1.1.1Lumen",
			"multiple decimal points 1.1.1.1Lumen",
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

func BenchmarkElectricCurrentSet(b *testing.B) {
	var err error
	var d Distance
	for i := 0; i < b.N; i++ {
		err = d.Set("1Foot")
		if err != nil {
			log.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%d", d)
}

func BenchmarkDistanceSet(b *testing.B) {
	var err error
	var e ElectricCurrent
	for i := 0; i < b.N; i++ {
		err = e.Set("1Amp")
		if err != nil {
			log.Fatal(err)
		}
	}
	b.StopTimer()
	_ = fmt.Sprintf("%d", e)
}
