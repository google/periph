# periph - Peripherals I/O in Go

* [doc/users/](doc/users/) for ready-to-use tools.
* [doc/apps/](doc/apps/) to use `periph` as a library. The complete API
  documentation, including examples, is at
  [![GoDoc](https://godoc.org/github.com/google/periph?status.svg)](https://godoc.org/github.com/google/periph).
* [doc/drivers/](doc/drivers/) to expand the list of supported hardware.


## Users

periph includes [many ready-to-use tools](cmd/)! See [doc/users/](doc/users/) for
more info on configuring the host and using the included tools.

```bash
go get github.com/google/periph/cmd/...
periph-info
headers-list
```


## Application developers

For [application developers](doc/apps/), `periph` provides OS-independent bus
interfacing. The following gets the current temperature, barometric pressure and
relative humidity using a bme280:

```go
package main

import (
    "fmt"
    "log"

    "github.com/google/periph/devices"
    "github.com/google/periph/devices/bme280"
    "github.com/google/periph/host"
)

func main() {
    // Load all the drivers:
    if _, err := host.Init(); err != nil {
        log.Fatal(err)
    }

    // Open a handle to the first available I²C bus. It could be via a FT232H
    // over USB or an I²C bus exposed on the host's headers, it doesn't matter.
    bus, err := i2c.New(-1)
    if err != nil {
        log.Fatal(err)
    }
    defer bus.Close()

    // Open a handle to a bme280 connected on the I²C bus using default settings:
    dev, err := bme280.NewI2C(bus, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer dev.Close()

    // Read temperature from the sensor:
    var env devices.Environment
    if err = dev.Sense(&env); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%8s %10s %9s\n", env.Temperature, env.Pressure, env.Humidity)
}
```

See more examples at [doc/apps/SAMPLES.md](doc/apps/SAMPLES.md)!


## Contributions

`periph` provides an extensible driver registry and common bus interfaces which
are explained in more details at [doc/drivers/](doc/drivers/). `periph` is
designed to work well with drivers living in external repositories so you are
not _required_ to fork the periph repository to load out-of-tree drivers for
your platform.

We gladly accept contributions from device driver developers via GitHub pull
requests, as long as the author has signed the Google Contributor License.
Please see [doc/drivers/CONTRIBUTING.md](doc/drivers/CONTRIBUTING.md) for more
details.


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
     [rpi.P1_3](https://godoc.org/github.com/google/periph/host/rpi#pkg-variables),
     [bcm283x.GPIO2](https://godoc.org/github.com/google/periph/host/bcm283x#Pin)
     or
     [bcm283x.I2C1_SDA](https://godoc.org/github.com/google/periph/host/bcm283x#pkg-variables)
     to refer to the exact same pin when I²C bus #1 is enabled on a Raspberry
     Pi.
3. The user can chose to optimize for performance instead of usability.
   * e.g.
     [apa102.Dev](https://godoc.org/github.com/google/periph/devices/apa102#Dev)
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
   registry](https://godoc.org/github.com/google/periph#Register).
   * e.g. a user can inject a custom driver to expose more pins, headers, etc.
     A USB device (like an FT232H) can expose headers _in addition_ to the
     headers found on the host.
6. The drivers must use the fastest possible implementation.
   * e.g. both
     [allwinner](https://godoc.org/github.com/google/periph/host/allwinner)
     and
     [bcm283x](https://godoc.org/github.com/google/periph/host/bcm283x)
     leverage sysfs gpio to expose interrupt driven edge detection, yet use
     memory mapped GPIO registers to perform single-cycle reads and writes.


## Authors

The main author is [Marc-Antoine Ruel](https://github.com/maruel). The full list
is in [AUTHORS](AUTHORS) and [CONTRIBUTORS](CONTRIBUTORS).


## Disclaimer

This is not an official Google product (experimental or otherwise), it
is just code that happens to be owned by Google.

This project is not affiliated with the Go project.
