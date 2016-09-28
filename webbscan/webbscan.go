/*
webbscan is a simple port scanner tool.

This is my first attempt at writing a Go program.  :-)
*/
package main

import (
	"flag"
	"fmt"

	"github.com/webbnh/DigitalOcean/tcpProbe"
	"github.com/webbnh/DigitalOcean/workflow"
)

// workItem represents an item to be passed to the workflow (it satisfies the
// workflow.Item interface), in this case it contains the number of a port to
// to be probed and a place to write the result.
type workItem struct {
	probeFunc func()
	port      int
	open      bool
}

// Do is the function which the workflow.Item interface uses to initiate the
// work on the item.
func (t workItem) Do() {
	t.probeFunc()
}

func init() {
}

func main() {
	host := flag.String("host", "127.0.0.1", "Host IP address")
	protocol := flag.String("protocol", "tcp",
		"Protocol (\"tcp\" or \"udp\")")
	flag.Parse()

	if *protocol != "tcp" {
		// UDP probe requires sending a packet.
		panic("Only TCP protocol is currently supported")
	}

	wf := workflow.New()

	wfItems := [65535]workItem{}

	for i := range wfItems {
		// Capture the host and port to be probed, as well as the
		// place to record the result, using a closure.
		wfItems[i].port = i + 1
		wfItems[i].probeFunc = func() {
			wfItems[i].open = tcpProbe.Probe(*host, wfItems[i].port)
		}

		// Send the item off to be independently executed.
		wf.Enqueue(wfItems[i])
	}

	wf.Wait()

	fmt.Printf("Open %s ports on %s:\n", *protocol, *host)
	for _, v := range wfItems {
		if v.open {
			fmt.Println(v.port)
		}
	}
}
