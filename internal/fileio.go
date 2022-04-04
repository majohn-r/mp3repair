package internal

import "os"

func Mkdir(dirName string) error {
	return os.Mkdir(dirName, 0755)
}
