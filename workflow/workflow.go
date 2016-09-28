// Package workflow provides an encapsulated and abstracted workflow model for
// executing a series of activities, possibly concurrently.
package workflow

import (
	_ "fmt"
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
}

// Create a new Workflow
func New() *Workflow {
	wf := new(Workflow)
	return wf
}

// Set characteristics of the workflow, such as maximum number of parallel
// actors and/or time-rate limits for starting items.
func (*Workflow) Set() {
}

// Enqueue() collects items to be executed in the specified workflow.
func (Workflow) Enqueue(item Item) {
	item.Do()
}

// Wait() causes the caller to block until the workflow items are complete.
func (Workflow) Wait() {
}
