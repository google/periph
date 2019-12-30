// Copyright 2019 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package inky drives an Inky pHAT or wHAT E ink display.
//
// Datasheet
//
// Inky lacks a true datasheet, so the code here is derived from the reference
// implementation by Pimoroni:
// https://github.com/pimoroni/inky
//
// The display seems to use a SSD1675 controller:
// https://www.china-epaper.com/uploads/soft/DEPG0420R01V3.0.pdf
package inky
