// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build usb

package usbbus

import (
	"log"
	"sort"
	"sync"

	gousb "github.com/kylelemons/gousb/usb"
	"periph.io/x/periph"
	"periph.io/x/periph/experimental/conn/usb"
)

// Desc represents the description of an USB peripheral on an USB bus.
type Desc struct {
	ID   usb.ID
	Bus  uint8
	Addr uint8
}

// All returns all the USB peripherals detected.
func All() []Desc {
	mu.Lock()
	defer mu.Unlock()
	// TODO(maruel): driver.Init() should skip scanning the USB bus unless
	// there's at least one USB driver registered. So in this case an USB scan
	// should be done synchronously.
	out := make([]Desc, len(all))
	copy(out, all)
	return out
}

//

var (
	newDriver = make(chan usb.Driver)

	mu      sync.Mutex
	all     descriptors
	drivers = map[usb.ID]usb.Opener{}
)

type descriptors []Desc

func (d descriptors) Len() int      { return len(d) }
func (d descriptors) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d descriptors) Less(i, j int) bool {
	if d[i].Bus < d[j].Bus {
		return true
	}
	if d[i].Bus > d[j].Bus {
		return false
	}
	return d[i].Addr < d[j].Addr
}

func fromDesc(d *gousb.Descriptor) Desc {
	return Desc{usb.ID{uint16(d.Vendor), uint16(d.Product)}, d.Bus, d.Address}
}

// dev is an open handle to an USB peripheral.
//
// The peripheral can disappear at any moment.
type dev struct {
	desc Desc
	name string
	d    *gousb.Device
	e    gousb.Endpoint
}

func (d *dev) String() string {
	return d.name
}

func (d *dev) Close() error {
	return d.d.Close()
}

func (d *dev) ID() *usb.ID {
	return &d.desc.ID
}

func (d *dev) Write(b []byte) (int, error) {
	return d.e.Write(b)
}

func (d *dev) Tx(w, r []byte) error {
	if _, err := d.e.Write(w); err != nil {
		return err
	}
	if len(r) == 0 {
		return nil
	}
	_, err := d.e.Read(r)
	return err
}

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "usb"
}

func (d *driver) Prerequisites() []string {
	return nil
}

func onNewDriver() {
	for d := range newDriver {
		mu.Lock()
		// The items are guaranteed to not have duplicates.
		drivers[d.ID] = d.Opener
		for _, devices := range all {
			if d.ID == devices.ID {
				// Only rescan if the peripheral had been detectd.
				scanDevices(map[usb.ID]usb.Opener{d.ID: d.Opener})
				break
			}
		}
		mu.Unlock()
	}
}

func (d *driver) Init() (bool, error) {
	// Gather all the previously registered peripheral drivers and do one scan
	// synchronously.
	//
	// Start one loop that will be called during the function call.
	var wg sync.WaitGroup
	wg.Add(1)
	quit := make(chan struct{})
	go func() {
		mu.Lock()
		defer mu.Unlock()
		wg.Done()
		for {
			select {
			case d := <-newDriver:
				// The items are guaranteed to not have duplicates.
				drivers[d.ID] = d.Opener
			case <-quit:
				return
			}
		}
	}()
	wg.Wait()
	usb.RegisterBus(newDriver)
	quit <- struct{}{}

	mu.Lock()
	defer mu.Unlock()
	scanDevices(drivers)

	// After this initial scan, scan asynchronously when drivers are registered.
	go onNewDriver()

	// TODO(maruel): Start an event loop when new peripherals are plugged in
	// without polling.
	// go func() { for { WaitForUSBBusEvents(); usb.OnDevice(...) } }()
	return true, nil
}

func scanDevices(m map[usb.ID]usb.Opener) error {
	// I'd much prefer something that just talks to the OS instead of using
	// libusb. Especially we only require a small API surface.
	ctx := gousb.NewContext()
	defer ctx.Close()
	all = nil
	devs, err := ctx.ListDevices(func(d *gousb.Descriptor) bool {
		// Return true to keep the peripheral open.
		desc := fromDesc(d)
		all = append(all, desc)
		_, ok := m[desc.ID]
		return ok
	})
	// This API is really poor as there can be multiple peripherals opened and you
	// don't know how many failed.
	// If the user needs root access, LIBUSB_ERROR_ACCESS (-3) will be returned.
	sort.Sort(all)
	for _, d := range devs {
		desc := fromDesc(d.Descriptor)
		name, err := d.GetStringDescriptor(1)
		if err != nil {
			// Sometimes the USB peripheral will return junk, default to the vendor
			// and peripheral ids.
			name = desc.ID.String()
		}
		// Control, isochronous or bulk?
		e, err := d.OpenEndpoint(1, 0, 0, 1|uint8(gousb.ENDPOINT_DIR_IN))
		if err != nil {
			log.Printf("Open: %v", err)
			d.Close()
			continue
		}
		if err := m[desc.ID](&dev{desc, name, d, e}); err != nil {
			log.Printf("opener: %v", err)
			d.Close()
			continue
		}
	}
	return err
}

func init() {
	periph.MustRegister(&driver{})
}

var _ periph.Driver = &driver{}
var _ usb.ConnCloser = &dev{}
