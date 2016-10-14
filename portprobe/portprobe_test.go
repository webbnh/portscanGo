// Unit tests for package portprobe
package portprobe

import (
	"errors"
	"net"
	"strconv"
	"testing"
	"time"
)

var results = []Result{closed, pending, open}

func TestIsComplete(t *testing.T) {
	for _, v := range results {
		expected := (v != pending)
		got := v.IsComplete()
		if got != expected {
			t.Errorf("%v.IsComplete() unexpectedly returned %v.\n",
				v, got)
		}
	}
}

func TestIsOpen(t *testing.T) {
	for _, v := range results {
		expected := (v == open)
		got := v.IsOpen()
		if got != expected {
			t.Errorf("%v.IsOpen() unexpectedly returned %v.\n",
				v, got)
		}
	}
}

func TestIsClosed(t *testing.T) {
	for _, v := range results {
		expected := (v == closed)
		got := v.IsClosed()
		if got != expected {
			t.Errorf("%v.IsClosed() unexpectedly returned %v.\n",
				v, got)
		}
	}
}

func TestString(t *testing.T) {
	for _, v := range results {
		var expected string
		switch v {
		case closed:
			expected = "closed"
		case pending:
			expected = "pending"
		case open:
			expected = "open"
		default:
			t.Fatalf("Unexpected value for type Result:  %v.\n", v)
		}
		got := v.String()
		if got != expected {
			t.Errorf("%#v.String() returned \"%s\" when expecting \"%s\".\n", v, got, expected)
		}
	}
}

// mockDialerTCP implements the NetDialerTCP interface
type mockDialerTCP struct {
	t               *testing.T
	expectedAddress string
	err             error
	conn            net.Conn
}

func (d mockDialerTCP) Dial(address string) (net.Conn, error) {
	if address != d.expectedAddress {
		d.t.Errorf("Dial(tcp) got address \"%s\"; expected \"%s\".\n",
			address, d.expectedAddress)
	}

	return d.conn, d.err
}

// mockDialerUDP implements the NetDialerUDP interface
type mockDialerUDP struct {
	t               *testing.T
	expectedAddress string
	err             error
	conn            NetUDPConn
}

func (d mockDialerUDP) Dial(address string) (NetUDPConn, error) {
	if address != d.expectedAddress {
		d.t.Errorf("Dial(udp) got address \"%s\"; expected \"%s\".\n",
			address, d.expectedAddress)
	}

	return d.conn, d.err
}

// mockAddr implements the net.Addr interface
type mockAddr struct {
	t     *testing.T
	laddr string
}

func (a mockAddr) String() string {
	return a.laddr
}

func (a mockAddr) Network() string {
	a.t.Fatal("mockAddr.Network() unexpectedly called.")
	return ""
}

// mockConn implements both of the net.Conn and net.PacketConn interfaces, and
// so it implements the NetUDPConn interface
type mockConn struct {
	t            *testing.T
	network      string
	laddr        string
	readN        int
	readErr      error
	calledClose  *bool
	calledWrite  *bool
	calledSetRDL *bool
	addr         mockAddr
}

func (d mockConn) Read(b []byte) (n int, err error) {
	d.t.Fatal("mockConn.Read() unexpectedly called.")
	return
}

func (d mockConn) Write(b []byte) (n int, err error) {
	if d.network != "udp" {
		d.t.Fatal("mockConn.Write() unexpectedly called.")
	}
	*d.calledWrite = true
	return len(b), nil
}

func (d mockConn) Close() error {
	*d.calledClose = true
	return nil
}

func (d mockConn) LocalAddr() (addr net.Addr) {
	if d.network != "udp" {
		d.t.Fatal("mockConn.LocalAddr() unexpectedly called.")
	}
	return mockAddr{d.t, d.laddr}
}

func (d mockConn) RemoteAddr() (addr net.Addr) {
	d.t.Fatal("mockConn.RemoteAddr() unexpectedly called.")
	return
}

func (d mockConn) SetDeadline(t time.Time) (err error) {
	d.t.Fatal("mockConn.SetDeadline() unexpectedly called.")
	return
}

func (d mockConn) SetReadDeadline(t time.Time) (err error) {
	if d.network != "udp" {
		d.t.Fatal("mockConn.SetReadDeadline() unexpectedly called.")
	}
	*d.calledSetRDL = true
	return
}

func (d mockConn) SetWriteDeadline(t time.Time) (err error) {
	d.t.Fatal("mockConn.SetWriteDeadline() unexpectedly called.")
	return
}

func (d mockConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	if d.network != "udp" {
		d.t.Fatal("mockConn.Read() unexpectedly called.")
	}
	return d.readN, nil, d.readErr
}

func (d mockConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	d.t.Fatal("mockConn.SetWriteDeadline() unexpectedly called.")
	return
}

func checkCalled(t *testing.T, got, exp bool, n int, tag string) {
	if got != exp {
		notStr := ""
		if !got {
			notStr = "not "
		}
		t.Errorf("Case #%d: %s() was unexpectedly %scalled.\n",
			n, tag, notStr)
	}
}

