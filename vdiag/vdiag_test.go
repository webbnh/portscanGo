// Unit tests for package vdiag.
package vdiag

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"reflect"
	"testing"
)

const testShortLevel = 2 // This should match vShortLevel, but it's not req'd

func TestVerbShortString(t *testing.T) {
	got := vShort.String()
	exp := "false"
	if got != exp {
		t.Errorf("Unexpected value, \"%s\"; expected \"%s\".\n",
			got, exp)
	}
}

func TestVerbShortIsBoolFlag(t *testing.T) {
	got := vShort.IsBoolFlag()
	if !got {
		t.Errorf("Unexpected value, \"%v\"; expected \"%v\".\n",
			got, true)
	}
}

func TestVerbShortSet(t *testing.T) {
	cases := []struct {
		testVerbosity     int
		testArg           string
		expectedErr       error
		expectedVerbosity int
	}{
		{testShortLevel, "dummy", nil, testShortLevel},
		{testShortLevel - 1, "dummy", nil, testShortLevel},
		{
			testShortLevel + 1,
			"dummy",
			errors.New("-v would reduce verbosity"),
			testShortLevel + 1,
		},
	}

	for _, v := range cases {
		verbosity = v.testVerbosity
		got := vShort.Set(v.testArg)

		if !reflect.DeepEqual(got, v.expectedErr) {
			t.Errorf("Got error \"%v\"; expected \"%v\" "+
				"(orignal verbosity: %d, argument: \"%s\").\n",
				got, v.expectedErr, v.testVerbosity, v.testArg)
		}
		if verbosity != v.expectedVerbosity {
			t.Errorf("Resulting verbosity %d; expected %d "+
				"(original verbosity: %d, argument: \"%s\").\n",
				verbosity, v.expectedVerbosity,
				v.testVerbosity, v.testArg)
		}
	}
}

// Utility routine for checking flag.Flag values
func checkFlag(t *testing.T, s, defVal string) {
	f := flag.Lookup(s)
	if f == nil {
		t.Errorf("Found no definition of flag \"%s\".\n", s)
		return
	}

	if f.Usage == "" {
		t.Errorf("Flag \"%s\" lacks a Usage string.\n", s)
	}

	// Checking the current value of the flag doesn't seem to be
	// very reliable in the context of running a test, so skip it.

	if f.DefValue != defVal {
		t.Errorf("Flag \"%s\" initial value is \"%s\"; "+
			"expected \"%s\".\n",
			s, f.DefValue, defVal)
	}
}

func TestInit(t *testing.T) {
	if !flag.Parsed() {
		t.Fatal("Cannot test initialization")
	}

	checkFlag(t, "v", "false")
	checkFlag(t, "verbose", "0")
}

func TestSet(t *testing.T) {
	cases := []struct {
		testVerbosity int
		testArg       int
	}{
		{0, 1},
		{1, 0},
		{0, 0},
		{1, 1},
		{5, 9},
		{9, 5},
	}

	for _, v := range cases {
		verbosity = v.testVerbosity
		Set(v.testArg)

		if verbosity != v.testArg {
			t.Errorf("Resulting verbosity %d; expected %d "+
				"(original verbosity: %d).\n",
				verbosity, v.testArg, v.testVerbosity)
		}
	}
}

func TestVerbosity(t *testing.T) {
	cases := []struct {
		testVerbosity int
	}{{7}, {1}, {0}, {5}, {9}}

	for _, v := range cases {
		verbosity = v.testVerbosity
		got := Verbosity()

		if got != v.testVerbosity {
			t.Errorf("Resulting verbosity %d; expected %d "+
				"(original verbosity: %d).\n",
				verbosity, v.testVerbosity)
		}
	}
}

func TestOut(t *testing.T) {
	generic := "message"
	cases := []struct {
		testVerbosity int
		reqVerbosity  int
		message       string
		result        string
	}{
		{0, 5, generic, ""},
		{3, 5, generic, ""},
		{5, 5, generic, generic},
		{7, 5, generic, generic},
		{9, 5, generic, generic},

		{5, 0, generic, generic},
		{5, 3, generic, generic},
		{5, 5, generic, generic},
		{5, 7, generic, ""},
		{5, 9, generic, ""},
	}

	var buf bytes.Buffer
	prefStr := "[%d]"
	w = &buf

	for _, v := range cases {
		buf.Reset()
		verbosity = v.testVerbosity

		Out(v.reqVerbosity, v.message)

		exp := ""
		if v.result != "" {
			exp = fmt.Sprintf(
				prefStr+"%s",
				v.reqVerbosity,
				v.result)
		}
		if bytes.Compare(buf.Bytes(), []byte(exp)) != 0 {
			t.Errorf("Got \"%s\"; expected \"%s\".\n",
				buf.Bytes(), exp)
		}
	}

	// Do a quick test to prove that Out() accepts a format string and
	// multiple arguments.
	buf.Reset()
	verbosity = 5
	req := 1
	formatStr := "Test %d%c%s arguments"
	exp := fmt.Sprintf(prefStr+formatStr, req, 4, 'm', "at")
	Out(req, formatStr, 4, 'm', "at")
	if bytes.Compare(buf.Bytes(), []byte(exp)) != 0 {
		t.Errorf("Got \"%s\"; expected \"%s\".\n",
			buf.Bytes(), exp)
	}
}
