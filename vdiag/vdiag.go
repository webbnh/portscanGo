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
	"os"
)

// verbosity is the current level of verbosity enabled for the program.
var verbosity int

// verbShort is a type which supports the methods require to get the -v switch
// to play nicely with the -verbose switch.
type verbShort bool

// vShort is just a dummy instantiation required to make the -v switch set the
// verbosity level to vShortLevel.
var vShort verbShort

const vShortLevel = 2

// String is used by the flag package.
func (v *verbShort) String() string {
	return fmt.Sprint(*v) // FIXME:  Should this be printing verbosity instead?
}

// vShort is a boolean flag.
func (v *verbShort) IsBoolFlag() bool {
	return true
}

// Set is used by the flag package; the custom method is provided here to
// allow the -v flag to interact with the -verbose flag appropriately.
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
// current setting.
func Out(level int, message string, v ...interface{}) {
	if level > verbosity {
		return
	}
	fmt.Fprintf(os.Stderr, "["+string('0'+level)+"]"+message, v...)
}
