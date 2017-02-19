# periph - Goals


## Abstract

Go developped a fairly large hardware hacker community in part because the
language and its tooling have the following properties:

* Easy to cross compile to ARM/Linux via `GOOS=linux GOARCH=arm go build .`.
* Significantly faster to execute than python and node.js.
* Significantly lighter in term of memory use than Java or node.js.
* Significantly more productive to code than C/C++.
* Builds reasonably fast on ARM.
* Fairly good OS level support: Debian pre-provided Go package (albeit a tad
  old) makes it easy to apt-get install on arm64, or arm32 users have access to
  package on [golang.org](https://golang.org).

Many Go packages, both generic and specialized, were created to fill the space.
This library came out of the desire to have a _designed_ API (contrary to
growing organically) with strict [code requirements](README.md#requirements) and
a [strong, opinionated philosophy](../../#philosophy) to enable long term
maintenance.


## Goals

`periph` was created as an anwer to specific goals:

* Not more abstract than absolutely needed. Use concrete types whenever
  possible.
* Orthogonality and composability
  * Each component must own an orthogonal part of the platform and each
    components can be composed together.
* Extensible:
  * Users can provide additional drivers that are seamlessly loaded
    with a structured ordering of priority.
* Performance:
  * Execution as performant as possible.
  * Overhead as minimal as possible, i.e. irrelevant driver are not be
    attempted to be loaded, uses memory mapped GPIO registers instead of sysfs
    whenever possible, etc.
* Coverage:
  * Be as OS agnostic as possible. Abstract OS specific concepts like
    [sysfs](https://godoc.org/periph.io/x/periph/host/sysfs).
  * Each driver implements and exposes as much of the underlying device
    capability as possible and relevant.
  * [cmd/](../../cmd/) implements useful directly usable tool.
  * [devices/](../../devices/) implements common device drivers.
  * [host/](../../host/) must implement a large base of common platforms that
    _just work_. This is in addition to extensibility.
* Simplicity:
  * Static typing is _thoroughly used_, to reduce the risk of runtime failure.
  * Minimal coding is needed to accomplish a task.
  * Use of the library is defacto portable.
  * Include fakes for buses and device interfaces to simplify the life of
    device driver developers.
* Stability
  * API must be stable without precluding core refactoring.
  * Breakage in the API should happen at a yearly parce at most once the library
    got to a stable state.
* Strong distinction about the driver (as a user of a
  [conn.Conn](https://godoc.org/periph.io/x/periph/conn#Conn)
  instance) and an application writer (as a user of a device driver). It's the
  _application_ that controls the objects' lifetime.
* Strong distinction between _enablers_ and _devices_. See
  [Background](README.md#background).


## Success criteria

* Preferred library used by first time Go users and by experts.
* Becomes the defacto HAL library.
* Becomes the central link for hardware support.


## Risks

The risks below are being addressed via a strong commitment to [driver lifetime
management](README.md#driver-lifetime-management) and having a high quality bar
via an explicit list of [requirements](README.md#requirements).

The enablers (boards, CPU, buses) is what will break or make this project.
Nobody want to do them but they are needed. You need a large base of enablers so
people can use anything yet they are hard to get right. You want them all in the
same repo so that when someone builds an app, it supports everything
transparently. It just works.

The device drivers do not need to all be in the same repo, that scales since
people know what is physically connected, but a large enough set of enablers is
needed to be in the base repository to enable seemlessness. People do not care
that a Pine64 has a different processor than a Rasberry Pi; both have the same
40 pins header and that's what they care about. So enablers need to be a great
HAL -> the right hardware abstraction layer (not too deep, not too light) is the
core here.

Devices need common interfaces to help with application developers (like
[devices.Display](https://godoc.org/periph.io/x/periph/devices#Display)
and
[devices.Environmental](https://godoc.org/periph.io/x/periph/devices#Environmental))
but the lack of core repository and coherency is less dramatic.


### Users

* The library is rejected by users as being too cryptic or hard to use.
* The device drivers are unreliable or non functional, as observed by users.
* Poor usability of the core interfaces.
* Missing drivers.


### Contributors

* Lack of API stability; high churn rate.
* Poor fitting of the core interfaces.
* No uptake in external contribution.
* Poor quality of contribution.
* Duplicate ways to accomplish the same thing, without a clear way to define the
  right way.
