package output

import (
	"fmt"
	"io"
	"os"
)

type (
	// Level is used to specify log levels for Bus.Log().
	Level uint32

	// Bus defines a set of functions for writing console messages and error
	// messages, and for providing access to the console writer, the error
	// writer, and a Logger instance; its primary use is to simplify how
	// application code handles console, error, and logged output, and its
	// secondary use is to make it easy to test output writing.
	Bus interface {
		Log(Level, string, map[string]any)
		WriteCanonicalConsole(string, ...any)
		WriteConsole(string, ...any)
		WriteCanonicalError(string, ...any)
		WriteError(string, ...any)
		ConsoleWriter() io.Writer
		ErrorWriter() io.Writer
		LogWriter() Logger
	}

	// Logger defines a set of functions for writing to a log at various log
	// levels
	Logger interface {
		Trace(msg string, fields map[string]any)
		Debug(msg string, fields map[string]any)
		Info(msg string, fields map[string]any)
		Warning(msg string, fields map[string]any)
		Error(msg string, fields map[string]any)
		Panic(msg string, fields map[string]any)
		Fatal(msg string, fields map[string]any)
	}

	bus struct {
		consoleChannel io.Writer
		errorChannel   io.Writer
		logChannel     Logger
		performWrites  bool
	}
)

// These are the different logging levels.
const (
	Fatal Level = iota
	Panic
	Error
	Warning
	Info
	Debug
	Trace
)

// NewDefaultBus returns an implementation of Bus that writes console messages
// to stdout and error messages to stderr.
func NewDefaultBus(l Logger) Bus {
	return NewCustomBus(os.Stdout, os.Stderr, l)
}

// NewCustomBus returns an implementation of Bus that lets the caller specify
// the console and error writers and the Logger.
func NewCustomBus(c, e io.Writer, l Logger) Bus {
	return &bus{
		consoleChannel: c,
		errorChannel:   e,
		logChannel:     l,
		performWrites:  true,
	}
}

// Log logs a message and map of fields at a specified log level.
func (b *bus) Log(l Level, msg string, args map[string]any) {
	if b.performWrites {
		switch l {
		case Trace:
			b.logChannel.Trace(msg, args)
		case Debug:
			b.logChannel.Debug(msg, args)
		case Info:
			b.logChannel.Info(msg, args)
		case Warning:
			b.logChannel.Warning(msg, args)
		case Error:
			b.logChannel.Error(msg, args)
		case Panic:
			b.logChannel.Panic(msg, args)
		case Fatal:
			b.logChannel.Fatal(msg, args)
		default:
			b.WriteCanonicalError("programming error: call to bus.Log() with invalid level value %d; message: '%s', args: '%v", l, msg, args)
		}
	}
}

// ConsoleWriter returns a writer for console output.
func (b *bus) ConsoleWriter() io.Writer {
	return b.consoleChannel
}

// ErrorWriter returns a writer for error output.
func (b *bus) ErrorWriter() io.Writer {
	return b.errorChannel
}

// LogWriter returns a Logger.
func (b *bus) LogWriter() Logger {
	return b.logChannel
}

// WriteCanonicalError writes error output in a canonical format.
func (b *bus) WriteCanonicalError(format string, a ...any) {
	if b.performWrites {
		fmt.Fprint(b.errorChannel, canonicalFormat(format, a...))
	}
}

// WriteError writes unedited error output.
func (b *bus) WriteError(format string, a ...any) {
	if b.performWrites {
		fmt.Fprintf(b.errorChannel, format, a...)
	}
}

// WriteCanonicalConsole writes output to a console in a canonical format.
func (b *bus) WriteCanonicalConsole(format string, a ...any) {
	if b.performWrites {
		fmt.Fprint(b.consoleChannel, canonicalFormat(format, a...))
	}
}

// WriteConsole writes output to a console.
func (b *bus) WriteConsole(format string, a ...any) {
	if b.performWrites {
		fmt.Fprintf(b.consoleChannel, format, a...)
	}
}
