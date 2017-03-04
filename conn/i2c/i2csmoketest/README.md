# i2c-test

Verifies that a DS2483 and an EEPROM can be accessed on the bus. Typically used
with the periph-tester board.

Requires a gpio pin to be tied to the EEPROM's write control (active low write
enable).

Sample output running on an Odroid-C1:

```
# ./periph-smoketest -v i2c-testboard -wc 83
20:29:14.241551 Using drivers:
20:29:14.241964 - sysfs-led
20:29:14.242056 - sysfs-thermal
20:29:14.242149 - sysfs-i2c
20:29:14.242264 - sysfs-gpio
20:29:14.242362 - sysfs-spi
20:29:14.242434 - odroid_c1
20:29:14.242511 Drivers skipped:
20:29:14.242591 - bcm283x-gpio: bcm283x CPU not detected
20:29:14.242685 - allwinner-gpio: Allwinner CPU not detected
20:29:14.242765 - chip: dependency not loaded: "allwinner-gpio"
20:29:14.242835 - rpi: dependency not loaded: "bcm283x-gpio"
20:29:14.242909 - allwinner-gpio-pl: dependency not loaded: "allwinner-gpio"
20:29:14.242988 - pine64: dependency not loaded: "allwinner-gpio-pl"
20:29:14.243421 i2c-smoke: random number seed 1479961754243306323
20:29:14.249073 i2c-smoke writing&reading EEPROM byte 0xcb
20:29:14.289290 i2c-smoke writing&reading EEPROM page 0x60
20:29:14.324772 Test i2c-testboard successful
```
