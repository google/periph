// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package ina219 controls a Texas Instruments ina219 high side current,
// voltage and power monitor IC over an i2c bus.
//
// Calibration
//
// Calibration is recommended for accurate current and power measurements.
// Voltage measurements do not require sensor calibration. To calibrate meansure
// the actual value of the shunt resistor.
//
// Datasheet
//
// http://www.ti.com/lit/ds/symlink/ina219.pdf
package ina219
