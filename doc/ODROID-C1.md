Periph on Odroid-C1
===================

The Odroid-C1, Odroid-C1+, and Odroid-C0 boards are supported by periph using the sysfs
drivers.  
These boards use an Amlogic S805 processor (called "meson_8b" in the linux kernel).
Currently no package for memory-mapped I/O has been written for this processor,
thus all gpio functions are implemented via sysfs. The functionality supported
is:
- 2x I²C buses
- 1x SPI bus with 1x chip-enable
- 25x GPIO pins on the main J2 header

In terms of headers, the `host/odroid_c1` package exports the main J2 header,
which is rPi compatible except for a couple of analog pins (which are not
currently supported).

Tips and tricks
---------------

The ODROID-C1+ is described on Hardkernel's web site:
[ODroid-C1+](http://www.hardkernel.com/main/products/prdt_info.php?g_code=G143703355573&tab_idx=2).

The best reference for actually using the various I/O buses is the wiki:
[Hardware wiki](http://odroid.com/dokuwiki/doku.php?id=en:c1_hardware).

The periph testing is done using a minimal (head-less) Ubuntu 16.04 build,
which can be downloaded from the
[Ubuntu 16.04 downlaod page](http://odroid.in/ubuntu_16.04lts/),
although more info may be available in this
[forum thread](http://forum.odroid.com/viewtopic.php?f=112&t=22789).

By default the i2c and spi drivers are not loaded and an unusable (with periph)
1-wire driver is loaded. Most likely you will have to adjust this. The
drivers are their peculiarities are described in the Ubuntu section of this
[wiki page](http://odroid.com/dokuwiki/doku.php?id=en:odroid-c1#ubuntu).

Driver cheat sheet:
- For I²C: `modprobe aml_i2c`, buses: i2c1 on pins 3&5, i2c2 on pins 27&28.
- For SPI: `modprobe spicc`, bus on pins 19, 21, 23 and chip enable pin 24.
- To free up gpio83 from 1-wire: `rmmod w1-gpio`

Interrupts on gpio pins are limited to 8 pins when using rising or falling edges
and 4 pins when using both edges.
