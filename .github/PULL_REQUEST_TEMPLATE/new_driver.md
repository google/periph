---
name: New driver
about: New code to support new hardware or new software

---

Please prefix the issue with the primary package affected. For example, if you
fixed something in
[periph.io/x/periph/host/sysfs](https://github.com/google/periph/tree/master/host/sysfs),
prefix the PR with `sysfs:`.

Please add new drivers under `experimental`. Wonder what it takes to promote a
driver as _stable_? See https://periph.io/project/#driver-lifetime-management. A
stable driver requires the smallest API surface, good unit test code coverage,
good documentation and a page in
[https://periph.io/device/](https://github.com/periph/website/tree/master/site/content/device)

**Mention the issue number it fixes or add the details of the changes if it
doesn't have a specific issue.**

Fixes #

<!--
example: Fixes #123
example: Helps with #123 but doesn't not completely fix it.
-->
