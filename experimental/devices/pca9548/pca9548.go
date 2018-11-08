package pca9548

import (
	// "periph.io/x/periph/conn/mmr"
	"errors"
	"strconv"
	"sync"

	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
)

type Dev struct {
	c       i2c.Bus
	address uint16
	ports   uint8
	mu      sync.Mutex
	port    uint8
}

var DefaultOpts = Opts{Address: 0x00}

type Opts struct {
	Address uint16
}

// New creates a new handel to a pca9548 i2c multiplexer.
func New(bus i2c.Bus, opts *Opts) (*Dev, error) {
	return &Dev{
		c:       bus,
		ports:   8,
		port:    0xFF,
		address: opts.Address,
	}, nil
}

// tx wraps the bus tx, it maintains which port that each bus is registered on
// os that communication from the master is always on the right port.
func (d *Dev) tx(port Port, address uint16, w, r []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if address == d.address {
		return errors.New("failed to write device address conflicts with mux address")
	}
	if port.number != d.port {
		err := d.c.Tx(d.address, []byte{uint8(1 << (port.number))}, nil)
		if err != nil {
			return errors.New("failed to write active channel on mux: " + err.Error())
		}
		d.port = port.number
	}
	return d.c.Tx(address, w, r)
}

// NewBus registers a new i2c bus on the mux.
func (d *Dev) NewBus(port uint8) (i2c.Bus, error) {
	if port >= d.ports {
		return Port{}, errors.New("port number must be between 0 and " + strconv.Itoa(int(d.ports-1)))
	}
	return Port{
		mux:    d,
		number: port,
	}, nil
}

// String gets the port number of the bus on the multiplexer
func (p Port) String() string { return "Port:" + strconv.Itoa(int(p.number)) }

// SetSpeed is no implemented as the port slaves the master port clock
func (p Port) SetSpeed(f physic.Frequency) error { return nil }

// Tx does a transaction on the multiplexer port it is register to.
func (p Port) Tx(addr uint16, w, r []byte) error { return p.mux.tx(p, addr, w, r) }

// Port is a i2c.Bus on the multiplexer
type Port struct {
	mux    *Dev
	number uint8
}
