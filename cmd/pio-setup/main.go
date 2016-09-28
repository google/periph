// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// pio-setup configures the host to make the most use of the underlying
// platform.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/google/pio/host"
	"github.com/google/pio/host/distro"
)

var (
	dryRun = flag.Bool("-dry-run", false, "print every changes that would be done but do not apply any")
)

// writeFile writes a file, only if it was not exactly the same content.
//
// Returns true if the content changed.
//
// Prints what would be written when in dry mode.
func writeFile(path, content string) (bool, error) {
	if actual, err := ioutil.ReadFile(path); err == nil && string(actual) == content {
		// No need to write anything.
		return false, nil
	}
	if *dryRun {
		fmt.Printf("DRY_RUN: writing to %s:\n--- CUT HERE ---\n%s\n-- CUT HERE ---\n", path, content)
		return true, nil
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return true, err
	}
	defer f.Close()
	_, err = f.Write([]byte(content))
	return true, err
}

// execCmd executes a command or prints what it would run when in dry mode.
func execCmd(args ...string) error {
	cmd := strings.Join(args, " ")
	if *dryRun {
		fmt.Printf("DRY_RUN: %s", cmd)
		return nil
	}
	if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
		return fmt.Errorf("%s failed:\n%s", cmd, out)
	}
	return nil
}

// Linux

// setupUdev updates /etc/udev/rules.d/99-pio.rules with the content specified.
//
// The actual content varies across linux distributions.
//
// If the file is updated, reloads the udev rules.
func setupUdev(content string) error {
	if changed, err := writeFile("/etc/udev/rules.d/99-pio.rules", content); err != nil {
		return err
	} else if changed {
		// Reload the rules if the file changed.
		if err := execCmd("udevadm", "control", "--reload-rules"); err != nil {
			return err
		}
		if err := execCmd("udevadm", "trigger"); err != nil {
			return err
		}
	}
	return nil
}

// setupRaspbian contains setup rules that only apply when running on Raspbian.
//
// https://raspbian.org
func setupRaspbian() error {
	// Call raspi-config with the relevant flags to enable I²C, SPI, etc.
	// TODO(maruel): The gpio rule must not be done on Raspbian but led should.
	return nil
}

// setupArmbian contains setup rules that only apply when running on Armbian.
//
// http://armbian.com/
func setupArmbian() error {
	// Add configuration as relevant here.
	return nil
}

// setupUbuntu contains setup rules that applies to Ubuntu based distributions
// that are not one of the above.
func setupUbuntu() error {
	// Add configuration as relevant here.
	return nil
}

// setupDebian contains setup rules that applies to debian based distributions
// that are not one of the above. Its the fallback category.
func setupDebian() error {
	// Add configuration as relevant here.
	return nil
}

// setupLinux contains setup rules that only apply when running on a linux
// based distribution.
func setupLinux() error {
	if err := setupUdev(""); err != nil {
		return err
	}
	if distro.IsRaspbian() {
		if err := setupRaspbian(); err != nil {
			return err
		}
	} else if distro.IsArmbian() {
		if err := setupArmbian(); err != nil {
			return err
		}
	} else if distro.IsUbuntu() {
		if err := setupUbuntu(); err != nil {
			return err
		}
	} else if distro.IsDebian() {
		if err := setupDebian(); err != nil {
			return err
		}
	} else {
		fmt.Printf("Unknown linux based distribution %q. Please update pio-setup to support this configuration.\n", distro.OSRelease()["NAME"])
	}
	return nil
}

// OSX

// setupOSX contains setup rules that only apply when running on OSX.
func setupOSX() error {
	// Add configuration as relevant here.
	return nil
}

// Windows

// setupWindows contains setup rules that only apply when running on Windows.
func setupWindows() error {
	// Add configuration as relevant here.
	return nil
}

// Main

func mainImpl() error {
	flag.Parse()
	if flag.NArg() != 0 {
		return errors.New("arguments not supported")
	}

	// Do not look at state, it is expected that drivers may fail, this tool is
	// the point of fixing these problems.
	if _, err := host.Init(); err != nil {
		return err
	}

	// The reason to use these constants instead of moving to OS specific build
	// files is to ensure all the code is valid and compilable, independent on
	// which OS the code is being compiled on. This is to simplify testing.
	if isLinux {
		if err := setupLinux(); err != nil {
			return err
		}
	}
	if isOSX {
		if err := setupOSX(); err != nil {
			return err
		}
	}
	if isWindows {
		if err := setupWindows(); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "pio-setup: %s.\n", err)
		os.Exit(1)
	}
}
