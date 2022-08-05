package internal

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
)

// PlainFileExists returns whether the specified file exists as a plain file
// (i.e., not a directory)
func PlainFileExists(path string) bool {
	f, err := os.Stat(path)
	if err == nil {
		return !f.IsDir()
	} else {
		return !errors.Is(err, os.ErrNotExist)
	}
}

// DirExists returns whether the specified file exists as a directory
func DirExists(path string) bool {
	f, err := os.Stat(path)
	if err == nil {
		return f.IsDir()
	} else {
		return !errors.Is(err, os.ErrNotExist)
	}
}

// Copy copies a file. Adapted from
// https://github.com/cleversoap/go-cp/blob/master/cp.go
func CopyFile(src, dest string) (err error) {
	var r *os.File
	r, err = os.Open(src)
	if err == nil {
		defer r.Close()
		var w *os.File
		w, err = os.Create(dest)
		if err == nil {
			defer w.Close()
			_, err = io.Copy(w, r)
		}
	}
	return
}

// Mkdir makes the specified directory; succeeds if the directory already
// exists. Fails if a plain file exists with the specified path.
func Mkdir(dirName string) (err error) {
	status, err := os.Stat(dirName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(dirName, 0755)
		}
		return
	}
	if !status.IsDir() {
		err = fmt.Errorf(ERROR_DIR_IS_FILE)
	}
	return
}

// ReadDirectory returns the contents of a specified directory
func ReadDirectory(o OutputBus, dir string) (files []fs.FileInfo, ok bool) {
	var err error
	if files, err = ioutil.ReadDir(dir); err != nil {
		o.LogWriter().Error(LE_CANNOT_READ_DIRECTORY, map[string]interface{}{
			FK_DIRECTORY: dir,
			FK_ERROR:     err,
		})
		o.WriteError(USER_CANNOT_READ_DIRECTORY, dir, err)
		return
	}
	ok = true
	return
}
