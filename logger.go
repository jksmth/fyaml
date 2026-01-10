package fyaml

import (
	"io"

	"github.com/jksmth/fyaml/internal/logger"
)

// Logger defines the logging interface for fyaml.
// All output is written to the configured io.Writer (typically os.Stderr).
type Logger interface {
	// Debugf logs verbose/debug information (shown when verbose enabled)
	Debugf(format string, args ...interface{})
	// Warnf logs warnings (always shown)
	Warnf(format string, args ...interface{})
}

// NewLogger creates a logger that writes to w.
// If verbose is true, Debugf messages are shown.
// Warnf messages are always shown.
func NewLogger(w io.Writer, verbose bool) Logger {
	return logger.New(w, verbose)
}

// NopLogger returns a no-op logger that discards all output.
func NopLogger() Logger {
	return logger.Nop()
}
