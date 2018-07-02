// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package bmxx80 controls a Bosch BMP180/BME280/BMP280 device over I²C, or SPI
// for the BMx280.
//
// More details
//
// See https://periph.io/device/bmxx80/ for more details about the device.
//
// Datasheets
//
// The URLs tend to rot, visit https://www.bosch-sensortec.com if they become
// invalid.
//
// BME280:
// https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME280_DS001-12.pdf
//
// BMP280:
// https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BMP280-DS001-19.pdf
//
// BMP180:
// https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BMP180-DS000-12.pdf
//
// The font the official datasheet on page 15 is hard to read, a copy with
// readable text can be found here:
//
// https://cdn-shop.adafruit.com/datasheets/BST-BMP180-DS000-09.pdf
//
// Notes on the BMP180 datasheet
//
// The results of the calculations in the algorithm on page 15 are partly
// wrong. It looks like the original authors used non-integer calculations and
// some nubers were rounded. Take the results of the calculations with a grain
// of salt.
package bmxx80
