// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package bcm283x exposes the BCM283x GPIO functionality.
//
// This driver implements memory-mapped GPIO pin manipulation and leverages
// sysfs-gpio for edge detection.
//
// Datasheet
//
// https://www.raspberrypi.org/wp-content/uploads/2012/02/BCM2835-ARM-Peripherals.pdf
//
// Its crowd-sourced errata: http://elinux.org/BCM2835_datasheet_errata
//
// Another doc about PCM and PWM:
// https://fr.scribd.com/doc/127599939/BCM2835-Audio-clocks
package bcm283x
