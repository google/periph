# headers-list

Prints physical headers pins and the associated functionality of each pin.

* Looking for the GPIO pins per functionality? Look at
  [gpio-list](../gpio-list).
* Looking for periph drivers loaded? Look at [periph-info](../periph-info).


## Example

Print the pins per their hardware location on the headers. This uses an
internal lookup table then query each pin. Here's an example on a host with two
SPI host and lirc enabled:

    $ headers-list
    AUDIO: 2 pins
      Pos  Name    Func
      1    GPIO41  PWM1_OUT
      2    GPIO40  PWM0_OUT

    HDMI: 1 pins
      Pos  Name    Func
      1    GPIO46  In/High

    P1: 40 pins
           Func    Name  Pos  Pos  Name   Func
                   V3_3    1  2    V5
       I2C1_SDA   GPIO2    3  4    V5
       I2C1_SCL   GPIO3    5  6    GROUND
        In/High   GPIO4    7  8    GPIO14 UART0_TXD
                 GROUND    9  10   GPIO15 UART0_RXD
         In/Low  GPIO17   11  12   GPIO18 Out/High
         In/Low  GPIO27   13  14   GROUND
         In/Low  GPIO22   15  16   GPIO23 In/Low
                   V3_3   17  18   GPIO24 In/Low
      SPI0_MOSI  GPIO10   19  20   GROUND
      SPI0_MISO   GPIO9   21  22   GPIO25 In/Low
       SPI0_CLK  GPIO11   23  24   GPIO8  Out/High
                 GROUND   25  26   GPIO7  Out/High
        In/High   GPIO0   27  28   GPIO1  In/High
        Out/Low   GPIO5   29  30   GROUND
        In/High   GPIO6   31  32   GPIO12 In/Low
        In/High  GPIO13   33  34   GROUND
      SPI1_MISO  GPIO19   35  36   GPIO16 In/Low
         In/Low  GPIO26   37  38   GPIO20 SPI1_MOSI
                 GROUND   39  40   GPIO21 SPI1_CLK
