package internal

import (
	"io/fs"
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
	fieldKeyVarName   = "environment variable"
)

// Reusable field keys for logs
const (
	FieldKeyDirectory = "directory"
	FieldKeyError     = "error"
	FieldKeyFileName  = "fileName"
	FieldKeySection   = "section"
	FieldKeyValue     = "value"
)

func configureLogging(path string) *cronowriter.CronoWriter {
	logFileTemplate := filepath.Join(path, logFileTemplate)
	symlink := filepath.Join(path, symlinkName)
	return cronowriter.MustNew(logFileTemplate, cronowriter.WithSymlink(symlink), cronowriter.WithInit())
}

func cleanupLogFiles(o OutputBus, path string) {
	if files, err := os.ReadDir(path); err != nil {
		o.LogWriter().Error(LogErrorCannotReadDirectory, map[string]any{
			FieldKeyDirectory: path,
			FieldKeyError:     err,
		})
		o.WriteError(UserLogDirCannotBeRead, path, err)
	} else {
		var fileMap map[time.Time]fs.DirEntry = make(map[time.Time]fs.DirEntry)
		var times []time.Time
		for _, file := range files {
			fileName := file.Name()
			if file.Type().IsRegular() &&
				strings.HasPrefix(fileName, logFilePrefix) &&
				strings.HasSuffix(fileName, logFileExtension) {
				if f, fErr := file.Info(); fErr == nil {
					modificationTime := f.ModTime()
					fileMap[modificationTime] = file
					times = append(times, modificationTime)
				}
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
					o.LogWriter().Error(LogErrorCannotDeleteFile, map[string]any{
						FieldKeyDirectory: path,
						FieldKeyFileName:  fileName,
						FieldKeyError:     err,
					})
					o.WriteError(UserLogFileCannotBeDeleted, logFilePath, err)
				} else {
					o.LogWriter().Info(LogInfoFileDeleted, map[string]any{
						FieldKeyDirectory: path,
						FieldKeyFileName:  fileName,
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
			o.WriteError(UserNoTempFolder)
			return false
		}
	}
	path := filepath.Join(CreateAppSpecificPath(tmpFolder), logDirName)
	if err := os.MkdirAll(path, 0755); err != nil {
		o.WriteError(UserCannotCreateDirectory, path, err)
		return false
	}
	logger = configureLogging(path)
	logrus.SetOutput(logger)
	cleanupLogFiles(o, path)
	return true
}

// Logger defines functions for writing to a log
type Logger interface {
	Info(msg string, fields map[string]any)
	Error(msg string, fields map[string]any)
}

type productionLogger struct{}

func (productionLogger) Info(msg string, fields map[string]any) {
	logrus.WithFields(fields).Info(msg)
}

func (productionLogger) Error(msg string, fields map[string]any) {
	logrus.WithFields(fields).Error(msg)
}
