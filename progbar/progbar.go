// Package progbar provides a simple ASCII progress bar
//
// After creating the bar with New(), the bar can be painted on the screen (or
// repainted again) using Paint(), it can be updated using Update(), and it
// can be finished with Done().  The function Spin() can be used to show
// intermediate activity by causing a spinning effect at the end of the bar.
package progbar

import (
	"io"
	"strings"
)

// The magic strings which, when printed in sequence, make the spinner appear
// to spin.
var spinStrs []string = []string{"-\b", "\\\b", "|\b", "/\b"}

// The Bar structure represents the parameters and state of the progress bar.
type Bar struct {
	// The width of the bar on the screen in columns.
	width int
	// The total size of the bar in "progress units".
	total int
	// The current amount of progress in the bar in "progress units".
	current int
	// The current orientation of the end-of-bar "spinner".
	curSpin int
	// The Writer used to display the bar.
	w io.Writer
}

// New creates a new bar which will grow to the specified width as the number
// of "progress units" approaches the specified size.  The specified Writer is
// used to display the bar.
func New(width int, size int, w io.Writer) *Bar {
	if width <= 0 || size <= 0 || w == nil {
		return nil
	}
	return &Bar{width: width, total: size, w: w}
}

// Paint displays an empty progress bar on the screen with vertical bars
// marking the beginning and end, fills the bar up to the current progress
// point, and leaves the cursor at the next column.
func (b *Bar) Paint() {
	s := []string{
		strings.Repeat(" ", b.width+1), // Move the cursor to the right
		"|\r|",                         // Last bar, return, first bar
		strings.Repeat("=", b.current*b.width/b.total), // Any progress
	}
	io.WriteString(b.w, strings.Join(s, ""))
}

// Update advances the bar by one "progress unit" and, if appropriate, adds a
// character to the bar on the screen.
func (b *Bar) Update() {
	b.current++
	if b.current%(b.total/b.width) == 0 {
		io.WriteString(b.w, "=")
	}
}

// Done erases the bar.
func (b *Bar) Done() {
	b.current = b.total // For completeness

	s := []string{
		"\r", // Return the cursor to the beginning of the line
		strings.Repeat(" ", b.width+2), // Clear the whole line
		"\r",                           // Like it was never there
	}
	io.WriteString(b.w, strings.Join(s, ""))
}

// Spin a little "wheel" at the end of the progress bar to indicate
// intermediate activity.
func (b *Bar) Spin() {
	b.curSpin = (b.curSpin + 1) & 0x3
	io.WriteString(b.w, spinStrs[b.curSpin])
}
