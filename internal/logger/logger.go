// Package logger provides a simple logging interface for fyaml.
package logger

import (
	"fmt"
	"io"
)

// Logger defines the logging interface for fyaml.
// All output is written to the configured io.Writer (typically os.Stderr).
type Logger interface {
	// Debugf logs verbose/debug information (shown when verbose enabled)
	Debugf(format string, args ...interface{})
	// Warnf logs warnings (always shown)
	Warnf(format string, args ...interface{})
}

// NoOpLogger discards all log output (zero allocation).
type NoOpLogger struct{}

// Debugf is a no-op.
func (NoOpLogger) Debugf(string, ...interface{}) {}

// Warnf is a no-op.
func (NoOpLogger) Warnf(string, ...interface{}) {}

// StdLogger writes to an io.Writer with optional verbose output.
type StdLogger struct {
	w       io.Writer
	verbose bool
}

// New creates a logger that writes to w.
// If verbose is true, Debugf messages are shown.
// Warnf messages are always shown.
func New(w io.Writer, verbose bool) Logger {
	return &StdLogger{w: w, verbose: verbose}
}

// Nop returns a no-op logger that discards all output.
func Nop() Logger {
	return NoOpLogger{}
}

// Debugf logs a debug message if verbose is enabled.
func (l *StdLogger) Debugf(format string, args ...interface{}) {
	if l.verbose {
		fmt.Fprintf(l.w, "[DEBUG] "+format+"\n", args...)
	}
}

// Warnf logs a warning message (always shown).
func (l *StdLogger) Warnf(format string, args ...interface{}) {
	fmt.Fprintf(l.w, "[WARN] "+format+"\n", args...)
}
