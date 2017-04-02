# 'onewire-testboard' smoke test

Verifies that a 1-wire bus with two devices can be searched and that a DS18b20
temperature sensor as well as a ds2431 eeprom can be accessed. Typically used
with the periph-tester board.

Example output running on a C.H.I.P.:

```
chip4 ~> sudo ./periph-smoketest -v onewire-testboard -i2cbus 1
05:47:18.561821 Using drivers:
05:47:18.562864 - allwinner-gpio
05:47:18.563359 - chip
05:47:18.563783 - sysfs-gpio
05:47:18.564144 - sysfs-i2c
05:47:18.564491 - sysfs-led
05:47:18.564830 - sysfs-spi
05:47:18.565161 - sysfs-thermal
05:47:18.565470 Drivers skipped:
05:47:18.566169 - allwinner-gpio-pl: A64 CPU not detected
05:47:18.566265 - bcm283x-gpio: bcm283x CPU not detected
05:47:18.566336 - odroid-c1: Hardkernel ODROID-C0/C1/C1+ board not detected
05:47:18.566403 - pine64: dependency not loaded: "allwinner-gpio-pl"
05:47:18.566466 - rpi: dependency not loaded: "bcm283x-gpio"
05:47:18.569578 onewire-smoke: random number seed 1481694438569396253
05:47:18.655117 onewire-smoke: found 2 devices 0xaf000001318c0128 0x28000014f3f0c52d
05:47:18.920729 onewire-smoke: temperature is 28.50°C
05:47:18.921942 Test onewire-testboard successful
```
