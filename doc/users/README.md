# periph - Users

Documentation for _users_ who want ready-to-use tools.


## Functionality included

[cmd/](../../cmd/) contains all the tools. Take a look first to see the included
functionality.


## Installing locally

The `periph` project doesn't release binaries at the moment, you are expected to
build from sources.


### Prerequisite

First, make sure to have Go installed. Get it from https://golang.org/dl/.

If you are running a Debian based distribution (Raspbian, Ubuntu, etc), you can
run `sudo apt-get install golang` to get the Go toolchain installed.


### Installation

It is as simple as:

```bash
go get -u github.com/google/periph/cmd/...
```

On many platforms, many tools requires running as root (via _sudo_) to have
access to the necessary CPU GPIO registers or even just kernel exposed APIs.


## Cross-compiling

To have faster builds, you may wish to build on a desktop and send the
executables to your ARM based micro computer (e.g.  Raspberry Pi).
[push.sh](https://github.com/google/periph/blob/master/cmd/push.sh) is included
to wrap this:

```bash
cd $GOPATH/src/github.com/google/periph/cmd
./push.sh raspberrypi bme280
```

It is basically a wrapper around `GOOS=linux GOARCH=arm go build .; scp <exe>
<host>:.`


## Configuring the host

More often than not on Debian based distros, you may have to run the executable
as root to be able to access the LEDs, GPIOs and other functionality.

It is possible to change the ACL on the _sysfs files_ via _udev_ rules. The
actual rules are OS specific.


### Debian

If you get `fatal error: libusb-1.0/libusb.h: No such file or directory`, run
`sudo apt-get install libusb-1.0.0-dev`
