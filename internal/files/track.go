package files

import (
	"fmt"
	"io/fs"
	"mp3/internal"
	"regexp"
	"strings"
	"time"

	"github.com/bogem/id3v2/v2"
	"github.com/sirupsen/logrus"
)

const (
	rawExtension             = "mp3"
	defaultFileExtension     = "." + rawExtension
	defaultTrackNamePattern  = "^\\d+[\\s-].+\\." + rawExtension + "$"
	trackDiffBadTags         = "cannot determine differences, tags were not recognized"
	trackDiffUnreadableTags  = "cannot determine differences, could not read tags"
	trackDiffUnreadTags      = "cannot determine differences, tags have not been read"
	trackUnknownTagsNotRead  = 0
	trackUnknownFormatError  = -1
	trackUnknownTagReadError = -2
)

// Track encapsulates data about a track in an album
type Track struct {
	path            string // full path to the file associated with the track, including the file itself
	name            string // name of the track, without the track number or file extension, e.g., "First Track"
	number          int    // number of the track
	containingAlbum *Album
	// these fields are populated when needed; acquisition is expensive
	title  string // track title per mp3 tag TIT2 frame
	track  int    // track number per mp3 tag TRCK frame - initially set to trackUnknownTagsNotRead
	album  string // album name per mp3 tag TALB frame
	artist string // artist name per mp3 tag TPE1 frame
}

// String returns the track's path (implementation of Stringer interface)
func (t *Track) String() string {
	return t.path
}

// Name returns the simple name of the track
func (t *Track) Name() string {
	return t.name
}

// Number returns the track's number per its filename
func (t *Track) Number() int {
	return t.number
}

func copyTrack(t *Track, a *Album) *Track {
	return &Track{
		path:            t.path,
		name:            t.name,
		number:          t.number,
		album:           t.album,
		artist:          t.artist,
		title:           t.title,
		track:           t.track,
		containingAlbum: a, // do not use source track's album!
	}
}

func newTrackFromFile(a *Album, f fs.FileInfo, simpleName string, trackNumber int) *Track {
	return NewTrack(a, f.Name(), simpleName, trackNumber)
}

// NewTrack creates a new instance of Track without (expensive) tag data
func NewTrack(a *Album, fullName string, simpleName string, trackNumber int) *Track {
	return &Track{
		path:            a.subDirectory(fullName),
		name:            simpleName,
		number:          trackNumber,
		track:           trackUnknownTagsNotRead,
		containingAlbum: a,
	}
}

// logic for sorting tracks spanning albums and artists
type Tracks []*Track

func (t Tracks) Len() int {
	return len(t)
}

func (t Tracks) Less(i, j int) bool {
	track1 := t[i]
	track2 := t[j]
	album1 := track1.containingAlbum
	album2 := track2.containingAlbum
	artist1 := album1.RecordingArtistName()
	artist2 := album2.RecordingArtistName()
	// compare artist name first
	if artist1 == artist2 {
		// artist names are the same ... try the album name next
		if album1.Name() == album2.Name() {
			// and album names are the same ... go by track number
			return track1.number < track2.number
		} else {
			return album1.Name() < album2.Name()
		}
	} else {
		return artist1 < artist2
	}
}

func (t Tracks) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

// TaggedTrackData contains raw tag frames data
type TaggedTrackData struct {
	album  string
	artist string
	title  string
	track  string
}

// NewTaggedTrackData creates a new instance of TaggedTrackData
func NewTaggedTrackData(albumFrame, artistFrame, titleFrame, numberFrame string) *TaggedTrackData {
	return &TaggedTrackData{
		album:  albumFrame,
		artist: artistFrame,
		title:  titleFrame,
		track:  numberFrame,
	}
}

var trackNameRegex *regexp.Regexp = regexp.MustCompile(defaultTrackNamePattern)

// BackupDirectory returns the path for this track
func (t *Track) BackupDirectory() string {
	return t.containingAlbum.subDirectory(backupDirName)
}

func (t *Track) needsTaggedData() bool {
	return t.track == trackUnknownTagsNotRead
}

