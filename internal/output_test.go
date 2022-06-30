package internal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestNewOutputDevice(t *testing.T) {
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
			if got := NewOutputDevice(os.Stdout, os.Stderr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOutputDevice() = %v, want %v", got, tt.want)
			}
			wStdout := &bytes.Buffer{}
			wStderr := &bytes.Buffer{}
			testDevice := NewOutputDevice(wStdout, wStderr)
			fmt.Fprint(testDevice.ConsoleWriter(), "hello to console")
			fmt.Fprintf(testDevice.ErrorWriter(), "hello to error")
			if gotWStdout := wStdout.String(); gotWStdout != tt.wantWStdout {
				t.Errorf("NewOutputDevice() = %v, want %v", gotWStdout, tt.wantWStdout)
			}
			if gotWStderr := wStderr.String(); gotWStderr != tt.wantWStderr {
				t.Errorf("NewOutputDevice() = %v, want %v", gotWStderr, tt.wantWStderr)
			}
			var o interface{} = testDevice
			if _, ok := o.(OutputBus); !ok {
				t.Errorf("NewOutputDevice() does not implement OutputBus")
			}
			// exercise log functionality
			testDevice.LogWriter().Info("info message", map[string]interface{}{"foo": "INFO"})
			testDevice.LogWriter().Warn("warn message", map[string]interface{}{"foo": "WARN"})
			testDevice.LogWriter().Error("errpr message", map[string]interface{}{"foo": "ERROR"})
		})
	}
}

func TestOutputDevice_ConsoleWriter(t *testing.T) {
	tests := []struct {
		name string
		o    *OutputDevice
		want io.Writer
	}{
		{
			name: "normal",
			o:    NewOutputDevice(os.Stdout, os.Stderr),
			want: os.Stdout,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.ConsoleWriter(); got != tt.want {
				t.Errorf("OutputDevice.ConsoleWriter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputDevice_ErrorWriter(t *testing.T) {
	tests := []struct {
		name string
		o    *OutputDevice
		want io.Writer
	}{
		{
			name: "normal",
			o:    NewOutputDevice(os.Stdout, os.Stderr),
			want: os.Stderr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.ErrorWriter(); got != tt.want {
				t.Errorf("OutputDevice.ErrorWriter() = %v, want %v", got, tt.want)
			}
		})
	}
}
