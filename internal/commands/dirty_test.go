package commands

import (
	"mp3/internal"
	"os"
	"path/filepath"
	"testing"

	"github.com/majohn-r/output"
)

func Test_markDirty(t *testing.T) {
	const fnName = "markDirty()"
	emptyDir := "empty"
	if err := internal.Mkdir(emptyDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, emptyDir, err)
	}
	filledDir := "filled"
	if err := internal.Mkdir(filledDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, filledDir, err)
	}
	if err := internal.CreateFileForTesting(filledDir, dirtyFileName); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, filepath.Join(filledDir, dirtyFileName), err)
	}
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, emptyDir)
		internal.DestroyDirectoryForTesting(fnName, filledDir)
	}()
	type args struct {
		cmd string
	}
	tests := []struct {
		name    string
		appPath string
		args
		output.WantedRecording
	}{
		{
			name:    "typical first use",
			appPath: emptyDir,
			args:    args{cmd: "calling command"},
			WantedRecording: output.WantedRecording{
				Log: "level='info' fileName='empty\\metadata.dirty' msg='metadata dirty file written'\n",
			},
		},
		{
			name:    "typical second use",
			appPath: filledDir,
			args:    args{cmd: "calling command"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := output.NewRecorder()
			internal.SetApplicationPathForTesting(tt.appPath)
			markDirty(o, tt.args.cmd)
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
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	oldAppPath := internal.SetApplicationPathForTesting(testDir)
	defer func() {
		internal.DestroyDirectoryForTesting(fnName, testDir)
		internal.SetApplicationPathForTesting(oldAppPath)
	}()

	tests := []struct {
		name      string
		wantDirty bool
	}{
		{
			name:      "file definitively exists",
			wantDirty: true,
		},
		{
			name:      "file definitely does not exist",
			wantDirty: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantDirty {
				if err := internal.CreateFileForTestingWithContent(testDir, dirtyFileName, []byte("dirty")); err != nil {
					t.Errorf("%s error creating %q: %v", fnName, dirtyFileName, err)
				}
			} else {
				os.Remove(filepath.Join(testDir, dirtyFileName))
			}
			if gotDirty := dirty(); gotDirty != tt.wantDirty {
				t.Errorf("%s = %t, want %t", fnName, gotDirty, tt.wantDirty)
			}
		})
	}
}

func Test_clearDirty(t *testing.T) {
	fnName := "clearDirty()"
	testDir := "clearDirty"
	if err := internal.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	oldAppPath := internal.ApplicationPath()
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
		internal.SetApplicationPathForTesting(oldAppPath)
		internal.DestroyDirectoryForTesting(fnName, testDir)
		openFile.Close()
		internal.DestroyDirectoryForTesting(fnName, testDir2)
	}()
	tests := []struct {
		name               string
		initialDirtyFolder string
		output.WantedRecording
	}{
		{
			name:               "successful removal",
			initialDirtyFolder: testDir,
			WantedRecording: output.WantedRecording{
				Log: "level='info' fileName='clearDirty\\metadata.dirty' msg='metadata dirty file deleted'\n",
			},
		},
		{
			name:               "nothing to remove",
			initialDirtyFolder: ".",
		},
		{
			name:               "unremovable file",
			initialDirtyFolder: testDir2,
			WantedRecording: output.WantedRecording{
				Error: "The file \"clearDirty2\\\\metadata.dirty\" cannot be deleted: remove clearDirty2\\metadata.dirty: The process cannot access the file because it is being used by another process.\n",
				Log:   "level='error' error='remove clearDirty2\\metadata.dirty: The process cannot access the file because it is being used by another process.' fileName='clearDirty2\\metadata.dirty' msg='cannot delete file'\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			internal.SetApplicationPathForTesting(tt.initialDirtyFolder)
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
