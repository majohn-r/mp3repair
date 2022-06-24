package internal

import (
	"io"

	"github.com/sirupsen/logrus"
)

const (
	INFO = iota
	WARN
	ERROR
)

const (
	fkLogLevel = "level"
)

type LogLevel int

type OutputBus interface {
	OutputWriter() io.Writer
	ErrorWriter() io.Writer
	Log(l LogLevel, msg string, fields map[string]interface{})
}

type OutputDevice struct {
	wOut io.Writer
	wErr io.Writer
}

func NewOutputDevice(wStdout io.Writer, wStderr io.Writer) *OutputDevice {
	return &OutputDevice{
		wOut: wStdout,
		wErr: wStderr,
	}
}

func (o *OutputDevice) OutputWriter() io.Writer {
	return o.wOut
}

func (o *OutputDevice) ErrorWriter() io.Writer {
	return o.wErr
}

func (o *OutputDevice) Log(l LogLevel, msg string, fields map[string]interface{}) {
	switch l {
	case INFO:
		logrus.WithFields(fields).Info(msg)
	case WARN:
		logrus.WithFields(fields).Warn(msg)
	case ERROR:
		logrus.WithFields(fields).Error(msg)
	default:
		logrus.WithFields(fields).Error(msg)
		logrus.WithFields(logrus.Fields{
			fkLogLevel: l,
		}).Error(LE_INVALID_LOG_LEVEL)
	}
}
