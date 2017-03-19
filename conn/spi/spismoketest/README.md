# spi-test

Verifies that a EEPROM can be accessed on the bus. Typically used with
the periph-tester board.

Requires a gpio pin to be tied to the EEPROM's write protect (active low write
protect).

Example output running on an Odroid-C1:

```
# ./periph-smoketest -v spi-testboard -wp 83
20:12:14.447484 Using drivers:
20:12:14.447557 - sysfs-led
20:12:14.447633 - sysfs-thermal
20:12:14.447718 - sysfs-spi
20:12:14.447792 - sysfs-i2c
20:12:14.447862 - sysfs-gpio
20:12:14.447913 Drivers skipped:
20:12:14.447982 - allwinner-gpio: Allwinner CPU not detected
20:12:14.448052 - bcm283x-gpio: bcm283x CPU not detected
20:12:14.448121 - allwinner-gpio-pl: dependency not loaded: "allwinner-gpio"
20:12:14.448191 - chip: dependency not loaded: "allwinner-gpio"
20:12:14.448273 - rpi: dependency not loaded: "bcm283x-gpio"
20:12:14.448342 - pine64: dependency not loaded: "allwinner-gpio-pl"
20:12:14.448771 spi-smoke: random number seed 1479960734448658003
20:12:14.454665 spi-smoke writing&reading EEPROM byte 0xee
20:12:14.468726 spi-smoke writing&reading EEPROM page 0x00c0
```
