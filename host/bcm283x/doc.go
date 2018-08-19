// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package bcm283x exposes the BCM283x GPIO functionality.
//
// This driver implements memory-mapped GPIO pin manipulation and leverages
// sysfs-gpio for edge detection.
//
// If you are looking for the actual implementation, open doc.go for further
// implementation details.
//
// GPIOs
//
// Aliases for GPCLK0, GPCLK1, GPCLK2 are created for corresponding CLKn pins.
// Same for PWM0_OUT and PWM1_OUT, which point respectively to PWM0 and PWM1.
//
// Datasheet
//
// https://www.raspberrypi.org/wp-content/uploads/2012/02/BCM2835-ARM-Peripherals.pdf
//
// Its crowd-sourced errata: http://elinux.org/BCM2835_datasheet_errata
//
// BCM2836:
// https://www.raspberrypi.org/documentation/hardware/raspberrypi/bcm2836/QA7_rev3.4.pdf
//
// Another doc about PCM and PWM:
// https://scribd.com/doc/127599939/BCM2835-Audio-clocks
//
// GPIO pad control:
// https://scribd.com/doc/101830961/GPIO-Pads-Control2
package bcm283x

// Other implementations details
//
// mainline:
// https://github.com/torvalds/linux/blob/master/drivers/dma/bcm2835-dma.c
// https://github.com/torvalds/linux/blob/master/drivers/gpio
//
// Raspbian kernel:
// https://github.com/raspberrypi/linux/blob/rpi-4.11.y/drivers/dma
// https://github.com/raspberrypi/linux/blob/rpi-4.11.y/drivers/gpio
