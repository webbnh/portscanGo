/*
webbscan is a simple port scanner tool.

This is my first attempt at writing a Go program.  :-)
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/webbnh/DigitalOcean/portprobe"
	"github.com/webbnh/DigitalOcean/portserv"
	"github.com/webbnh/DigitalOcean/progbar"
	"github.com/webbnh/DigitalOcean/vdiag"
	"github.com/webbnh/DigitalOcean/workflow"
)

var progressBar *progbar.Bar = progbar.New(70, 65535, os.Stderr)

// workItem represents an item to be passed to the workflow (it satisfies the
// workflow.Item interface), in this case it contains the number of a port to
// to be probed and a place to write the result.
type workItem struct {
	// Closure which invokes the appropriate probe function using the
	// requested parameters (e.g., the host)
	probeFunc func(*workItem)
	// Port to be probed
	port int
	// Result of probe (e.g., open, closed, pending)
	result portprobe.Result
}

// Do is the function which the workflow.Item interface uses to initiate the
// work on the item.  Here it calls a closure which relieves us from having to
// include more fields in the item.
func (t workItem) Do(output chan<- workflow.Item) {
	vdiag.Out(8, "In Do() for port %d\n", t.port)
	t.probeFunc(&t)
	progressBar.Spin()
	output <- t
	vdiag.Out(8, "Leaving Do() for port %d, result is %v\n",
		t.port, t.result)
}

func main() {
	// Command line flags
	var (
		host     string
		protocol string
		agents   int
		rate     int
	)

	flag.StringVar(&host, "host", "127.0.0.1", "Host IP address")
	flag.StringVar(&protocol, "protocol", "tcp",
		"Protocol (\"tcp\" or \"udp\")")
	flag.IntVar(&agents, "agents", 8, "Number of concurrent probes")
	flag.IntVar(&rate, "rate", 0, "Maximum number of probes per second (0: unlimited)")
	flag.Parse()

	switch protocol {
	case "tcp":
	case "udp":
	default:
		fmt.Fprintf(os.Stderr, "\"%s\" protocol is not supported.\n",
			protocol)
		os.Exit(-1)
	}

	vdiag.Out(1, "Scanning for open %s ports on %s using %d agents.\n",
		protocol, host, agents)
	if rate != 0 {
		vdiag.Out(1, "Probe rate limited to %d probes per second.\n",
			rate)
	}
	if vdiag.Verbosity() > 0 {
		fmt.Printf("(Diagnostic messages verbosity level %d.)\n",
			vdiag.Verbosity())
	}

	wfItems := [65535]workItem{}
	wf := workflow.New(cap(wfItems), agents, rate)

	progressBar.Paint()

	start := time.Now()
	// Request a scan of each (and all) of the ports.
	for i := range wfItems {
		wfItems[i].port = i + 1 // Initialize for later

		// Capture the host and port to be probed, as well as the
		// place to record the result, using a closure.
		port := wfItems[i].port
		wfItems[i].probeFunc = func(item *workItem) {
			vdiag.Out(7, "Calling probe for %s:%d\n", host, port)
			// FIXME:  This logic should be moved to portprobe.
			switch protocol {
			case "tcp":
				item.result = portprobe.Tcp(
					portprobe.DialerTCP{}, host, port)
			case "udp":
				item.result = portprobe.Udp(
					portprobe.DialerUDP{}, host, port)
			}
		}

		// Send the item off to be independently executed.
		vdiag.Out(6, "Queuing item %d.\n", i)
		wf.Enqueue(wfItems[i])
	}

	// Wait for the scans to complete.
	for i := range wfItems {
		progressBar.Update()

		// Since the items are executed concurrently, they may
		// complete out of order.  We're done when all the scans have
		// finished, so check each one.  If the target item is not
		// complete, delay by trying to get a completed one.  What we
		// actually get back is a copy, so propagate its result into
		// the appropriate slot in the array.
		for !wfItems[i].result.IsComplete() {
			vdiag.Out(5, "Scan of port %d is not ready, waiting...",
				wfItems[i].port)
			item := wf.Dequeue().(workItem)
			wfItems[item.port-1].result = item.result
			vdiag.Out(5, "got %v.\n", item)
		}
	}

	elapsed := time.Now().Sub(start)
	progressBar.Done()

	// Print the result
	printedHeader := false
	for _, v := range wfItems {
		if v.result.IsOpen() {
			if !printedHeader {
				fmt.Printf("Open %s ports on %s:\n",
					protocol, host)
				printedHeader = true
			}
			portStr := portserv.Tcp(v.port)
			if portStr == "" {
				fmt.Println(v.port)
			} else {
				fmt.Printf("%d (%s)\n", v.port, portStr)
			}
		}
	}
	if !printedHeader {
		fmt.Printf("No open %s ports on %s.\n", protocol, host)
	}

	wf.Destroy()
	vdiag.Out(1, "Elapsed time: %v.\n", elapsed)
	if len(wfItems)/int(elapsed/time.Second) > 0 {
		vdiag.Out(1, "Average probe rate: %d probes/second.\n",
			len(wfItems)/int(elapsed/time.Second))
	} else {
		vdiag.Out(1, "Average probe rate: %v/probe.\n",
			elapsed/time.Duration(len(wfItems)))
	}
}
