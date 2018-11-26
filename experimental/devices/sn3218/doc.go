// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.
//
// Package sn3218 controls a SN3218 LED driver with 18 LEDs over an i2c bus.
// See cmd/sn3218/main.go for a usage example. Make sure to run Reset() after
// New(), so that the state of the LEDs match the register of the chip.
//
// Datasheet
// https://github.com/pimoroni/sn3218/blob/master/datasheets/sn3218-datasheet.pdf
package sn3218
