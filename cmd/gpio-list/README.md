# gpio-list

Prints GPIO pins either by pin number or by functionality (e.g. configured as
I²C or SPI pin).

* Looking for the location of the pin on the header to connect your GPIO? Look
  at [headers-list](../headers-list).
* Looking for periph drivers loaded? Look at [periph-info](../periph-info).


## Examples

* Use `gpio-list -help` for help
* Use `-n` to print pins that are not connected or in INVALID state
* Use `-a` to print everything at once

The followings were captured on a Raspberry Pi 3 with I2C1, SPI0 and SPI1
enabled, lirc (IR) enabled and Bluetooth disabled with the following in
`/boot/config.txt`:

    dtparam=i2c_arm=on
    dtparam=spi=on
    dtoverlay=lirc-rpi,gpio_out_pin=5,gpio_in_pin=6,gpio_in_pull=high
    dtoverlay=spi1-1cs
    dtoverlay=pi3-disable-bt

then running:

    sudo systemctl disable hciuart

For more information for enabling functional pins, see
[![GoDoc](https://godoc.org/github.com/google/periph/host/rpi?status.svg)](https://godoc.org/github.com/google/periph/host/rpi).


### Aliases

When possible, aliases are created per functionality. Print the GPIO aliases
with:

    $ gpio-list -l
    GPCLK1   : GPIO42
    GPCLK2   : GPIO43
    I2C1_SCL : GPIO3
    I2C1_SDA : GPIO2
    PWM0_OUT : GPIO40
    PWM1_OUT : GPIO41
    SPI0_CLK : GPIO11
    SPI0_MISO: GPIO9
    SPI0_MOSI: GPIO10
    SPI1_CLK : GPIO21
    SPI1_MISO: GPIO19
    SPI1_MOSI: GPIO20
    UART0_RXD: GPIO15
    UART0_TXD: GPIO14


### GPIO

Print the GPIO pins per number:

    $ gpio-list -g
    GPIO0 : In/High
    GPIO1 : In/High
    GPIO2 : I2C1_SDA
    GPIO3 : I2C1_SCL
    GPIO4 : In/High
    GPIO5 : Out/Low
    GPIO6 : In/High
    GPIO7 : Out/High
    GPIO8 : Out/High
    GPIO9 : SPI0_MISO
    GPIO10: SPI0_MOSI
    GPIO11: SPI0_CLK
    GPIO12: In/Low
    GPIO13: In/High
    GPIO14: UART0_TXD
    GPIO15: UART0_RXD
    GPIO16: In/Low
    GPIO17: In/Low
    GPIO18: Out/High
    GPIO19: SPI1_MISO
    GPIO20: SPI1_MOSI
    GPIO21: SPI1_CLK
    GPIO22: In/Low
    GPIO23: In/Low
    GPIO24: In/Low
    GPIO25: In/Low
    GPIO26: In/Low
    GPIO27: In/Low
    GPIO40: PWM0_OUT
    GPIO41: PWM1_OUT
    GPIO46: In/High
