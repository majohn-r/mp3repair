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

// constants for log fields
const (
	LOG_ALBUM_FILTER  string = "albumFilter"
	LOG_ALBUM_NAME    string = "albumName"
	LOG_ARTIST_FILTER string = "artistFilter"
	LOG_ARTIST_NAME   string = "artistName"
	LOG_COMMAND_NAME  string = "command"
	LOG_DIRECTORY     string = "directory"
	LOG_ERROR         string = "error"
	LOG_EXTENSION     string = "extension"
	LOG_FILE_NAME     string = "fileName"
	LOG_FILTER        string = "filter"
	LOG_FLAG          string = "flag"
	LOG_PATH          string = "path"
	LOG_TRACK_NAME    string = "trackName"
	LOG_VALUE         string = "value"
)

func ConfigureLogging(path string) *cronowriter.CronoWriter {
	logFileTemplate := filepath.Join(path, logFileTemplate)
	symlink := filepath.Join(path, symlinkName)
	return cronowriter.MustNew(logFileTemplate, cronowriter.WithSymlink(symlink), cronowriter.WithInit())
}

func CleanupLogFiles(path string) {
	if files, err := ioutil.ReadDir(path); err != nil {
		logrus.WithFields(logrus.Fields{LOG_PATH: path, LOG_ERROR: err}).Warn(LOG_CANNOT_READ_DIRECTORY)
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
					logrus.WithFields(logrus.Fields{LOG_PATH: logFilePath, LOG_ERROR: err}).Warn(LOG_CANNOT_DELETE_FILE)
				} else {
					logrus.WithFields(logrus.Fields{LOG_PATH: logFilePath}).Info(LOG_FILE_DELETED)
				}
			}
		}
	}
}
