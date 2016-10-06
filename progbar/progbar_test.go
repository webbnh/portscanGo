// Unit tests for package progbar.
package progbar

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)

const (
	expectedWidth = 72
	expectedSize  = 1001
)

func TestNew(t *testing.T) {
	expectedWriter := &os.Stderr

	if New(0, expectedSize, *expectedWriter) != nil {
		t.Error("New() succeeded with zero width.")
	}
	if New(expectedWidth, 0, *expectedWriter) != nil {
		t.Error("New() succeeded with zero size.")
	}
	if New(expectedWidth, expectedSize, nil) != nil {
		t.Error("New() succeeded with nil Writer.")
	}

	bar := New(expectedWidth, expectedSize, *expectedWriter)
	if bar == nil {
		t.Fatal("New() unexpectedly failed.")
	}
	if bar.width != expectedWidth {
		t.Errorf("bar.width is %d; expected %d.\n",
			bar.width, expectedWidth)
	}
	if bar.total != expectedSize {
		t.Errorf("bar.size is %d; expected %d.\n",
			bar.total, expectedSize)
	}
	if !reflect.DeepEqual(bar.w, *expectedWriter) {
		t.Errorf("bar.w is %v; expected %v.\n",
			bar.w, *expectedWriter)
	}
	if bar.current != 0 {
		t.Errorf("bar.current is %d; expected zero\n", bar.current)
	}
}

// Test Paint
//
// TODO:  test "repaint" scenario
func TestPaint(t *testing.T) {
	var buf bytes.Buffer

	bar := New(expectedWidth, expectedSize, &buf)

	bar.Paint()

	got := buf.Bytes()
	exp := bytes.Repeat([]byte{' '}, expectedWidth)
	exp = append(exp, "|\r|"...)

	if bytes.Compare(got, exp) != 0 {
		for i := range got {
			if i >= len(exp) {
				t.Errorf("Got excess character ('%c') at position %d.\n",
					got[i], i)
				continue
			}
			if got[i] != exp[i] {
				t.Errorf("Got '%c' at position %d; expected '%c'.\n",
					got[i], i, exp[i])
			}
		}
		if len(got) < len(exp) {
			t.Errorf("Characters missing after position %d; expected %d.\n",
				len(got), len(exp))
		}
	}
}

func TestUpdate(t *testing.T) {
	t.Error("Not implemented")
}

func TestDone(t *testing.T) {
	t.Error("Not implemented")
}

func TestSpin(t *testing.T) {
	t.Error("Not implemented")
}
