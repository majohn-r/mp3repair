package commands

import (
	"mp3/internal"
	"os"
	"path/filepath"
	"testing"

	"github.com/majohn-r/output"
)

func Test_findAppFolder(t *testing.T) {
	fnName := "findAppFolder()"
	savedDirtyFolderFound := dirtyFolderFound
	savedDirtyFolder := dirtyFolder
	savedDirtyFolderValid := dirtyFolderValid
	savedAppSpecificPath, savedAppSpecificPathValid := internal.AppSpecificPath()
	defer func() {
		dirtyFolderFound = savedDirtyFolderFound
		dirtyFolder = savedDirtyFolder
		dirtyFolderValid = savedDirtyFolderValid
		internal.SetAppSpecificPathForTesting(savedAppSpecificPath, savedAppSpecificPathValid)
	}()
	tests := []struct {
		name                        string
		initialAppSpecificPath      string
		initialAppSpecificPathValid bool
		initialDirtyFolder          string
		initialDirtyFolderFound     bool
		initialDirtyFolderValid     bool
		wantFolder                  string
		wantOk                      bool
	}{
		{
			name:                        "expected first call",
			initialAppSpecificPath:      ".",
			initialAppSpecificPathValid: true,
			initialDirtyFolder:          "",
			initialDirtyFolderFound:     false,
			initialDirtyFolderValid:     false,
			wantFolder:                  ".",
			wantOk:                      true,
		},
		{
			name:                        "expected second call",
			initialAppSpecificPath:      ".",
			initialAppSpecificPathValid: true,
			initialDirtyFolder:          "",
			initialDirtyFolderFound:     true,
			initialDirtyFolderValid:     false,
			wantFolder:                  "",
			wantOk:                      false,
		},
		{
			name:                        "invalid folder",
			initialAppSpecificPath:      "no such file/mp3",
			initialAppSpecificPathValid: true,
			initialDirtyFolder:          "",
			initialDirtyFolderFound:     false,
			initialDirtyFolderValid:     false,
			wantFolder:                  "",
			wantOk:                      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirtyFolder = tt.initialDirtyFolder
			dirtyFolderFound = tt.initialDirtyFolderFound
			dirtyFolderValid = tt.initialDirtyFolderValid
			internal.SetAppSpecificPathForTesting(tt.initialAppSpecificPath, tt.initialAppSpecificPathValid)
			gotFolder, gotOk := findAppFolder()
			if gotFolder != tt.wantFolder {
				t.Errorf("%s gotFolder = %v, want %v", fnName, gotFolder, tt.wantFolder)
			}
			if gotOk != tt.wantOk {
				t.Errorf("%s gotOk = %v, want %v", fnName, gotOk, tt.wantOk)
			}
		})
	}
}

func Test_markDirty(t *testing.T) {
	fnName := "markDirty()"
	savedDirtyFolderFound := dirtyFolderFound
	savedDirtyFolder := dirtyFolder
	savedDirtyFolderValid := dirtyFolderValid
	testDir := "markDirty"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		dirtyFolderFound = savedDirtyFolderFound
		dirtyFolder = savedDirtyFolder
		dirtyFolderValid = savedDirtyFolderValid
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	tests := []struct {
		name               string
		command            string
		initialDirtyFolder string
		output.WantedRecording
	}{
		{
			name:               "typical first use",
			command:            "calling command",
			initialDirtyFolder: testDir,
			WantedRecording: output.WantedRecording{
				Log: "level='info' fileName='markDirty\\metadata.dirty' msg='metadata dirty file written'\n",
			},
		},
		{
			name:               "error first case",
			command:            "calling command",
			initialDirtyFolder: "no such dir",
			WantedRecording: output.WantedRecording{
				Error: "The file \"no such dir\\\\metadata.dirty\" cannot be created: open no such dir\\metadata.dirty: The system cannot find the path specified.\n",
				Log:   "level='error' command='calling command' error='open no such dir\\metadata.dirty: The system cannot find the path specified.' fileName='no such dir\\metadata.dirty' msg='cannot create file'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirtyFolderFound = true // short-circuit finding the folder
			dirtyFolderValid = true // yes, of course it's valid (even if it isn't)
			dirtyFolder = tt.initialDirtyFolder
			o := output.NewRecorder()
			markDirty(o, tt.command)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}

func Test_dirty(t *testing.T) {
	fnName := "dirty()"
	testDir := "dirty"
	savedDirtyFolderFound := dirtyFolderFound
	savedDirtyFolder := dirtyFolder
	savedDirtyFolderValid := dirtyFolderValid
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	defer func() {
		dirtyFolderFound = savedDirtyFolderFound
		dirtyFolder = savedDirtyFolder
		dirtyFolderValid = savedDirtyFolderValid
		internal.DestroyDirectoryForTesting(fnName, testDir)
	}()
	if err := internal.CreateFileForTestingWithContent(testDir, dirtyFileName, []byte("dirty")); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, dirtyFileName, err)
	}
	tests := []struct {
		name                    string
		initialDirtyFolder      string
		initialDirtyFolderValid bool
		wantDirty               bool
	}{
		{
			name:                    "file definitively exists",
			initialDirtyFolder:      testDir,
			initialDirtyFolderValid: true,
			wantDirty:               true,
		},
		{
			name:                    "file definitely does not exist",
			initialDirtyFolder:      ".",
			initialDirtyFolderValid: true,
			wantDirty:               false,
		},
		{
			name:                    "don't know",
			initialDirtyFolder:      "",
			initialDirtyFolderValid: false,
			wantDirty:               true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirtyFolderFound = true // short-circuit finding the folder
			dirtyFolderValid = tt.initialDirtyFolderValid
			dirtyFolder = tt.initialDirtyFolder
			if gotDirty := dirty(); gotDirty != tt.wantDirty {
				t.Errorf("%s = %t, want %t", fnName, gotDirty, tt.wantDirty)
			}
		})
	}
}

