package files

import (
	"fmt"
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

type taggedTrackData struct {
	album  string
	artist string
	title  string
	number string
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

func (t *Track) needsTaggedData() bool {
	return t.TaggedTrack == trackUnknownTagsNotRead
}

func (t *Track) setTagReadError() {
	t.TaggedTrack = trackUnknownTagReadError
}

func (t *Track) setTagFormatError() {
	t.TaggedTrack = trackUnknownFormatError
}

func toTrackNumber(s string) (i int, err error) {
	if strings.HasPrefix(s, "-") {
		err = fmt.Errorf("invalid format: %q", s)
		return
	}
	i, err = strconv.Atoi(s)
	return
}

func (t *Track) setTags(d *taggedTrackData) {
	if trackNumber, err := toTrackNumber(d.number); err != nil {
		logrus.WithFields(logrus.Fields{"trackTag": d.number, internal.LOG_ERROR: err}).Warn("invalid track tag")
		t.setTagFormatError()
	} else {
		t.TaggedAlbum = d.album
		t.TaggedArtist = d.artist
		t.TaggedTitle = d.title
		t.TaggedTrack = trackNumber
	}
}

func rawReadTags(path string) (d *taggedTrackData, err error) {
	var tag *id3v2.Tag
	if tag, err = id3v2.Open(path, id3v2.Options{Parse: true}); err != nil {
		return
	}
	defer tag.Close()
	d = &taggedTrackData{
		album:  tag.Album(),
		artist: tag.Artist(),
		title:  tag.Title(),
		number: tag.GetTextFrame("TRCK").Text,
	}
	return
}

func (t *Track) ReadMP3Tags() {
	t.readTags(rawReadTags)
}

func (t *Track) readTags(reader func(string) (*taggedTrackData, error)) {
	if t.needsTaggedData() {
		if tags, err := reader(t.fullPath); err != nil {
			logrus.WithFields(logrus.Fields{internal.LOG_PATH: t.fullPath, internal.LOG_ERROR: err}).Warn(internal.LOG_CANNOT_READ_FILE)
			t.setTagReadError()
		} else {
			t.setTags(tags)
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
