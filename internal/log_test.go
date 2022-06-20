package internal

import (
	"bytes"
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
	fnName := "configureLogging()"
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
				DestroyDirectoryForTesting(fnName, tt.args.path)
			}()
			got := ConfigureLogging(tt.args.path)
			if got == nil {
				t.Errorf("%s: returned nil cronoWriter.CronoWriter", fnName)
			}
			got.Close()
			var latestFound bool
			var mp3logFound bool
			if files, err := ioutil.ReadDir(tt.args.path); err != nil {
				t.Errorf("%s: cannot read path %v: %v", fnName, tt.args.path, err)
			} else {
				for _, file := range files {
					mode := file.Mode()
					fileName := file.Name()
					if mode.IsRegular() && strings.HasPrefix(fileName, logFilePrefix) && strings.HasSuffix(fileName, logFileExtension) {
						mp3logFound = true

					} else if fs.ModeSymlink == (mode&fs.ModeSymlink) && fileName == symlinkName {
						latestFound = true
					}
				}
			}
			if !latestFound {
				t.Errorf("%s: %s not found", fnName, symlinkName)
			}
			if !mp3logFound {
				t.Errorf("%s: no %s*%s files found", fnName, logFilePrefix, logFileExtension)
			}
		})
	}
}

func TestCleanupLogFiles(t *testing.T) {
	fnName := "CleanupLogFiles()"
	type args struct {
		path string
	}
	tests := []struct {
		name          string
		fileCount     int
		createFolder  bool
		lockFiles     bool
		args          args
		wantFileCount int
		wantWErr      string
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
			name:          "locked",
			fileCount:     maxLogFiles + 2,
			createFolder:  true,
			lockFiles:     true,
			args:          args{path: "testlogs"},
			wantFileCount: maxLogFiles + 2,
			wantWErr:      "The log file \"testlogs\\\\mp3.00.log\" cannot be deleted: remove testlogs\\mp3.00.log: The process cannot access the file because it is being used by another process.\nThe log file \"testlogs\\\\mp3.01.log\" cannot be deleted: remove testlogs\\mp3.01.log: The process cannot access the file because it is being used by another process.\n",
		},
		{
			name:     "missing path",
			args:     args{path: "testlogs"},
			wantWErr: "The log file directory \"testlogs\" cannot be read: open testlogs: The system cannot find the file specified.\n",
		},
	}
	for _, tt := range tests {
		defer func() {
			DestroyDirectoryForTesting(fnName, tt.args.path)
		}()
		var filesToClose []*os.File
		if tt.createFolder {
			if err := os.MkdirAll(tt.args.path, 0755); err != nil {
				t.Errorf("%s: cannot create %v: %v", fnName, tt.args.path, err)
			}
			// create required files
			for k := 0; k < tt.fileCount; k++ {
				filename := filepath.Join(tt.args.path, logFilePrefix+fmt.Sprintf("%02d", k)+logFileExtension)
				if file, err := os.Create(filename); err != nil {
					t.Errorf("%s: cannot create log file %q: %v", fnName, filename, err)
				} else {
					tm := time.Now().Add(time.Hour * time.Duration(k))
					if err := os.Chtimes(filename, tm, tm); err != nil {
						t.Errorf("%s: cannot set access and modification times on %s: %v", fnName, filename, err)
					}
					if !tt.lockFiles {
						file.Close()
					} else {
						// this delay in closing the files will exercise the
						// path where the cleanup code cannot delete files
						// because they're locked
						filesToClose = append(filesToClose, file)
					}
				}
			}
		} else {
			if err := os.RemoveAll(tt.args.path); err != nil {
				t.Errorf("%s error destroying test directory %q: %v", fnName, tt.args.path, err)
			}
		}
		t.Run(tt.name, func(t *testing.T) {
			wErr := &bytes.Buffer{}
			CleanupLogFiles(wErr, tt.args.path)
			gotWErr := wErr.String()
			if strings.Contains(gotWErr, ".log") {
				for k := 0; k < tt.fileCount; k++ {
					// search := fmt.Sprintf("mp3.%02d.log", k)
					// gotWErr = strings.ReplaceAll(gotWErr, search, "mp3.xx.log")
				}
			}
			if gotWErr != tt.wantWErr {
				t.Errorf("%s = %v, want %v", fnName, gotWErr, tt.wantWErr)
			}
			if tt.createFolder {
				if files, err := ioutil.ReadDir(tt.args.path); err != nil {
					t.Errorf("%s: cannot read directory %s: %v", fnName, tt.args.path, err)
				} else {
					var gotFileCount int
					for _, file := range files {
						fileName := file.Name()
						if file.Mode().IsRegular() && strings.HasPrefix(fileName, logFilePrefix) && strings.HasSuffix(fileName, logFileExtension) {
							gotFileCount++
						}
					}
					if gotFileCount != tt.wantFileCount {
						t.Errorf("%s: file count got %d, want %d", fnName, gotFileCount, tt.wantFileCount)
					}
				}
				for _, file := range filesToClose {
					file.Close()
				}
			} else {
				if _, err := os.Stat(tt.args.path); err != nil {
					if !os.IsNotExist(err) {
						t.Errorf("%s: expected %s to not exist, but got error %v", fnName, tt.args.path, err)
					}
				} else {
					t.Errorf("%s: expected %s to not exist!", fnName, tt.args.path)
				}
			}
		})
	}
}
