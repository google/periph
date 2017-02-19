# periph - Samples

[README.md](README.md) contains general information for application developpers.
The complete API documentation, including examples, is at
[![GoDoc](https://godoc.org/periph.io/x/periph?status.svg)](https://godoc.org/periph.io/x/periph).

You are encouraged to look at tools in [cmd/](cmd/). These can be used as the
basis of your projects.

To try the following samples, put the code into a file named `sample.go` then
execute `go run sample.go`.


## Toggle a LED

_Purpose:_ Simplest example

`periph` doesn't expose any _toggle_-like functionality on purpose, it is as
stateless as possible.


```go
package main

import (
    "log"
    "time"

    "periph.io/x/periph/conn/gpio"
    "periph.io/x/periph/host"
    "periph.io/x/periph/host/rpi"
)

func main() {
    // Load all the drivers:
    if _, err := host.Init(); err != nil {
        log.Fatal(err)
    }

    l := gpio.Low
    for {
        // Lookup a pin by its location on the board:
        if err := rpi.P1_33.Out(l); err != nil {
            log.Fatal(err)
        }
        l = !l
        time.Sleep(500 * time.Millisecond)
    }
}
```


## IR (infra red remote)

_Purpose:_ display IR remote keys.

This sample uses lirc (http://www.lirc.org/). This assumes you installed lirc
and configured it. See
[devices/lirc](https://godoc.org/periph.io/x/periph/devices/lirc)
for more information.

```go
package main

import (
    "fmt"
    "log"

    "periph.io/x/periph/devices/lirc"
    "periph.io/x/periph/host"
)

func main() {
    // Load all the drivers:
    if _, err := host.Init(); err != nil {
        log.Fatal(err)
    }

    // Open a handle to lircd:
    conn, err := lirc.New()
    if err != nil {
        log.Fatal(err)
    }

    // Open a channel to receive IR messages and print them out as they are
    // received, skipping repeated messages:
    for msg := range conn.Channel() {
        if !msg.Repeat {
            fmt.Printf("%12s from %12s\n", msg.Key, msg.RemoteType)
        }
    }
}
```


## OLED 128x64 display

_Purpose:_ display an animated GIF.

This sample uses a
[ssd1306](https://godoc.org/periph.io/x/periph/devices/ssd1306).
The frames in the GIF are resized and centered first to reduce the CPU overhead.

```go
package main

import (
    "image"
    "image/draw"
    "image/gif"
    "log"
    "os"
    "time"

    "periph.io/x/periph/devices/ssd1306"
    "periph.io/x/periph/host"
    "github.com/nfnt/resize"
)

// convertAndResizeAndCenter takes an image, resizes and centers it on a
// image.Gray of size w*h.
func convertAndResizeAndCenter(w, h int, src image.Image) *image.Gray {
    src = resize.Thumbnail(uint(w), uint(h), src, resize.Bicubic)
    img := image.NewGray(image.Rect(0, 0, w, h))
    r := src.Bounds()
    r = r.Add(image.Point{(w - r.Max.X) / 2, (h - r.Max.Y) / 2})
    draw.Draw(img, r, src, image.Point{}, draw.Src)
    return img
}

func main() {
    // Load all the drivers:
    if _, err := host.Init(); err != nil {
        log.Fatal(err)
    }

    // Open a handle to the first available I²C bus:
    bus, err := i2c.New(-1)
    if err != nil {
        log.Fatal(err)
    }

    // Open a handle to a ssd1306 connected on the I²C bus:
    dev, err := ssd1306.NewI2C(bus, 128, 64, false)
    if err != nil {
        log.Fatal(err)
    }

    // Decodes an animated GIF as specified on the command line:
    if len(os.Args) != 2 {
        log.Fatal("please provide the path to an animated GIF")
    }
    f, err := os.Open(os.Args[1])
    if err != nil {
        log.Fatal(err)
    }
    g, err := gif.DecodeAll(f)
    f.Close()
    if err != nil {
        log.Fatal(err)
    }

    // Converts every frame to image.Gray and resize them:
    imgs := make([]*image.Gray, len(g.Image))
    for i := range g.Image {
        imgs[i] = convertAndResizeAndCenter(dev.W, dev.H, g.Image[i])
    }

    // Display the frames in a loop:
    for i := 0; ; i++ {
        index := i % len(imgs)
        c := time.After(time.Duration(10*g.Delay[index]) * time.Millisecond)
        img := imgs[index]
        dev.Draw(img.Bounds(), img, image.Point{})
        <-c
    }
}
```

## GPIO Edge detection

_Purpose:_ Signals when a button was pressed or a motion detector detected a
movement.

The
[gpio.PinIn.Edge()](https://godoc.org/periph.io/x/periph/conn/gpio#PinIn)
function permits a edge detection without a busy loop. This is useful for
**motion detectors**, **buttons** and other kinds of inputs where a busy loop
would burn CPU for no reason.

```go
package main

import (
    "fmt"
    "log"

    "periph.io/x/periph/host"
    "periph.io/x/periph/conn/gpio"
)

func main() {
    // Load all the drivers:
    if _, err := host.Init(); err != nil {
        log.Fatal(err)
    }

    // Lookup a pin by its number:
    p, err := gpio.ByNumber(16)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s: %s\n", p, p.Function())

    // Set it as input, with an internal pull down resistor:
    if err = p.In(gpio.Down, gpio.BothEdges); err != nil {
        log.Fatal(err)
    }

    // Wait for edges as detected by the hardware, and print the value read:
    for {
        p.WaitForEdge()
        fmt.Printf("-> %s\n", p.Read())
    }
}
```


## Measuring weather

_Purpose:_ gather temperature, pressure and relative humidity.

This sample uses a
[bme280](https://godoc.org/periph.io/x/periph/devices/bme280).

```go
package main

import (
    "fmt"
    "log"

    "periph.io/x/periph/devices"
    "periph.io/x/periph/devices/bme280"
    "periph.io/x/periph/host"
)

func main() {
    // Load all the drivers:
    if _, err := host.Init(); err != nil {
        log.Fatal(err)
    }

    // Open a handle to the first available I²C bus:
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
