package output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"unicode"
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

	outputDevice struct {
		consoleWriter io.Writer
		errorWriter   io.Writer
		logWriter     Logger
		performWrites bool
	}

	// Recorder is an implementation of Bus that simply records its inputs; it's
	// intended for unit tests, where you can provide the code under test with
	// an instance of Recorder and then verify that the code produces the
	// expected console, error, and log output.
	Recorder struct {
		consoleWriter *bytes.Buffer
		errorWriter   *bytes.Buffer
		logWriter     *RecordingLogger
	}

	// WantedRecording is intended to be used in unit tests as part of the test
	// structure; it allows simple capturing of what the test wants the console,
	// error, and log output to contain.
	WantedRecording struct {
		Console string
		Error   string
		Log     string
	}

	// RecordingLogger is a simple logger intended for use in unit tests; it
	// records the output given to it. Caveats: your production log may not
	// actually do anything with some calls into it - for instance, many logging
	// frameworks allow you to limit the severity of what is logger, e.g., only
	// warnings or worse; RecordingLogger will record every call made into it.
	// In addition, the output recorded cannot be guaranteed to match exactly
	// what your logging code records - but it will include the log level, the
	// message, and all field-value pairs. And, finally, the RecordingLogger
	// will behave differently to a logging mechanism that supports panic and
	// fatal logs, in that a production logger will probably call panic in
	// processing a panic log, and will probably exit the program on a fatal
	// log. RecordingLogger does neither of those.
	RecordingLogger struct {
		writer *bytes.Buffer
	}

	// NilWriter is a writer that does nothing at all; its intended use is to
	// pass into code used in testing where the side effect of writing to the
	// console or writing error output is of no interest whatsoever.
	NilWriter struct{}

	// NilLogger is a logger that does nothing at all; its intended use is to
	// pass into code used in testing where the side effect of logging is of no
	// interest whatsoever.
	NilLogger struct{}
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
	return &outputDevice{
		consoleWriter: c,
		errorWriter:   e,
		logWriter:     l,
		performWrites: true,
	}
}

// NewRecorder returns a recording implementation of Bus.
func NewRecorder() *Recorder {
	return &Recorder{
		consoleWriter: &bytes.Buffer{},
		errorWriter:   &bytes.Buffer{},
		logWriter:     NewRecordingLogger(),
	}
}

// NewNilBus returns an implementation of Bus that records and writes nothing.
func NewNilBus() Bus {
	nw := NilWriter{}
	return &outputDevice{
		consoleWriter: nw,
		errorWriter:   nw,
		logWriter:     NilLogger{},
		performWrites: false,
	}
}

// Log logs a message and map of fields at a specified log level.
func (o *outputDevice) Log(l Level, msg string, args map[string]any) {
	if o.performWrites {
		switch l {
		case Trace:
			o.logWriter.Trace(msg, args)
		case Debug:
			o.logWriter.Debug(msg, args)
		case Info:
			o.logWriter.Info(msg, args)
		case Warning:
			o.logWriter.Warning(msg, args)
		case Error:
			o.logWriter.Error(msg, args)
		case Panic:
			o.logWriter.Panic(msg, args)
		case Fatal:
			o.logWriter.Fatal(msg, args)
		default:
			o.WriteCanonicalError("programming error: call to outputDevice.Log() with invalid level value %d; message: '%s', args: '%v", l, msg, args)
		}
	}
}

// ConsoleWriter returns a writer for console output.
func (o *outputDevice) ConsoleWriter() io.Writer {
	return o.consoleWriter
}

// ErrorWriter returns a writer for error output.
func (o *outputDevice) ErrorWriter() io.Writer {
	return o.errorWriter
}

// LogWriter returns a Logger.
func (o *outputDevice) LogWriter() Logger {
	return o.logWriter
}

// WriteCanonicalError writes error output in a canonical format.
func (o *outputDevice) WriteCanonicalError(format string, a ...any) {
	if o.performWrites {
		fmt.Fprint(o.errorWriter, canonicalFormat(format, a...))
	}
}

// WriteError writes unedited error output.
func (o *outputDevice) WriteError(format string, a ...any) {
	if o.performWrites {
		fmt.Fprintf(o.errorWriter, format, a...)
	}
}

// WriteCanonicalConsole writes output to a console in a canonical format.
func (o *outputDevice) WriteCanonicalConsole(format string, a ...any) {
	if o.performWrites {
		fmt.Fprint(o.consoleWriter, canonicalFormat(format, a...))
	}
}

// WriteConsole writes output to a console.
func (o *outputDevice) WriteConsole(format string, a ...any) {
	if o.performWrites {
		fmt.Fprintf(o.consoleWriter, format, a...)
	}
}

// Log records a message and map of fields at a specified log level.
func (r *Recorder) Log(l Level, msg string, args map[string]any) {
	switch l {
	case Trace:
		r.logWriter.Trace(msg, args)
	case Debug:
		r.logWriter.Debug(msg, args)
	case Info:
		r.logWriter.Info(msg, args)
	case Warning:
		r.logWriter.Warning(msg, args)
	case Error:
		r.logWriter.Error(msg, args)
	case Panic:
		r.logWriter.Panic(msg, args)
	case Fatal:
		r.logWriter.Fatal(msg, args)
	default:
		r.WriteCanonicalError("programming error: call to Recorder.Log() with invalid level value %d; message: '%s', args: '%v", l, msg, args)
	}
}

// ConsoleWriter returns the internal console writer.
func (r *Recorder) ConsoleWriter() io.Writer {
	return r.consoleWriter
}

