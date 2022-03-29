package main

import (
	"errors"
	"os"
	"testing"
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
		if err := os.RemoveAll("mp3"); err != nil {
			t.Errorf("%s: %v", fnName, err)
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
