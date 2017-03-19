# gpio-test

Verifies that the library physically work. It requires the user to connect two
GPIO pins together and provide their pin number at the command line.

Example output running on a Raspberry Pi:

```
$ gpio-test 12 6
Using drivers:
  - bcm283x
  - rpi
  - sysfs-gpio
  - sysfs-spi
  - sysfs-i2c
Using pins and their current state:
- GPIO12: In/High
- GPIO6: In/High

Testing GPIO6 -> GPIO12
  Testing base functionality
    GPIO12.In(Float)
    GPIO6.Out(Low)
    -> GPIO12: In/Low
    -> GPIO6: Out/Low
    GPIO6.Out(High)
    -> GPIO12: In/High
    -> GPIO6: Out/High
  Testing edges
    GPIO12.Edges()
    GPIO6.Out(Low)
    Low <- GPIO12
    GPIO6.Out(High)
    High <- GPIO12
    GPIO6.Out(Low)
    Low <- GPIO12
    GPIO12.DisableEdges()
  Testing pull resistor
    GPIO6.In(Down)
    -> GPIO12: In/Low
    -> GPIO6: In/Low
    GPIO6.In(Up)
    -> GPIO12: In/High
    -> GPIO6: In/High
Testing GPIO12 -> GPIO6
  Testing base functionality
    GPIO6.In(Float)
    GPIO12.Out(Low)
    -> GPIO6: In/Low
    -> GPIO12: Out/Low
    GPIO12.Out(High)
    -> GPIO6: In/High
    -> GPIO12: Out/High
  Testing edges
    GPIO6.Edges()
    GPIO12.Out(Low)
    Low <- GPIO6
    GPIO12.Out(High)
    High <- GPIO6
    GPIO12.Out(Low)
    Low <- GPIO6
    GPIO6.DisableEdges()
  Testing pull resistor
    GPIO12.In(Down)
    -> GPIO6: In/Low
    -> GPIO12: In/Low
    GPIO12.In(Up)
    -> GPIO6: In/High
    -> GPIO12: In/High
```