func Test_clearDirty(t *testing.T) {
	fnName := "clearDirty()"
	testDir := "clearDirty"
	savedDirtyFolderFound := dirtyFolderFound
	savedDirtyFolder := dirtyFolder
	savedDirtyFolderValid := dirtyFolderValid
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	if err := internal.CreateFileForTestingWithContent(testDir, dirtyFileName, []byte("dirty")); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, dirtyFileName, err)
	}
	// create another file structure with a dirty file that is open for reading
	testDir2 := "clearDirty2"
	if err := internal.Mkdir(testDir2); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir2, err)
	}
	if err := internal.CreateFileForTestingWithContent(testDir2, dirtyFileName, []byte("dirty")); err != nil {
		t.Errorf("%s error creating second %q: %v", fnName, dirtyFileName, err)
	}
	openFile, err := os.Open(filepath.Join(testDir2, dirtyFileName))
	if err != nil {
		t.Errorf("%s error opening second %q: %v", fnName, dirtyFileName, err)
	}
	defer func() {
		dirtyFolderFound = savedDirtyFolderFound
		dirtyFolder = savedDirtyFolder
		dirtyFolderValid = savedDirtyFolderValid
		internal.DestroyDirectoryForTesting(fnName, testDir)
		openFile.Close()
		internal.DestroyDirectoryForTesting(fnName, testDir2)
	}()
	tests := []struct {
		name                    string
		initialDirtyFolder      string
		initialDirtyFolderValid bool
		output.WantedRecording
	}{
		{
			name:                    "successful removal",
			initialDirtyFolder:      testDir,
			initialDirtyFolderValid: true,
			WantedRecording: output.WantedRecording{
				Log: "level='info' fileName='clearDirty\\metadata.dirty' msg='metadata dirty file deleted'\n",
			},
		},
		{
			name:                    "nothing to remove",
			initialDirtyFolder:      ".",
			initialDirtyFolderValid: true,
		},
		{
			name:                    "unremovable file",
			initialDirtyFolder:      testDir2,
			initialDirtyFolderValid: true,
			WantedRecording: output.WantedRecording{
				Error: "The file \"clearDirty2\\\\metadata.dirty\" cannot be deleted: remove clearDirty2\\metadata.dirty: The process cannot access the file because it is being used by another process.\n",
				Log:   "level='error' error='remove clearDirty2\\metadata.dirty: The process cannot access the file because it is being used by another process.' fileName='clearDirty2\\metadata.dirty' msg='cannot delete file'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirtyFolderFound = true // short-circuit finding the folder
			dirtyFolderValid = tt.initialDirtyFolderValid
			dirtyFolder = tt.initialDirtyFolder
			o := output.NewRecorder()
			clearDirty(o)
			if issues, ok := o.Verify(tt.WantedRecording); !ok {
				for _, issue := range issues {
					t.Errorf("%s %s", fnName, issue)
				}
			}
		})
	}
}
