// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package nrzled is a driver for LEDs ws2811/ws2812/ws2812b and compatible
// devices like sk6812 and ucs1903 that uses a single wire NRZ encoded
// communication protocol.
//
// Note that some ICs are 7 bits with the least significant bit ignored, others
// are using a real 8 bits PWM. The PWM frequency varies across ICs.
//
// Datasheet
//
// This directory contains datasheets for ws2812, ws2812b, ucs190x and various
// sk6812.
//
// https://github.com/cpldcpu/light_ws2812/tree/master/Datasheets
//
// UCS1903 datasheet
//
// http://www.bestlightingbuy.com/pdf/UCS1903%20datasheet.pdf
package nrzled
