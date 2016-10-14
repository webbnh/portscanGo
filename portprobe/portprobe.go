// Package portprobe provides support for probing ports on a host.
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

// Result is the result of the probe; the appropriate "IsXXXX()" function
// should be used to evaluate it.
type Result int

// IsClosed returns a boolean indicating whether the port is closed.
func (r Result) IsClosed() bool { return r == closed }

// IsComplete returns a boolean indicating whether the probe has completed.
func (r Result) IsComplete() bool { return r != pending }

// IsClosed returns a boolean indicating whether the port is open.
func (r Result) IsOpen() bool { return r == open }

// String returns the probe result as a string.
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

// netDialerTCP wraps the TCP version of net.Dial() in an interface so that we
// can mock it for testing.
type netDialerTCP interface {
	Dial(address string) (net.Conn, error)
}

// dialerTCP implements the netDialerTCP interface by invoking the
// corresponding functions from package net.
type dialerTCP struct{}

func (d dialerTCP) Dial(address string) (net.Conn, error) {
	return net.Dial("tcp", address)
}

// probeTcp determines whether the indicated TCP port on the target host is
// open.
func probeTcp(d netDialerTCP, node string, port int) Result {
	address := fmt.Sprintf("%s:%d", node, port)
	conn, err := d.Dial(address)
	if err != nil {
		vdiag.Out(6, "Dial(tcp:%s) returned \"%v\".\n", address, err)
		return closed
	}
	conn.Close()
	return open
}

// netUDPConn is the interface which net.UDPConn implements; any type which
// implements this interface implements both net.Conn and net.PacketConn
type netUDPConn interface {
	net.PacketConn
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	RemoteAddr() net.Addr
}

// netDialerUDP wraps the UDP support in package net in an interface so that
// we can mock it for testing.
type netDialerUDP interface {
	Dial(address string) (netUDPConn, error)
}

// dialerUDP implements the netDialerUDP interface by invoking the
// corresponding functions from package net.
type dialerUDP struct{}

func (d dialerUDP) Dial(address string) (netUDPConn, error) {
	conn, err := net.Dial("udp", address)
	return conn.(*net.UDPConn), err
}

// Probe determines whether the indicated UDP port on the target host is open.
func probeUdp(d netDialerUDP, node string, port int) Result {
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

// Instances of the probe functions which can be overridden for unit testing.
var (
	probeFuncTCP = probeTcp
	probeFuncUDP = probeUdp
)

// Probe determines whether the specified port on the on the specified host is
// potentially accepting input via the specified network protocol.
func Probe(protocol, host string, port int) Result {
	switch protocol {
	case "tcp":
		return probeFuncTCP(dialerTCP{}, host, port)
	case "udp":
		return probeFuncUDP(dialerUDP{}, host, port)
	default:
		vdiag.Out(2, "Probe:  unexpected protocol, \"%s\".'n", protocol)
		return pending
	}
}
