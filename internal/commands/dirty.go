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
	dirtyFolderFound = false
	dirtyFolder      = ""
	dirtyFolderValid = false
)

func markDirty(o output.Bus, cmd string) {
	if path, ok := findAppFolder(); ok {
		dirtyFile := filepath.Join(path, dirtyFileName)
		if _, err := os.Stat(dirtyFile); err != nil && errors.Is(err, os.ErrNotExist) {
			if writeErr := os.WriteFile(dirtyFile, []byte("dirty"), internal.StdFilePermissions); writeErr != nil {
				reportFileCreationFailure(o, cmd, dirtyFile, writeErr)
			} else {
				o.Log(output.Info, "metadata dirty file written", map[string]any{"fileName": dirtyFile})
			}
		}
	}
}

func findAppFolder() (folder string, ok bool) {
	if !dirtyFolderFound {
		if path, ok := internal.AppSpecificPath(); ok {
			if info, err := os.Stat(path); err == nil {
				if info.IsDir() {
					dirtyFolder = path
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

func clearDirty(o output.Bus) {
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

func dirty() bool {
	if path, ok := findAppFolder(); ok {
		return internal.PlainFileExists(filepath.Join(path, dirtyFileName))
	}
	// can't find the app folder, so must assume same problem would have
	// occurred when attempting to write the 'dirty' file
	return true
}