func (t *Track) setTagReadError() {
	t.track = trackUnknownTagReadError
}

func (t *Track) setTagFormatError() {
	t.track = trackUnknownFormatError
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

// SetTags sets track frame fields
func (t *Track) SetTags(d *TaggedTrackData) {
	if trackNumber, err := toTrackNumber(d.track); err != nil {
		logrus.WithFields(logrus.Fields{
			internal.LOG_PATH:  t.String,
			"trackTag":         d.track,
			internal.LOG_ERROR: err}).Warn("invalid track tag")
		t.setTagFormatError()
	} else {
		t.album = removeLeadingBOMs(d.album)
		t.artist = removeLeadingBOMs(d.artist)
		t.title = removeLeadingBOMs(d.title)
		t.track = trackNumber
	}
}

// depending on encoding, frame values may begin with a BOM (byte order mark)
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

type taggedTrackState struct {
	tagFormatError     bool
	tagReadError       bool
	noTags             bool
	numberingConflict  bool
	trackNameConflict  bool
	albumNameConflict  bool
	artistNameConflict bool
}

// HasNumberingConflict returns true if there is a conflict between the track
// number (as derived from the track's file name) and the value of the track's
// TRCK frame.
func (s taggedTrackState) HasNumberingConflict() bool {
	return s.numberingConflict
}

// HasTrackNameConflict returns true if there is a conflict between the track
// name (as derived from the track's file name) and the value of the track's
// TIT2 frame.
func (s taggedTrackState) HasTrackNameConflict() bool {
	return s.trackNameConflict
}

// HasAlbumNameConflict returns true if there is a conflict between the name of
// the album the track is associated with and the value of the track's TALB
// frame.
func (s taggedTrackState) HasAlbumNameConflict() bool {
	return s.albumNameConflict
}

// HasArtistNameConflict returns true if there is a conflict between the track's
// recording artist and the value of the track's TPE1 frame.
func (s taggedTrackState) HasArtistNameConflict() bool {
	return s.artistNameConflict
}

// HasTaggingConflicts returns true if there are any conflicts between the
// track's frame values and their corresponding file-based values.
func (s taggedTrackState) HasTaggingConflicts() bool {
	return s.numberingConflict || s.trackNameConflict || s.albumNameConflict || s.artistNameConflict
}

// AnalyzeIssues determines whether there are problems with the track's
// frame-based values.
func (t *Track) AnalyzeIssues() taggedTrackState {
	switch t.track {
	case trackUnknownFormatError:
		return taggedTrackState{tagFormatError: true}
	case trackUnknownTagReadError:
		return taggedTrackState{tagReadError: true}
	case trackUnknownTagsNotRead:
		return taggedTrackState{noTags: true}
	default:
		return taggedTrackState{
			numberingConflict:  t.track != t.number,
			trackNameConflict:  !isComparable(nameTagPair{name: t.name, tag: t.title}),
			albumNameConflict:  !isComparable(nameTagPair{name: t.containingAlbum.Name(), tag: t.album}),
			artistNameConflict: !isComparable(nameTagPair{name: t.containingAlbum.RecordingArtistName(), tag: t.artist}),
		}
	}
}

// FindDifferences returns strings describing the problems found by calling
// AnalyzeIssues.
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
			fmt.Sprintf("track number %d does not agree with track tag %d", t.number, t.track))
	}
	if s.HasTrackNameConflict() {
		differences = append(differences,
			fmt.Sprintf("title %q does not agree with title tag %q", t.name, t.title))
	}
	if s.HasAlbumNameConflict() {
		differences = append(differences,
			fmt.Sprintf("album %q does not agree with album tag %q", t.containingAlbum.Name(), t.album))
	}
	if s.HasArtistNameConflict() {
		differences = append(differences,
			fmt.Sprintf("artist %q does not agree with artist tag %q", t.containingAlbum.RecordingArtistName(), t.artist))
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

// RawReadTags reads the tag from an MP3 file and collects interesting frame
// values.
func RawReadTags(path string) (d *TaggedTrackData, err error) {
	var tag *id3v2.Tag
	if tag, err = id3v2.Open(path, id3v2.Options{Parse: true, ParseFrames: stdFrames}); err != nil {
		return
	}
	defer tag.Close()
	d = &TaggedTrackData{
		album:  tag.Album(),
		artist: tag.Artist(),
		title:  tag.Title(),
		track:  tag.GetTextFrame("TRCK").Text,
	}
	return
}

// EditTags rewrites tag frames to match file-based values and saves (re-writes)
// the associated MP3 file.
func (t *Track) EditTags() error {
	a := t.AnalyzeIssues()
	if !a.HasTaggingConflicts() {
		return fmt.Errorf("track %d %q of album %q by artist %q has no tagging conflicts, no edit needed",
			t.number, t.name, t.containingAlbum.Name(), t.containingAlbum.RecordingArtistName())
	}
	tag, err := id3v2.Open(t.path, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()
	tag.SetDefaultEncoding(id3v2.EncodingUTF8)
	if a.HasAlbumNameConflict() {
		tag.SetAlbum(t.containingAlbum.Name())
	}
	if a.HasArtistNameConflict() {
		tag.SetArtist(t.containingAlbum.RecordingArtistName())
	}
	if a.HasTrackNameConflict() {
		tag.SetTitle(t.name)
	}
	if a.HasNumberingConflict() {
		tag.AddTextFrame("TRCK", tag.DefaultEncoding(), fmt.Sprintf("%d", t.number))
	}
	return tag.Save()
}

// use of semaphores nicely documented here:
// https://gist.github.com/repejota/ed9070d57c23102d50c94e1a126b2f5b

type empty struct{}

var semaphores = make(chan empty, 20) // 20 is a typical limit for open files

func (t *Track) readTags(reader func(string) (*TaggedTrackData, error)) {
	if t.needsTaggedData() {
		semaphores <- empty{} // block while full
		go func() {
			defer func() {
				<-semaphores // read to release a slot
			}()
			if tags, err := reader(t.path); err != nil {
				logrus.WithFields(logrus.Fields{internal.LOG_PATH: t.String, internal.LOG_ERROR: err}).Warn(internal.LOG_CANNOT_READ_FILE)
				t.setTagReadError()
			} else {
				t.SetTags(tags)
			}
		}()
	}
}

// UpdateTracks reads the MP3 tags for all the associated tracks.
func UpdateTracks(artists []*Artist, reader func(string) (*TaggedTrackData, error)) {
	for _, artist := range artists {
		for _, album := range artist.Albums() {
			for _, track := range album.Tracks() {
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

// ParseTrackNameForTesting parses a name into its simple form (no leading track
// number, no file extension); it is for testing only
func ParseTrackNameForTesting(name string) (simpleName string, trackNumber int) {
	simpleName, trackNumber, _ = parseTrackName(name, nil, defaultFileExtension)
	return
}

func parseTrackName(name string, album *Album, ext string) (simpleName string, trackNumber int, valid bool) {
	if !trackNameRegex.MatchString(name) {
		logrus.WithFields(logrus.Fields{internal.LOG_TRACK_NAME: name, internal.LOG_ALBUM_NAME: album.name, internal.LOG_ARTIST_NAME: album.RecordingArtistName()}).Warn(internal.LOG_INVALID_TRACK_NAME)
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

// AlbumPath returns the path to the track's album
func (t *Track) AlbumPath() string {
	if t.containingAlbum == nil {
		return ""
	}
	return t.containingAlbum.path
}

// AlbumName returns the name of the track's album
func (t *Track) AlbumName() string {
	if t.containingAlbum == nil {
		return ""
	}
	return t.containingAlbum.name
}

// RecordingArtist returns the name of the artist on whose album this track
// appears
func (t *Track) RecordingArtist() string {
	if t.containingAlbum == nil {
		return ""
	}
	return t.containingAlbum.RecordingArtistName()
}

// Copy copies the track to a specified destination path
func (t *Track) Copy(destination string) error {
	return internal.CopyFile(t.path, destination)
}
