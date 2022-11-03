package output

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
)

type (
	// Recorder is an implementation of Bus that simply records its inputs; it's
	// intended for unit tests, where you can provide the code under test (that
	// needs a Bus) with an instance of Recorder and then verify that the code
	// produces the expected console, error, and log output.
	Recorder struct {
		consoleChannel *bytes.Buffer
		errorChannel   *bytes.Buffer
		logChannel     *RecordingLogger
	}

	// WantedRecording is intended to be used in unit tests as part of the test
	// structure; it allows the test writer to capture what the test wants the
	// console, error, and log output to contain.
	WantedRecording struct {
		Console string
		Error   string
		Log     string
	}

	// RecordingLogger is a simple logger intended for use in unit tests; it
	// records the output given to it.
	//
	// Caveats:
	//
	// Your production log may not actually do anything with some calls into it
	// - for instance, many logging frameworks allow you to limit the severity
	// of what is logged, e.g., only warnings or worse; RecordingLogger will
	// record every call made into it.
	//
	// The output recorded cannot be guaranteed to match exactly what your
	// logging code records - but it will include the log level, the message,
	// and all field-value pairs.
	//
	// The RecordingLogger will probably behave differently to a logging
	// mechanism that supports panic and fatal logs, in that a production logger
	// will probably call panic in processing a panic log, and will probably
	// exit the program on a fatal log. RecordingLogger does neither of those.
	RecordingLogger struct {
		writer *bytes.Buffer
	}
)

// NewRecorder returns a recording implementation of Bus.
func NewRecorder() *Recorder {
	return &Recorder{
		consoleChannel: &bytes.Buffer{},
		errorChannel:   &bytes.Buffer{},
		logChannel:     NewRecordingLogger(),
	}
}

// Log records a message and map of fields at a specified log level.
func (r *Recorder) Log(l Level, msg string, fields map[string]any) {
	switch l {
	case Trace:
		r.logChannel.Trace(msg, fields)
	case Debug:
		r.logChannel.Debug(msg, fields)
	case Info:
		r.logChannel.Info(msg, fields)
	case Warning:
		r.logChannel.Warning(msg, fields)
	case Error:
		r.logChannel.Error(msg, fields)
	case Panic:
		r.logChannel.Panic(msg, fields)
	case Fatal:
		r.logChannel.Fatal(msg, fields)
	default:
		r.WriteCanonicalError("programming error: call to Recorder.Log() with invalid level value %d; message: '%s', args: '%v", l, msg, fields)
	}
}

// ConsoleWriter returns the internal console writer.
func (r *Recorder) ConsoleWriter() io.Writer {
	return r.consoleChannel
}

// ErrorWriter returns the internal error writer.
func (r *Recorder) ErrorWriter() io.Writer {
	return r.errorChannel
}

// LogWriter returns the internal logger.
func (r *Recorder) LogWriter() Logger {
	return r.logChannel
}

// WriteCanonicalError records data written as an error.
func (r *Recorder) WriteCanonicalError(format string, a ...any) {
	fmt.Fprint(r.errorChannel, canonicalFormat(format, a...))
}

// WriteError records un-edited data written as an error.
func (r *Recorder) WriteError(format string, a ...any) {
	fmt.Fprintf(r.errorChannel, format, a...)
}

// WriteCanonicalConsole records data written to the console.
func (r *Recorder) WriteCanonicalConsole(format string, a ...any) {
	fmt.Fprint(r.consoleChannel, canonicalFormat(format, a...))
}

// WriteConsole records data written to the console.
func (r *Recorder) WriteConsole(format string, a ...any) {
	fmt.Fprintf(r.consoleChannel, format, a...)
}

// ConsoleOutput returns the data written as console output.
func (r *Recorder) ConsoleOutput() string {
	return r.consoleChannel.String()
}

// ErrorOutput returns the data written as error output.
func (r *Recorder) ErrorOutput() string {
	return r.errorChannel.String()
}

// LogOutput returns the data written to a log.
func (r *Recorder) LogOutput() string {
	return r.logChannel.writer.String()
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
