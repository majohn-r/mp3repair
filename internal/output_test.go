package internal

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestNewOutputDevice(t *testing.T) {
	fnName := "NewOutputDevice()"
	tests := []struct {
		name        string
		want        *OutputDevice
		wantWStdout string
		wantWStderr string
	}{
		{
			name:        "normal",
			want:        &OutputDevice{consoleWriter: os.Stdout, errorWriter: os.Stderr, logWriter: productionLogger{}},
			wantWStdout: "hello to console",
			wantWStderr: "hello to error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewOutputDevice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
			testDevice := NewOutputDevice()
			var o any = testDevice
			if _, ok := o.(OutputBus); !ok {
				t.Errorf("%s: does not implement OutputBus", fnName)
			}
			// exercise log functionality
			testDevice.LogWriter().Info("info message", map[string]any{"foo": "INFO"})
			testDevice.LogWriter().Error("errpr message", map[string]any{"foo": "ERROR"})
		})
	}
}

func TestOutputDevice_ErrorWriter(t *testing.T) {
	fnName := "OutputDevice.ErrorWriter()"
	tests := []struct {
		name string
		o    *OutputDevice
		want io.Writer
	}{
		{
			name: "normal",
			o:    NewOutputDevice(),
			want: os.Stderr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.ErrorWriter(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
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
		w    *bytes.Buffer
		args
		want string
	}{
		{
			name: "broad test",
			w:    &bytes.Buffer{},
			args: args{
				format: "test format %d %q %v..?!..?\n\n\n\n",
				a:      []any{25, "foo", 1.245},
			},
			want: "Test format 25 \"foo\" 1.245?\n",
		},
		{
			name: "narrow test",
			w:    &bytes.Buffer{},
			args: args{
				format: "1. test format %d %q %v",
				a:      []any{25, "foo", 1.245},
			},
			want: "1. test format 25 \"foo\" 1.245.\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OutputDevice{errorWriter: tt.w}
			o.WriteError(tt.args.format, tt.args.a...)
			if got := tt.w.String(); got != tt.want {
				t.Errorf("%s got %q want %q", fnName, got, tt.want)
			}
		})
	}
}

func TestOutputDevice_WriteConsole(t *testing.T) {
	fnName := "OutputDevice.WriteConsole()"
	type args struct {
		strict bool
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
				strict: true,
				format: "test %s...\n\n",
				a:      []any{"foo."},
			},
			want: "Test foo.\n",
		},
		{
			name: "lax rules",
			w:    &bytes.Buffer{},
			args: args{
				strict: false,
				format: "test %s...\n\n",
				a:      []any{"foo."},
			},
			want: "test foo....\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OutputDevice{consoleWriter: tt.w}
			o.WriteConsole(tt.args.strict, tt.args.format, tt.args.a...)
			if got := tt.w.String(); got != tt.want {
				t.Errorf("%s: got %q want %q", fnName, got, tt.want)
			}
		})
	}
}
