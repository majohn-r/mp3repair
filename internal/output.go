package internal

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type OutputBus interface {
	ConsoleWriter() io.Writer
	WriteError(string, ...any)
	ErrorWriter() io.Writer
	LogWriter() Logger
}

type OutputDevice struct {
	consoleWriter io.Writer
	errorWriter   io.Writer
	logWriter     Logger
}

func NewOutputDevice() *OutputDevice {
	return &OutputDevice{
		consoleWriter: os.Stdout,
		errorWriter:   os.Stderr,
		logWriter:     productionLogger{},
	}
}

func (o *OutputDevice) ConsoleWriter() io.Writer {
	return o.consoleWriter
}

func (o *OutputDevice) ErrorWriter() io.Writer {
	return o.errorWriter
}

func (o *OutputDevice) LogWriter() Logger {
	return o.logWriter
}

func (o *OutputDevice) WriteError(format string, a ...any) {
	fmt.Fprintln(o.errorWriter, createErrorOutput(format, a...))
}

func createErrorOutput(format string, a ...any) string {
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
