package files

import (
	"fmt"
	"mp3/internal"
	"regexp"
	"strings"
	"time"

	"github.com/bogem/id3v2/v2"
	"github.com/sirupsen/logrus"
)

const (
	rawExtension            string = "mp3"
	DefaultFileExtension    string = "." + rawExtension
	defaultTrackNamePattern string = "^\\d+[\\s-].+\\." + rawExtension + "$"
	trackDiffBadTags        string = "cannot determine differences, tags were not recognized"
	trackDiffUnreadableTags string = "cannot determine differences, could not read tags"
	trackDiffUnreadTags     string = "cannot determine differences, tags have not been read"
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
	// this is more complicated than I wanted, because some mp3 rippers produce
	// track numbers like "12/14", meaning 12th track of 14
	if len(s) == 0 {
		err = fmt.Errorf("invalid format: %q", s)
		return
	}
	n := 0
	bs := []byte(s)
	for j, b := range bs {
		c := int(b)
		if c >= '0' && c <= '9' {
			n *= 10
			n += c - '0'
		} else {
			switch j {
			case 0: // never saw a digit
				err = fmt.Errorf("invalid format: %q", s)
				return
			default: // found something other than a digit, but read at least one
				i = n
				return
			}
		}
	}
	// normal path, whole string was digits
	i = n
	return
}

func (t *Track) setTags(d *taggedTrackData) {
	if trackNumber, err := toTrackNumber(d.number); err != nil {
		logrus.WithFields(logrus.Fields{"trackTag": d.number, internal.LOG_ERROR: err}).Warn("invalid track tag")
		t.setTagFormatError()
	} else {
		t.TaggedAlbum = removeLeadingBOMs(d.album)
		t.TaggedArtist = removeLeadingBOMs(d.artist)
		t.TaggedTitle = removeLeadingBOMs(d.title)
		t.TaggedTrack = trackNumber
	}
}

// randomly, some tags - particularly titles - begin with a BOM (byte order mark)
func removeLeadingBOMs(s string) string {
	r := []rune(s)
	if r[0] == '\ufeff' {
		for r[0] == '\ufeff' {
			r = r[1:]
		}
		return string(r)
	}
	return s
}

type nameTagPair struct {
	name string
	tag  string
}

func (t *Track) FindDifferences() []string {
	switch t.TaggedTrack {
	case trackUnknownFormatError:
		return []string{trackDiffBadTags}
	case trackUnknownTagReadError:
		return []string{trackDiffUnreadableTags}
	case trackUnknownTagsNotRead:
		return []string{trackDiffUnreadTags}
	default:
		var differences []string
		if t.TaggedTrack != t.TrackNumber {
			differences = append(differences,
				fmt.Sprintf("track number %d does not agree with track tag %d", t.TrackNumber, t.TaggedTrack))
		}
		if !isComparable(nameTagPair{name: t.Name, tag: t.TaggedTitle}) {
			differences = append(differences,
				fmt.Sprintf("title %q does not agree with title tag %q", t.Name, t.TaggedTitle))
		}
		if !isComparable(nameTagPair{name: t.ContainingAlbum.Name, tag: t.TaggedAlbum}) {
			differences = append(differences,
				fmt.Sprintf("album %q does not agree with album tag %q", t.ContainingAlbum.Name, t.TaggedAlbum))
		}
		if !isComparable(nameTagPair{name: t.ContainingAlbum.RecordingArtist.Name, tag: t.TaggedArtist}) {
			differences = append(differences,
				fmt.Sprintf("artist %q does not agree with artist tag %q", t.ContainingAlbum.RecordingArtist.Name, t.TaggedArtist))
		}
		return differences
	}
}

func isComparable(p nameTagPair) bool {
	fileName := strings.ToLower(p.name)
	tag := strings.ToLower(p.tag)
	// strip off illegal end characters from the tag
	for strings.HasSuffix(tag, ".") || strings.HasSuffix(tag, " ") {
		tag = tag[:len(tag)-1]
	}
	if fileName == tag {
		return true
	}
	tagAsRunes := []rune(tag)
	nameAsRunes := []rune(fileName)
	if len(tagAsRunes) != len(nameAsRunes) {
		return false
	}
	for index, c := range tagAsRunes {
		if !isIllegalRuneForFileNames(c) && nameAsRunes[index] != c {
			return false
		}
	}
	return true // rune by rune comparison was successful
}

// per https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file
func isIllegalRuneForFileNames(r rune) bool {
	if r >= 0 && r <= 31 {
		return true
	}
	switch r {
	case '<', '>', ':', '"', '/', '\\', '|', '?', '*':
		return true
	default:
		return false
	}
}

var stdFrames = []string{"TALB", "TIT2", "TPE1", "TRCK"} // album, title, artist, track, in that order

func RawReadTags(path string) (d *taggedTrackData, err error) {
	var tag *id3v2.Tag
	if tag, err = id3v2.Open(path, id3v2.Options{Parse: true, ParseFrames: stdFrames}); err != nil {
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

// use of semaphores nicely documented here:
// https://gist.github.com/repejota/ed9070d57c23102d50c94e1a126b2f5b

type empty struct{}

var semaphores = make(chan empty, 20) // 20 is a typical limit for open files

func (t *Track) readTags(reader func(string) (*taggedTrackData, error)) {
	if t.needsTaggedData() {
		semaphores <- empty{} // block while full
		go func() {
			defer func() {
				<-semaphores // read to release a slot
			}()
			if tags, err := reader(t.fullPath); err != nil {
				logrus.WithFields(logrus.Fields{internal.LOG_PATH: t.fullPath, internal.LOG_ERROR: err}).Warn(internal.LOG_CANNOT_READ_FILE)
				t.setTagReadError()
			} else {
				t.setTags(tags)
			}
		}()
	}
}

func UpdateTracks(artists []*Artist, reader func(string) (*taggedTrackData, error)) {
	for _, artist := range artists {
		for _, album := range artist.Albums {
			for _, track := range album.Tracks {
				track.readTags(reader)
			}
		}
	}
	waitForSemaphoresDrained()
}

func waitForSemaphoresDrained() {
	for len(semaphores) != 0 {
		time.Sleep(1 * time.Microsecond)
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
