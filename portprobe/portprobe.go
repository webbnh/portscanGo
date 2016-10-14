// Package portprobe provides functions for probing ports on a host.
package portprobe

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/webbnh/DigitalOcean/vdiag"
)

// Port statuses
const (
	closed = iota - 1
	pending
	open
)

const readTimeout = 1 * time.Second

type Result int

func (r Result) IsComplete() bool { return r != pending }
func (r Result) IsOpen() bool     { return r == open }
func (r Result) IsClosed() bool   { return r == closed }

func (r Result) String() string {
	switch r {
	case closed:
		return "closed"
	case pending:
		return "pending"
	case open:
		return "open"
	}
	return "<unrecognized value>"
}

// Wrap the TCP version of net.Dial() in an interface so that we can mock it
// for testing.
type NetDialerTCP interface {
	Dial(address string) (net.Conn, error)
}

// DialerTCP implements the NetDialerTCP interface by invoking the
// corresponding functions from package net.
type DialerTCP struct{}

func (d DialerTCP) Dial(address string) (net.Conn, error) {
	return net.Dial("tcp", address)
}

// Probe determines whether the indicated TCP port on the target host is open.
func Tcp(d NetDialerTCP, node string, port int) Result {
	address := fmt.Sprintf("%s:%d", node, port)
	conn, err := d.Dial(address)
	if err != nil {
		vdiag.Out(6, "Dial(tcp:%s) returned \"%v\".\n", address, err)
		return closed
	}
	conn.Close()
	return open
}

// NetUDPConn is the interface which net.UDPConn implements; any type which
// implements this interface implements both net.Conn and net.PacketConn
type NetUDPConn interface {
	net.PacketConn
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	RemoteAddr() net.Addr
}

// Wrap the UDP support in package net in an interface so that we can mock it
// for testing.
type NetDialerUDP interface {
	Dial(address string) (NetUDPConn, error)
}

// DialerUDP implements the NetDialerUDP interface by invoking the
// corresponding functions from package net.
type DialerUDP struct{}

func (d DialerUDP) Dial(address string) (NetUDPConn, error) {
	conn, err := net.Dial("udp", address)
	return conn.(*net.UDPConn), err
}

// Probe determines whether the indicated UDP port on the target host is open.
func Udp(d NetDialerUDP, node string, port int) Result {
	address := fmt.Sprintf("%s:%d", node, port)
	conn, err := d.Dial(address)
	if err != nil {
		vdiag.Out(6, "Dial(udp:%s) returned \"%v\".\n", address, err)
		// We failed to establish a connection...if this can ever
		// happen, assume it means the port is closed.
		return closed
	}
	defer conn.Close()

	// If the target socket just happens to be the one that we're sending
	// from (e.g., the target host is localhost), then that socket would
	// be closed if we weren't using it.
	if address == conn.LocalAddr().String() {
		vdiag.Out(5, "Probing myself!\n")
		return closed
	}

	m := []byte("Some UDP message")
	n, err := conn.Write(m)
	if err != nil || n != len(m) {
		panic(fmt.Sprintf("UDP Write() returned %d, \"%v\".\n", n, err))
	}

	err = conn.SetReadDeadline(time.Now().Add(readTimeout))
	if err != nil {
		panic(fmt.Sprintf("UDP SetReadDeadline() returned \"%v\".\n",
			err))
	}

	var buf bytes.Buffer
	n, _, err = conn.ReadFrom(buf.Bytes())
	if err != nil {
		if err.(net.Error).Timeout() {
			// There was something there, but it declined to
			// respond in a timely fashion:  assume the port is
			// open.
			return open
		}

		// There was an error (likely "connection refused") accessing
		// the port:  assume that it is closed.
		vdiag.Out(5, "ReadFrom(%d) returned %d, \"%v\".\n",
			port, n, err)
		return closed
	}

	if n > 0 {
		// Something actually responded to our message!  The port must
		// be open.
		vdiag.Out(5, "ReadFrom(%d) returned %d, \"%v\".\n",
			port, n, buf.Bytes())
		return open
	}

	// We got a zero-length read...assume failure and that the port is
	// closed.
	vdiag.Out(5, "ReadFrom(%d) returned zero without error.\n", port)
	return closed
}
