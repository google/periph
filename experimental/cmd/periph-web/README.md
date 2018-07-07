# periph-web

Web UI and JSON API server for [periph.io](https://periph.io).

- The web UI doesn't depend on external web resources, so it can be used on
  networks without internet access!
- The Go code doesn't depend on any external library besides the Go standard
  library.

Try it now:

```
go get -u periph.io/x/periph/experimental/cmd/periph-web
periph-web
```


# API

periph-web exposes a [JSON API](webapi.go) to enable the web UI:

Getters (read only, no side effect):

- `/api/periph/v1/gpio/aliases`: returns GPIO aliases in
  [gpioreg](https://periph.io/x/periph/conn/gpio/gpioreg).
- `/api/periph/v1/gpio/list`: returns all registered GPIOs in
  [gpioreg](https://periph.io/x/periph/conn/gpio/gpioreg), even the ones not on
  board headers ([pinreg](https://periph.io/x/periph/conn/pin/pinreg)).
- `/api/periph/v1/header/list`: returns all registered headers in
  [pinreg](https://periph.io/x/periph/conn/pin/pinreg).
- `/api/periph/v1/i2c/list`: returns all registered I²C buses in
  [i2creg](https://periph.io/x/periph/conn/i2c/i2creg).
- `/api/periph/v1/spi/list`: returns all registered SPI ports in
  [spireg](https://periph.io/x/periph/conn/spi/spireg).
- `/api/periph/v1/server/state`: returns the loaded periph drivers.

Actions:

- `/api/periph/v1/gpio/read`: accepts a list of GPIO names and return their
  [gpio.Level](https://periph.io/x/periph/conn/gpio) as 0 or 1.
- `/api/periph/v1/gpio/out`: sets the output of GPIOs specified as a dict of
  {gpio pin name: level at 0 or 1}.
- `/raw/periph/v1/xsrf_token`: returns a fresh XSRF token as a raw string. This
  is not a JSON API.


## Using with curl

The API is protected via a XSRF token. It is valid for 24 hours and needs
to be refreshed. Request without a token will be denied with 401.

Here's how to access the HTTP JSON API via bash:

```
export TARGET_HOST=raspberrypi:7080
export XSRF_TOKEN="$(curl -s -X POST http://$TARGET_HOST/raw/periph/v1/xsrf_token)"
curl -s -b "XSRF-TOKEN=$XSRF_TOKEN" -d '{}' -H Content-Type:application/json http://$TARGET_HOST/api/periph/v1/header/list | python -mjson.tool
```

Read the values of GPIO5 and GPIO6:

```
curl -s -b "XSRF-TOKEN=$XSRF_TOKEN" -d '["GPIO5","GPIO6"]' -H Content-Type:application/json http://$TARGET_HOST/api/periph/v1/gpio/read
```

Set the GPIO6 as output to High:

```
curl -s -b "XSRF-TOKEN=$XSRF_TOKEN" -d '{"GPIO6":1}' -H Content-Type:application/json http://$TARGET_HOST/api/periph/v1/gpio/out
```


By default, the HTTP server binds to localhost. If you want to access it from
another host, pass the argument `-http=0.0.0.0:7080` or the port of your
choosing.


## Extended support

If you want to play with a [FTDI FT232H/FT232R](https://periph.io/device/ftdi/),
you have to build with [periph.io/x/extra](https://periph.io/x/extra) built in.
You can do with:

```
go get -u -tags periphextra periph.io/x/periph/experimental/cmd/periph-web
```

Cross-compiling won't work with extra.
