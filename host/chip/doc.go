// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package chip contains NextThing Co's C.H.I.P. hardware logic. It is intrinsically
// related to package allwinner.
//
// References: http://www.chip-community.org/index.php/Hardware_Information and
// http://docs.getchip.com/chip.html#chip-hardware
//
// The platform detection assumes the following info coming from the device tree:
//   root@chip2:/proc/device-tree# od -c compatible
//   0000000   n   e   x   t   t   h   i   n   g   ,   c   h   i   p  \0   a
//   0000020   l   l   w   i   n   n   e   r   ,   s   u   n   5   i   -   r
//   0000040   8  \0
//   root@chip2:/proc/device-tree# od -c model
//   0000000   N   e   x   t   T   h   i   n   g       C   .   H   .   I   .
//   0000020   P   .  \0
package chip
