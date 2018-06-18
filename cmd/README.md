# periph/cmd - read-to-use executables

This directory contains directly usable tools installable via:

```
go get periph.io/x/periph/cmd/...
```

Most of the tools can optionally leverage drivers in
[hostextra](https://periph.io/x/extra/hostextra) when the build tag
`periphextra` is defined:

```
go get -u -tags periphextra periph.io/x/periph/cmd/...
```

This permits taking advantage of drivers like FTDI's
[D2XX](https://periph.io/x/extra/hostextra/d2xx).


## Push

If you prefer to build on your workstation and push the binaries to the micro
computer, install `push` from [periph.io/x/bootstrap](
https://github.com/periph/bootstrap) to cross compile and efficiently push via
rsync:

```
go get -u periph.io/x/bootstrap/cmd/push
push -host pi@raspberrypi periph.io/x/periph/cmd/...
```

## Recommended first use

Try first `periph-info`. It will print out if any driver failed to run, for
example if you have to run as root to access certain drivers.

Then run `headers-list` to list all the headers on your board and confirm that
you get the expected output. If your board is missing, you can [contribute
it](https://periph.io/project/contributing/).


## Devices

- [apa102](apa102): Writes to a LED strip of APA-102 (sometimes called Dotstar).
  Can show an image animating on the Y axis.
- [bmxx80](bmxx80): Reads the temperature, pressure and humidity off a
  bmp180/bme280/bmp280. Humidity sensing is only supported on bme280.
- [ir](ir): Reads codes (button presses) on an InfraRed remote sensor.
- [led](led): Reads the state of on-board LEDs.
- [ssd1306](ssd1306): Writes text, an image or an animated GIF to an OLED
  display.
- [tm1637](tm1637): Writes to a segment digits display.


## Buses

- [gpio-list](gpio-list): Looking for the GPIO pins per functionality?
  Prints the state of each GPIO pin.
- [gpio-read](gpio-read): Read the input value of a GPIO pin and change
  input resistor.
- [gpio-write](gpio-write): Change the output value of a GPIO pin.
- [headers-list](headers-list): Pinrts the location of the pin on the header to
  connect your GPIO. This is the perfect tool to know where to connect the
  wires.
- [i2c-io](i2c-io): Reads and/or writes to an I²C device.
- [i2c-list](i2c-list): Lists which I²C buses are enabled and where the pins
  are.
- [spi-io](spi-io): Reads and/or writes to an SPI device.
- [spi-list](spi-list): Lists which SPI ports are enabled and where the pins
  are.


## Other

- [periph-info](periph-info): Lists which periph drivers loaded and which
  failed.
- [periph-smoketest](periph-smoketest): Runs one of the smoke test for the
  drivers. The smoke test differs from unit tests as they require real hardware
  to confirm that the driver being tested works.


## Troubleshooting

Having getting the tools to run? See [users/](https://groups.google.com/forum/#!forum/periph-users) for more
documentation.
