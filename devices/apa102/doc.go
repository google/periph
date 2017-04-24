// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package apa102 drives a strip of APA102 LEDs connected on a SPI bus.
//
// These peripherals are interesting because they have 2 PWMs: one global of 5
// bits of resolution and one per channel of 8 bits of resolution. This means
// that the dynamic range is of 13 bits.
//
// This driver handles color intensity and temperature correction and uses the
// full near 8000:1 dynamic range as supported by the peripheral.
//
// Datasheet
//
// https://cpldcpu.files.wordpress.com/2014/08/apa-102c-super-led-specifications-2014-en.pdf
package apa102
