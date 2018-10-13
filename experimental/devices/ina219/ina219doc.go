// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package ina219 a device driver for an i2c high side current shunt and power
// monitor ic. Calibration is recommended for accurate current and power
// measurements. Voltage measurements do not require sensor calibration.
//
//
// Datasheet
// http://www.ti.com/lit/ds/symlink/ina219.pdf
//
// Slave Address:
// Depending which pins the A1, A0 pins are connected to will change the slave
// address as bellow. Default configuration is address 0x40
//
//
package ina219
