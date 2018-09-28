// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package ws2812b drives a strip of Worldsemi ws2812b LEDs connected to the MOSI pin of a SPI port.
// This approach avoids the need for root access, and leverages the SPI hardware for timing and shifting. 
// 
// These lights are popularized by Adafruit under the brand name Neopixels.
// Their datasheet can be found at https://cdn-shop.adafruit.com/datasheets/WS2812B.pdf
//
// This driver is sensitive to variations in SPI clock speed.
// On the Raspberry Pi, you will need to add `core_freq=250`
// to /boot/config.txt to prevent glitching.
//
// You may also need to increase your SPI buffer size to 12*num_pixels+3, or just max it out
// with `spidev.bufsize=65536`. That should allopw you to buffer over 5400 Neopixels.
package ws2812b
