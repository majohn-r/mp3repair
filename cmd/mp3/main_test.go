package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
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

func Test_initEnv(t *testing.T) {
	type args struct {
		lookup func(w io.Writer) bool
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		wantW string
	}{
		{name: "no errors", args: args{lookup: func(w io.Writer) bool { return true }}, want: true},
		{name: "process errors", args: args{lookup: func(w io.Writer) bool {
			return false
		}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if got := initEnv(w, tt.args.lookup); got != tt.want {
				t.Errorf("initEnv() = %v, want %v", got, tt.want)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("initEnv() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}

func Test_initLogging(t *testing.T) {
	defer func() {
		logger.Close()
		dirName := "mp3"
		if err := os.RemoveAll(dirName); err != nil {
			t.Errorf("initLogging() error destroying test directory %q: %v", dirName, err)
		}
	}()
	type args struct {
		parentDir string
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		wantW string
	}{
		{name: "current directory", args: args{parentDir: "."}, want: true, wantW: ""},
		{
			name:  "non-existent directory",
			args:  args{parentDir: "main_test.go"},
			want:  false,
			wantW: `The directory "main_test.go\\mp3\\logs" cannot be created: mkdir main_test.go: The system cannot find the path specified..` + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if got := initLogging(w, tt.args.parentDir); got != tt.want {
				t.Errorf("initLogging() = %v, want %v", got, tt.want)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("initLogging() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
