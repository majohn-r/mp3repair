package files_test

import (
	"mp3repair/internal/files"
	"path/filepath"
	"testing"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func TestMarkDirty(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	const fnName = "MarkDirty()"
	emptyDir := "empty"
	cmd_toolkit.Mkdir(emptyDir)
	filledDir := "filled"
	cmd_toolkit.Mkdir(filledDir)
	createFile(filledDir, files.DirtyFileName)
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
			if differences, verified := o.Verify(tt.WantedRecording); !verified {
				for _, difference := range differences {
					t.Errorf("%s %s", fnName, difference)
				}
			}
		})
	}
}

func TestDirty(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	const fnName = "Dirty()"
	testDir := "dirty"
	cmd_toolkit.Mkdir(testDir)
	oldAppPath := cmd_toolkit.SetApplicationPath(testDir)
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
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
				if fileErr := createFileWithContent(testDir, files.DirtyFileName,
					[]byte("dirty")); fileErr != nil {
					t.Errorf("%s error creating %q: %v", fnName, files.DirtyFileName, fileErr)
				}
			} else {
				cmd_toolkit.FileSystem().Remove(filepath.Join(testDir, files.DirtyFileName))
			}
			if gotDirty := files.Dirty(); gotDirty != tt.want {
				t.Errorf("%s = %t, want %t", fnName, gotDirty, tt.want)
			}
		})
	}
}

func TestClearDirty(t *testing.T) {
	originalFileSystem := cmd_toolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmd_toolkit.AssignFileSystem(originalFileSystem)
	}()
	const fnName = "ClearDirty()"
	testDir := "clearDirty"
	cmd_toolkit.Mkdir(testDir)
	oldAppPath := cmd_toolkit.ApplicationPath()
	createFileWithContent(testDir, files.DirtyFileName, []byte("dirty"))
	// create another file structure with a dirty file that is open for reading
	uncleanable := "clearDirty2"
	cmd_toolkit.Mkdir(uncleanable)
	createFileWithContent(uncleanable, files.DirtyFileName, []byte("dirty"))
	// open file locks file from being deleted
	f, _ := cmd_toolkit.FileSystem().Open(filepath.Join(uncleanable, files.DirtyFileName))
	defer func() {
		cmd_toolkit.SetApplicationPath(oldAppPath)
		f.Close()
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
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cmd_toolkit.SetApplicationPath(tt.initialDirtyFolder)
			o := output.NewRecorder()
			files.ClearDirty(o)
			if differences, verified := o.Verify(tt.WantedRecording); !verified {
				for _, difference := range differences {
					t.Errorf("%s %s", fnName, difference)
				}
			}
		})
	}
}
