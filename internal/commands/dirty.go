package commands

import (
	"errors"
	"mp3/internal"
	"os"
	"path/filepath"

	"github.com/majohn-r/output"
)

const (
	dirtyFileName = "metadata.dirty"
)

func markDirty(o output.Bus, cmd string) {
	f := filepath.Join(internal.ApplicationPath(), dirtyFileName)
	if _, err := os.Stat(f); err != nil && errors.Is(err, os.ErrNotExist) {
		// ignore error - if the file didn't exist a moment ago, there is no
		// reason to assume that the file cannot be written to
		_ = os.WriteFile(f, []byte("dirty"), internal.StdFilePermissions)
		o.Log(output.Info, "metadata dirty file written", map[string]any{"fileName": f})
	}
}

func clearDirty(o output.Bus) {
	f := filepath.Join(internal.ApplicationPath(), dirtyFileName)
	if internal.PlainFileExists(f) {
		if err := os.Remove(f); err != nil {
			reportFileDeletionFailure(o, f, err)
		} else {
			o.Log(output.Info, "metadata dirty file deleted", map[string]any{"fileName": f})
		}
	}
}

func dirty() bool {
	return internal.PlainFileExists(filepath.Join(internal.ApplicationPath(), dirtyFileName))
}
