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

All submissions, including submissions by project members, require review. We
use Github pull requests for this purpose.


## Code quality

All submissions, including submissions by project members, require abiding to
high code quality. See [Requirements](README.md#requirements) for the
expectations and take a look at the [driver lifetime
management](README.md#driver-lifetime-management) to learn how to contribute new
device support.


## Testing

We use [sci](https://github.com/maruel/sci) for automated testing on devices.
The devices run unit tests and [periph-smoketest](../../cmd/periph-smoketest).

The fleet currently include Raspberry Pi 3 and a Pine 64. The tests must not be
broken by a PR.


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
