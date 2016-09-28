// Unit tests for webbscan
package main

import (
	"testing"

	"github.com/webbnh/DigitalOcean/tcpProbe"
	"github.com/webbnh/DigitalOcean/workflow"
)

func TestDo(t *testing.T) {
	cases := []struct {
		testPort  int
		expResult tcpProbe.Result
	}{
		{1, -1},
		{2, 1},
		{3, 0},
		{8, 1},
		{5, -1},
	}

	for _, v := range cases {
		item := workItem{
			probeFunc: func(i *workItem) {
				v := v // Capture for the closure.
				if i.port != v.testPort {
					t.Errorf("Got port %d; expected %d "+
						"(case %v).\n",
						i.port, v.testPort, v)
				}
				i.result = v.expResult
			},
			port: v.testPort,
		}
		outChan := make(chan workflow.Item, 10)

		item.Do(outChan)

		switch len(outChan) {
		case 0:
			t.Errorf("Work item %v failed to produce an output.\n",
				v)
			continue
		default:
			t.Errorf("Work item %v produced %d outputs "+
				"(expected only 1).\n",
				v, len(outChan))
			continue
		case 1: // Expected result; execute the rest of the loop
			break
		}

		outItem := (<-outChan).(workItem)

		if outItem.result != v.expResult {
			t.Errorf("Got result \"%v\"; expected \"%v\" "+
				"(case %v).\n",
				outItem.result, v.expResult, v)
		}
	}
}

// Test the main function
func TestWebbscan(t *testing.T) {
	t.Log("I punted on unit-testing main() -- " +
		"I'll leave that to the \"integration\" suite.\n")
}
