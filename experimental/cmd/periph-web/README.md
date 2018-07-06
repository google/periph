# periph-web

Simple webcomponents based self-contained web UI for periph.


## Using with curl

The API is protected via a simple XSRF token. It is valid for 24 hours and needs
to be refreshed. Request without a token will be denied with 401.

Here's how to access the HTTP API via bash:

```
export TARGET_HOST=raspberrypi:7080
export XSRF_TOKEN="$(curl -s -X POST http://${TARGET_HOST}/api/periph/v1/xsrf_token/raw)"
curl -s -b "XSRF-TOKEN=${XSRF_TOKEN}" -d '{}' -H 'Content-Type: application/json' http://${TARGET_HOST}/api/periph/v1/header/_all | python -mjson.tool
```

By default, the HTTP server binds to localhost. If you want to access it from
another host, pass the argument `-http=0.0.0.0:7080` or the port of your
choosing.
