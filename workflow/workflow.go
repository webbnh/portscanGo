// Package workflow provides an encapsulated and abstracted workflow model for
// executing a series of activities, possibly concurrently.
package workflow

import (
	"sync/atomic"
	"time"

	"github.com/webbnh/DigitalOcean/vdiag"
)

// A type which satisfies the workflow.Item interface can be handed off to this
// package to be executed independently of the caller and other items in the
// workflow.
type Item interface {
	// The Do method initiates the work on the item and sends the result
	// to the provided chanel.
	Do(chan<- Item)
}

// Workflow represents and controls the flow of work.  Multiple independent
// workflows may be created and active concurrently.
type Workflow struct {
	// Queues of pending and completed work items
	input, output chan Item
	// Interval between starting work items for rate throttling
	interval time.Duration
	// Timer used for rate throttling
	throttle *time.Timer
	// Number of completed items
	done int32
	// Number of items initiated without throttling
	unthrottled int32
}

// Create a new Workflow, specifying total number of items, the maximum number
// of items to be executed concurrently, and the maximum number of items to
// start per second.
func New(size, maxActors, maxRate int) *Workflow {
	wf := new(Workflow)
	wf.input = make(chan Item, size)
	wf.output = make(chan Item, size)
	if maxRate != 0 {
		wf.interval = time.Second / time.Duration(maxRate)
		wf.throttle = time.NewTimer(wf.interval)
	}
	for i := 0; i < maxActors; i++ {
		go wf.act()
	}
	return wf
}

// Destroy the workflow, releasing its resources for garbage collection.
func (wf Workflow) Destroy() {
	close(wf.input)
	close(wf.output)
	vdiag.Out(3, "Performed %d operations, %d without throttling.\n",
		wf.done, wf.unthrottled)
}

// act pulls items from the input queue and executes them until the flow is
// complete.
func (wf *Workflow) act() {
	for {
		t := int32(1) // For the unthrottled & no-throttle cases
		if wf.throttle != nil {
			select {
			case <-wf.throttle.C:
				// We were able to receive without
				// blocking, so we were not throttled.
			default:
				// We are being throttled -- wait for
				// our turn
				t = 0
				<-wf.throttle.C
			}

			// Request a new interval timer for the next actor.
			wf.throttle = time.NewTimer(wf.interval)
		}

		// Get an item from the input queue and execute it (which
		// should queue it to the output queue); if the input queue is
		// closed, exit.
		item, ok := <-wf.input
		if !ok {
			return
		}
		item.Do(wf.output)
		atomic.AddInt32(&wf.done, 1)
		atomic.AddInt32(&wf.unthrottled, t)
	}
}

// Enqueue() collects items to be executed in the specified workflow.
func (wf Workflow) Enqueue(item Item) {
	wf.input <- item
}

// Dequeue() returns a completed items from the specified workflow.
func (wf Workflow) Dequeue() Item {
	return <-wf.output
}

// Wait() causes the caller to block until the workflow items are complete.
func (wf *Workflow) Wait() {
	// As long as there is pending input items or active executions,
	// wait for completions.
	for wf.done < int32(cap(wf.output)) {
		<-wf.output
	}
}
