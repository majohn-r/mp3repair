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

var (
	markDirtyAttempted = false
	dirtyFolderFound   = false
	dirtyFolder        = ""
	dirtyFolderValid   = false
)

// MarkDirty creates the 'dirty' file if it doesn't already exist.
func MarkDirty(o output.Bus, cmd string) {
	if !markDirtyAttempted {
		if path, ok := findAppFolder(); ok {
			dirtyFile := filepath.Join(path, dirtyFileName)
			if _, err := os.Stat(dirtyFile); err != nil && errors.Is(err, os.ErrNotExist) {
				if writeErr := os.WriteFile(dirtyFile, []byte("dirty"), 0o644); writeErr != nil {
					reportFileCreationFailure(o, cmd, dirtyFile, writeErr)
				} else {
					o.Log(output.Info, "metadata dirty file written", map[string]any{"fileName": dirtyFile})
				}
			}
		}
	}
	markDirtyAttempted = true
}

func findAppFolder() (folder string, ok bool) {
	if !dirtyFolderFound {
		appSpecificPath, appSpecificPathValid := internal.AppSpecificPath()
		if appSpecificPathValid {
			if info, err := os.Stat(appSpecificPath); err == nil {
				if info.IsDir() {
					dirtyFolder = appSpecificPath
					dirtyFolderValid = true
				}
			}
		}
		dirtyFolderFound = true
	}
	folder = dirtyFolder
	ok = dirtyFolderValid
	return
}

// ClearDirty deletes the 'dirty' file if it exists.
func ClearDirty(o output.Bus) {
	if path, ok := findAppFolder(); ok {
		dirtyFile := filepath.Join(path, dirtyFileName)
		if internal.PlainFileExists(dirtyFile) {
			if err := os.Remove(dirtyFile); err != nil {
				reportFileDeletionFailure(o, dirtyFile, err)
			} else {
				o.Log(output.Info, "metadata dirty file deleted", map[string]any{"fileName": dirtyFile})
			}
		}
	}
}

// Dirty returns false only if the 'dirty' file could exist, but does not.
func Dirty() (dirty bool) {
	if path, ok := findAppFolder(); ok {
		dirty = internal.PlainFileExists(filepath.Join(path, dirtyFileName))
	} else {
		// can't find the app folder, so must assume same problem would have
		// occurred when attempting to write the 'dirty' file
		dirty = true
	}
	return
}
