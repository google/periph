// Copyright 2020 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package tlv493d implements interfacing code to the Infineon TLV493D haff effect sensor.
//
// Features of the device:
//  3-dimensional hall effect sensor, measures up to +/-130 mT magnetic flux.
//  temperature sensor
//  i2c interface
//  12-bit resolution
//  low power consumption
//
// Features of the driver:
//  Implemented all options of the device
//  Power modes described in the documentation are defined as constants
//  2 precisions: high precision (12 bits), where all registers are read or low precision, which saves 50% of I2C bandwidth, but without temperature and only 8-bit resolution
//  Continuous reading mode
//
// Datasheet and application notes:
// https://www.infineon.com/cms/en/product/sensor/magnetic-sensors/magnetic-position-sensors/3d-magnetics/tlv493d-a1b6/
//
package tlv493d
