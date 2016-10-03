// Package tcpProbe provides a function for probing TCP ports on a host.
package tcpProbe

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

// Probe determines whether the indicated port on the target host is open.
func Probe(node string, port int) Result {
	address := fmt.Sprintf("%s:%d", node, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return closed
	}
	conn.Close()
	return open
}
