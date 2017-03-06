# Contributing to Periph

Want to contribute? Great! First, read this page (including the small print at
the end).

## Before you contribute

Before we can use your code, you must sign the [Google Individual Contributor
License Agreement] (https://cla.developers.google.com/about/google-individual)
(CLA), which you can do online. The CLA is necessary mainly because you own the
copyright to your changes, even after your contribution becomes part of our
codebase, so we need your permission to use and distribute your code. We also
need to be sure of various other things—for instance that you'll tell us if you
know that your code infringes on other people's patents. You don't have to sign
the CLA until after you've submitted your code for review and a member has
approved it, but you must do it before we can put your code into our codebase.
Before you start working on a larger contribution, you should get in touch with
us first through the issue tracker with your idea so that we can help out and
possibly guide you. Coordinating up front makes it much easier to avoid
frustration later on.


## Code reviews

All submissions, including submissions by project members, require review. The
`periph` project uses Github pull requests for this purpose.


## Code quality

All submissions, including submissions by project members, require abiding to
high code quality. See [Requirements](../#requirements) for the
expectations and take a look at the [driver lifetime
management](../#driver-lifetime-management) to learn how to contribute new
device support.


## Testing

The `periph` project uses use [gohci](https://github.com/periph/gohci) for
automated testing on devices. The devices run unit tests, `go vet` and
[periph-smoketest](https://github.com/google/periph/tree/master/cmd/periph-smoketest).

The fleet currently is currently hosted by [maruel](https://github.com/maruel):

- [Raspberry Pi 3](https://www.raspberrypi.org/) running [Raspbian Jessie
  Lite](https://www.raspberrypi.org/downloads/raspbian/)
- [C.H.I.P.](https://getchip.com/pages/chip) running Debian headless image
  provided by NTC
- Windows 10 VM

The tests must not be broken by a PR.


## Conduct

While this project is not related to the Go project itself, `periph` abides to
the same code of conduct as the Go project as described at
https://golang.org/conduct. `periph` doesn't yet have a formal committee (help's
appreciated), please email directly `maruel@chromium.org` for issues
encountered.


## The small print

Contributions made by corporations are covered by a different agreement than
the one above, the [Software Grant and Corporate Contributor License Agreement]
(https://cla.developers.google.com/about/google-corporate).
