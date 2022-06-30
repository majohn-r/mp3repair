package internal

import (
	"io"
	"os"
)

type OutputBus interface {
	ConsoleWriter() io.Writer
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
