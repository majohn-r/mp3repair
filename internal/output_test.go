package internal

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestNewDefaultOutputBus(t *testing.T) {
	fnName := "NewDefaultOutputBus()"
	tests := []struct {
		name              string
		want              OutputBus
		wantConsoleWriter io.Writer
		wantErrorWriter   io.Writer
		wantLogWriter     Logger
	}{
		{
			name:              "normal",
			want:              NewCustomOutputBus(os.Stdout, os.Stderr, NilLogger{}),
			wantConsoleWriter: os.Stdout,
			wantErrorWriter:   os.Stderr,
			wantLogWriter:     NilLogger{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDefaultOutputBus(NilLogger{})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			if w := got.ConsoleWriter(); w != tt.wantConsoleWriter {
				t.Errorf("%s got console writer %v, want %v", fnName, w, tt.wantConsoleWriter)
			}
			if w := got.ErrorWriter(); w != tt.wantErrorWriter {
				t.Errorf("%s got error writer %v, want %v", fnName, w, tt.wantErrorWriter)
			}
			if w := got.LogWriter(); w != tt.wantLogWriter {
				t.Errorf("%s got log writer %v, want %v", fnName, w, tt.wantLogWriter)
			}
		})
	}
}

func Test_outputDevice_Log(t *testing.T) {
	fnName := "outputDevice.Log()"
	type args struct {
		l    Level
		msg  string
		args map[string]any
	}
	tests := []struct {
		name string
		args
		wantLogOutput   string
		wantErrorOutput string
	}{
		{
			name: "trace",
			args: args{
				l:    Trace,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantLogOutput: "level='trace' f='v' msg='hello'\n",
		},
		{
			name: "debug",
			args: args{
				l:    Debug,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantLogOutput: "level='debug' f='v' msg='hello'\n",
		},
		{
			name: "info",
			args: args{
				l:    Info,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantLogOutput: "level='info' f='v' msg='hello'\n",
		},
		{
			name: "warning",
			args: args{
				l:    Warning,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantLogOutput: "level='warning' f='v' msg='hello'\n",
		},
		{
			name: "error",
			args: args{
				l:    Error,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantLogOutput: "level='error' f='v' msg='hello'\n",
		},
		{
			name: "panic",
			args: args{
				l:    Panic,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantLogOutput: "level='panic' f='v' msg='hello'\n",
		},
		{
			name: "fatal",
			args: args{
				l:    Fatal,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantLogOutput: "level='fatal' f='v' msg='hello'\n",
		},
		{
			name: "illegal",
			args: args{
				l:    Trace + 1,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantErrorOutput: "Programming error: call to outputDevice.Log() with invalid level value 7; message: 'hello', args: 'map[f:v].\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eW := &bytes.Buffer{}
			l := NewRecordingLogger()
			o := NewCustomOutputBus(nil, eW, l)
			o.Log(tt.args.l, tt.args.msg, tt.args.args)
			if got := l.writer.String(); got != tt.wantLogOutput {
				t.Errorf("%s got log %q want %q", fnName, got, tt.wantLogOutput)
			}
			if got := eW.String(); got != tt.wantErrorOutput {
				t.Errorf("%s got error %q want %q", fnName, got, tt.wantErrorOutput)
			}
		})
	}
}

func TestOutputDevice_WriteCanonicalError(t *testing.T) {
	fnName := "OutputDevice.WriteCanonicalError()"
	type args struct {
		format string
		a      []any
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "broad test",
			args: args{
				format: "test format %d %q %v..?!..?\n\n\n\n",
				a:      []any{25, "foo", 1.245},
			},
			want: "Test format 25 \"foo\" 1.245?\n",
		},
		{
			name: "narrow test",
			args: args{
				format: "1. test format %d %q %v",
				a:      []any{25, "foo", 1.245},
			},
			want: "1. test format 25 \"foo\" 1.245.\n",
		},
		{
			name: "nothing but newlines",
			args: args{
				format: "\n\n\n\n\n\n\n\n\n\n\n\n\n\n",
			},
			want: "\n",
		},
		{
			name: "nothing but terminal punctuation",
			args: args{
				format: "!!?.!?.",
			},
			want: ".\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			o := &outputDevice{errorWriter: w, performWrites: true}
			o.WriteCanonicalError(tt.args.format, tt.args.a...)
			if got := w.String(); got != tt.want {
				t.Errorf("%s got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestOutputDevice_WriteError(t *testing.T) {
	fnName := "OutputDevice.WriteError()"
	type args struct {
		format string
		a      []any
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "broad test",
			args: args{
				format: "test format %d %q %v..?!..?\n\n\n\n",
				a:      []any{25, "foo", 1.245},
			},
			want: "test format 25 \"foo\" 1.245..?!..?\n\n\n\n",
		},
		{
			name: "narrow test",
			args: args{
				format: "1. test format %d %q %v",
				a:      []any{25, "foo", 1.245},
			},
			want: "1. test format 25 \"foo\" 1.245",
		},
		{
			name: "nothing but newlines",
			args: args{
				format: "\n\n\n\n\n\n\n\n\n\n\n\n\n\n",
			},
			want: "\n\n\n\n\n\n\n\n\n\n\n\n\n\n",
		},
		{
			name: "nothing but terminal punctuation",
			args: args{
				format: "!!?.!?.",
			},
			want: "!!?.!?.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			o := &outputDevice{errorWriter: w, performWrites: true}
			o.WriteError(tt.args.format, tt.args.a...)
			if got := w.String(); got != tt.want {
				t.Errorf("%s got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestOutputDevice_WriteCanonicalConsole(t *testing.T) {
	fnName := "OutputDevice.WriteCanonicalConsole()"
	type args struct {
		format string
		a      []any
	}
	tests := []struct {
		name string
		w    *bytes.Buffer
		args
		want string
	}{
		{
			name: "strict rules",
			w:    &bytes.Buffer{},
			args: args{
				format: "test %s...\n\n",
				a:      []any{"foo."},
			},
			want: "Test foo.\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &outputDevice{consoleWriter: tt.w, performWrites: true}
			o.WriteCanonicalConsole(tt.args.format, tt.args.a...)
			if got := tt.w.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestOutputDevice_WriteConsole(t *testing.T) {
	fnName := "OutputDevice.WriteConsole()"
	type args struct {
		format string
		a      []any
	}
	tests := []struct {
		name string
		w    *bytes.Buffer
		args
		want string
	}{
		{
			name: "lax rules",
			w:    &bytes.Buffer{},
			args: args{
				format: "test %s...\n\n",
				a:      []any{"foo."},
			},
			want: "test foo....\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &outputDevice{consoleWriter: tt.w, performWrites: true}
			o.WriteConsole(tt.args.format, tt.args.a...)
			if got := tt.w.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestRecordingLogger_Trace(t *testing.T) {
	fnName := "RecordingLogger.Trace()"
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := []struct {
		name string
		tl   RecordingLogger
		args
		want string
	}{
		{
			name: "simple test",
			tl:   RecordingLogger{writer: &bytes.Buffer{}},
			args: args{
				msg:    "simple message",
				fields: map[string]any{"f1": 1, "f2": true, "f3": "v"},
			},
			want: "level='trace' f1='1' f2='true' f3='v' msg='simple message'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tl.Trace(tt.args.msg, tt.args.fields)
			if got := tt.tl.writer.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestRecordingLogger_Debug(t *testing.T) {
	fnName := "RecordingLogger.Debug()"
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := []struct {
		name string
		tl   RecordingLogger
		args
		want string
	}{
		{
			name: "simple test",
			tl:   RecordingLogger{writer: &bytes.Buffer{}},
			args: args{
				msg:    "simple message",
				fields: map[string]any{"f1": 1, "f2": true, "f3": "v"},
			},
			want: "level='debug' f1='1' f2='true' f3='v' msg='simple message'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tl.Debug(tt.args.msg, tt.args.fields)
			if got := tt.tl.writer.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestRecordingLogger_Info(t *testing.T) {
	fnName := "RecordingLogger.Info()"
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := []struct {
		name string
		tl   RecordingLogger
		args
		want string
	}{
		{
			name: "simple test",
			tl:   RecordingLogger{writer: &bytes.Buffer{}},
			args: args{
				msg:    "simple message",
				fields: map[string]any{"f1": 1, "f2": true, "f3": "v"},
			},
			want: "level='info' f1='1' f2='true' f3='v' msg='simple message'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tl.Info(tt.args.msg, tt.args.fields)
			if got := tt.tl.writer.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestRecordingLogger_Warning(t *testing.T) {
	fnName := "RecordingLogger.Warning()"
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := []struct {
		name string
		tl   RecordingLogger
		args
		want string
	}{
		{
			name: "simple test",
			tl:   RecordingLogger{writer: &bytes.Buffer{}},
			args: args{
				msg:    "simple message",
				fields: map[string]any{"f1": 1, "f2": true, "f3": "v"},
			},
			want: "level='warning' f1='1' f2='true' f3='v' msg='simple message'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tl.Warning(tt.args.msg, tt.args.fields)
			if got := tt.tl.writer.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestRecordingLogger_Error(t *testing.T) {
	fnName := "RecordingLogger.Error()"
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := []struct {
		name string
		tl   RecordingLogger
		args
		want string
	}{
		{
			name: "simple test",
			tl:   RecordingLogger{writer: &bytes.Buffer{}},
			args: args{
				msg:    "simple message",
				fields: map[string]any{"f1": 1, "f2": true, "f3": "v"},
			},
			want: "level='error' f1='1' f2='true' f3='v' msg='simple message'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tl.Error(tt.args.msg, tt.args.fields)
			if got := tt.tl.writer.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestRecordingLogger_Panic(t *testing.T) {
	fnName := "RecordingLogger.Panic()"
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := []struct {
		name string
		tl   RecordingLogger
		args
		want string
	}{
		{
			name: "simple test",
			tl:   RecordingLogger{writer: &bytes.Buffer{}},
			args: args{
				msg:    "simple message",
				fields: map[string]any{"f1": 1, "f2": true, "f3": "v"},
			},
			want: "level='panic' f1='1' f2='true' f3='v' msg='simple message'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tl.Panic(tt.args.msg, tt.args.fields)
			if got := tt.tl.writer.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestRecordingLogger_Fatal(t *testing.T) {
	fnName := "RecordingLogger.Fatal()"
	type args struct {
		msg    string
		fields map[string]any
	}
	tests := []struct {
		name string
		tl   RecordingLogger
		args
		want string
	}{
		{
			name: "simple test",
			tl:   RecordingLogger{writer: &bytes.Buffer{}},
			args: args{
				msg:    "simple message",
				fields: map[string]any{"f1": 1, "f2": true, "f3": "v"},
			},
			want: "level='fatal' f1='1' f2='true' f3='v' msg='simple message'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tl.Fatal(tt.args.msg, tt.args.fields)
			if got := tt.tl.writer.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestNewRecordingOutputBus(t *testing.T) {
	fnName := "NewRecordingOutputBus()"
	tests := []struct {
		name            string
		canonicalWrites bool
		consoleFmt      string
		consoleArgs     []any
		errorFmt        string
		errorArgs       []any
		logMessage      string
		logArgs         map[string]any
		WantedOutput
	}{
		{
			name:        "non-canonical test",
			consoleFmt:  "%s %d %t",
			consoleArgs: []any{"hello", 42, true},
			errorFmt:    "%d %t %s",
			errorArgs:   []any{24, false, "bye"},
			logMessage:  "hello!",
			logArgs:     map[string]any{"field": "value"},
			WantedOutput: WantedOutput{
				WantConsoleOutput: "hello 42 true",
				WantErrorOutput:   "24 false bye",
				WantLogOutput:     "level='error' field='value' msg='hello!'\n",
			},
		},
		{
			name:            "canonical test",
			canonicalWrites: true,
			consoleFmt:      "%s %d %t",
			consoleArgs:     []any{"hello", 42, true},
			errorFmt:        "%d %t %s",
			errorArgs:       []any{24, false, "bye"},
			logMessage:      "hello!",
			logArgs:         map[string]any{"field": "value"},
			WantedOutput: WantedOutput{
				WantConsoleOutput: "Hello 42 true.\n",
				WantErrorOutput:   "24 false bye.\n",
				WantLogOutput:     "level='error' field='value' msg='hello!'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewRecordingOutputBus()
			var i any = o
			if _, ok := i.(OutputBus); !ok {
				t.Errorf("%s: RecordingOutputBus does not implement OutputBus", fnName)
			}
			if o.ConsoleWriter() == nil {
				t.Errorf("%s: console writer is nil", fnName)
			}
			if o.ErrorWriter() == nil {
				t.Errorf("%s: error writer is nil", fnName)
			}
			if o.LogWriter() == nil {
				t.Errorf("%s: log writer is nil", fnName)
			}
			if tt.canonicalWrites {
				o.WriteCanonicalConsole(tt.consoleFmt, tt.consoleArgs...)
				o.WriteCanonicalError(tt.errorFmt, tt.errorArgs...)
			} else {
				o.WriteConsole(tt.consoleFmt, tt.consoleArgs...)
				o.WriteError(tt.errorFmt, tt.errorArgs...)
			}
			o.LogWriter().Error(tt.logMessage, tt.logArgs)
			if issues, ok := o.VerifyOutput(tt.WantedOutput); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func TestRecordingOutputBus_VerifyOutput(t *testing.T) {
	fnName := "RecordingOutputBus.VerifyOutput()"
	type args struct {
		o *RecordingOutputBus
		w WantedOutput
	}
	tests := []struct {
		name string
		args
		wantIssues []string
		wantOk     bool
	}{
		{name: "normal", args: args{o: NewRecordingOutputBus(), w: WantedOutput{}}, wantOk: true},
		{
			name: "errors",
			args: args{
				o: NewRecordingOutputBus(),
				w: WantedOutput{
					WantConsoleOutput: "unexpected console output",
					WantErrorOutput:   "unexpected error output",
					WantLogOutput:     "unexpected log output",
				},
			},
			wantIssues: []string{
				"console output = \"\", want \"unexpected console output\"",
				"error output = \"\", want \"unexpected error output\"",
				"log output = \"\", want \"unexpected log output\"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIssues, gotOk := tt.args.o.VerifyOutput(tt.args.w)
			if !reflect.DeepEqual(gotIssues, tt.wantIssues) {
				t.Errorf("%s gotIssues = %v, want %v", fnName, gotIssues, tt.wantIssues)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
		})
	}
}

func TestRecordingOutputBus_Log(t *testing.T) {
	fnName := "RecordingOutputBus.Log()"
	type args struct {
		l    Level
		msg  string
		args map[string]any
	}
	tests := []struct {
		name string
		args
		wantLogOutput   string
		wantErrorOutput string
	}{
		{
			name: "trace",
			args: args{
				l:    Trace,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantLogOutput: "level='trace' f='v' msg='hello'\n",
		},
		{
			name: "debug",
			args: args{
				l:    Debug,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantLogOutput: "level='debug' f='v' msg='hello'\n",
		},
		{
			name: "info",
			args: args{
				l:    Info,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantLogOutput: "level='info' f='v' msg='hello'\n",
		},
		{
			name: "warning",
			args: args{
				l:    Warning,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantLogOutput: "level='warning' f='v' msg='hello'\n",
		},
		{
			name: "error",
			args: args{
				l:    Error,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantLogOutput: "level='error' f='v' msg='hello'\n",
		},
		{
			name: "panic",
			args: args{
				l:    Panic,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantLogOutput: "level='panic' f='v' msg='hello'\n",
		},
		{
			name: "fatal",
			args: args{
				l:    Fatal,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantLogOutput: "level='fatal' f='v' msg='hello'\n",
		},
		{
			name: "illegal",
			args: args{
				l:    Trace + 1,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			wantErrorOutput: "Programming error: call to RecordingOutputBus.Log() with invalid level value 7; message: 'hello', args: 'map[f:v].\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRecordingOutputBus()
			r.Log(tt.args.l, tt.args.msg, tt.args.args)
			if got := r.LogOutput(); got != tt.wantLogOutput {
				t.Errorf("%s got log %q want %q", fnName, got, tt.wantLogOutput)
			}
			if got := r.ErrorOutput(); got != tt.wantErrorOutput {
				t.Errorf("%s got error %q want %q", fnName, got, tt.wantErrorOutput)
			}
		})
	}
}

func TestNewNilOutputBus(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "simple test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewNilOutputBus()
			o.WriteCanonicalConsole("%s %d %t", "foo", 42, true)
			o.WriteConsole("%s %d %t", "foo", 42, true)
			o.WriteCanonicalError("%s %d %t", "foo", 42, true)
			o.WriteError("%s %d %t", "foo", 42, true)
			o.LogWriter().Error("error message", map[string]any{"field1": "value"})
		})
	}
}

func TestNilWriter_Write(t *testing.T) {
	fnName := "NilWriter.Write()"
	type args struct {
		p []byte
	}
	tests := []struct {
		name string
		nw   NilWriter
		args
		wantN   int
		wantErr bool
	}{
		{
			name: "a few bytes",
			nw:   NilWriter{},
			args: args{
				p: []byte{0, 1, 2},
			},
			wantN:   3,
			wantErr: false,
		},
		{
			name: "nil",
			nw:   NilWriter{},
			args: args{
				p: nil,
			},
			wantN:   0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotN, err := tt.nw.Write(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("%s = %v, want %v", fnName, gotN, tt.wantN)
			}
		})
	}
}
