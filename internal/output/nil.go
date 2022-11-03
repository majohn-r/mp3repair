package output

type (

	// NilWriter is a writer that does nothing at all; its intended use is for
	// test code where the side effect of writing to the console or writing
	// error output is of no interest whatsoever.
	NilWriter struct{}

	// NilLogger is a logger that does nothing at all; its intended use is for
	// test code where the side effect of logging is of no interest whatsoever.
	NilLogger struct{}
)

// NewNilBus returns an implementation of Bus that does nothing.
func NewNilBus() Bus {
	nw := NilWriter{}
	return &bus{
		consoleChannel: nw,
		errorChannel:   nw,
		logChannel:     NilLogger{},
		performWrites:  false,
	}
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
