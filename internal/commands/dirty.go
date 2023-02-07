package commands

import (
	"errors"
	"os"
	"path/filepath"

	tools "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

const (
	dirtyFileName = "metadata.dirty"
)

func markDirty(o output.Bus, cmd string) {
	f := filepath.Join(tools.ApplicationPath(), dirtyFileName)
	if _, err := os.Stat(f); err != nil && errors.Is(err, os.ErrNotExist) {
		// ignore error - if the file didn't exist a moment ago, there is no
		// reason to assume that the file cannot be written to
		_ = os.WriteFile(f, []byte("dirty"), tools.StdFilePermissions)
		o.Log(output.Info, "metadata dirty file written", map[string]any{"fileName": f})
	}
}

func clearDirty(o output.Bus) {
	f := filepath.Join(tools.ApplicationPath(), dirtyFileName)
	if tools.PlainFileExists(f) {
		if err := os.Remove(f); err != nil {
			tools.ReportFileDeletionFailure(o, f, err)
		} else {
			o.Log(output.Info, "metadata dirty file deleted", map[string]any{"fileName": f})
		}
	}
}

func dirty() bool {
	return tools.PlainFileExists(filepath.Join(tools.ApplicationPath(), dirtyFileName))
}
