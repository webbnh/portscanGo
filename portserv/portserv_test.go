// Unit tests for package portserv.
package portserv

import (
	"net"
	"testing"
)

func check(t *testing.T, network string, getService func(int) string) {
	var checked int
	for i := 1; i <= 65535; i++ {
		service := getService(i)
		if service != "" {
			port, err := net.LookupPort(network, service)
			if err != nil {
				t.Errorf("Port %d(%s):  error looking up port for returned service \"%s\":  %v\n", i, network, service, err)
			} else if port != i {
				t.Errorf("Port %d(%s):  got mismatched port (%d) for returned service \"%s\".\n", i, network, port, service)
			}
			checked++
		}
	}

	if checked == 0 {
		t.Errorf("No services returned for %s ports.\n", network)
	} else {
		t.Logf("Confirmed matches for %d %s ports/services.\n",
			checked, network)
	}
}

func TestTcp(t *testing.T) {
	check(t, "tcp", Tcp)
}

func TestUdp(t *testing.T) {
	check(t, "udp", Udp)
}
