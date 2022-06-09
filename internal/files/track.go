package files

import (
	"fmt"
	"mp3/internal"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bogem/id3v2/v2"
	"github.com/sirupsen/logrus"
)

const (
	rawExtension             = "mp3"
	DefaultFileExtension     = "." + rawExtension
	defaultTrackNamePattern  = "^\\d+[\\s-].+\\." + rawExtension + "$"
	trackDiffBadTags         = "cannot determine differences, tags were not recognized"
	trackDiffUnreadableTags  = "cannot determine differences, could not read tags"
	trackDiffUnreadTags      = "cannot determine differences, tags have not been read"
	BackupDirName            = "pre-repair-backup"
	trackUnknownFormatError  = -1
	trackUnknownTagsNotRead  = -2
	TrackUnknownTagReadError = -3
)

type Track struct {
	Path            string // full path to the file associated with the track, including the file itself
	Name            string // name of the track, without the track number or file extension, e.g., "First Track"
	TrackNumber     int    // number of the track
	ContainingAlbum *Album
	TaggedTitle     string // track title per mp3 tag TRCK frame
	TaggedTrack     int    // track number per mp3 tag TIT2 frame
	TaggedAlbum     string // album name per mp3 tag TALB frame
	TaggedArtist    string // artist name per mp3 tag TPE1 frame
}

// logic for sorting tracks spanning albums and artists
type Tracks []*Track

func (t Tracks) Len() int {
	return len(t)
}

func (t Tracks) Less(i, j int) bool {
	track1 := t[i]
	track2 := t[j]
	album1 := track1.ContainingAlbum
	album2 := track2.ContainingAlbum
	artist1 := album1.RecordingArtist
	artist2 := album2.RecordingArtist
	// compare artist name first
	if artist1.Name == artist2.Name {
		// artist names are the same ... try the album name next
		if album1.Name() == album2.Name() {
			// and album names are the same ... go by track number
			return track1.TrackNumber < track2.TrackNumber
		} else {
			return album1.Name() < album2.Name()
		}
	} else {
		return artist1.Name < artist2.Name
	}
}

func (t Tracks) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type taggedTrackData struct {
	album  string
	artist string
	title  string
	number string
}

var trackNameRegex *regexp.Regexp = regexp.MustCompile(defaultTrackNamePattern)

// BackupDirectory returns the path for this track
func (t *Track) BackupDirectory() string {
	return filepath.Join(t.ContainingAlbum.Path, BackupDirName)
}

func (t *Track) needsTaggedData() bool {
	return t.TaggedTrack == trackUnknownTagsNotRead
}

func (t *Track) setTagReadError() {
	t.TaggedTrack = TrackUnknownTagReadError
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
	s = removeLeadingBOMs(s)
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
		logrus.WithFields(logrus.Fields{
			internal.LOG_PATH:  t.Path,
			"trackTag":         d.number,
			internal.LOG_ERROR: err}).Warn("invalid track tag")
		t.setTagFormatError()
	} else {
		t.TaggedAlbum = removeLeadingBOMs(d.album)
		t.TaggedArtist = removeLeadingBOMs(d.artist)
		t.TaggedTitle = removeLeadingBOMs(d.title)
		t.TaggedTrack = trackNumber
	}
}

// randomly, some tags - particularly titles - begin with a BOM (byte order
// mark)
func removeLeadingBOMs(s string) string {
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	if r[0] == '\ufeff' {
		return removeLeadingBOMs(string(r[1:]))
	}
	return s
}

type nameTagPair struct {
	name string
	tag  string
}

type TrackState struct {
	tagFormatError     bool
	tagReadError       bool
	noTags             bool
	numberingConflict  bool
	trackNameConflict  bool
	albumNameConflict  bool
	artistNameConflict bool
}

func (s TrackState) HasNumberingConflict() bool {
	return s.numberingConflict
}

func (s TrackState) HasTrackNameConflict() bool {
	return s.trackNameConflict
}

