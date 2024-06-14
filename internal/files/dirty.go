package files

import (
	"errors"
	"os"
	"path/filepath"

	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

const (
	DirtyFileName = "metadata.dirty"
)

func MarkDirty(o output.Bus) {
	fs := cmdtoolkit.FileSystem()
	f := dirtyPath()
	if _, fileErr := fs.Stat(f); fileErr != nil && errors.Is(fileErr, os.ErrNotExist) {
		// ignore error - if the file didn't exist a moment ago, there is no
		// reason to assume that the file cannot be written to
		_ = afero.WriteFile(fs, f, []byte("dirty"), cmdtoolkit.StdFilePermissions)
		o.Log(output.Info, "metadata dirty file written", map[string]any{"fileName": f})
	}
}

func ClearDirty(o output.Bus) {
	f := dirtyPath()
	if !cmdtoolkit.PlainFileExists(f) {
		return
	}
	fs := cmdtoolkit.FileSystem()
	// best effort
	_ = fs.Remove(f)
	o.Log(output.Info, "metadata dirty file deleted", map[string]any{"fileName": f})
}

func Dirty() bool {
	return cmdtoolkit.PlainFileExists(dirtyPath())
}

func dirtyPath() string {
	return filepath.Join(cmdtoolkit.ApplicationPath(), DirtyFileName)
}
