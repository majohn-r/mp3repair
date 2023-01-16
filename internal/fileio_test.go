package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/majohn-r/output"
)

func TestMkdir(t *testing.T) {
	const fnName = "Mkdir()"
	topDir := "artificalDir"
	defer func() {
		if err := os.RemoveAll(topDir); err != nil {
			t.Errorf("%s error destroying test directory %q: %v", fnName, topDir, err)
		}
	}()
	type args struct {
		dirName string
	}
	tests := map[string]struct {
		args
		wantErr bool
	}{
		"failure":          {args: args{dirName: "testutilities_test.go"}, wantErr: true},
		"success":          {args: args{dirName: topDir}, wantErr: false},
		"directory exists": {args: args{dirName: "."}, wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := Mkdir(tt.args.dirName); (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
			}
		})
	}
}

func TestPlainFileExists(t *testing.T) {
	const fnName = "PlainFileExists()"
	type args struct {
		path string
	}
	tests := map[string]struct {
		args
		want bool
	}{
		"plain file":        {args: args{path: "fileio_test.go"}, want: true},
		"non-existent file": {args: args{path: "no such file"}, want: false},
		"directory":         {args: args{path: "."}, want: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := PlainFileExists(tt.args.path); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestDirExists(t *testing.T) {
	const fnName = "DirExists()"
	type args struct {
		path string
	}
	tests := map[string]struct {
		args
		want bool
	}{
		"no such directory":               {args: args{path: "no such dir"}, want: false},
		"file exists, is not a directory": {args: args{path: "fileio_test.go"}, want: false},
		"file exists, is a directory":     {args: args{path: ".."}, want: true},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := DirExists(tt.args.path); got != tt.want {
				t.Errorf("%s = %v, want %v", fnName, got, tt.want)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	const fnName = "CopyFile()"
	topDir := "copies"
	if err := Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	srcName := "source.txt"
	srcPath := filepath.Join(topDir, srcName)
	if err := CreateFileForTesting(topDir, srcName); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, srcPath, err)
	}
	defer func() {
		DestroyDirectoryForTesting(fnName, topDir)
	}()
	type args struct {
		src  string
		dest string
	}
	tests := map[string]struct {
		args
		wantErr bool
	}{
		"copy non-existent file":   {args: args{src: "foo.txt2", dest: "f.txt"}, wantErr: true},
		"copy to non-existent dir": {args: args{src: srcPath, dest: filepath.Join(topDir, "non-existent-dir", "foo.txt")}, wantErr: true},
		"good copy":                {args: args{src: srcPath, dest: filepath.Join(topDir, "new.txt")}, wantErr: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := CopyFile(tt.args.src, tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("%s error = %v, wantErr %v", fnName, err, tt.wantErr)
			}
		})
	}
}

func TestReadDirectory(t *testing.T) {
	const fnName = "ReadDirectory()"
	// generate test data
	topDir := "loadTest"
	if err := Mkdir(topDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, topDir, err)
	}
	defer func() {
		DestroyDirectoryForTesting(fnName, topDir)
	}()
	type args struct {
		dir string
	}
	tests := map[string]struct {
		args
		wantFiles []fs.DirEntry
		wantOk    bool
		output.WantedRecording
	}{
		"default": {args: args{topDir}, wantFiles: []fs.DirEntry{}, wantOk: true},
		"non-existent dir": {
			args: args{"non-existent directory"},
			WantedRecording: output.WantedRecording{
				Error: "The directory \"non-existent directory\" cannot be read: open non-existent directory: The system cannot find the file specified.\n",
				Log:   "level='error' directory='non-existent directory' error='open non-existent directory: The system cannot find the file specified.' msg='cannot read directory'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			gotFiles, gotOk := ReadDirectory(o, tt.args.dir)
			if !reflect.DeepEqual(gotFiles, tt.wantFiles) {
				t.Errorf("%s gotFiles = %v, want %v", fnName, gotFiles, tt.wantFiles)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}
