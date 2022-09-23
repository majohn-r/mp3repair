package internal

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// OutputBus defines a set of functions for writing to the console, to the
// standard error devices, and to a logger; its primary use is to make it much
// easier to test output writing
type OutputBus interface {
	WriteConsole(bool, string, ...any)
	WriteError(string, ...any)
	ErrorWriter() io.Writer
	LogWriter() Logger
}

type outputDevice struct {
	consoleWriter io.Writer
	errorWriter   io.Writer
	logWriter     Logger
}

// NewOutputBus returns an implementation of OutputBus
func NewOutputBus() OutputBus {
	return &outputDevice{
		consoleWriter: os.Stdout,
		errorWriter:   os.Stderr,
		logWriter:     productionLogger{},
	}
}

// ErrorWriter returns a writer to stderr
func (o *outputDevice) ErrorWriter() io.Writer {
	return o.errorWriter
}

// LogWriter returns a logger
func (o *outputDevice) LogWriter() Logger {
	return o.logWriter
}

// WriteError writes output to stderr
func (o *outputDevice) WriteError(format string, a ...any) {
	fmt.Fprintln(o.errorWriter, createStrictOutput(format, a...))
}

// WriteConsole writes output to stdout
func (o *outputDevice) WriteConsole(strict bool, format string, a ...any) {
	fmt.Fprint(o.consoleWriter, createConsoleOutput(strict, format, a...))
}

func createConsoleOutput(strict bool, format string, a ...any) string {
	if strict {
		return createStrictOutput(format, a...) + "\n"
	}
	return fmt.Sprintf(format, a...)
}

func createStrictOutput(format string, a ...any) string {
	s := fmt.Sprintf(format, a...)
	// strip off trailing newlines
	for strings.HasSuffix(s, "\n") {
		s = strings.TrimSuffix(s, "\n")
	}
	lastChar := s[len(s)-1:]
	finalChar := lastChar
	// trim off trailing sentence termination characters
	for lastChar == "." || lastChar == "!" || lastChar == "?" {
		s = strings.TrimSuffix(s, lastChar)
		lastChar = s[len(s)-1:]
	}
	if finalChar == "." || finalChar == "!" || finalChar == "?" {
		s = s + finalChar
	} else {
		s = s + "."
	}
	b := []byte(s)
	if b[0] >= 'a' && b[0] <= 'z' {
		b[0] -= 0x20
		s = string(b)
	}
	return s
}