// ErrorWriter returns the internal error writer.
func (r *Recorder) ErrorWriter() io.Writer {
	return r.errorWriter
}

// LogWriter returns the internal logger.
func (r *Recorder) LogWriter() Logger {
	return r.logWriter
}

// WriteCanonicalError records data written as an error.
func (r *Recorder) WriteCanonicalError(format string, a ...any) {
	fmt.Fprint(r.errorWriter, canonicalFormat(format, a...))
}

// WriteError records un-edited data written as an error.
func (r *Recorder) WriteError(format string, a ...any) {
	fmt.Fprintf(r.errorWriter, format, a...)
}

// WriteCanonicalConsole records data written to the console.
func (r *Recorder) WriteCanonicalConsole(format string, a ...any) {
	fmt.Fprint(r.consoleWriter, canonicalFormat(format, a...))
}

// WriteConsole records data written to the console.
func (r *Recorder) WriteConsole(format string, a ...any) {
	fmt.Fprintf(r.consoleWriter, format, a...)
}

// ConsoleOutput returns the data written as console output.
func (r *Recorder) ConsoleOutput() string {
	return r.consoleWriter.String()
}

// ErrorOutput returns the data written as error output.
func (r *Recorder) ErrorOutput() string {
	return r.errorWriter.String()
}

// LogOutput returns the data written to a log.
func (r *Recorder) LogOutput() string {
	return r.logWriter.writer.String()
}

// NewRecordingLogger returns a recording implementation of Logger.
func NewRecordingLogger() *RecordingLogger {
	return &RecordingLogger{writer: &bytes.Buffer{}}
}

// Trace records a trace log message.
func (rl *RecordingLogger) Trace(msg string, fields map[string]any) {
	rl.log("trace", msg, fields)
}

// Debug records a debug log message.
func (rl *RecordingLogger) Debug(msg string, fields map[string]any) {
	rl.log("debug", msg, fields)
}

// Info records an info log message.
func (rl *RecordingLogger) Info(msg string, fields map[string]any) {
	rl.log("info", msg, fields)
}

// Warning records a warning log message.
func (rl *RecordingLogger) Warning(msg string, fields map[string]any) {
	rl.log("warning", msg, fields)
}

// Error records an error log message.
func (rl *RecordingLogger) Error(msg string, fields map[string]any) {
	rl.log("error", msg, fields)
}

// Panic records a panic log message and does not call panic().
func (rl *RecordingLogger) Panic(msg string, fields map[string]any) {
	rl.log("panic", msg, fields)
}

// Fatal records a fatal log message and does not terminate the program.
func (rl *RecordingLogger) Fatal(msg string, fields map[string]any) {
	rl.log("fatal", msg, fields)
}

func (rl *RecordingLogger) log(level string, msg string, fields map[string]any) {
	var parts []string
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s='%v'", k, v))
	}
	sort.Strings(parts)
	fmt.Fprintf(rl.writer, "level='%s' %s msg='%s'\n", level, strings.Join(parts, " "), msg)
}

// Trace does nothing.
func (nl NilLogger) Trace(msg string, fields map[string]any) {
}

// Debug does nothing.
func (nl NilLogger) Debug(msg string, fields map[string]any) {
}

// Info does nothing.
func (nl NilLogger) Info(msg string, fields map[string]any) {
}

// Warning does nothing.
func (nl NilLogger) Warning(msg string, fields map[string]any) {
}

// Error does nothing.
func (nl NilLogger) Error(msg string, fields map[string]any) {
}

// Panic does nothing.
func (nl NilLogger) Panic(msg string, fields map[string]any) {
}

// Fatal does nothing.
func (nl NilLogger) Fatal(msg string, fields map[string]any) {
}

// Write does nothing except return the expected values
func (nw NilWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Verify verifies the recorded output against the expected output and returns
// any differences found.
func (r *Recorder) Verify(w WantedRecording) (issues []string, ok bool) {
	ok = true
	if got := r.ConsoleOutput(); got != w.Console {
		issues = append(issues, fmt.Sprintf("console output = %q, want %q", got, w.Console))
		ok = false
	}
	if got := r.ErrorOutput(); got != w.Error {
		issues = append(issues, fmt.Sprintf("error output = %q, want %q", got, w.Error))
		ok = false
	}
	if got := r.LogOutput(); got != w.Log {
		issues = append(issues, fmt.Sprintf("log output = %q, want %q", got, w.Log))
		ok = false
	}
	return
}

func canonicalFormat(format string, a ...any) string {
	s := fmt.Sprintf(format, a...)
	// strip off trailing newlines
	for len(s) > 0 && s[len(s)-1:] == "\n" {
		s = s[:len(s)-1]
	}
	if len(s) == 0 {
		return "\n"
	}
	lastChar := s[len(s)-1:]
	finalChar := lastChar
	if !isSentenceTerminatingPunctuation(lastChar) {
		finalChar = "."
	}
	// trim off trailing sentence termination characters
	for len(s) > 0 && isSentenceTerminatingPunctuation(lastChar) {
		s = s[:len(s)-1]
		if len(s) > 0 {
			lastChar = s[len(s)-1:]
		}
	}
	s = s + finalChar
	r := []rune(s)
	if unicode.IsLower(r[0]) {
		r[0] = unicode.ToUpper(r[0])
		s = string(r)
	}
	s = s + "\n"
	return s
}

func isSentenceTerminatingPunctuation(s string) bool {
	switch s {
	case ".", "!", "?":
		return true
	}
	return false
}
