// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package netlink

import (
	"errors"
	"fmt"
	"sync"
	"syscall"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/onewire"
	"periph.io/x/periph/conn/onewire/onewirereg"
)

// NewOneWire opens a 1-wire bus via its netlink interface as described at
// https://www.kernel.org/doc/Documentation/w1/w1.netlink
//
// masterID is the bus number reported by the netlink W1_LIST_MASTERS command.
// The resulting object is safe for concurrent use.
//
// NOTE: the Linux 1-wire netlink API does not support strong pull-ups after
// write operations. Hence this driver does not support this feature either. The
// pull-up argument passed to Tx() is ignored. Devices may need to be powered
// externally to work with this driver.
//
// Do not use netlink.NewOneWire() directly as package netink is providing a
// Linux-specific implementation. Instead, use
// https://periph.io/x/periph/conn/onewire/onewirereg#Open. This permits it to
// work on all operating systems.
func NewOneWire(masterID uint32) (*OneWire, error) {
	if isLinux {
		return newOneWire(masterID)
	}
	return nil, errors.New("netlink-onewire: is not supported on this platform")
}

// OneWire is a 1-wire bus via netlink.
//
// It can be used to communicate with multiple devices from multiple goroutines.
type OneWire struct {
	masterID uint32

	mu  sync.Mutex
	s   *socket
	seq uint32
}

// Close closes the handle to the 1-wire driver. It is not a requirement to
// close before process termination.
func (o *OneWire) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if err := o.s.close(); err != nil {
		return fmt.Errorf("netlink-onewire: %v", err)
	}
	return nil
}

// String returns the name of the OneWire instance.
func (o *OneWire) String() string {
	return fmt.Sprintf("Netlink-OneWire%d", o.masterID)
}

