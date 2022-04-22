package files

import (
	"mp3/internal"
	"regexp"
	"strconv"
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
	fullPath        string // full path to the file associated with the track, including the file itself
	Name            string // name of the track, without the track number or file extension, e.g., "First Track"
	TrackNumber     int    // number of the track
	ContainingAlbum *Album
	TaggedTitle     string // track title per mp3 tag
	TaggedTrack     int    // track number per mp3 tag
	TaggedAlbum     string // album name per mp3 tag
	TaggedArtist    string // artist name per mp3 tag
}

const (
	trackUnknownFormatError  int = -1
	trackUnknownTagsNotRead  int = -2
	trackUnknownTagReadError int = -3
)

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
	if track.TaggedTrack == trackUnknownTagsNotRead {
		tag, err := id3v2.Open(track.fullPath, id3v2.Options{Parse: true})
		if err != nil {
			logrus.WithFields(logrus.Fields{internal.LOG_PATH: track.fullPath, internal.LOG_ERROR: err}).Warn(internal.LOG_CANNOT_READ_FILE)
			track.TaggedTrack = trackUnknownTagReadError
		} else {
			defer tag.Close()

			// Read tags.
			track.TaggedAlbum = tag.Album()
			track.TaggedArtist = tag.Artist()
			track.TaggedTitle = tag.Title()
			rawTrackTag := tag.GetTextFrame("TRCK").Text
			if track.TaggedTrack, err = strconv.Atoi(rawTrackTag); err != nil || strings.HasPrefix(rawTrackTag, "-") {
				logrus.WithFields(logrus.Fields{"trackTag": rawTrackTag, internal.LOG_ERROR: err}).Warn("invalid track tag")
				track.TaggedTrack = trackUnknownFormatError
			}
		}
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
