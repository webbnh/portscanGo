// tcpProbe is a package which provides functions for probing TCP ports on a host
package tcpProbe

import (
	"fmt"
	"net"
)

// Probe tries to determine whether the indicated port on the target host is open
func Probe(node string, port int) bool {
	address := fmt.Sprintf("%s:%d", node, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
