package files

import (
	"errors"
	"os"
	"path/filepath"

	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

const (
	DirtyFileName = "metadata.dirty"
)

func MarkDirty(o output.Bus) {
	f := filepath.Join(cmd_toolkit.ApplicationPath(), DirtyFileName)
	if _, err := os.Stat(f); err != nil && errors.Is(err, os.ErrNotExist) {
		// ignore error - if the file didn't exist a moment ago, there is no
		// reason to assume that the file cannot be written to
		_ = os.WriteFile(f, []byte("dirty"), cmd_toolkit.StdFilePermissions)
		o.Log(output.Info, "metadata dirty file written", map[string]any{"fileName": f})
	}
}

func ClearDirty(o output.Bus) {
	f := filepath.Join(cmd_toolkit.ApplicationPath(), DirtyFileName)
	if cmd_toolkit.PlainFileExists(f) {
		if err := os.Remove(f); err != nil {
			cmd_toolkit.ReportFileDeletionFailure(o, f, err)
		} else {
			o.Log(output.Info, "metadata dirty file deleted", map[string]any{
				"fileName": f,
			})
		}
	}
}

func Dirty() bool {
	return cmd_toolkit.PlainFileExists(filepath.Join(
		cmd_toolkit.ApplicationPath(), DirtyFileName))
}
