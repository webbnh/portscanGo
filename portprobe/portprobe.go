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

const readTimeout = 1*time.Second

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

// Wrap net.Dial() in an interface so that we can mock it for testing.
type NetDialer interface {
	Dial(network, address string) (net.Conn, error)
}

// Dialer implements the NetDialer interface by invoking net.Dial().
type Dialer struct{}

func (d Dialer) Dial(network, address string) (net.Conn, error) {
	return net.Dial(network, address)
}

// Probe determines whether the indicated TCP port on the target host is open.
func Tcp(d NetDialer, node string, port int) Result {
	address := fmt.Sprintf("%s:%d", node, port)
	conn, err := d.Dial("tcp", address)
	if err != nil {
		return closed
	}
	conn.Close()
	return open
}

// Probe determines whether the indicated UDP port on the target host is open.
func Udp(d NetDialer, node string, port int) Result {
	address := fmt.Sprintf("%s:%d", node, port)
	conn, err := d.Dial("udp", address)
	if err != nil {
		vdiag.Out(6, "UDP Dial() returned \"%v\".\n", err)
		return closed
	}
	defer conn.Close()

	var buf bytes.Buffer
	m := []byte("Some UDP message")

	n, err := conn.Write(m)
	if err != nil || n != len(m) {
		vdiag.Out(4, "UDP Write() returned %d, \"%v\".\n", n, err)
	}

	err = conn.SetReadDeadline(time.Now().Add(readTimeout))
	if err != nil {
		vdiag.Out(4, "UDP SetReadDeadline() returned \"%v\".\n",
			err)
	}

	n, err = conn.Read(buf.Bytes())
	if err != nil {
		vdiag.Out(4, "UDP Read() returned %d, \"%v\".\n", n, err)

		if err.(net.Error).Timeout() {
			return open
		}
	}

	return closed
}
