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
	logFilePrefix    string = "mp3."
	logFileExtension string = ".log"
	logFileTemplate  string = logFilePrefix + "%Y%m%d" + logFileExtension
	symlinkName      string = "latest.log"
	maxLogFiles      int    = 10
)

func ConfigureLogging(path string) *cronowriter.CronoWriter {
	logFileTemplate := filepath.Join(path, logFileTemplate)
	symlink := filepath.Join(path, symlinkName)
	return cronowriter.MustNew(logFileTemplate, cronowriter.WithSymlink(symlink), cronowriter.WithInit())
}

func CleanupLogFiles(path string) {
	if files, err := ioutil.ReadDir(path); err != nil {
		logrus.WithFields(logrus.Fields{
			"path":  path,
			"error": err,
		}).Warn("cannot read log path")
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
				logFilePath := filepath.Join(path, fileMap[times[k]].Name())
				if err := os.Remove(logFilePath); err != nil {
					logrus.WithFields(logrus.Fields{
						"path":  logFilePath,
						"error": err,
					}).Warn("cannot delete old log file")
				} else {
					logrus.WithFields(logrus.Fields{
						"path": logFilePath,
					}).Info("deleted old log file")
				}
			}
		}
	}
}
