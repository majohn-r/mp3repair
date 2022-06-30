package internal

import (
	"io"
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

// TODO [#86] os.Stderr and os.Stdout should be hardwired
func NewOutputDevice(wStdout io.Writer, wStderr io.Writer) *OutputDevice {
	return &OutputDevice{
		consoleWriter: wStdout,
		errorWriter:   wStderr,
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
