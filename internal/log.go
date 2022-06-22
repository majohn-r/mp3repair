package internal

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/utahta/go-cronowriter"
)

const (
	logFilePrefix    = AppName + "."
	logFileExtension = ".log"
	logFileTemplate  = logFilePrefix + "%Y%m%d" + logFileExtension
	symlinkName      = "latest" + logFileExtension
	maxLogFiles      = 10
)

// constants for log field keys
const (
	FK_DIRECTORY = "directory"
	FK_ERROR     = "error"
	FK_FILE_NAME = "fileName"
	fkVarName    = "environment variable"
)

// ConfigureLogging sets up logging
func ConfigureLogging(path string) *cronowriter.CronoWriter {
	logFileTemplate := filepath.Join(path, logFileTemplate)
	symlink := filepath.Join(path, symlinkName)
	return cronowriter.MustNew(logFileTemplate, cronowriter.WithSymlink(symlink), cronowriter.WithInit())
}

// CleanupLogFiles cleans up old log files
func CleanupLogFiles(wErr io.Writer, path string) {
	if files, err := ioutil.ReadDir(path); err != nil {
		logrus.WithFields(logrus.Fields{
			FK_DIRECTORY: path,
			FK_ERROR:     err,
		}).Warn(LW_CANNOT_READ_DIRECTORY)
		fmt.Fprintf(wErr, USER_LOG_DIR_CANNOT_BE_READ, path, err)
	} else {
		var fileMap map[time.Time]fs.FileInfo = make(map[time.Time]fs.FileInfo)
		var times []time.Time
		for _, file := range files {
			fileName := file.Name()
			if file.Mode().IsRegular() && strings.HasPrefix(fileName, logFilePrefix) && strings.HasSuffix(fileName, logFileExtension) {
				modificationTime := file.ModTime()
				fileMap[modificationTime] = file
				times = append(times, modificationTime)
			}
		}
		if len(times) > maxLogFiles {
			sort.Slice(times, func(i, j int) bool {
				return times[i].Before(times[j])
			})
			limit := len(times) - maxLogFiles
			for k := 0; k < limit; k++ {
				fileName := fileMap[times[k]].Name()
				logFilePath := filepath.Join(path, fileName)
				if err := os.Remove(logFilePath); err != nil {
					logrus.WithFields(logrus.Fields{
						FK_DIRECTORY: path,
						FK_FILE_NAME: fileName,
						FK_ERROR:     err,
					}).Warn(LW_CANNOT_DELETE_FILE)
					fmt.Fprintf(wErr, USER_LOG_FILE_CANNOT_BE_DELETED, logFilePath, err)
				} else {
					logrus.WithFields(logrus.Fields{
						FK_DIRECTORY: path,
						FK_FILE_NAME: fileName,
					}).Info(LI_FILE_DELETED)
				}
			}
		}
	}
}
