package internal

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConfigureLogging(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "standard test case", args: args{path: "testlogs"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if err := os.RemoveAll(tt.args.path); err != nil {
					t.Errorf("ConfigureLogging(): cannot clean up %v: %v", tt.args.path, err)
				}
			}()
			got := ConfigureLogging(tt.args.path)
			if got == nil {
				t.Error("ConfigureLogging(): returned nil coronWriter.CronoWriter")
			}
			got.Close()
			var latestFound bool
			var mp3logFound bool
			if files, err := ioutil.ReadDir(tt.args.path); err != nil {
				t.Errorf("ConfigureLogging(): cannot read path %v: %v", tt.args.path, err)
			} else {
				for _, file := range files {
					mode := file.Mode()
					fileName := file.Name()
					if mode.IsRegular() && strings.HasPrefix(fileName, logFilePrefix) {
						mp3logFound = true

					} else if fs.ModeSymlink == (mode&fs.ModeSymlink) && fileName == symlinkName {
						latestFound = true
					}
				}
			}
			if !latestFound {
				t.Errorf("ConfigureLogging(): %s not found", symlinkName)
			}
			if !mp3logFound {
				t.Errorf("ConfigureLogging(): no %s* files found", logFilePrefix)
			}
		})
	}
}

func TestCleanupLogFiles(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name          string
		fileCount     int
		createFolder  bool
		args          args
		wantFileCount int
	}{
		{
			name:          "no work to do",
			fileCount:     maxLogFiles,
			createFolder:  true,
			args:          args{path: "testlogs"},
			wantFileCount: maxLogFiles,
		},
		{
			name:          "files to delete",
			fileCount:     maxLogFiles + 2,
			createFolder:  true,
			args:          args{path: "testlogs"},
			wantFileCount: maxLogFiles,
		},
		{
			name: "missing path",
			args: args{path: "testlogs"},
		},
	}
	for _, tt := range tests {
		defer func() {
			if err := os.RemoveAll(tt.args.path); err != nil {
				t.Errorf("CleanupLogFiles(): cannot clean up %v: %v", tt.args.path, err)
			}
		}()
		if tt.createFolder {
			if err := os.MkdirAll(tt.args.path, 0755); err != nil {
				t.Errorf("CleanupLogFiles(): cannot create %v: %v", tt.args.path, err)
			}
			// create required files
			for k := 0; k < tt.fileCount; k++ {
				filename := filepath.Join(tt.args.path, logFilePrefix+fmt.Sprintf("%2d", k))
				if file, err := os.Create(filename); err != nil {
					t.Errorf("CleanupLogFiles(): cannot create log file %q: %v", filename, err)
				} else {
					tm := time.Now().Add(time.Hour * time.Duration(k))
					if err := os.Chtimes(filename, tm, tm); err != nil {
						t.Errorf("CleanupLogFiles(): cannot set access and modification times on %s: %v", filename, err)
					}
					file.Close()
				}
			}
		} else {
			if err := os.RemoveAll(tt.args.path); err != nil {
				t.Errorf("CleanupLogFiles(): cannot ensure %v does not exist: %v", tt.args.path, err)
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			CleanupLogFiles(tt.args.path)
			if tt.createFolder {
				if files, err := ioutil.ReadDir(tt.args.path); err != nil {
					t.Errorf("CleanupLogFiles(): cannot read directory %s: %v", tt.args.path, err)
				} else {
					var gotFileCount int
					for _, file := range files {
						if file.Mode().IsRegular() && strings.HasPrefix(file.Name(), logFilePrefix) {
							gotFileCount++
						}
					}
					if gotFileCount != tt.wantFileCount {
						t.Errorf("CleanupLogFiles(): file count got %d, want %d", gotFileCount, tt.wantFileCount)
					}
				}
			} else {
				if _, err := os.Stat(tt.args.path); err != nil {
					if !os.IsNotExist(err) {
						t.Errorf("CleanupLogFiles(): expected %s to not exist, but got error %v", tt.args.path, err)
					}
				} else {
					t.Errorf("CleanupLogFiles(): expected %s to not exist!", tt.args.path)
				}
			}
		})
	}
}
