package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

func Test_initLogging(t *testing.T) {
	fnName := "initLogging()"
	type args struct {
		parentDir string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "current directory", args: args{parentDir: "."}, want: true},
		{name: "non-existent directory", args: args{parentDir: "main_test.go"}, want: false},
	}
	defer func() {
		logger.Close()
		dirName := "mp3"
		if err := os.RemoveAll(dirName); err != nil {
			t.Errorf("%s error destroying test directory %q: %v", fnName, dirName, err)
		}
	}()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := initLogging(tt.args.parentDir); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func Test_initEnv(t *testing.T) {
	fnName := "initEnv()"
	type args struct {
		lookup func() []error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "no errors", args: args{lookup: func() []error { return nil }}, want: true},
		{name: "process errors", args: args{lookup: func() []error {
			var e []error
			e = append(e, errors.New("error 1"))
			e = append(e, errors.New("error 2"))
			return e
		}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := initEnv(tt.args.lookup); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

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
		{name: "failure", args: args{returnValue: 1}, wantW: fmt.Sprintf(statusFormat, version, creation)},
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
