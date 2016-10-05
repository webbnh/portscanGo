// Unit tests for package workflow.
//
// TODO:  These tests are not exhaustive (e.g., they don't test several
// corner cases such as enqueuing too many items, nor many cases which
// block or are otherwise asynchronous such as waiting in Dequeue()).
// And, there's no test for Destroy().
package workflow

import (
	"testing"
)

// Struct testItem implements workflow.Item
type testItem struct {
	id   int
	done bool
}

// Do marks the item as done and puts it on the output queue
func (item testItem) Do(output chan<- Item) {
	item.done = true
	output <- item
}

// Test New()
func TestNew(t *testing.T) {
	const expectedSize = 10

	wf := New(expectedSize, 0, 0)

	if wf == nil {
		t.Error("New returned nil.")
		return
	}

	if wf.input == nil {
		t.Error("input channel is nil.")
	} else if cap(wf.input) != expectedSize {
		t.Errorf("input channel size is %d; expected %d\n",
			cap(wf.input), expectedSize)
	}

	if wf.output == nil {
		t.Error("output chanel is nil.")
	} else if cap(wf.output) != expectedSize {
		t.Errorf("output channel size is %d; expected %d\n",
			cap(wf.output), expectedSize)
	}

	if wf.interval != 0 {
		t.Errorf("Rate is %v; expected zero\n", wf.interval)
	}
}

// Test Enqueue(), act(), and Dequeue() processing the queue serially.
func TestFlowSerial(t *testing.T) {
	const itemCount = 10

	// Start no actors:  we'll call act() directly
	wf := New(itemCount, 0, 0)

	for i := 1; i <= itemCount; i++ {
		wf.Enqueue(testItem{id: i})
	}

	close(wf.input) // Cause act() to return when finished
	wf.act()

	for i := 1; i <= itemCount; i++ {
		item := wf.Dequeue().(testItem)
		if !item.done {
			t.Errorf("Item #%d was completed but not marked done.\n",
				i)
		}
		if item.id != i {
			t.Errorf("Item #%d has mismatched id (%d).\n",
				i, item.id)
		}
	}

	if wf.done != itemCount {
		t.Errorf("Done is %d; expected %d.\n", wf.done, itemCount)
	}
}

// Test Enqueue() and Dequeue() (and act()) processing the queue concurrently
// with the enqueue and dequeue operations.
func TestFlowConcurrent(t *testing.T) {
	const itemCount = 10

	// Start only one actor to ensure the items are execute in sequence.
	wf := New(itemCount, 1, 0)

	for i := 1; i <= itemCount; i++ {
		wf.Enqueue(testItem{id: i})
	}

	for i := 1; i <= itemCount; i++ {
		item := wf.Dequeue().(testItem)
		if !item.done {
			t.Errorf("Item #%d was completed but not marked done.\n",
				i)
		}
		if item.id != i {
			t.Errorf("Item #%d has mismatched id (%d).\n",
				i, item.id)
		}
	}

	if wf.done != itemCount {
		t.Errorf("Done is %d; expected %d.\n", wf.done, itemCount)
	}
}

// Test Wait()
func TestWait(t *testing.T) {
	const itemCount = 10

	wf := New(itemCount, 1, 0) // Create one actor

	for i := 1; i <= itemCount; i++ {
		wf.Enqueue(testItem{id: i})
	}

	wf.Wait()

	if wf.done != itemCount {
		t.Errorf("Done is %d; expected %d.\n", wf.done, itemCount)
	}
}
