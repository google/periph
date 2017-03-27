// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package odroidc1 contains header definitions for Hardkernel's ODROID C0, C1,
// and C1+ boards.
//
// These boards use an Amlogic S805 processor (called "meson_8b" in the linux
// kernel). Currently no package for memory-mapped I/O has been written for
// this processor, thus all gpio functions are implemented via sysfs.
//
// This package only exports the main J2 header, which is rPi compatible except
// for a couple of analog pins (which are not currently supported). The J2
// header has two I²C buses on header pins 3/5 and 27/28, the I²C functionality
// can be enabled by loading the aml_i2c kernel module. It has one SPI bus on
// header pins 19/21/23/24. The onewire gpio driver appears to be loaded by
// default on header pin 7.
//
// References
//
// Product page: http://www.hardkernel.com/main/products/prdt_info.php?g_code=G143703355573&tab_idx=2
//
// Hardware wiki: http://odroid.com/dokuwiki/doku.php?id=en:c1_hardware
//
// Ubuntu drivers: http://odroid.com/dokuwiki/doku.php?id=en:odroid-c1#ubuntu
package odroidc1
