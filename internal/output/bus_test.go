package output

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestNewDefaultBus(t *testing.T) {
	fnName := "NewDefaultBus()"
	tests := []struct {
		name              string
		want              Bus
		wantConsoleWriter io.Writer
		wantErrorWriter   io.Writer
		wantLogWriter     Logger
	}{
		{
			name:              "normal",
			want:              NewCustomBus(os.Stdout, os.Stderr, NilLogger{}),
			wantConsoleWriter: os.Stdout,
			wantErrorWriter:   os.Stderr,
			wantLogWriter:     NilLogger{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDefaultBus(NilLogger{})
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

func Test_bus_Log(t *testing.T) {
	fnName := "bus.Log()"
	type args struct {
		l    Level
		msg  string
		args map[string]any
	}
	tests := []struct {
		name string
		args
		Log   string
		Error string
	}{
		{
			name: "trace",
			args: args{
				l:    Trace,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Log: "level='trace' f='v' msg='hello'\n",
		},
		{
			name: "debug",
			args: args{
				l:    Debug,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Log: "level='debug' f='v' msg='hello'\n",
		},
		{
			name: "info",
			args: args{
				l:    Info,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Log: "level='info' f='v' msg='hello'\n",
		},
		{
			name: "warning",
			args: args{
				l:    Warning,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Log: "level='warning' f='v' msg='hello'\n",
		},
		{
			name: "error",
			args: args{
				l:    Error,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Log: "level='error' f='v' msg='hello'\n",
		},
		{
			name: "panic",
			args: args{
				l:    Panic,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Log: "level='panic' f='v' msg='hello'\n",
		},
		{
			name: "fatal",
			args: args{
				l:    Fatal,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Log: "level='fatal' f='v' msg='hello'\n",
		},
		{
			name: "illegal",
			args: args{
				l:    Trace + 1,
				msg:  "hello",
				args: map[string]any{"f": "v"},
			},
			Error: "Programming error: call to bus.Log() with invalid level value 7; message: 'hello', args: 'map[f:v].\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eW := &bytes.Buffer{}
			l := NewRecordingLogger()
			o := NewCustomBus(nil, eW, l)
			o.Log(tt.args.l, tt.args.msg, tt.args.args)
			if got := l.writer.String(); got != tt.Log {
				t.Errorf("%s got log %q want %q", fnName, got, tt.Log)
			}
			if got := eW.String(); got != tt.Error {
				t.Errorf("%s got error %q want %q", fnName, got, tt.Error)
			}
		})
	}
}

func Test_bus_WriteCanonicalError(t *testing.T) {
	fnName := "bus.WriteCanonicalError()"
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
			o := &bus{errorChannel: w, performWrites: true}
			o.WriteCanonicalError(tt.args.format, tt.args.a...)
			if got := w.String(); got != tt.want {
				t.Errorf("%s got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func Test_bus_WriteError(t *testing.T) {
	fnName := "bus.WriteError()"
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
			o := &bus{errorChannel: w, performWrites: true}
			o.WriteError(tt.args.format, tt.args.a...)
			if got := w.String(); got != tt.want {
				t.Errorf("%s got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func Test_bus_WriteCanonicalConsole(t *testing.T) {
	fnName := "bus.WriteCanonicalConsole()"
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
			o := &bus{consoleChannel: tt.w, performWrites: true}
			o.WriteCanonicalConsole(tt.args.format, tt.args.a...)
			if got := tt.w.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func Test_bus_WriteConsole(t *testing.T) {
	fnName := "bus.WriteConsole()"
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
			o := &bus{consoleChannel: tt.w, performWrites: true}
			o.WriteConsole(tt.args.format, tt.args.a...)
			if got := tt.w.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}