// Tx performs reset, write and (if len(r) > 0) read operations on the 1-wire
// bus.
//
// NOTE: the Linux 1-wire netlink API does not support requesting strong
// pull-ups after write operations. Hence this driver does not support this
// feature either. The pull-up argument passed to Tx() is ignored.
func (o *OneWire) Tx(w, r []byte, _ onewire.Pullup) error {
	// Grouping multiple commands into a single netlink message appears to make
	// bus transactions significantly more stable.
	cmds := []*w1Cmd{w1CmdReset(), w1CmdWrite(w)}
	if l := len(r); l > 0 {
		cmds = append(cmds, w1CmdRead(l))
	}
	m := &w1Msg{
		typ:      W1_MASTER_CMD,
		masterID: o.masterID,
		cmds:     cmds,
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	o.seq++

	d, err := o.s.sendAndRecv(o.seq, m)
	if err != nil {
		return fmt.Errorf("netlink-onewire: transaction failed: %v", err)
	}
	copy(r, d)

	return nil
}

// Search performs a device search operation on the 1-wire bus. The resulting
// device addresses are returned. If alarmOnly is true, only devices in alarm
// state are returned.
func (o *OneWire) Search(alarmOnly bool) ([]onewire.Address, error) {
	cmd := w1CmdSearch()
	if alarmOnly {
		cmd = w1CmdAlarmSearch()
	}
	m := &w1Msg{
		typ:      W1_MASTER_CMD,
		masterID: o.masterID,
		cmds:     []*w1Cmd{cmd},
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	o.seq++

	d, err := o.s.sendAndRecv(o.seq, m)
	if err != nil {
		return nil, fmt.Errorf("netlink-onewire: search command failed: %v", err)
	}

	l := len(d)
	if l%8 != 0 {
		return nil, fmt.Errorf("netlink-onewire: search payload size %d is not a multiple of 8", l)
	}

	var addrs []onewire.Address
	for len(d) > 0 {
		addr := getUint64(d)
		addrs = append(addrs, onewire.Address(addr))

		d = d[8:]
	}
	return addrs, nil
}

// Private details.

func newOneWire(masterID uint32) (*OneWire, error) {
	s, err := newSocket()
	if err != nil {
		return nil, fmt.Errorf("netlink-onewire: failed to create socket: %v", err)
	}

	return &OneWire{
		masterID: masterID,
		s:        s,
	}, nil
}

//

const (
	// Supported netlink message type.
	NLMSG_DONE = uint16(0x3)

	// 1-Wire connector IDs.
	CN_W1_IDX = uint32(0x3)
	CN_W1_VAL = uint32(0x1)
)

type msgType uint8

// Supported netlink message types.
const (
	W1_MASTER_CMD   = msgType(4)
	W1_LIST_MASTERS = msgType(6)
)

// w1Msg holds the information required to create a buffer that represents a C
// struct w1_netlink_msg with zero or more 1-wire commands.
type w1Msg struct {
	typ      msgType
	masterID uint32
	cmds     []*w1Cmd
}

// serialize returns a buffer with the same memory layout as a w1_netlink_msg
// struct.
func (m *w1Msg) serialize() []byte {
	var l int
	for _, cmd := range m.cmds {
		l += cmd.len()
	}

	b := make([]byte, 12+l)
	b[0] = byte(m.typ)
	putUint16(b[2:4], uint16(l))
	putUint32(b[4:8], m.masterID)

	d := b[12:]
	for _, cmd := range m.cmds {
		copy(d, cmd.serialize())
		d = d[cmd.len():]
	}
	return b
}

type cmdType uint8

// Supported OneWire commands.
const (
	W1_CMD_READ         = cmdType(0)
	W1_CMD_WRITE        = cmdType(1)
	W1_CMD_SEARCH       = cmdType(2)
	W1_CMD_ALARM_SEARCH = cmdType(3)
	W1_CMD_RESET        = cmdType(5)
)

// w1Cmd holds the information required to create a buffer that represents a C
// struct w1_netlink_cmd.
type w1Cmd struct {
	typ cmdType
	// For read and write commands.
	payloadLen int
	// For write commands.
	payload []byte
	// True if the command is expected to triggers a response from the kernel.
	wantResponse bool
}

// serialize returns a buffer with a memory layout that matches struct
// w1_netlink_cmd.
func (c *w1Cmd) serialize() []byte {
	b := make([]byte, 4+c.payloadLen)
	b[0] = byte(c.typ)
	// b[1]: reserved
	putUint16(b[2:4], uint16(c.payloadLen))
	if len(c.payload) > 0 {
		copy(b[4:], c.payload)
	}
	return b
}

func (c *w1Cmd) len() int {
	return 4 + c.payloadLen
}

func w1CmdReset() *w1Cmd { return &w1Cmd{typ: W1_CMD_RESET} }

func w1CmdSearch() *w1Cmd { return &w1Cmd{typ: W1_CMD_SEARCH, wantResponse: true} }

func w1CmdAlarmSearch() *w1Cmd { return &w1Cmd{typ: W1_CMD_ALARM_SEARCH, wantResponse: true} }

func w1CmdRead(l int) *w1Cmd { return &w1Cmd{typ: W1_CMD_READ, payloadLen: l, wantResponse: true} }

func w1CmdWrite(d []byte) *w1Cmd { return &w1Cmd{typ: W1_CMD_WRITE, payloadLen: len(d), payload: d} }

//

// socket is a netlink connector socket for reading and writing 1-wire messages.
type socket struct {
	fd int
}

// newSocket returns a socket instance.
func newSocket() (*socket, error) {
	// Open netlink socket.
	fd, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_DGRAM, syscall.NETLINK_CONNECTOR)
	if err != nil {
		return nil, fmt.Errorf("failed to open netlink socket: %v", err)
	}

	if err := syscall.Bind(fd, &syscall.SockaddrNetlink{Family: syscall.AF_NETLINK}); err != nil {
		return nil, fmt.Errorf("failed to bind netlink socket: %v", err)
	}

	return &socket{fd: fd}, nil
}

func (s *socket) sendAndRecv(seq uint32, m *w1Msg) ([]byte, error) {
	if err := s.sendMsg(m.serialize(), seq); err != nil {
		return nil, fmt.Errorf("failed to send W1 message: %v", err)
	}

	var data []byte
	for _, cmd := range m.cmds {
		if cmd.wantResponse {
			d, err := s.recvCmd(seq, seq+1, m.typ, cmd.typ)
			if err != nil {
				return nil, fmt.Errorf("failed to receive command payload: %v", err)
			}
			data = d
		}

		// Every command is ack'ed with a separate message, including commands
		// for which a response has already been returned.
		if _, err := s.recvCmd(seq, 0, m.typ, cmd.typ); err != nil {
			return nil, fmt.Errorf("failed to receive command ack message: %v", err)
		}
	}

	return data, nil
}

// sendMsg wraps the given data in a netlink header and connector message, and
// writes it to the socket. seq is the sequence number in the connector message
// (C struct cn_msg). The same sequence number must be passed to subsequent
// readMsg or readCmd calls.
func (s *socket) sendMsg(data []byte, seq uint32) error {
	dataLen := len(data)

	// Total size of message, with padding for 4 byte alignment.
	nlLen := syscall.SizeofNlMsghdr + 20 + dataLen

	// Populate required fields of struct nlmsghdr.
	nl := make([]byte, nlLen+(4-nlLen%4)%4)
	putUint32(nl[0:4], uint32(nlLen))
	putUint16(nl[4:6], NLMSG_DONE)
	putUint32(nl[8:12], seq)

	// Populate required fields of struct cn_msg.
	cn := nl[16:]
	putUint32(cn[0:4], CN_W1_IDX)
	putUint32(cn[4:8], CN_W1_VAL)
	putUint32(cn[8:12], seq)
	putUint16(cn[16:18], uint16(dataLen))

	// Append payload.
	copy(cn[20:], data)

	if err := syscall.Sendto(s.fd, nl, 0, &syscall.SockaddrNetlink{Family: syscall.AF_NETLINK}); err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	return nil
}

// recvMsg performs a single read from the socket, verifies and removes netlink,
// connector, and W1 message headers, and returns the W1 message payload. The
// netlink connector header must have sequence and acknowledgement numbers equal
// to wantSeq and wantAck, respectively. The W1 message must have type wantType.
// recvMsg returns an error if either of these conditions are not satisfied.
// Multiple (bundled) messages are not supported.
func (s *socket) recvMsg(wantSeq, wantAck uint32, wantType msgType) ([]byte, error) {
	data := make([]byte, 1024)

	n, _, err := syscall.Recvfrom(s.fd, data, 0)
	if err != nil {
		return nil, err
	}
	b := data[:n]

	// Check struct nlmsghdr fields len and type, and strip off netlink header.
	nlLen := int(getUint32(b[0:4]))
	if n < nlLen {
		return nil, fmt.Errorf("received message size (%d bytes) < netlink header length (%d bytes)", n, nlLen)
	}
	if gotType, wantType := getUint16(b[4:6]), NLMSG_DONE; gotType != wantType {
		return nil, fmt.Errorf("received netlink message type %d, want %d", gotType, wantType)
	}
	b = b[syscall.SizeofNlMsghdr:nlLen]

	if l := len(b); l < 20 {
		return nil, fmt.Errorf("incomplete netlink connector message; got %d bytes, want 20", l)
	}

	// Check struct cn_msg fields idx, val, seq and ack.
	if gotIdx, wantIdx := getUint32(b[0:4]), CN_W1_IDX; gotIdx != wantIdx {
		return nil, fmt.Errorf("got connector index %d, want %d", gotIdx, wantIdx)
	}
	if gotSeq := getUint32(b[8:12]); gotSeq != wantSeq {
		return nil, fmt.Errorf("received connector seq %d, want %d", gotSeq, wantSeq)
	}
	if gotAck := getUint32(b[12:16]); gotAck != wantAck {
		return nil, fmt.Errorf("received connector ack %d, want %d", gotAck, wantAck)
	}

	// Check payload length and strip off struct cn_msg.
	wantLen := getUint16(b[16:18])
	b = b[20:]
	if gotLen := len(b); gotLen != int(wantLen) {
		return nil, fmt.Errorf("invalid w1_netlink_msg length %d, want %d", gotLen, wantLen)
	}
	if wantLen == 0 {
		return nil, errors.New("empty connector message")
	}
	if wantLen < 12 {
		return nil, fmt.Errorf("incomplete w1_netlink_msg; got %d bytes, want at least 12", wantLen)
	}

	// Check w1_netlink_msg type, status, and payload length.
	if gotType := msgType(b[0]); gotType != wantType {
		return nil, fmt.Errorf("invalid w1_netlink_msg type %v, want %v", gotType, wantType)
	}
	if status := b[1]; status != 0 {
		return nil, fmt.Errorf("invalid w1_netlink_msg status %d", status)
	}
	wantLen = getUint16(b[2:4])
	b = b[12:]
	if gotLen := len(b); gotLen != int(wantLen) {
		return nil, fmt.Errorf("invalid w1_netlink_msg payload length %d, want %d", gotLen, wantLen)
	}

	return b, nil
}

// recvCmd performs a single read from the socket, verifies and removes netlink,
// connector, W1 message, and W1 command headers, and returns the W1 command
// payload. wantSeq and wantAck are the expected connector sequence and
// acknowledgement numbers, respectively. wantMsgType and wantCmdType are the
// expected W1 message and command types, respectively. An error is returned if
// the received data does not match either of these values.
func (s *socket) recvCmd(wantSeq, wantAck uint32, wantMsgType msgType, wantCmdType cmdType) ([]byte, error) {
	b, err := s.recvMsg(wantSeq, wantAck, wantMsgType)
	if err != nil {
		return nil, err
	}
	if l := len(b); l < 4 {
		return nil, fmt.Errorf("incomplete w1_netlink_cmd; got %d bytes, want at least 4", l)
	}

	// Check w1_netlink_cmd type and payload length.
	if gotCmdType := cmdType(b[0]); gotCmdType != wantCmdType {
		return nil, fmt.Errorf("invalid w1_netlink_cmd type %v, want %v", gotCmdType, wantCmdType)
	}
	wantLen := getUint16(b[2:4])
	b = b[4:]
	if gotLen := len(b); gotLen != int(wantLen) {
		return nil, fmt.Errorf("invalid w1_netlink_cmd payload length %d, want %d", gotLen, wantLen)
	}

	return b, nil
}

func (s *socket) close() error {
	fd := s.fd
	s.fd = 0
	return syscall.Close(fd)
}

//

// driver1W implements periph.Driver.
type driver1W struct {
	buses []string
}

func (d *driver1W) String() string {
	return "netlink-onewire"
}

func (d *driver1W) Prerequisites() []string {
	return nil
}

func (d *driver1W) After() []string {
	return nil
}

func (d *driver1W) Init() (bool, error) {
	s, err := newSocket()
	if err != nil {
		return false, fmt.Errorf("netlink-onewire: failed to open socket: %v", err)
	}
	defer s.close()

	// Find bus masters.
	m := &w1Msg{typ: W1_LIST_MASTERS}
	if err := s.sendMsg(m.serialize(), 0); err != nil {
		return false, fmt.Errorf("netlink-onewire: failed to send list bus msg: %v", err)
	}

	b, err := s.recvMsg(0, 1, W1_LIST_MASTERS)
	if err != nil {
		return false, fmt.Errorf("netlink-onewire: failed to receive bus IDs: %v", err)
	}

	l := len(b)
	if l%4 != 0 {
		return false, fmt.Errorf("netlink-onewire: data size %d is not a multiple of 4", l)
	}

	var ids []uint32
	for len(b) > 0 {
		ids = append(ids, getUint32(b))
		b = b[4:]
	}

	for _, id := range ids {
		bus := int(id)
		name := fmt.Sprintf("netlink-w1-master %d", bus)
		d.buses = append(d.buses, name)
		aliases := []string{fmt.Sprintf("OneWire%d", bus)}
		if err := onewirereg.Register(name, aliases, bus, openerOneWire(bus).Open); err != nil {
			return true, err
		}
	}
	return true, nil
}

//

func putUint16(b []byte, v uint16) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
}

func putUint32(b []byte, v uint32) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

func getUint16(b []byte) uint16 {
	return uint16(b[0]) +
		uint16(b[1])<<8
}

func getUint32(b []byte) uint32 {
	return uint32(b[0]) +
		uint32(b[1])<<8 +
		uint32(b[2])<<16 +
		uint32(b[3])<<24
}

func getUint64(b []byte) uint64 {
	return uint64(b[0]) +
		uint64(b[1])<<8 +
		uint64(b[2])<<16 +
		uint64(b[3])<<24 +
		uint64(b[4])<<32 +
		uint64(b[5])<<40 +
		uint64(b[6])<<48 +
		uint64(b[7])<<56
}

//

type openerOneWire int

func (o openerOneWire) Open() (onewire.BusCloser, error) {
	b, err := NewOneWire(uint32(o))
	if err != nil {
		return nil, err
	}
	return b, nil
}

func init() {
	if isLinux {
		periph.MustRegister(&drvOneWire)
	}
}

var drvOneWire driver1W

var _ onewire.Bus = &OneWire{}
var _ onewire.BusCloser = &OneWire{}
