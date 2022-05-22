package internal

import (
	"os"
	"testing"
)

func TestMkdir(t *testing.T) {
	fnName := "Mkdir()"
	topDir := "artificalDir"
	defer func() {
		if err := os.RemoveAll(topDir); err != nil {
			t.Errorf("%s error destroying test directory %q: %v", fnName, topDir, err)
		}
	}()
	type args struct {
		dirName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "failure", args: args{dirName: "testutilities_test.go"}, wantErr: true},
		{name: "success", args: args{dirName: topDir}, wantErr: false},
		// previous test will have created 'topDir' ... subsequent attempt should not fail
		{name: "directory exists", args: args{dirName: topDir}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Mkdir(tt.args.dirName); (err != nil) != tt.wantErr {
				t.Errorf("%q error = %v, wantErr %v", fnName, err, tt.wantErr)
			}
		})
	}
}

func TestPlainFileExists(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "plain file", args: args{path: "fileio_test.go"}, want: true},
		{name: "non-existent file", args: args{path: "no such file"}, want: false},
		{name: "directory", args: args{path: "."}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PlainFileExists(tt.args.path); got != tt.want {
				t.Errorf("PlainFileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
