# Periph on C.H.I.P

The NextThing Co's C.H.I.P. board is supported by periph using sysfs drivers
as well as using memory-mapped I/O for gpio pins. The CHIPs use an
Allwinner R8 processor. The following functionality is supported:

- 3x I²C buses
- 1x SPI bus with 1x chip-enable
- 8x GPIO pins via pcf8574 I²C I/O extender ("XIO" pins)
- 43x memory-mapped GPIO pins (the "LCD" and "CSI" pins and a few more)

In terms of headers, the `host/chip` package exports the two U13 and U14
headers.


## Tips and tricks

CHIP is described at NextThing's [product
page](https://www.getchip.com/pages/chip) and in much more detail in the [CHIP
Hardware](http://docs.getchip.com/chip.html#chip-hardware) section of the
documentation.

For in-depth information about the hardware the best reference is in the
[community wiki](http://www.chip-community.org/index.php/Hardware_Information),
which also has a section on [building kernels and device
drivers](http://www.chip-community.org/index.php/Kernel_Hacking).

The periph testing is done using the headless Debian image provided by NTC
in the [CHIP flasher](http://flash.getchip.com/).

The headless image released by NTC in Nov 2016 using kernel 4.4.13-ntc-mlc
has the i2c driver loaded by default and exposes all three I²C buses.
The SPI driver is also included, but a DTBO (Device Tree Binary Overlay)
is required in order to create the /dev/spi32766.0 device and connect it
to the pins.

Buses cheat sheet:

- i2c0: not available on the headers but has axp209 power control chip
- i2c1: U13 pins 9 & 11
- i2c2: U14 pins 25 & 26, has pcf8574 I/O extender
- spi2.0 or spi32766.0: U14 pins 27, 28, 29, 30; only a single
  chip-select is supported

GPIO edge detection (using interrupts) is only supported on a few of the
processor's pins: AP-EINT1(PG1), AP-EINT3(PB3), CSIPCK(PE0), and CSICK(PE1).
Edge detection is also supported on the XIO pins, but this feature is
rather limited due to the device and the driver (for example, the driver
interrupts on all edges).
