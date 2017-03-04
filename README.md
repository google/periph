# periph - Peripherals I/O in Go


## Features

* Continuously tested on [Raspberry Pi](https://raspberrypi.org),
  [C.H.I.P.](https://getchip.com/) and Windows 10 via
  [gohci](https://github.com/periph/gohci).
* Interfaces: I²C, SPI, gpio (both low latency memory mapped registers and
  zero-CPU edge detection), 1-wire and more.
* Devices: apa102, bme280, ds18b20, ssd1306, tm1637, and more coming.


## Quick links

* [doc/users/](doc/users/) for ready-to-use tools.
* [doc/apps/](doc/apps/) to use `periph` as a library. The complete API
  documentation, including examples, is at
  [![GoDoc](https://godoc.org/periph.io/x/periph?status.svg)](https://godoc.org/periph.io/x/periph).
* [doc/drivers/](doc/drivers/) to expand the list of supported hardware.


## Users

periph includes [many ready-to-use tools](cmd/)! See [doc/users/](doc/users/)
for more info on configuring the host and using the included tools.

```bash
# Retrieve and install all the commands at once:
go get periph.io/x/periph/cmd/...
# List the host drivers registered and/or initialized:
periph-info
# List the known headers:
headers-list
# List the known GPIO state:
gpio-list
```


## Application developers

For [application developers](doc/apps/), `periph` provides OS-independent bus
interfacing. It really tries hard to _get out of the way_. Here's the canonical
"toggle a LED" sample:


```go
package main

import (
    "time"
    "periph.io/x/periph/conn/gpio"
    "periph.io/x/periph/host"
)

func main() {
    host.Init()
    for l := gpio.Low; ; l = !l {
        gpio.ByNumber(13).Out(l)
        time.Sleep(500 * time.Millisecond)
    }
}
```

The following are synonyms, use the form you prefer:

* Runtime discovery:
  * [`gpio.ByNumber(13)`](https://godoc.org/periph.io/x/periph/conn/gpio/#ByNumber) or [`gpio.ByName("13")`](https://godoc.org/periph.io/x/periph/conn/gpio/#ByName)
  *  [`gpio.ByName("GPIO13")`](https://godoc.org/periph.io/x/periph/conn/gpio/#ByName)
* Using global variables:
  * [`rpi.P1_33`](https://godoc.org/periph.io/x/periph/host/rpi#/P1_33) to select the pin via its position on the board
  * [`bcm283x.GPIO13`](https://godoc.org/periph.io/x/periph/host/bcm283x/#GPIO13)

This example uses basically no CPU: the `Out()` call doesn't call into the
kernel. Instead it directly changes the GPIO memory mapped register.


### Samples

See [doc/apps/SAMPLES.md](doc/apps/SAMPLES.md) for more examples.


## Contributions

`periph` provides an extensible driver registry and common bus interfaces which
are explained in more details at [doc/drivers/](doc/drivers/). `periph` is
designed to work well with drivers living in external repositories so you are
not _required_ to fork the periph repository to load out-of-tree drivers for
your platform.

**Every commit is [tested on real hardware](doc/drivers/CONTRIBUTING.md#testing)
via [gohci](https://github.com/periph/gohci) workers.**

We gladly accept contributions for documentation improvements and from device
driver developers via GitHub pull requests, as long as the author has signed the
Google Contributor License. Please see
[doc/drivers/CONTRIBUTING.md](doc/drivers/CONTRIBUTING.md) for more details.


## Philosophy

1. Optimize for simplicity, correctness and usability in that order.
   * e.g. everything, interfaces and structs, uses strict typing, there's no
     `interface{}` in sight.
2. OS agnostic. Clear separation of interfaces in [conn/](conn/),
   enablers in [host/](host) and device drivers in [devices/](devices/).
   * e.g. no devfs or sysfs path in sight.
   * e.g. conditional compilation enables only the relevant drivers to be loaded
     on each platform.
3. ... yet doesn't get in the way of platform specific code.
   * e.g. A user can use statically typed global variables
     [rpi.P1_3](https://godoc.org/periph.io/x/periph/host/rpi#pkg-variables),
     [bcm283x.GPIO2](https://godoc.org/periph.io/x/periph/host/bcm283x#Pin)
     to refer to the exact same pin on a Raspberry Pi.
3. The user can chose to optimize for performance instead of usability.
   * e.g.
     [apa102.Dev](https://godoc.org/periph.io/x/periph/devices/apa102#Dev)
     exposes both high level
     [draw.Image](https://golang.org/pkg/image/draw/#Image) to draw an image and
     low level [io.Writer](https://golang.org/pkg/io/#Writer) to write raw RGB
     24 bits pixels. The user chooses.
4. Use a divide and conquer approach. Each component has exactly one
   responsibility.
   * e.g. instead of having a driver per "platform", there's a driver per
     "component": one for the CPU, one for the board headers, one for each
     bus and sensor, etc.
5. Extensible via a [driver
   registry](https://godoc.org/periph.io/x/periph#Register).
   * e.g. a user can inject a custom driver to expose more pins, headers, etc.
     A USB device (like an FT232H) can expose headers _in addition_ to the
     headers found on the host.
6. The drivers must use the fastest possible implementation.
   * e.g. both
     [allwinner](https://godoc.org/periph.io/x/periph/host/allwinner)
     and
     [bcm283x](https://godoc.org/periph.io/x/periph/host/bcm283x)
     leverage sysfs gpio to expose interrupt driven edge detection, yet use
     memory mapped GPIO registers to perform single-cycle reads and writes.


## Authors

`periph` was initiated by [Marc-Antoine Ruel](https://github.com/maruel). The
full list of contributors is in [AUTHORS](AUTHORS) and
[CONTRIBUTORS](CONTRIBUTORS).


## Disclaimer

This is not an official Google product (experimental or otherwise), it
is just code that happens to be owned by Google.

This project is not affiliated with the Go project.
