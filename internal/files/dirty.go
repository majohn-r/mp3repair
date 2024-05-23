package files

import (
	"errors"
	"os"
	"path/filepath"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
	"github.com/spf13/afero"
)

const (
	DirtyFileName = "metadata.dirty"
)

func MarkDirty(o output.Bus) {
	fs := cmd_toolkit.FileSystem()
	f := dirtyPath()
	if _, fileErr := fs.Stat(f); fileErr != nil && errors.Is(fileErr, os.ErrNotExist) {
		// ignore error - if the file didn't exist a moment ago, there is no
		// reason to assume that the file cannot be written to
		afero.WriteFile(fs, f, []byte("dirty"), cmd_toolkit.StdFilePermissions)
		o.Log(output.Info, "metadata dirty file written", map[string]any{"fileName": f})
	}
}

func ClearDirty(o output.Bus) {
	f := dirtyPath()
	if !cmd_toolkit.PlainFileExists(f) {
		return
	}
	fs := cmd_toolkit.FileSystem()
	// best effort
	fs.Remove(f)
	o.Log(output.Info, "metadata dirty file deleted", map[string]any{"fileName": f})
}

func Dirty() bool {
	return cmd_toolkit.PlainFileExists(dirtyPath())
}

func dirtyPath() string {
	return filepath.Join(cmd_toolkit.ApplicationPath(), DirtyFileName)
}
