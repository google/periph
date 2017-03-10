# periph - Peripherals I/O in Go

Documentation is at https://periph.io


## Sample

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

Visit https://periph.io/doc/apps/samples/ for more samples.


## Authors

`periph` was initiated with ❤️️ and passion by [Marc-Antoine
Ruel](https://github.com/maruel). The full list of contributors is in
[AUTHORS](https://github.com/google/periph/blob/master/AUTHORS) and
[CONTRIBUTORS](https://github.com/google/periph/blob/master/CONTRIBUTORS).


## Disclaimer

This is not an official Google product (experimental or otherwise), it
is just code that happens to be owned by Google.

This project is not affiliated with the Go project.
