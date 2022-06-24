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
			want:        &OutputDevice{wOut: os.Stdout, wErr: os.Stderr},
			wantWStdout: "hello to stdout",
			wantWStderr: "hello to stderr",
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
			fmt.Fprint(testDevice.OutputWriter(), "hello to stdout")
			fmt.Fprintf(testDevice.ErrorWriter(), "hello to stderr")
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
			testDevice.Log(INFO, "info message", map[string]interface{}{"foo": INFO})
			testDevice.Log(WARN, "warn message", map[string]interface{}{"foo": WARN})
			testDevice.Log(ERROR, "errpr message", map[string]interface{}{"foo": ERROR})
			testDevice.Log(ERROR+WARN+INFO, "info message", map[string]interface{}{"foo": ERROR + WARN + INFO})
		})
	}
}

func TestOutputDevice_OutputWriter(t *testing.T) {
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
			if got := tt.o.OutputWriter(); got != tt.want {
				t.Errorf("OutputDevice.OutputWriter() = %v, want %v", got, tt.want)
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

func TestOutputDevice_Log(t *testing.T) {
	type args struct {
		l      LogLevel
		msg    string
		fields map[string]interface{}
	}
	tests := []struct {
		name string
		o    *OutputDevice
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.o.Log(tt.args.l, tt.args.msg, tt.args.fields)
		})
	}
}
