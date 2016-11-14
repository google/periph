# periph - Application developers

Documentation for _application developers_ who want to write Go applications
leveraging `periph`.

The complete API documentation, including examples, is at
[![GoDoc](https://godoc.org/github.com/google/periph?status.svg)](https://godoc.org/github.com/google/periph).


## Introduction

`periph` uses a driver registry to efficiently load the relevant drivers on the
host it is running on. It differentiates between drivers that _enable_
functionality on the host and drivers for devices connected _to_ the host.

Most micro computers expose at least some of the following:
[I²C bus](https://godoc.org/github.com/google/periph/conn/i2c#Bus),
[SPI bus](https://godoc.org/github.com/google/periph/conn/spi#Conn),
[gpio
pins](https://godoc.org/github.com/google/periph/conn/gpio#PinIO),
[analog
pins](https://godoc.org/github.com/google/periph/conn/analog),
[UART](https://godoc.org/github.com/google/periph/conn/uart), I2S
and PWM.

* The interfaces are defined in [conn/](../../conn/).
* The concrete objects _implementing_ the interfaces are in
  [host/](../../host/).
* The device drivers _using_ these interfaces are located in
  [devices/](../../devices/).

A device can be connected on a bus, let's say an LED strip connected over SPI.
In this case the application needs to obtain a handle to the SPI bus and then
connect the LED device driver to the SPI bus handle.


## Project state

The library is **not stable** yet and breaking changes continuously happen.
Please version the libary using [one of the go vendoring
tools](https://github.com/golang/go/wiki/PackageManagementTools) and sync
frequently.


## Initialization

The function to initialize the drivers registered by default is
[host.Init()](https://godoc.org/github.com/google/periph/host#Init). It
returns a
[periph.State](https://godoc.org/github.com/google/periph#State):

```go
state, err := host.Init()
```

[periph.State](https://godoc.org/github.com/google/periph#State) contains
information about:

* The drivers loaded and active.
* The drivers skipped, because the relevant hardware wasn't found.
* The drivers that failed to load due to an error. The app may still run without
  these drivers.

In addition,
[host.Init()](https://godoc.org/github.com/google/periph/host#Init) may
return an error when there's a structural issue, for example two drivers with
the same name were registered. This is a fatal failure. The package
[host](https://godoc.org/github.com/google/periph/host) registers all the
drivers under [host/](../../host/).


## Connection

A connection
[conn.Conn](https://godoc.org/github.com/google/periph/conn#Conn)
is a **point-to-point** connection between the host and a device where the
application is the master driving the I/O.

A `Conn` can be multiplexed over the underlying bus. For example an I²C bus
[i2c.Bus](https://godoc.org/github.com/google/periph/conn/i2c#Bus) may have
multiple connections (slaves) to the master, each addressed by the device
address.


### SPI connection

An
[spi.Conn](https://godoc.org/github.com/google/periph/conn/spi#Conn)
**is** a
[conn.Conn](https://godoc.org/github.com/google/periph/conn#Conn).


### I²C connection

An [i2c.Bus](https://godoc.org/github.com/google/periph/conn/i2c#Bus) is **not**
a [conn.Conn](https://godoc.org/github.com/google/periph/conn#Conn).
This is because an I²C bus is **not** a point-to-point connection but instead is
a real bus where multiple devices can be connected simultaneously, like a USB
bus. To create a point-to-point connection to a device which does implement
[conn.Conn](https://godoc.org/github.com/google/periph/conn#Conn) use
[i2c.Dev](https://godoc.org/github.com/google/periph/conn/i2c#Dev), which embeds
the device's address:

```go
// Open the first available I²C bus:
bus, _ := i2c.New(-1)
// Address the device with address 0x76 on the I²C bus:
dev := i2c.Dev{bus, 0x76}
// This is now a point-to-point connection and implements conn.Conn:
var _ conn.Conn = &dev
```

Since many devices have their address hardcoded, it's up to the device driver to
specify the address.


### GPIO

[gpio pins](https://godoc.org/github.com/google/periph/conn/gpio#PinIO)
can be leveraged for arbitrary uses, such as buttons, LEDs, relays, etc. 


## Samples

See [SAMPLES.md](SAMPLES.md) for various examples.


## Using out-of-tree drivers

Out of tree drivers can be seamlessly loaded by an application. The example
below shows a hypothetical driver located at github.com/example/virtual_i2c that
exposes an I²C bus over a REST API to a remote device.

This driver can be used with periph as if it were built into periph as follows:

```go
package main

import (
    "log"

    "github.com/example/virtual_i2c"
    "github.com/google/periph"
    "github.com/google/periph/host"
    "github.com/google/periph/conn/i2c"
)

type driver struct{}

func (d *driver) String() string          { return "virtual_i2c" }
func (d *driver) Type() periph.Type       { return periph.Bus }
func (d *driver) Prerequisites() []string { return nil }

func (d *driver) Init() (bool, error) {
    // Load the driver. Note that drivers are loaded *concurrently* by periph.
    if err := virtual_i2c.Load(); err != nil {
        return true, err
    }
    err := i2c.Register("foo", 10, func() (i2c.BusCloser, error) {
        // You may have to create a struct to convert the API:
        return virtual_i2c.Open()
    })
    // If this Init() function returns an error, it will be in the state
    // returned by host.Init():
    return true, err
}

func main() {
    // Register your driver in the registry:
    if _, err := drivers.Register(driver); err != nil {
        log.Fatal(err)
    }
    // Initialize normally. Your driver will be loaded:
    if _, err := host.Init(); err != nil {
        log.Fatal(err)
    }

    bus, err := i2c.New(10)
    if err != nil {
        log.Fatal(err)
    }
    defer bus.Close()

    // Use your bus driver like if it had been provided by periph.
}
```
