package internal

import (
	"os"
	"path/filepath"
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
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
			}
		})
	}
}

func TestPlainFileExists(t *testing.T) {
	fnName := "PlainFileExists()"
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
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestDirExists(t *testing.T) {
	fnName := "DirExists()"
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "no such directory", args: args{path: "no such dir"}, want: false},
		{name: "file exists, is not a directory", args: args{path: "fileio_test.go"}, want: false},
		{name: "file exists, is a directory", args: args{path: ".."}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DirExists(tt.args.path); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	fnName := "CopyFile()"
	topDir := "copies"
	if err := Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	defer func() {
		DestroyDirectoryForTesting(fnName, topDir)
	}()
	srcName := "source.txt"
	srcPath := filepath.Join(topDir, srcName)
	if err := CreateFileForTesting(topDir, srcName); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, srcPath, err)
	}
	type args struct {
		src  string
		dest string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "copy non-existent file", args: args{src: "foo.txt2", dest: "f.txt"}, wantErr: true},
		{
			name:    "copy to non-existent dir",
			args:    args{src: srcPath, dest: filepath.Join(topDir, "non-existent-dir", "foo.txt")},
			wantErr: true,
		},
		{name: "good copy", args: args{src: srcPath, dest: filepath.Join(topDir, "new.txt")}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CopyFile(tt.args.src, tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
			}
		})
	}
}
