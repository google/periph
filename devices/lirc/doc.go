// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package lirc implements InfraRed receiver support through native linux app
// lirc.
//
// Configuration
//
// lircd MUST be configured via TWO files: /etc/lirc/hardware.conf and
// /etc/lirc/lircd.conf.
//
// See http://www.lirc.org/ for more details about daemon configuration.
//
// /etc/lirc/hardware.conf
//
// This file contains the interaction between the lircd process and the kernel
// driver, if any. This is the link between the physical signal and decoding
// pulses.
//
// /etc/lirc/lircd.conf
//
// This file contains all the known IR codes for the remotes you plan to use
// and convert into key codes. This means you need to "train" lircd with the
// remotes you plan to use.
//
// Keys are listed at
// http://www.lirc.org/api-docs/html/input__map_8inc_source.html
//
// Debugging
//
// Here's a quick recipe to train a remote:
//
//     # Detect your remote
//     irrecord -a -d /var/run/lirc/lircd ~/lircd.conf
//     # Grep for key names you found to find the remote in the remotes library
//     grep -R '<hex value>' /usr/share/lirc/remotes/
//     # Listen and send command to the server
//     nc -U /var/run/lirc/lircd
//     # List all valid key names
//     irrecord -l
//     grep -hoER '(BTN|KEY)_\w+' /usr/share/lirc/remotes | sort | uniq | less
//
// Raspbian
//
// Please see documentation of package periph/host/rpi for details on how to set
// it up.
//
// Hardware
//
// A good peripheral is the VS1838. Then you need peripheral driver for hardware
// accelerated signal decoding, that lircd will then leverage to decode the
// keypresses.
package lirc
