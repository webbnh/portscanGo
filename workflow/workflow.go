// Package workflow provides an encapsulated and abstracted workflow model for
// executing a series of activities, possibly concurrently.
package workflow

import (
	"fmt"
	"sync/atomic"
	"time"
)

// A type which satisfies the workflow.Item interface can be handed off to this
// package to be executed independently of the caller and other items in the
// workflow.
type Item interface {
	// The Do method initiates the work on the item.
	Do()
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

	// Performance counters
	unthrottled int32
	done        int32
}

// Create a new Workflow, specifying total number of items, the maximum number
// of items to be executed concurrently, and the maximum number of items to
// start per second.
func New(size, maxActors, maxRate int) *Workflow {
	wf := new(Workflow)
	wf.input = make(chan Item, size)
	wf.output = make(chan Item, size)
	if maxRate == 0 {
		wf.interval = 0
	} else {
		wf.interval = time.Second / time.Duration(maxRate)
	}
	wf.throttle = time.NewTimer(wf.interval)
	for i := 0; i < maxActors; i++ {
		go wf.act()
	}
	return wf
}

// Destroy the workflow, releasing its resources for garbage collection
func (wf Workflow) Destroy() {
	close(wf.input)
	close(wf.output)
	fmt.Printf("Performed %d operations, %d without throttling.\n",
		wf.done, wf.unthrottled)
}

// act pulls items from the input queue and executes them until the flow is
// complete.
func (wf *Workflow) act() {
	for {
		// The length of an unbuffered channel is either zero or one;
		// count how many times we proceeded unthrottled
		t := int32(len(wf.throttle.C))

		// Wait for an interval (to avoid issuing requests too quickly)
		// then request a new interval timer for the next actor
		<-wf.throttle.C
		wf.throttle = time.NewTimer(wf.interval)

		// Get an item from the input queue, execute it, and queue it
		// to the output queue; if the input queue is closed, exit.
		item, ok := <-wf.input
		if !ok {
			return
		}
		item.Do()
		atomic.AddInt32(&wf.unthrottled, t)
		atomic.AddInt32(&wf.done, 1)
		wf.output <- item
	}
}

// Enqueue() collects items to be executed in the specified workflow.
func (wf Workflow) Enqueue(item Item) {
	wf.input <- item
}

// Wait() causes the caller to block until the workflow items are complete.
func (wf *Workflow) Wait() {
	// As long as there is pending input items or active executions,
	// wait for completions
	for wf.done < int32(cap(wf.output)) {
		<-wf.output
	}
}
