// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ht16k33

import (
	"fmt"
	"strconv"

	"periph.io/x/periph/conn/i2c"
)

var digitValues = map[rune]uint16{
	' ':  0x0,
	'!':  0x6,
	'"':  0x220,
	'#':  0x12ce,
	'$':  0x12ed,
	'%':  0xc24,
	'&':  0x235d,
	'\'': 0x400,
	'(':  0x2400,
	')':  0x900,
	'*':  0x3fc0,
	'+':  0x12c0,
	',':  0x800,
	'-':  0xc0,
	'.':  0x4000,
	'/':  0xc00,
	'0':  0xc3f,
	'1':  0x6,
	'2':  0xdb,
	'3':  0x8f,
	'4':  0xe6,
	'5':  0x2069,
	'6':  0xfd,
	'7':  0x7,
	'8':  0xff,
	'9':  0xef,
	':':  0x1200,
	';':  0xa00,
	'<':  0x2400,
	'=':  0xc8,
	'>':  0x900,
	'?':  0x1083,
	'@':  0x2bb,
	'A':  0xf7,
	'B':  0x128f,
	'C':  0x39,
	'D':  0x120f,
	'E':  0xf9,
	'F':  0x71,
	'G':  0xbd,
	'H':  0xf6,
	'I':  0x1200,
	'J':  0x1e,
	'K':  0x2470,
	'L':  0x38,
	'M':  0x536,
	'N':  0x2136,
	'O':  0x3f,
	'P':  0xf3,
	'Q':  0x203f,
	'R':  0x20f3,
	'S':  0xed,
	'T':  0x1201,
	'U':  0x3e,
	'V':  0xc30,
	'W':  0x2836,
	'X':  0x2d00,
	'Y':  0x1500,
	'Z':  0xc09,
	'[':  0x39,
	'\\': 0x2100,
	']':  0xf,
	'^':  0xc03,
	'_':  0x8,
	'`':  0x100,
	'a':  0x1058,
	'b':  0x2078,
	'c':  0xd8,
	'd':  0x88e,
	'e':  0x858,
	'f':  0x71,
	'g':  0x48e,
	'h':  0x1070,
	'i':  0x1000,
	'j':  0xe,
	'k':  0x3600,
	'l':  0x30,
	'm':  0x10d4,
	'n':  0x1050,
	'o':  0xdc,
	'p':  0x170,
	'q':  0x486,
	'r':  0x50,
	's':  0x2088,
	't':  0x78,
	'u':  0x1c,
	'v':  0x2004,
	'w':  0x2814,
	'x':  0x28c0,
	'y':  0x200c,
	'z':  0x848,
	'{':  0x949,
	'|':  0x1200,
	'}':  0x2489,
	'~':  0x520,
}

// Display is a handler to control an alphanumeric display based on ht16k33.
type Display struct {
	dev *Dev
}

// NewAlphaNumericDisplay returns a Display object that communicates over I2C to ht16k33.
//
// To use on the default address, ht16k33.I2CAddr must be passed as argument.
func NewAlphaNumericDisplay(bus i2c.Bus, address uint16) (*Display, error) {
	dev, err := NewI2C(bus, address)
	if err != nil {
		return nil, err
	}

	display := &Display{dev: dev}
	return display, nil
}

// SetDigit at position to provided value.
func (d *Display) SetDigit(pos int, digit rune, decimal bool) error {
	val := digitValues[digit]
	if decimal {
		val |= digitValues['.']
	}
	return d.dev.WriteColumn(pos, val)
}

// DisplayString print string of values to the display.
//
// Characters in the string should be any ASCII value 32 to 127 (printable ASCII).
func (d *Display) DisplayString(value string, justifyRight bool) error {
	err := d.dev.Clear()
	if err != nil {
		return err
	}
	// Calculate starting position of digits based on justification.
	pos := (4 - len(value))
	if !justifyRight || pos < 0 {
		pos = 0
	}
	// Go through each character and print it on the display.
	for _, ch := range value {
		if ch == '.' {
			// Print decimal points on the previous digit.
			c := rune(value[pos-1])
			if err := d.SetDigit(pos-1, c, true); err != nil {
				return err
			}
		} else {
			if err := d.SetDigit(pos, ch, false); err != nil {
				return err
			}
			pos++
		}
	}
	return nil
}

// DisplayInt print a string of numeric values to the display.
func (d *Display) DisplayInt(value int, justifyRight bool) error {
	str := strconv.Itoa(value)
	return d.DisplayString(str, justifyRight)
}

// DisplayFloat print a string of numeric values to the display.
func (d *Display) DisplayFloat(value float64, justifyRight bool) error {
	str := fmt.Sprintf("%5f", value)
	return d.DisplayString(str, justifyRight)
}

// Halt clear all the display.
func (d *Display) Halt() error {
	return d.dev.Clear()
}
