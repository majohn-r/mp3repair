package files

import (
	"mp3/internal"
	"regexp"
	"strings"

	"github.com/bogem/id3v2/v2"
	"github.com/sirupsen/logrus"
)

const (
	rawExtension            string = "mp3"
	DefaultFileExtension    string = "." + rawExtension
	defaultTrackNamePattern string = "^\\d+[\\s-].+\\." + rawExtension + "$"
)

type Track struct {
	fullPath        string
	fileName        string
	Name            string
	TrackNumber     int
	ContainingAlbum *Album
}

type Album struct {
	Name            string
	Tracks          []*Track
	RecordingArtist *Artist
}

type Artist struct {
	Name   string
	Albums []*Album
}

var trackNameRegex *regexp.Regexp = regexp.MustCompile(defaultTrackNamePattern)

func ReadMP3Data(track *Track) {
	tag, err := id3v2.Open(track.fullPath, id3v2.Options{Parse: true})
	if err != nil {
		logrus.WithFields(logrus.Fields{internal.LOG_PATH: track.fullPath, internal.LOG_ERROR: err}).Warn(internal.LOG_CANNOT_READ_FILE)
	} else {
		defer tag.Close()

		// Read tags.
		// TODO: this is temporary, and so does not use official log field names
		logrus.WithFields(logrus.Fields{
			"fileSystemTrackName":   track.Name,
			"fileSystemTrackNumber": track.TrackNumber,
			"fileSystemArtistName":  track.ContainingAlbum.RecordingArtist.Name,
			"fileSystemAlbumName":   track.ContainingAlbum.Name,
			"metadataTrackName":     tag.Title(),
			"metadataTrackNumber":   tag.GetTextFrame("TRCK").Text,
			"metadataArtistName":    tag.Artist(),
			"metadataAlbumName":     tag.Album(),
		}).Info("track data")
	}
}

// accessible outside the package for test purposes
func ParseTrackName(name string, album string, artist string, ext string) (simpleName string, trackNumber int, valid bool) {
	if !trackNameRegex.MatchString(name) {
		logrus.WithFields(logrus.Fields{internal.LOG_TRACK_NAME: name, internal.LOG_ALBUM_NAME: album, internal.LOG_ARTIST_NAME: artist}).Warn(internal.LOG_INVALID_TRACK_NAME)
		return
	}
	wantDigit := true
	runes := []rune(name)
	for i, r := range runes {
		if wantDigit {
			if r >= '0' && r <= '9' {
				trackNumber *= 10
				trackNumber += int(r - '0')
			} else {
				wantDigit = false
			}
		} else {
			simpleName = strings.TrimSuffix(string(runes[i:]), ext)
			break
		}
	}
	valid = true
	return
}
