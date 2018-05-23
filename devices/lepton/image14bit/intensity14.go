// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package image14bit

import (
	"image/color"
	"strconv"
)

// Intensity14 is a 14-bit grayscale implementation of color.Color.
//
// Valid range is between 0 and 16383 (inclusive).
type Intensity14 uint16

// RGBA returns a grayscale result.
func (g Intensity14) RGBA() (uint32, uint32, uint32, uint32) {
	b := uint32(g) & 1
	i := uint32(g)<<2 | b<<1 | b
	return i, i, i, 65535
}

func (g Intensity14) String() string {
	return "Intensity14(" + strconv.Itoa(int(g)) + ")"
}

// Intensity14Model is the color Model for 14-bit grayscale.
var Intensity14Model = color.ModelFunc(convert)

func convert(c color.Color) color.Color {
	return convertIntensity14(c)
}

func convertIntensity14(c color.Color) Intensity14 {
	switch t := c.(type) {
	case Intensity14:
		return t
	default:
		r, g, b, _ := c.RGBA()
		// Use the same coefficients as color.GrayModel.
		y := (19595*r + 38470*g + 7471*b + 1<<15) >> 18
		return Intensity14(y)
	}
}
