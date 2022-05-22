package internal

import (
	"errors"
	"fmt"
	"os"
)

func PlainFileExists(path string) bool {
	f, err := os.Stat(path)
	if err == nil {
		return !f.IsDir()
	} else {
		return !errors.Is(err, os.ErrNotExist)
	}
}

func Mkdir(dirName string) (err error) {
	status, err := os.Stat(dirName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(dirName, 0755)
		}
		return
	}
	if !status.IsDir() {
		err = fmt.Errorf("%q exists and is not a directory", dirName)
	}
	return
}
