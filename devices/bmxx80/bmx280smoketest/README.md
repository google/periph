# 'bmx280' smoke test

Verifies that two BME280/BMP280, one over I²C, one over SPI, can read roughly
the same temperature, humidity and pressure.

It can also be leveraged to record the I/O to write playback unit tests. It is a
good example to reuse to write other device driver unit test.
