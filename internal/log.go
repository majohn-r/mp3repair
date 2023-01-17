package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/majohn-r/output"
	"github.com/sirupsen/logrus"
	"github.com/utahta/go-cronowriter"
)

const (
	logDirName       = "logs"
	logFilePrefix    = AppName + "."
	logFileExtension = ".log"
	symlinkName      = "latest" + logFileExtension
	maxLogFiles      = 10
)

func configure(path string) *cronowriter.CronoWriter {
	return cronowriter.MustNew(
		filepath.Join(path, logFilePrefix+"%Y%m%d"+logFileExtension),
		cronowriter.WithSymlink(filepath.Join(path, symlinkName)),
		cronowriter.WithInit())
}

func cleanup(o output.Bus, path string) {
	if files, err := os.ReadDir(path); err != nil {
		LogUnreadableDirectory(o, path, err)
		o.WriteCanonicalError("The log file directory %q cannot be read: %v", path, err)
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
				logFile := filepath.Join(path, fileName)
				if err := os.Remove(logFile); err != nil {
					LogFileDeletionFailure(o, logFile, err)
					o.WriteCanonicalError("The log file %q cannot be deleted: %v", logFile, err)
				} else {
					o.Log(output.Info, "successfully deleted file", map[string]any{
						"directory": path,
						"fileName":  fileName,
					})
				}
			}
		}
	}
}

// exposed so that unit tests can close the writer!
var logger *cronowriter.CronoWriter

// InitLogging sets up logging
func InitLogging(o output.Bus) bool {
	if tmpFolder, found := findTemp(o); !found {
		return false
	} else {
		path := filepath.Join(CreateAppSpecificPath(tmpFolder), logDirName)
		if err := os.MkdirAll(path, stdDirPermissions); err != nil {
			WriteDirectoryCreationError(o, path, err)
			return false
		}
		logger = configure(path)
		logrus.SetOutput(logger)
		cleanup(o, path)
	}
	return true
}

func findTemp(o output.Bus) (string, bool) {
	for _, v := range []string{"TMP", "TEMP"} {
		if tmpFolder, found := os.LookupEnv(v); found {
			return tmpFolder, found
		}
	}
	o.WriteCanonicalError("Neither the TMP nor TEMP environment variables are defined")
	return "", false
}

// ProductionLogger is the production implementation of the Logger interface
type ProductionLogger struct{}

// Trace outputs a trace log message
func (pl ProductionLogger) Trace(msg string, fields map[string]any) {
	logrus.WithFields(fields).Trace(msg)
}

// Debug outputs a debug log message
func (pl ProductionLogger) Debug(msg string, fields map[string]any) {
	logrus.WithFields(fields).Debug(msg)
}

// Info outputs an info log message
func (pl ProductionLogger) Info(msg string, fields map[string]any) {
	logrus.WithFields(fields).Info(msg)
}

// Warning outputs a warning log message
func (pl ProductionLogger) Warning(msg string, fields map[string]any) {
	logrus.WithFields(fields).Warning(msg)
}

// Error outputs an error log message
func (pl ProductionLogger) Error(msg string, fields map[string]any) {
	logrus.WithFields(fields).Error(msg)
}

// Panic outputs a panic log message and calls panic()
func (pl ProductionLogger) Panic(msg string, fields map[string]any) {
	logrus.WithFields(fields).Panic(msg)
}

// Fatal outputs a fatal log message and terminates the program
func (pl ProductionLogger) Fatal(msg string, fields map[string]any) {
	logrus.WithFields(fields).Fatal(msg)
}
