# periph - Peripherals I/O in Go

[![mascot](https://raw.githubusercontent.com/periph/website/master/site/static/img/periph-mascot-280.png)](https://periph.io/)

Documentation is at https://periph.io

[![GoDoc](https://godoc.org/periph.io/x/periph?status.svg)](https://godoc.org/periph.io/x/periph)
[![Go Report Card](https://goreportcard.com/badge/periph.io/x/periph)](https://goreportcard.com/report/periph.io/x/periph)
[![Coverage Status](https://codecov.io/gh/google/periph/graph/badge.svg)](https://codecov.io/gh/google/periph)
[![Build Status](https://travis-ci.org/google/periph.svg)](https://travis-ci.org/google/periph)
[![Gitter chat](https://badges.gitter.im/google/periph.png)](https://gitter.im/periph-io/Lobby)


## Example

~~~go
package main

import (
    "time"
    "periph.io/x/periph/conn/gpio"
    "periph.io/x/periph/host"
    "periph.io/x/periph/host/rpi"
)

func main() {
    host.Init()
    for l := gpio.Low; ; l = !l {
        rpi.P1_33.Out(l)
        time.Sleep(500 * time.Millisecond)
    }
}
~~~

Curious? Look at [supported devices](https://periph.io/device/) for more
examples!


## Authors

`periph` was initiated with ❤️️ and passion by [Marc-Antoine
Ruel](https://github.com/maruel). The full list of contributors is in
[AUTHORS](https://github.com/google/periph/blob/master/AUTHORS) and
[CONTRIBUTORS](https://github.com/google/periph/blob/master/CONTRIBUTORS).


## Disclaimer

This is not an official Google product (experimental or otherwise), it
is just code that happens to be owned by Google.

This project is not affiliated with the Go project.
