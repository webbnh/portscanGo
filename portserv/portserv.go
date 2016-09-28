// Package portserv provides interfaces which translate TCP and UDP ports into
// service names.
package portserv

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
)

var tcpServices map[int]string = make(map[int]string)
var udpServices map[int]string = make(map[int]string)

func init() {
	re := regexp.MustCompile(`\n(\S+)\s+(\d+)/(tcp|udp)`)
	text, err := ioutil.ReadFile("/etc/services")
	if err != nil {
		panic(err)
	}
	matches := re.FindAllStringSubmatch(string(text), -1)
	if matches == nil {
		panic("Failed to parse services file")
	}

	for _, v := range matches {
		port, err := strconv.Atoi(v[2])
		if err != nil {
			panic(fmt.Sprintf("regex match returned non-numeric number: \"%s\"", v[1]))
		}
		switch v[3] {
		case "tcp":
			tcpServices[port] = v[1]
		case "udp":
			udpServices[port] = v[1]
		default:
			panic(fmt.Sprintf("Unexpected protocol value: \"%s\"",
				v[3]))
		}
	}
}

func Tcp(port int) string { return tcpServices[port] }
func Udp(port int) string { return udpServices[port] }