func TestProbeTcp(t *testing.T) {
	const node = "127.0.0.1"
	const port = 0

	address := node + ":" + strconv.Itoa(port)
	cases := []struct {
		network     string
		address     string
		err         error
		result      Result
		calledClose bool
	}{
		{"tcp", address, nil, open, true},
		{"tcp", address, errors.New("Connection failed"), closed, false},
	}

	for i, v := range cases {
		calledClose := false
		dialer := mockDialerTCP{t, v.address, v.err,
			mockConn{t, v.network, "", 0, nil, &calledClose, nil,
				nil, mockAddr{}}}
		got := probeTcp(dialer, node, port)
		if got != v.result {
			t.Errorf("Case #%d: Probe returned %v; expected %v for error \"%v\".\n",
				i, got, v.result, v.err)
		}
		checkCalled(t, calledClose, v.calledClose, i, "Close")
	}
}

type timeoutErr struct {
	t          *testing.T
	wasTimeout bool
}

func (te timeoutErr) Timeout() bool {
	return te.wasTimeout
}

func (te timeoutErr) Temporary() bool {
	te.t.Fatal("timeoutErr.Temporary() was unexpectedly called.")
	return false
}

func (te timeoutErr) Error() string {
	te.t.Fatal("timeoutErr.Temporary() was unexpectedly called.")

	if te.wasTimeout {
		return "There was a timeout."
	}
	return "There was no timeout."
}

func TestProbeUdp(t *testing.T) {
	const node = "127.0.0.1"
	const lport = 0
	const rport = 1

	raddress := node + ":" + strconv.Itoa(rport)
	laddress := node + ":" + strconv.Itoa(lport)

	cases := []struct {
		network      string
		raddr        string
		laddr        string
		dialErr      error
		readErr      error
		readRet      int
		result       Result
		calledClose  bool
		calledWrite  bool
		calledSetRDL bool
	}{
		// Dial() fails (returns non-nil error), result: closed
		{"udp", raddress, laddress, errors.New("Connection failed"),
			nil, 0, closed, false, false, false},
		// Target address equals source address, result: closed
		{"udp", raddress, raddress, nil, nil, 0, closed, true, false,
			false},
		// The read times out, result: open
		{"udp", raddress, laddress, nil, timeoutErr{t, true}, 0, open,
			true, true, true},
		// The read returns (non-timeout) error, result: closed
		{"udp", raddress, laddress, nil, timeoutErr{t, false}, 0,
			closed, true, true, true},
		// The read succeeds (non-zero length), result: open
		{"udp", raddress, laddress, nil, nil, 10, open, true, true,
			true},
		// The read returns zero length, result closed
		{"udp", raddress, laddress, nil, nil, 0, closed, true, true,
			true},
	}

	for i, v := range cases {
		t.Logf("Testing case #%d.\n", i)
		calledClose := false
		calledWrite := false
		calledSetRDL := false
		dialer := mockDialerUDP{t, v.raddr, v.dialErr,
			&mockConn{t, v.network, v.laddr, v.readRet,
				v.readErr, &calledClose, &calledWrite,
				&calledSetRDL, mockAddr{t, v.laddr}}}
		got := probeUdp(dialer, node, rport)
		if got != v.result {
			t.Errorf("Case #%d: Probe returned %v; expected %v.\n",
				i, got, v.result)
		}
		checkCalled(t, calledClose, v.calledClose, i, "Close")
		checkCalled(t, calledWrite, v.calledWrite, i, "Write")
		checkCalled(t, calledSetRDL, v.calledSetRDL, i,
			"SetReadDeadline")
	}
}

func TestProbe(t *testing.T) {
	host := "localhost"
	port := 42
	cases := []struct {
		protocol       string
		result         Result
		calledProbeTcp bool
		calledProbeUdp bool
	}{
		{"tcp", open, true, false},
		{"tcp", closed, true, false},
		{"udp", open, false, true},
		{"udp", closed, false, true},
		{"bad", pending, false, false},
	}

	for i, v := range cases {
		calledProbeTcp := false
		calledProbeUdp := false
		t.Logf("Testing case #%d.\n", i)

		checkHostPort := func(gotHost string, gotPort int) {
			if gotHost != host {
				t.Errorf("Case #%d: "+
					"Got host \"%s\"; expected \"%s\".\n",
					i, gotHost, host)
			}
			if gotPort != port {
				t.Errorf("Case #%d: "+
					"Got port \"%d\"; expected \"%d\".\n",
					i, gotPort, port)
			}
		}

		probeFuncTCP = func(d NetDialerTCP, gotHost string, gotPort int) Result {
			// I assume the compiler checking will suffice for the
			// dialer parameter.
			checkHostPort(gotHost, gotPort)
			calledProbeTcp = true
			return v.result
		}

		probeFuncUDP = func(d NetDialerUDP, gotHost string, gotPort int) Result {
			// I assume the compiler checking will suffice for the
			// dialer parameter.
			checkHostPort(gotHost, gotPort)
			calledProbeUdp = true
			return v.result
		}

		got := Probe(v.protocol, host, port)
		if got != v.result {
			t.Errorf("Case #%d: Probe returned %v; expected %v.\n",
				i, got, v.result)
		}

		checkCalled(t, calledProbeTcp, v.calledProbeTcp, i, "probeTcp")
		checkCalled(t, calledProbeUdp, v.calledProbeUdp, i, "probeUdp")
	}
}
