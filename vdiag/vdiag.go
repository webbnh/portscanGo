// Package vdiag provides a facility which simplifies and centralizes the
// production of diagnostic messages with various levels of verbosity
//
// Importing the package automagically adds two command line flags, "-v" and
// "-verbose": the latter takes an integer setting the desired level of
// verbosity, while the former is shorthand for setting it to a pre-defined
// (but non-zero) low level.
//
// Diagnostics are issued by calling the Out() function, specifying the
// verbosity level of the requested message (which is a printf format string).
// If the requested verbosity is too high, then the message is not printed.
package vdiag

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

// Place to send diagnostic output (overwritten for testing)
var w io.Writer = os.Stderr

// verbosity is the current level of verbosity enabled for the program.
var verbosity int

// verbShort implements the flag.Value interface.  It is used to get the -v
// switch to play nicely with the -verbose switch.
type verbShort struct{}

// vShort is just a dummy instantiation required to make the -v switch set the
// verbosity level to vShortLevel -- the set method will actually modify the 
// verbosity level instead of changing the value of vShort.
var vShort verbShort

const vShortLevel = 2

// String is used by the flag package.
func (v *verbShort) String() string {
	return fmt.Sprint(verbosity >= vShortLevel)
}

// verbShort functions as a boolean flag.
//
// Note: this interface function is supposed to indicated to the flag package
// that this custom flag is a boolean flag; however, my version of Go is
// apparently too old to support this, so I have no idea if it works properly
// in an up-to-date installation.
func (v *verbShort) IsBoolFlag() bool {
	return true
}

// Set is used by the flag package; the custom method is provided here to
// allow the -v flag to interact with the -verbose flag appropriately.
//
// Note: the argument is ignored becuase it is presumed that the presence of
// the -v flag (regardless of what value (if any) the user might be forced to
// give it), is enough to set the verbosity level.
func (v *verbShort) Set(value string) error {
	switch {
	case verbosity == vShortLevel:
		break
	case verbosity < vShortLevel:
		Set(vShortLevel)
	default:
		return errors.New("-v would reduce verbosity")
	}
	return nil
}

// Package initialization function:  add our command line flags.
func init() {
	flag.Var(&vShort, "v", "Enable basic verbose output")
	flag.IntVar(&verbosity, "verbose", 0, "Set level of verbosity")
}

// Set sets the program's verbosity level.
func Set(level int) {
	verbosity = level
}

// Verbosity returns the program's current verbosity level.
func Verbosity() int {
	return verbosity
}

// Out prints the specified message (treating it like a printf format string)
// if the specified verbosity level is less than or equal to the program's
// current setting.  (The message is prefixed with the requested verbosity
// level.)
func Out(level int, message string, v ...interface{}) {
	if level > verbosity {
		return
	}
	fmt.Fprintf(w, "["+string('0'+level)+"]"+message, v...)
}
