// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package apa102 drives a strip of APA102 LEDs connected on a SPI bus.
//
// It handles color intensity and temperature correction and uses the full
// 8000:1 dynamic range as supported by the device.
//
// Datasheet
//
// https://cpldcpu.files.wordpress.com/2014/08/apa-102c-super-led-specifications-2014-en.pdf
package apa102
