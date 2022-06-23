package main

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func Test_run(t *testing.T) {
	type args struct {
		cmdlineArgs []string
	}
	tests := []struct {
		name            string
		args            args
		wantReturnValue int
	}{
		{name: "failure", args: args{cmdlineArgs: []string{"./mp3", "foo"}}, wantReturnValue: 1},
		{name: "success", args: args{cmdlineArgs: []string{"./mp3"}}, wantReturnValue: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotReturnValue := run(tt.args.cmdlineArgs); gotReturnValue != tt.wantReturnValue {
				t.Errorf("run() = %v, want %v", gotReturnValue, tt.wantReturnValue)
			}
		})
	}
}

func Test_report(t *testing.T) {
	creation = time.Now().Format(time.RFC3339)
	version = "test"
	type args struct {
		returnValue int
	}
	tests := []struct {
		name  string
		args  args
		wantW string
	}{
		{name: "success", args: args{returnValue: 0}, wantW: ""},
		{name: "failure", args: args{returnValue: 1}, wantW: fmt.Sprintf(statusFormat, "mp3", version, creation)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			report(w, tt.args.returnValue)
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("report() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
