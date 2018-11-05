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
		{"-0.00001%rH", decimal{"1", -5, negative}, 8},
		{"0.00001%rH", decimal{"1", -5, positive}, 7},
		{"1.0", decimal{"1", 0, positive}, 3},
		{"0.10001", decimal{"10001", -5, positive}, 7},
		{"-0.10001", decimal{"10001", -5, negative}, 8},
		{"1n", decimal{"1", 0, positive}, 1},
		{"200n", decimal{"2", 2, positive}, 3},
		{".01", decimal{"1", -2, positive}, 3},
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
