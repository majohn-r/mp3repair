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
			got := configureLogging(tt.args.path)
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
		name              string
		fileCount         int
		createFolder      bool
		lockFiles         bool
		args              args
		wantFileCount     int
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
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
			wantLogOutput: "level='info' directory='testlogs' fileName='mp3.00.log' msg='successfully deleted file'\n" +
				"level='info' directory='testlogs' fileName='mp3.01.log' msg='successfully deleted file'\n",
		},
		{
			name:            "locked",
			fileCount:       maxLogFiles + 2,
			createFolder:    true,
			lockFiles:       true,
			args:            args{path: "testlogs"},
			wantFileCount:   maxLogFiles + 2,
			wantErrorOutput: "The log file \"testlogs\\\\mp3.00.log\" cannot be deleted: remove testlogs\\mp3.00.log: The process cannot access the file because it is being used by another process.\nThe log file \"testlogs\\\\mp3.01.log\" cannot be deleted: remove testlogs\\mp3.01.log: The process cannot access the file because it is being used by another process.\n",
			wantLogOutput: "level='warn' directory='testlogs' error='remove testlogs\\mp3.00.log: The process cannot access the file because it is being used by another process.' fileName='mp3.00.log' msg='cannot delete file'\n" +
				"level='warn' directory='testlogs' error='remove testlogs\\mp3.01.log: The process cannot access the file because it is being used by another process.' fileName='mp3.01.log' msg='cannot delete file'\n",
		},
		{
			name:            "missing path",
			args:            args{path: "testlogs"},
			wantErrorOutput: "The log file directory \"testlogs\" cannot be read: open testlogs: The system cannot find the file specified.\n",
			wantLogOutput:   "level='warn' directory='testlogs' error='open testlogs: The system cannot find the file specified.' msg='cannot read directory'\n",
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
			outputDevice := NewOutputDeviceForTesting()
			cleanupLogFiles(outputDevice, tt.args.path)
			if gotErrorOutput := outputDevice.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("%s error output = %v, want %v", fnName, gotErrorOutput, tt.wantErrorOutput)
			}
			if gotConsoleOutput := outputDevice.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("%s console output = %v, want %v", fnName, gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotLogOutput := outputDevice.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("%s log output = %v, want %v", fnName, gotLogOutput, tt.wantLogOutput)
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

func TestInitLogging(t *testing.T) {
	savedStates := []*SavedEnvVar{SaveEnvVarForTesting("TMP"), SaveEnvVarForTesting("TEMP")}
	dirName := filepath.Join(".", "InitLogging")
	workDir := SecureAbsolutePathForTesting(dirName)
	if err := Mkdir(workDir); err != nil {
		t.Errorf("InitLogging: failed to create test directory")
	}
	defer func() {
		for _, state := range savedStates {
			state.RestoreForTesting()
		}
		logger.Close()
		if err := os.RemoveAll(workDir); err != nil {
			t.Errorf("InitLogging() error destroying test directory %q: %v", workDir, err)
		}
	}()
	thisFile := SecureAbsolutePathForTesting("log_test.go")
	tests := []struct {
		name              string
		state             []*SavedEnvVar
		want              bool
		wantConsoleOutput string
		wantErrorOutput   string
		wantLogOutput     string
	}{
		{
			name:  "useTmp",
			state: []*SavedEnvVar{{Name: "TMP", Value: workDir, Set: true}, {Name: "TEMP"}},
			want:  true,
		},
		{
			name:  "useTemp",
			state: []*SavedEnvVar{{Name: "TEMP", Value: workDir, Set: true}, {Name: "TEMP"}},
			want:  true,
		},
		{
			name:            "no temps",
			state:           []*SavedEnvVar{{Name: "TMP"}, {Name: "TEMP"}},
			wantErrorOutput: USER_NO_TEMP_FOLDER,
		},
		{
			name:  "cannot create dir",
			state: []*SavedEnvVar{{Name: "TMP", Value: thisFile, Set: true}, {Name: "TEMP"}},
			wantErrorOutput: fmt.Sprintf(
				"The directory %q cannot be created: mkdir %s: The system cannot find the path specified..\n",
				filepath.Join(thisFile, AppName, logDirName), thisFile),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if logger != nil {
				logger.Close()
			}
			for _, s := range tt.state {
				s.RestoreForTesting()
			}
			outputDevice := NewOutputDeviceForTesting()
			if got := InitLogging(outputDevice); got != tt.want {
				t.Errorf("InitLogging() = %v, want %v", got, tt.want)
			}
			if gotConsoleOutput := outputDevice.ConsoleOutput(); gotConsoleOutput != tt.wantConsoleOutput {
				t.Errorf("InitLogging() output = %v, want %v", gotConsoleOutput, tt.wantConsoleOutput)
			}
			if gotErrorOutput := outputDevice.ErrorOutput(); gotErrorOutput != tt.wantErrorOutput {
				t.Errorf("InitLogging() error = %v, want %v", gotErrorOutput, tt.wantErrorOutput)
			}
			if gotLogOutput := outputDevice.LogOutput(); gotLogOutput != tt.wantLogOutput {
				t.Errorf("InitLogging() log output = %v, want %v", gotLogOutput, tt.wantLogOutput)
			}
			if logger != nil {
				logger.Close()
			}
		})
	}
}
