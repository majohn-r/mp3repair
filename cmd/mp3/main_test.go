package main

import (
	"fmt"
	"mp3/internal"
	"strings"
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
		hasOutput       bool
		wantErr         string
		wantLogPrefix   string
		wantLogSuffix   string
	}{
		{
			name:            "failure",
			args:            args{cmdlineArgs: []string{"./mp3", "foo"}},
			wantReturnValue: 1,
			wantErr:         "There is no command named \"foo\"; valid commands include [check ls postRepair repair].\n",
			wantLogPrefix: "level='info' args='[./mp3 foo]' timeStamp='' version='unknown version!' msg='execution starts'\n" +
				"level='info' duration='",
			wantLogSuffix: "' exitCode='1' msg='execution ends'\n",
		},
		{
			name:            "success",
			args:            args{cmdlineArgs: []string{"./mp3"}},
			wantReturnValue: 0,
			wantLogPrefix: "level='info' args='[./mp3]' timeStamp='' version='unknown version!' msg='execution starts'\n" +
				"level='info' duration='",
			wantLogSuffix: "' exitCode='0' msg='execution ends'\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			if gotReturnValue := run(o, tt.args.cmdlineArgs); gotReturnValue != tt.wantReturnValue {
				t.Errorf("run() = %v, want %v", gotReturnValue, tt.wantReturnValue)
			}
			if gotErr := o.Stderr(); gotErr != tt.wantErr {
				t.Errorf("run() error output = %v, want %v", gotErr, tt.wantErr)
			}
			gotOut := o.Stdout()
			if tt.hasOutput && len(gotOut) == 0 {
				t.Errorf("run() console output = %t, want %t", len(gotOut) == 0, tt.hasOutput)
			}
			gotLog := o.LogOutput()
			if !strings.HasPrefix(gotLog, tt.wantLogPrefix) {
				t.Errorf("run() log output %q does not start with %q", gotLog, tt.wantLogPrefix)
			}
			if !strings.HasSuffix(gotLog, tt.wantLogSuffix) {
				t.Errorf("run() log output %q does not end with %q", gotLog, tt.wantLogSuffix)
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
		name    string
		args    args
		wantOut string
		wantErr string
		wantLog string
	}{
		{name: "success", args: args{returnValue: 0}, wantErr: ""},
		{name: "failure", args: args{returnValue: 1}, wantErr: fmt.Sprintf(statusFormat, "mp3", version, creation)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := internal.NewOutputDeviceForTesting()
			report(o, tt.args.returnValue)
			if gotOut := o.Stdout(); gotOut != tt.wantOut {
				t.Errorf("report() console output = %v, want %v", gotOut, tt.wantOut)
			}
			if gotErr := o.Stderr(); gotErr != tt.wantErr {
				t.Errorf("report() error output = %v, want %v", gotErr, tt.wantErr)
			}
			if gotLog := o.LogOutput(); gotLog != tt.wantLog {
				t.Errorf("report() log output = %v, want %v", gotLog, tt.wantLog)
			}
		})
	}
}
