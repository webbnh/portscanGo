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

// mockDialer implements the portprobe.NetDialer interface
type mockDialer struct {
	t               *testing.T
	expectedNetwork string
	expectedAddress string
	conn            net.Conn
	err             error
}

func (d mockDialer) Dial(network, address string) (net.Conn, error) {
	if network != d.expectedNetwork {
		d.t.Errorf("Dial() received network \"%s\"; expected \"%s\".\n",
			network, d.expectedNetwork)
	}
	if address != d.expectedAddress {
		d.t.Errorf("Dial() received address \"%s\"; expected \"%s\".\n",
			address, d.expectedAddress)
	}

	return d.conn, d.err
}

// Mock Conn implements the net.Conn interface
type mockConn struct {
	t           *testing.T
	network string
	calledClose *bool
}

func (d mockConn) Read(b []byte) (n int, err error) {
	if d.network != "udp" {
		d.t.Fatal("mockConn.Read() unexpectedly called.")
	}
	return
}

func (d mockConn) Write(b []byte) (n int, err error) {
	if d.network != "udp" {
		d.t.Fatal("mockConn.Write() unexpectedly called.")
	}
	return
}

func (d mockConn) Close() error {
	*d.calledClose = true
	return nil
}

func (d mockConn) LocalAddr() (addr net.Addr) {
	d.t.Fatal("mockConn.LocalAddr() unexpectedly called.")
	return
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
	return
}

func (d mockConn) SetWriteDeadline(t time.Time) (err error) {
	d.t.Fatal("mockConn.SetWriteDeadline() unexpectedly called.")
	return
}

func TestTcp(t *testing.T) {
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

	for _, v := range cases {
		t.Logf("Testing %s(%s) with error \"%v\".\n",
			v.address, v.network, v.err)
		calledClose := false
		dialer := mockDialer{t, v.network, v.address,
			mockConn{t, v.network, &calledClose}, v.err}
		got := Tcp(dialer, node, port)
		if got != v.result {
			t.Errorf("Probe returned %v; expected %v for error \"%v\".\n",
				got, v.result, v.err)
		} else if calledClose != v.calledClose {
			notStr := ""
			if v.calledClose {
				notStr = "not "
			}
			t.Errorf("Close() was unexpectedly %scalled.\n", notStr)
		}
	}
}

func TestUdp(t *testing.T) {
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
		{"udp", address, nil, open, true},
		{"udp", address, errors.New("Connection failed"), closed, false},
	}

	for _, v := range cases {
		t.Logf("Testing %s(%s) with error \"%v\".\n",
			v.address, v.network, v.err)
		calledClose := false
		dialer := mockDialer{t, v.network, v.address,
			mockConn{t, v.network, &calledClose}, v.err}
		got := Udp(dialer, node, port)
		if got != v.result {
			t.Errorf("Probe returned %v; expected %v for error \"%v\".\n",
				got, v.result, v.err)
		} else if calledClose != v.calledClose {
			notStr := ""
			if v.calledClose {
				notStr = "not "
			}
			t.Errorf("Close() was unexpectedly %scalled.\n", notStr)
		}
	}
}
