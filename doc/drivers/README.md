# periph - Device driver developers

Documentation for _device driver developers_ who either want to develop a
device driver in their own code base or want to submit a contribution to extend
the hardware supported by `periph`.


## Background

The main purpose of `periph` is to provide interfaces to assemble components
to communicate with hardware peripherals. As such, it splits boards
into their individual components: CPU, buses, physical headers, etc, instead of
representing each board as a whole object.

Read more about the [goals](goals/). This should be read before contributing to
understand the project rationale.

Read more about the [detailed design](design/) to understand more about how
things are done.


### Classes of hardware

This document distinguishes two classes of drivers:

- Enablers: they are what make the interconnects work, so that you can then
  use real stuff. That's both point-to-point connections (GPIO, UART, TCP) and
  buses (I²C, SPI, BT) where individual devices can be addressed.  They enable
  you to do something but are not the essence of what you want to do.
- Devices: they are the end goal, to do something functional. There are multiple
  subclasses of devices like sensors, output devices, etc.


## Driver lifetime management

Proper driver lifetime management is key to the success of this project. There
must be clear expectations to add, update and remove drivers for the core
project. As described in the [Risks section in goals/](goals/#risks) below,
poor drivers or high churn rate will destroy the value proposition.

This is critical as drivers can be silently broken by seemingly innocuous
changes. Because the testing of hardware drivers is significantly harder than
that of software-only projects, there’s an inherent faith in the quality of the
code that must be asserted.


### Experimental

Any driver can be requested to be added to the library under the
[experimental/](https://github.com/google/periph/tree/master/experimental/)
directory. The following process must be followed:

- Create a driver out of tree an make it work.
- Improve the driver so it meets a minimal quality bar under the promise of
  being improved. See [Requirements](#requirements) for the extensive list.
- Follow the [contributing/](contributing/) requirements.
- Create a Pull Request for integration under
  [experimental/](https://github.com/google/periph/tree/master/experimental/)
  and respond to the code review.

At this point, it is available for use to everyone but it is not loaded by
default by [host.Init()](https://godoc.org/periph.io/x/periph/host#Init).

There is no API compatibility guarantee for drivers under
[experimental/](https://github.com/google/periph/tree/master/experimental/).


### Stable

A driver in
[experimental/](https://github.com/google/periph/tree/master/experimental/) can
be promoted to stable in either
[devices/](https://github.com/google/periph/tree/master/devices/) or
[host/](https://github.com/google/periph/tree/master/host/) as appropriate. The
following process must be followed:

- Declare at least one (or multiple) owners that are responsive to
  feature requests and bug reports.
  - There could be a threshold, > _TO BE DETERMINED_ lines, where more than one
    owner is required.
  - The owners commit to support the driver for the foreseeable future and
    **promptly** do code reviews to keep the driver quality at the expected
    standard.
- There are multiple reports that the driver is functioning as expected.
- If another driver exists for an intersecting class of devices, the other
  driver must enter deprecation phase.
- At this point the driver must maintain an API compatibility promise.


### Deprecation

A driver can be subsumed by a newer driver with a better core implementation or
a new breaking API. The previous driver must be deprecated, moved back to
[experimental/](https://github.com/google/periph/tree/master/experimental/) and
announced to be deleted after _TO BE DETERMINED_ amount of time.


### Contributing a new driver

A new proposed driver must first be implemented out of tree and fit all the
items in [Requirements](#requirements) listed below. It can then be proposed as
[Experimental](#experimental), and finally requested to be promoted to
[Stable](#stable).


## Requirements

All the code must fit the following requirements.

**Fear not!** We know the list _is_ daunting but as you create your pull request
to add something in
[experimental/](https://github.com/google/periph/tree/master/experimental/)
we'll happily guide you in the process to help improve the code to meet the
expected standard. The end goal is to write *high quality maintainable code* and
use this as a learning experience.

- The code must be Go idiomatic.
  - Constructor `NewXXX()` returns an object of concrete type.
  - Functions accept interfaces.
  - Leverage standard interfaces like
    [io.Writer](https://golang.org/pkg/io/#Writer) and
    [image.Image](https://golang.org/pkg/image/#Image) where possible.
  - No `interface{}` unless strictly required.
  - Minimal use of factories except for protocol level registries.
  - No `init()` code that accesses peripherals on process startup. These belong
    to
    [Driver.Init()](https://godoc.org/periph.io/x/periph#Driver).
- Exact naming
  - Driver for a chipset must have the name of the chipset or the chipset
    family. Don't use `oleddisplay`, use `ssd1306`.
  - Driver must use the real chip name, not a marketing name by a third party.
    Don't use `dotstar` (as marketed by Adafruit), use `apa102` (as created
    by APA Electronic co. LTD.).
  - A link to the datasheet must be included in the package doc unless NDA'ed
    or inaccessible.
- Testability
  - Code must be testable and unit tested without a device. The unit tests are
    meant to run as part of `go test`.
  - When relevant, include a smoke test. The smoke test tests a real device to
    confirm the driver physically works for devices. It must be under the
    package being tested, named as `foosmoketest` for package `foo`. Modify
    [periph-smoketests/](https://github.com/google/periph/tree/master/cmd/periph-smoketests/)
    to expose this smoke test.
- Usability
  - Provide a standalone executable in
    [cmd/](https://github.com/google/periph/tree/master/cmd/) to expose the
    functionality.  It is acceptable to only expose a small subset of the
    functionality but _the tool must have purpose_.
  - Provide a `func Example()` along your test to describe basic usage of your
    driver. See the official [testing
    package](https://golang.org/pkg/testing/#hdr-Examples) for more details.
- Performance
  - Drivers controlling an output device must have a fast path that can be used
    to directly write in the device's native format, e.g.
    [io.Writer](https://golang.org/pkg/io/#Writer).
  - Drivers controlling an output device must have a generic path accepting
    a higher level interface when found in the stdlib, e.g.
    [image.Image](https://golang.org/pkg/image/#Image).
  - Floating point arithmetic should only be used when absolutely necesary in
    the driver code. Most of the cases can be replaced by fixed point
    arithmetic, for example
    [devices.Milli](https://godoc.org/periph.io/x/periph/devices#Milli).
    Floating point arithmetic is acceptable in the unit tests and tools in
    [cmd/](https://github.com/google/periph/tree/master/cmd/) but should not be
    abused.
  - Drivers must be implemented with performance in mind. For example I²C
    operations should be batched to minimize overhead.
  - Benchmark must be implemented for non trivial processing running on the
    host.
- Code must compile on all OSes, with minimal use of OS-specific thunk as
  strictly needed. Take advantage of constructs like `if isArm { ...}` where the
  conditional is optimized away at compile time via dead code elimination
  and `isArm` is a simple boolean constant defined in relevant .go files
  having a build constraint.
- Struct implementing an interface must validate at compile time with `var _
  <Interface> = &<Type>{}`.
- License is Apache v2.0.


## Code style

- The code tries to follow Go code style as described at
  https://github.com/golang/go/wiki/CodeReviewComments
- Top level comments are expected to be wrapped at 80 cols. Indented comments
  should be wrapped at reasonable width.
- Comments should start with a capitalized letter and end with a period.
- Markdown style is a "try to similar to the current doc".
