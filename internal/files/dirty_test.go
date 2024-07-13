package files

import (
	"path/filepath"
	"testing"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

func TestMarkDirty(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	emptyDir := "empty"
	_ = cmdtoolkit.Mkdir(emptyDir)
	filledDir := "filled"
	_ = cmdtoolkit.Mkdir(filledDir)
	_ = createFile(filledDir, dirtyFileName)
	tests := map[string]struct {
		appPath string
		output.WantedRecording
	}{
		"typical first use": {
			appPath: emptyDir,
			WantedRecording: output.WantedRecording{
				Log: "" +
					"level='info'" +
					" fileName='empty\\metadata.dirty'" +
					" msg='metadata dirty file written'\n",
			},
		},
		"typical second use": {appPath: filledDir},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			o := output.NewRecorder()
			cmdtoolkit.SetApplicationPath(tt.appPath)
			MarkDirty(o)
			o.Report(t, "MarkDirty()", tt.WantedRecording)
		})
	}
}

func TestDirty(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	testDir := "dirty"
	_ = cmdtoolkit.Mkdir(testDir)
	oldAppPath := cmdtoolkit.SetApplicationPath(testDir)
	defer func() {
		_ = cmdtoolkit.AssignFileSystem(originalFileSystem)
		_ = cmdtoolkit.SetApplicationPath(oldAppPath)
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
				if fileErr := createFileWithContent(testDir, dirtyFileName,
					[]byte("dirty")); fileErr != nil {
					t.Errorf("Dirty() error creating %q: %v", dirtyFileName, fileErr)
				}
			} else {
				_ = cmdtoolkit.FileSystem().Remove(filepath.Join(testDir, dirtyFileName))
			}
			if gotDirty := Dirty(); gotDirty != tt.want {
				t.Errorf("Dirty() = %t, want %t", gotDirty, tt.want)
			}
		})
	}
}

func TestClearDirty(t *testing.T) {
	originalFileSystem := cmdtoolkit.AssignFileSystem(afero.NewMemMapFs())
	defer func() {
		cmdtoolkit.AssignFileSystem(originalFileSystem)
	}()
	testDir := "clearDirty"
	_ = cmdtoolkit.Mkdir(testDir)
	oldAppPath := cmdtoolkit.ApplicationPath()
	_ = createFileWithContent(testDir, dirtyFileName, []byte("dirty"))
	// create another file structure with a dirty file that is open for reading
	permanentlyDirtyDirectory := "clearDirty2"
	_ = cmdtoolkit.Mkdir(permanentlyDirtyDirectory)
	_ = createFileWithContent(permanentlyDirtyDirectory, dirtyFileName, []byte("dirty"))
	// open file locks file from being deleted
	f, _ := cmdtoolkit.FileSystem().Open(filepath.Join(permanentlyDirtyDirectory, dirtyFileName))
	defer func() {
		cmdtoolkit.SetApplicationPath(oldAppPath)
		_ = f.Close()
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
			cmdtoolkit.SetApplicationPath(tt.initialDirtyFolder)
			o := output.NewRecorder()
			ClearDirty(o)
			o.Report(t, "ClearDirty()", tt.WantedRecording)
		})
	}
}
