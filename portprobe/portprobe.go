// Package portprobe provides functions for probing ports on a host.
package portprobe

import (
	"fmt"
	"net"
)

// Port statuses
const (
	closed = iota - 1
	pending
	open
)

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

// Probe determines whether the indicated port on the target host is open.
func Probe(d NetDialer, node string, port int) Result {
	address := fmt.Sprintf("%s:%d", node, port)
	conn, err := d.Dial("tcp", address)
	if err != nil {
		return closed
	}
	conn.Close()
	return open
}
