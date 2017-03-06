# periph platforms

Periph supports the Linux sysfs drivers for gpio pins, I²C buses, SPI buses,
LEDs, and thermal sensors. This means that periph supports most **any Linux
platform**. It is tested to be compatible to Windows. In addition, periph has
special support for a small number of platforms. For some, such as Odroid-C1,
the additional support means that the headers, pins, and bus names are
predefined and that all changes are tested on the actual hardware. For others,
such as the various Raspberry Pi versions, CHIP, and Pine64, the additional
support means that in addition to the above high-speed memory-mapped I/O has
been implemented.

Current platforms:

- Raspberry Pi
- [NextThing C.H.I.P.](chip)
- Pine 64
- [Odoid-C1](odroid-c1/)
- Generic Linux
