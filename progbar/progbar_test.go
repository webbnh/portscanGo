// Unit tests for package progbar.
package progbar

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"
)

// Arbitrary values for the progress bar width and size
const (
	expectedWidth = 69
	expectedSize  = 1001
)

func TestNew(t *testing.T) {
	expectedWriter := &ioutil.Discard // Arbitrary choice

	if New(0, expectedSize, *expectedWriter) != nil {
		t.Error("New() succeeded with zero width.")
	}
	if New(expectedWidth, 0, *expectedWriter) != nil {
		t.Error("New() succeeded with zero size.")
	}
	if New(expectedWidth, expectedWidth-1, *expectedWriter) != nil {
		t.Error("New() succeeded with width greater than size.")
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
	// We don't really care about the initial value of Bar.curSpin.
}

// Utility routine for checking the results of a Paint() call.
func checkPaint(t *testing.T, got, exp []byte) {
	if bytes.Compare(got, exp) != 0 {
		for i := range got {
			if i >= len(exp) {
				t.Errorf("Got excess character ('%c') "+
					"at position %d.\n",
					got[i], i)
				continue
			}
			if got[i] != exp[i] {
				t.Errorf("Got '%c' at position %d; "+
					"expected '%c'.\n",
					got[i], i, exp[i])
			}
		}
		if len(got) < len(exp) {
			t.Errorf("Characters missing after position %d; "+
				"expected %d.\n",
				len(got), len(exp))
		}
	}
}

func TestPaint(t *testing.T) {
	var buf bytes.Buffer

	bar := New(expectedWidth, expectedSize, &buf)

	// Check the initial display of a(n empty) bar
	bar.Paint()
	got := buf.Bytes()
	exp := bytes.Repeat([]byte{' '}, expectedWidth+1)
	exp = append(exp, "|\r|"...)
	checkPaint(t, got, exp)

	// Add some progress and check the display again
	const newProg = 4
	bar.current += (newProg*bar.total + bar.width - 1) / bar.width
	buf.Reset()
	bar.Paint()
	got = buf.Bytes()
	exp = append(exp, bytes.Repeat([]byte{'='}, newProg)...)
	checkPaint(t, got, exp)
}

func TestUpdate(t *testing.T) {
	var buf bytes.Buffer

	bar := New(expectedWidth, expectedSize, &buf)
	got := make([]byte, 80)

	for i := 1; i <= expectedSize; i++ {
		bar.Update()

		if bar.current != i {
			t.Fatalf("Bar.current is %d; expected %d.\n",
				bar.current, i)
		}

		n, err := buf.Read(got)
		if bar.current%(bar.total/bar.width) == 0 {
			switch {
			case n == 0:
				t.Errorf("Bar was not extended at update %d "+
					"(total: %d, width: %d).\n",
					i, bar.total, bar.width)
			case n > 1:
				t.Errorf("Bar was extended by %d characters "+
					"(expected only 1).\n",
					n)
			case n == 1:
				if got[0] != '=' {
					t.Errorf("Bar was extended with "+
						"character '%c'"+
						"instead of '='.\n",
						got[0])
				}
			default:
				t.Fatalf("Test bug:  "+
					"Unexpected return from bytes.Read():"+
					" %d, \"%v\".\n",
					n, err)
			}
		} else if n != 0 {
			t.Errorf("Bar was unexpectedly extended, update %d.\n",
				i)
		}
	}
}

func TestDone(t *testing.T) {
	var buf bytes.Buffer

	bar := New(expectedWidth, expectedSize, &buf)

	bar.Done()

	if bar.current != bar.total {
		t.Fatalf("Bar.current is %d; expected %d.\n",
			bar.current, bar.total)
	}

	got := buf.Bytes()
	exp := []byte{'\r'}
	exp = append(exp, bytes.Repeat([]byte{' '}, expectedWidth+2)...)
	exp = append(exp, '\r')
	checkPaint(t, got, exp)
}

func TestSpin(t *testing.T) {
	var buf bytes.Buffer

	bar := New(expectedWidth, expectedSize, &buf)

	// Starting at the end makes the loops cleaner
	bar.curSpin = len(spinStrs) - 1

	// Test two cycles to confirm the reset
	for j := 0; j < 2; j++ {
		for i, v := range spinStrs {
			buf.Reset()
			bar.Spin()

			if bar.curSpin != i {
				t.Error("Bar.curSpin is %d; expected %d "+
					"(pass %d).\n",
					bar.curSpin, i, j)
			}
			got := buf.Bytes()
			exp := []byte(v)
			if bytes.Compare(got, exp) != 0 {
				t.Errorf("Got '%v'; expected '%v' "+
					"(call %d, pass %d).\n",
					got, exp)
			}
		}
	}
}