func (s TrackState) HasAlbumNameConflict() bool {
	return s.albumNameConflict
}

func (s TrackState) HasArtistNameConflict() bool {
	return s.artistNameConflict
}

func (s TrackState) HasTaggingConflicts() bool {
	return s.numberingConflict || s.trackNameConflict || s.albumNameConflict || s.artistNameConflict
}

func (t *Track) AnalyzeIssues() TrackState {
	switch t.TaggedTrack {
	case trackUnknownFormatError:
		return TrackState{tagFormatError: true}
	case TrackUnknownTagReadError:
		return TrackState{tagReadError: true}
	case trackUnknownTagsNotRead:
		return TrackState{noTags: true}
	default:
		return TrackState{
			numberingConflict:  t.TaggedTrack != t.TrackNumber,
			trackNameConflict:  !isComparable(nameTagPair{name: t.Name, tag: t.TaggedTitle}),
			albumNameConflict:  !isComparable(nameTagPair{name: t.ContainingAlbum.Name(), tag: t.TaggedAlbum}),
			artistNameConflict: !isComparable(nameTagPair{name: t.ContainingAlbum.RecordingArtist.Name, tag: t.TaggedArtist}),
		}
	}
}

func (t *Track) FindDifferences() []string {
	s := t.AnalyzeIssues()
	if s.tagFormatError {
		return []string{trackDiffBadTags}
	}
	if s.tagReadError {
		return []string{trackDiffUnreadableTags}
	}
	if s.noTags {
		return []string{trackDiffUnreadTags}
	}
	if !s.HasTaggingConflicts() {
		return nil
	}
	var differences []string
	if s.HasNumberingConflict() {
		differences = append(differences,
			fmt.Sprintf("track number %d does not agree with track tag %d", t.TrackNumber, t.TaggedTrack))
	}
	if s.HasTrackNameConflict() {
		differences = append(differences,
			fmt.Sprintf("title %q does not agree with title tag %q", t.Name, t.TaggedTitle))
	}
	if s.HasAlbumNameConflict() {
		differences = append(differences,
			fmt.Sprintf("album %q does not agree with album tag %q", t.ContainingAlbum.Name(), t.TaggedAlbum))
	}
	if s.HasArtistNameConflict() {
		differences = append(differences,
			fmt.Sprintf("artist %q does not agree with artist tag %q", t.ContainingAlbum.RecordingArtist.Name, t.TaggedArtist))
	}
	return differences
}

func isComparable(p nameTagPair) bool {
	fileName := strings.ToLower(p.name)
	tag := strings.ToLower(p.tag)
	// strip off illegal end characters from the tag
	for strings.HasSuffix(tag, " ") {
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

func (t *Track) EditTags() error {
	a := t.AnalyzeIssues()
	if !a.HasTaggingConflicts() {
		return fmt.Errorf("track %d %q of album %q by artist %q has no tagging conflicts, no edit needed",
			t.TrackNumber, t.Name, t.ContainingAlbum.Name(), t.ContainingAlbum.RecordingArtist.Name)
	}
	tag, err := id3v2.Open(t.Path, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()
	tag.SetDefaultEncoding(id3v2.EncodingUTF8)
	if a.HasAlbumNameConflict() {
		tag.SetAlbum(t.ContainingAlbum.Name())
	}
	if a.HasArtistNameConflict() {
		tag.SetArtist(t.ContainingAlbum.RecordingArtist.Name)
	}
	if a.HasTrackNameConflict() {
		tag.SetTitle(t.Name)
	}
	if a.HasNumberingConflict() {
		tag.AddTextFrame("TRCK", tag.DefaultEncoding(), fmt.Sprintf("%d", t.TrackNumber))
	}
	return tag.Save()
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
			if tags, err := reader(t.Path); err != nil {
				logrus.WithFields(logrus.Fields{internal.LOG_PATH: t.Path, internal.LOG_ERROR: err}).Warn(internal.LOG_CANNOT_READ_FILE)
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
