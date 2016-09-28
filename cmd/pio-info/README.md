# pio-info

Prints the lists of drivers that were loaded, the ones skipped and the one that
failed to load, if any.

* Looking for the GPIO pins per functionality? Look at
  [gpio-list](../gpio-list).
* Looking for the location of the pin on the header to connect your GPIO? Look
  at [headers-list](../headers-list).


## Example

On a [Raspberry Pi](https://www.raspberrypi.org/) running
[Raspbian](https://raspbian.org/):

    $ pio-info
    Drivers loaded and their dependencies, if any:
    - bcm283x
    - rpi          : [bcm283x]
    - sysfs-gpio
    - sysfs-i2c
    - sysfs-led
    - sysfs-spi
    - sysfs-thermal
    Drivers skipped and the reason why:
    - allwinner   : Allwinner CPU not detected
    - allwinner_pl: dependency not loaded: "allwinner"
    - pine64      : dependency not loaded: "allwinner_pl"
    Drivers failed to load and the error:
      <none>

On a [Pine64](https://www.pine64.org/) running [Armbian](http://armbian.com)
running **as a user** (not root):

    $ pio-info
    Drivers loaded and their dependencies, if any:
    - pine64       : [allwinner_pl]
    - sysfs-i2c
    - sysfs-thermal
    Drivers skipped and the reason why:
    - bcm283x  : bcm283x CPU not detected
    - rpi      : dependency not loaded: "bcm283x"
    - sysfs-led: no LED found
    - sysfs-spi: no SPI bus found
    Drivers failed to load and the error:
    - allwinner   : need more access, try as root: open /dev/mem: permission denied
    - allwinner_pl: need more access, try as root: open /dev/mem: permission denied
    - sysfs-gpio  : need more access, try as root or setup udev rules: open /sys/class/gpio/export: permission denied

On a [Pine64](https://www.pine64.org/) running [Armbian](http://armbian.com) **as
root**:

    $ sudo pio-info
    Drivers loaded and their dependencies, if any:
    - allwinner
    - allwinner_pl : [allwinner]
    - pine64       : [allwinner_pl]
    - sysfs-gpio
    - sysfs-i2c
    - sysfs-thermal
    Drivers skipped and the reason why:
    - bcm283x  : bcm283x CPU not detected
    - rpi      : dependency not loaded: "bcm283x"
    - sysfs-led: no LED found
    - sysfs-spi: no SPI bus found
    Drivers failed to load and the error:
      <none>

On some platforms, more driver can be loaded when running as root, improving
performance and adding some features, like input pull resistor support.
