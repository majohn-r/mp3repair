package internal

import (
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
	logDirName       = "logs"
	logFilePrefix    = AppName + "."
	logFileExtension = ".log"
	logFileTemplate  = logFilePrefix + "%Y%m%d" + logFileExtension
	symlinkName      = "latest" + logFileExtension
	maxLogFiles      = 10
	// constants for log field keys
	FK_DIRECTORY = "directory"
	FK_ERROR     = "error"
	FK_FILE_NAME = "fileName"
	FK_SECTION   = "section"
	fkVarName    = "environment variable"
)

func configureLogging(path string) *cronowriter.CronoWriter {
	logFileTemplate := filepath.Join(path, logFileTemplate)
	symlink := filepath.Join(path, symlinkName)
	return cronowriter.MustNew(logFileTemplate, cronowriter.WithSymlink(symlink), cronowriter.WithInit())
}

func cleanupLogFiles(o OutputBus, path string) {
	if files, err := ioutil.ReadDir(path); err != nil {
		o.LogWriter().Warn(LW_CANNOT_READ_DIRECTORY, map[string]interface{}{
			FK_DIRECTORY: path,
			FK_ERROR:     err,
		})
		o.WriteError(USER_LOG_DIR_CANNOT_BE_READ, path, err)
	} else {
		var fileMap map[time.Time]fs.FileInfo = make(map[time.Time]fs.FileInfo)
		var times []time.Time
		for _, file := range files {
			fileName := file.Name()
			if file.Mode().IsRegular() &&
				strings.HasPrefix(fileName, logFilePrefix) &&
				strings.HasSuffix(fileName, logFileExtension) {
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
					o.LogWriter().Warn(LW_CANNOT_DELETE_FILE, map[string]interface{}{
						FK_DIRECTORY: path,
						FK_FILE_NAME: fileName,
						FK_ERROR:     err,
					})
					o.WriteError(USER_LOG_FILE_CANNOT_BE_DELETED, logFilePath, err)
				} else {
					o.LogWriter().Info(LI_FILE_DELETED, map[string]interface{}{
						FK_DIRECTORY: path,
						FK_FILE_NAME: fileName,
					})
				}
			}
		}
	}
}

// exposed so that unit tests can close the writer!
var logger *cronowriter.CronoWriter

// InitLogging sets up logging
func InitLogging(o OutputBus) bool {
	var tmpFolder string
	var found bool
	if tmpFolder, found = os.LookupEnv("TMP"); !found {
		if tmpFolder, found = os.LookupEnv("TEMP"); !found {
			o.WriteError(USER_NO_TEMP_FOLDER)
			return false
		}
	}
	path := filepath.Join(CreateAppSpecificPath(tmpFolder), logDirName)
	if err := os.MkdirAll(path, 0755); err != nil {
		o.WriteError(USER_CANNOT_CREATE_DIRECTORY, path, err)
		return false
	}
	logger = configureLogging(path)
	logrus.SetOutput(logger)
	cleanupLogFiles(o, path)
	return true
}

type Logger interface {
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
}

type productionLogger struct{}

func (productionLogger) Info(msg string, fields map[string]interface{}) {
	logrus.WithFields(fields).Info(msg)
}

func (productionLogger) Warn(msg string, fields map[string]interface{}) {
	logrus.WithFields(fields).Warn(msg)
}

func (productionLogger) Error(msg string, fields map[string]interface{}) {
	logrus.WithFields(fields).Error(msg)
}
