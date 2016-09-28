// workflow is a package which provides an encapsulated and abstracted workflow
// model for executing a series of activities (presented as a slice), possibly
// concurrently.
package workflow

import (
	_ "fmt"
)

// Item provides the interface required to execute each item in the
// workflow.
type Item interface {
	Do()
}

type Workflow struct {
}

// Create a new Workflow
func New() *Workflow {
	wf := new(Workflow)
	return wf
}

// Todo:  this function will be use to set characteristics of the workflow,
// such as maximum number of parallel actors and/or time-rate limits.
func (*Workflow) Set() {
}

// Enqueue() collects items to be executed in the specified Workflow; items may
// be executed immediately or executed concurrently with the caller.
func (Workflow) Enqueue(item Item) {
	item.Do()
}

// Wait() causes the caller to block until the workflow items are complete.
func (Workflow) Wait() {
}
