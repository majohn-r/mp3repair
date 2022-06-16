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
	logFilePrefix    = appName + "."
	logFileExtension = ".log"
	logFileTemplate  = logFilePrefix + "%Y%m%d" + logFileExtension
	symlinkName      = "latest" + logFileExtension
	maxLogFiles      = 10
)

// constants for log field keys
const (
	FK_ALBUM_FILTER_FLAG       = "-albumFilter"
	FK_ANNOTATE_LISTINGS_FLAG  = "-annotate"
	FK_ARTIST_FILTER_FLAG      = "-artistFilter"
	FK_DRY_RUN_FLAG            = "-dryRun"
	FK_EMPTY_FOLDERS_FLAG      = "-emptyFolders"
	FK_FILE_EXTENSION_FLAG     = "-ext"
	FK_GAP_ANALYSIS_FLAG       = "-gapAnalysis"
	FK_INCLUDE_ALBUMS_FLAG     = "-includeAlbums"
	FK_INCLUDE_ARTISTS_FLAG    = "-includeArtists"
	FK_INCLUDE_TRACKS_FLAG     = "-includeTracks"
	FK_INTEGRITY_ANALYSIS_FLAG = "-integrityAnalysis"
	FK_TOP_DIR_FLAG            = "-topDir"
	FK_TRACK_SORTING_FLAG      = "-trackSorting"
	FK_ALBUM_NAME              = "albumName"
	FK_ARTIST_NAME             = "artistName"
	FK_COMMAND_LINE_ARGUMENTS  = "args"
	FK_COMMAND_NAME            = "command"
	FK_COUNT                   = "count"
	FK_DESTINATION             = "destination"
	FK_DIRECTORY               = "directory"
	FK_DURATION                = "duration"
	FK_ERROR                   = "error"
	FK_FILE_NAME               = "fileName"
	FK_KEY                     = "key"
	FK_SOURCE                  = "source"
	FK_TIMESTAMP               = "timeStamp"
	FK_TRACK_NAME              = "trackName"
	FK_TRCK_FRAME              = "TRCK"
	FK_TYPE                    = "type"
	FK_VALUE                   = "value"
	FK_VERSION                 = "version"
)

// ConfigureLogging sets up logging
func ConfigureLogging(path string) *cronowriter.CronoWriter {
	logFileTemplate := filepath.Join(path, logFileTemplate)
	symlink := filepath.Join(path, symlinkName)
	return cronowriter.MustNew(logFileTemplate, cronowriter.WithSymlink(symlink), cronowriter.WithInit())
}

// CleanupLogFiles cleans up old log files
func CleanupLogFiles(path string) {
	if files, err := ioutil.ReadDir(path); err != nil {
		logrus.WithFields(logrus.Fields{
			FK_DIRECTORY: path,
			FK_ERROR:     err,
		}).Warn(LW_CANNOT_READ_DIRECTORY)
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
