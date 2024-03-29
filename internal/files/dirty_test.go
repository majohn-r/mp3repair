package files_test

import (
	"mp3repair/internal/files"
	"os"
	"path/filepath"
	"testing"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

func TestMarkDirty(t *testing.T) {
	const fnName = "MarkDirty()"
	emptyDir := "empty"
	if err := cmd_toolkit.Mkdir(emptyDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, emptyDir, err)
	}
	filledDir := "filled"
	if err := cmd_toolkit.Mkdir(filledDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, filledDir, err)
	}
	if err := createFile(filledDir, files.DirtyFileName); err != nil {
		t.Errorf("%s error creating %q: %v", fnName,
			filepath.Join(filledDir, files.DirtyFileName), err)
	}
	defer func() {
		destroyDirectory(fnName, emptyDir)
		destroyDirectory(fnName, filledDir)
	}()
	type args struct {
		cmd string
	}
	tests := map[string]struct {
		appPath string
		args
		output.WantedRecording
	}{
		"typical first use": {
			appPath: emptyDir,
			args:    args{cmd: "calling command"},
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" fileName='empty\\metadata.dirty'" +
					" msg='metadata dirty file written'\n",
			},
		},
		"typical second use": {
			appPath: filledDir,
			args:    args{cmd: "calling command"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmd_toolkit.SetApplicationPath(tt.appPath)
			files.MarkDirty(o)
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("%s %s", fnName, difference)
				}
			}
		})
	}
}

func TestDirty(t *testing.T) {
	const fnName = "Dirty()"
	testDir := "dirty"
	if err := cmd_toolkit.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	oldAppPath := cmd_toolkit.SetApplicationPath(testDir)
	defer func() {
		destroyDirectory(fnName, testDir)
		cmd_toolkit.SetApplicationPath(oldAppPath)
	}()
	tests := map[string]struct {
		want bool
	}{
		"file definitively exists":       {want: true},
		"file definitely does not exist": {want: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.want {
				if err := createFileWithContent(testDir, files.DirtyFileName,
					[]byte("dirty")); err != nil {
					t.Errorf("%s error creating %q: %v", fnName, files.DirtyFileName, err)
				}
			} else {
				os.Remove(filepath.Join(testDir, files.DirtyFileName))
			}
			if gotDirty := files.Dirty(); gotDirty != tt.want {
				t.Errorf("%s = %t, want %t", fnName, gotDirty, tt.want)
			}
		})
	}
}

func TestClearDirty(t *testing.T) {
	const fnName = "ClearDirty()"
	testDir := "clearDirty"
	if err := cmd_toolkit.Mkdir(testDir); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, testDir, err)
	}
	oldAppPath := cmd_toolkit.ApplicationPath()
	if err := createFileWithContent(testDir, files.DirtyFileName,
		[]byte("dirty")); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, files.DirtyFileName, err)
	}
	// create another file structure with a dirty file that is open for reading
	uncleanable := "clearDirty2"
	if err := cmd_toolkit.Mkdir(uncleanable); err != nil {
		t.Errorf("%s error creating %q: %v", fnName, uncleanable, err)
	}
	if err := createFileWithContent(uncleanable, files.DirtyFileName,
		[]byte("dirty")); err != nil {
		t.Errorf("%s error creating second %q: %v", fnName, files.DirtyFileName, err)
	}
	f, err := os.Open(filepath.Join(uncleanable, files.DirtyFileName))
	if err != nil {
		t.Errorf("%s error opening second %q: %v", fnName, files.DirtyFileName, err)
	}
	defer func() {
		cmd_toolkit.SetApplicationPath(oldAppPath)
		destroyDirectory(fnName, testDir)
		f.Close()
		destroyDirectory(fnName, uncleanable)
	}()
	tests := map[string]struct {
		initialDirtyFolder string
		output.WantedRecording
	}{
		"successful removal": {
			initialDirtyFolder: testDir,
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" fileName='clearDirty\\metadata.dirty'" +
					" msg='metadata dirty file deleted'\n",
			},
		},
		"nothing to remove": {
			initialDirtyFolder: ".",
		},
		"unremovable file": {
			initialDirtyFolder: uncleanable,
			WantedRecording: output.WantedRecording{
				Error: "The file \"clearDirty2\\\\metadata.dirty\" cannot be deleted:" +
					" remove clearDirty2\\metadata.dirty: The process cannot access the file" +
					" because it is being used by another process.\n",
				Log: "" +
					"level='error'" +
					" error='remove clearDirty2\\metadata.dirty: The process cannot access" +
					" the file because it is being used by another process.'" +
					" fileName='clearDirty2\\metadata.dirty'" +
					" msg='cannot delete file'\n",
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd_toolkit.SetApplicationPath(tt.initialDirtyFolder)
			o := output.NewRecorder()
			files.ClearDirty(o)
			if differences, ok := o.Verify(tt.WantedRecording); !ok {
				for _, difference := range differences {
					t.Errorf("%s %s", fnName, difference)
				}
			}
		})
	}
}
