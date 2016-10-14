// Copyright 2016 The PIO Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package rpi contains Raspberry Pi hardware logic. It is intrinsically
// related to package bcm283x.
//
// Requires Raspbian Jessie.
//
// Contains both low level GPIO logic (including edge triggering) and higher
// level communication like InfraRed, SPI (2 host) and I²C (2 host).
//
// No code in this module is "thread-safe".
//
// Configuration
//
// The pins function can be affected by device overlays as defined in
// /boot/config.txt. The full documentation of overlays is at
// https://github.com/raspberrypi/firmware/blob/master/boot/overlays/README.
// Documentation for the file format at
// https://www.raspberrypi.org/documentation/configuration/device-tree.md#part3
//
// I²C
//
// The BCM238x has 2 I²C host.
//
// - /dev/i2c-1 can be enabled with:
//    dtparam=i2c=on
// - /dev/i2c-0 can be enabled with the following but be warned that it
// conflicts with HAT EEPROM detection at boot
// https://github.com/raspberrypi/hats
//    dtparam=i2c_vc=on
//    dtoverlay=i2c0-bcm2708 (Confirm?)
//
// I2S
//
// Can be enabled with:
//     dtparam=i2s=on
//
// IR
//
// Raspbian has a specific device tree overlay named "lirc-rpi" to enable
// hardware based decoding of IR signals. This loads a kernel module that
// exposes itself at /dev/lirc0. You can add the following in your
// /boot/config.txt:
//
//     dtoverlay=lirc-rpi,gpio_out_pin=5,gpio_in_pin=13,gpio_in_pull=high
//
// Default pins 17 and 18 clashes with SPI1 so change the pin if you plan to
// enable both SPI buses.
//
// See
// https://github.com/raspberrypi/firmware/blob/master/boot/overlays/README for
// more details on configuring the kernel module.
//
// /etc/lirc/hardware.conf
//
// Once the kernel module is configure, you need to point lircd to it. Run the
// following as root to point lircd to use lirc_rpi kernel module:
//
//    sed -i s'/DRIVER="UNCONFIGURED"/DRIVER="default"/' /etc/lirc/hardware.conf
//    sed -i s'/DEVICE=""/DEVICE="\/dev\/lirc0"/' /etc/lirc/hardware.conf
//    sed -i s'/MODULES=""/MODULES="lirc_rpi"/' /etc/lirc/hardware.conf
//
// IR/Sources
//
// https://github.com/raspberrypi/linux/blob/rpi-4.8.y/drivers/staging/media/lirc/lirc_rpi.c
//
// Someone made a version that supports multiple devices:
// https://github.com/bengtmartensson/lirc_rpi
//
// PWM
//
// To take back control to use as general purpose PWM, comment out the
// following line:
//     dtparam=audio=on
//
// SPI
//
// The BCM238x has 3 SPI host but only two are usable. The pins implementing
// bcm283x.SPI2_xxx pins are not physically usable because they are not
// connected to hardware pins on the board.
//
// - /dev/spidev0.0 and /dev/spidev0.1 can be enabled with:
//     dtparam=spi=on
// - /dev/spidev1.0 can be enabled with:
//     dtoverlay=spi1-1cs
// On rPi3, bluetooth must be disabled with:
//     dtoverlay=pi3-disable-bt
// and bluetooth UART service needs to be disabled with:
//     sudo systemctl disable hciuart
//
// UART
//
// Kernel boot messages go to the UART (0 or 1, depending on Pi version) at
// 115200 bauds.
//
// On Rasberry Pi 1 and 2, UART0 is used.
//
// On Raspberry Pi 3, UART0 is connected to bluetooth so the console is
// connected to UART1 instead. Disabling bluetooth also reverts to use UART0
// and not UART1.
//
// UART0 can be disabled with:
//     dtparam=uart0=off
// UART1 can be enabled with:
//     dtoverlay=uart1
//
// Physical
//
// The physical pin out is based on http://www.raspberrypi.org information but
// http://pinout.xyz/ has a nice interactive web page.
package rpi
