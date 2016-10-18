# pio - Device driver developpers

Documentation for _device driver developers_ who either wants to developper a
device driver in their own code base or want to submit a contribution to extend
the supported hardware.


## Background

The main purpose of `pio` it to provide interfaces to assemble components
together to communicate with hardware peripherals. As such, it splits boards
into its individual components: CPU, buses, physical headers, etc, instead of
thinking of each board as a whole object.

Read more about the goals at [GOALS.md](GOALS.md).

Read more about detailed design at [DESIGN.md](DESIGN.md).


### Classes of hardware

This document distinguishes two classes of drivers:

* Enablers: they are what make the interconnects work, so that you can then
  use real stuff. That's both point-to-point connections (GPIO, UART, TCP) and
  buses (I²C, SPI, BT) where individual devices can be addressed.  They enable
  you to do something but are not the essence of what you want to do.
* Devices: they are the end goal, to do something functional. There are multiple
  subclasses of devices like sensors, output devices, etc.


## Driver lifetime management

Proper driver lifetime management is key to the success of this project. There
must be clear expectations to add, update and remove drivers for the core
project. As described in the [Risks section in GOALS.md](GOALS.md#risk) below,
poor drivers or high churn rate will destroy the value proposition.

This is critical as drivers can be silently broken by seemingly innocuous
changes. Because the testing story of hardware is significantly harder than
software-only projects, there’s an inherent faith in the quality of the code
that must be asserted.


### Experimental

Any driver can be requested to be added to the library under
[experimental/](../../experimental/) directory. The following process must be
followed:
* One or multiple developers have created a driver out of tree.
* The driver is deemed to work.
* The driver meets minimal quality bar under the promise of being improved. See
  [Requirements](#requirements) for the extensive list.
* Follow [CONTRIBUTING.md](CONTRIBUTING.md) demands.
* Create a Pull Request for integration under
  [experimental/](../../experimental/) and respond to the code review.

At this point, it is available for use to everyone but is not loaded defacto by
[host.Init()](https://godoc.org/github.com/google/pio/host#Init).

There is no API compatibility guarantee for drivers under
[experimental/](../../experimental/).


### Stable

A driver in [experimental/](../../experimental/) can be promoted to stable in
either [devices/](../../devices/) or [host/](../../host/) as relevant. The
following process must be followed:
* Declare at least one (or multiple) owners that are responsive to reply to
  feature requests and bug reports.
  * There could be a threshold, > _TO BE DETERMINED_ lines, where more than one
    owner is required.
  * Contributors commit to support the driver for the foreseeable future and
    **promptly** do code reviews to keep the driver quality to the expected
    standard.
* There are multiple reports that the driver is functioning as expected.
* If another driver exists for an intersecting class of devices, the other
  driver must enter deprecation phase.
* At this point the driver must maintain its API compatibility promise.


### Deprecation

A driver can be subsumed by a newer driver with a better core implementation or
a new breaking API. The previous driver must be deprecated, moved back to
[experimental/](../../experimental/) and announced to be deleted after _TO BE
DETERMINED_ amount of time.


### Contributing a new driver

A new proposed driver must be first implemented out of tree and fit all the
items in [Requirements](#requirements) listed below. First propose it as
[Experimental](#experimental), then ask to promote it to [Stable](#stable).


## Requirements

All the code must fit the following requirements.

**Fear not!** We know the list _is_ daunting but as you create your pull request
to add something at [experimental/](../../experimental/), we'll happily guide
you in the process to help improve the code to meet the expected standard. The
end goal is to write *high quality maintainable code* and use this as a learning
experience.

* The code must be Go idiomatic.
  * Constructor `NewXXX()` returns an object of concrete type.
  * Functions accept interfaces.
  * Leverage standard interfaces like
    [io.Writer](https://golang.org/pkg/io/#Writer) and
    [image.Image](https://golang.org/pkg/image/#Image) where possible.
  * No `interface{}` unless strictly required.
  * Minimal use of factories except for protocol level registries.
  * No `init()` code that accesses peripherals on process startup. These belongs
    to
    [Driver.Init()](https://godoc.org/github.com/google/pio#Driver).
* Exact naming
  * Driver for a chipset must have the name of the chipset or the chipset
    family. Don't use `oleddisplay`, use `ssd1306`.
  * Driver must use the real chip name, not a marketing name by a third party.
    Don't use `dotstar` (as marketed by Adafruit), use `apa102` (as created
    by APA Electronic co. LTD.).
  * A link to the datasheet must be included in the package doc unless NDA'ed
    or inaccessible.
* Testability
  * Code must be testable and tested without a device.
  * When relevant, include a smoke test under [tests/](../../tests/). The smoke
    test tests a real device to confirm the driver physically works for devices.
* Usability
  * Provide a standalone executable in [cmd/](../../cmd/) to expose the
    functionality.  It is acceptable to only expose a small subset of the
    functionality but _the tool must have purpose_.
  * Provide a `func Example()` along your test to describe basic usage of your
    driver. See the official [testing
    package](https://golang.org/pkg/testing/#hdr-Examples) for more details.
* Performance
  * Drivers controling an output device must have a fast path that can be used
    to directly write in the device's native format, e.g.
    [io.Writer](https://golang.org/pkg/io/#Writer).
  * Drivers controling an output device must have a generic path accepting
    higher level interface when found in the stdlib, e.g.
    [image.Image](https://golang.org/pkg/image/#Image).
  * Floating point arithmetic should only be used when absolutely necesary in
    the driver code. Most of the cases can be replaced with fixed point
    arithmetic, for example
    [devices.Milli](https://godoc.org/github.com/google/pio/devices#Milli).
    Floating point arithmetic is acceptable in the unit tests and tools in
    [cmd/](../../cmd/) but should not be abused.
  * Drivers must be implemented with performance in mind. For example I²C
    operations should be batched to minimize overhead.
  * Benchmark must be implemented for non trivial processing running on the host.
* Code must compile on all OSes, with minimal use of OS-specific thunk as
  strictly needed.
* Struct implementing an interface must validate at compile time with `var _
  <Interface> = &<Type>{}`.
* License is Apache v2.0.


## Code style

* The code tries to follow Go code style as described at
  https://github.com/golang/go/wiki/CodeReviewComments
* Top level comments are expected to be wrapped at 80 cols. Indented comments
  should be wrapped at reasonable width.
* Comments should start with a capitalized letter and end with a period.
* Markdown tries to follow [Google Markdown
  style](https://github.com/google/styleguide/blob/gh-pages/docguide/style.md)
