package internal

import (
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
			var o interface{} = testDevice
			if _, ok := o.(OutputBus); !ok {
				t.Errorf("%s: does not implement OutputBus", fnName)
			}
			// exercise log functionality
			testDevice.LogWriter().Info("info message", map[string]interface{}{"foo": "INFO"})
			testDevice.LogWriter().Warn("warn message", map[string]interface{}{"foo": "WARN"})
			testDevice.LogWriter().Error("errpr message", map[string]interface{}{"foo": "ERROR"})
		})
	}
}

func TestOutputDevice_ConsoleWriter(t *testing.T) {
	fnName := "OutputDevice.ConsoleWriter()"
	tests := []struct {
		name string
		o    *OutputDevice
		want io.Writer
	}{
		{
			name: "normal",
			o:    NewOutputDevice(),
			want: os.Stdout,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.o.ConsoleWriter(); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
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
