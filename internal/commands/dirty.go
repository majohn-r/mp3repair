package commands

import (
	"errors"
	"mp3/internal"
	"os"
	"path/filepath"
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
func MarkDirty(o internal.OutputBus) {
	if !markDirtyAttempted {
		if path, ok := findAppFolder(); ok {
			dirtyFile := filepath.Join(path, dirtyFileName)
			if _, err := os.Stat(dirtyFile); err != nil && errors.Is(err, os.ErrNotExist) {
				if writeErr := os.WriteFile(dirtyFile, []byte("dirty"), 0644); writeErr != nil {
					o.WriteError(internal.USER_CANNOT_CREATE_FILE, dirtyFile, writeErr)
					o.LogWriter().Error(internal.LE_CANNOT_CREATE_FILE, map[string]any{
						internal.FK_FILE_NAME: dirtyFile,
						internal.FK_ERROR:     writeErr,
					})
				} else {
					o.LogWriter().Info(internal.LI_DIRTY_FILE_WRITTEN, map[string]any{
						internal.FK_FILE_NAME: dirtyFile,
					})
				}
			}
		}
	}
	markDirtyAttempted = true
}

func findAppFolder() (folder string, ok bool) {
	if !dirtyFolderFound {
		appSpecificPath, appSpecificPathValid := internal.GetAppSpecificPath()
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
func ClearDirty(o internal.OutputBus) {
	if path, ok := findAppFolder(); ok {
		dirtyFile := filepath.Join(path, dirtyFileName)
		if internal.PlainFileExists(dirtyFile) {
			if err := os.Remove(dirtyFile); err != nil {
				o.WriteError(internal.USER_CANNOT_DELETE_FILE, dirtyFile, err)
				o.LogWriter().Error(internal.LE_CANNOT_DELETE_FILE, map[string]any{
					internal.FK_FILE_NAME: dirtyFile,
					internal.FK_ERROR:     err,
				})
			} else {
				o.LogWriter().Info(internal.LI_DIRTY_FILE_DELETED, map[string]any{
					internal.FK_FILE_NAME: dirtyFile,
				})
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
