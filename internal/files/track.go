package files

import (
	"fmt"
	"io/fs"
	"mp3/internal"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bogem/id3v2/v2"
)

const (
	albumFrame              = "TALB"
	artistFrame             = "TPE1"
	defaultFileExtension    = "." + rawExtension
	defaultTrackNamePattern = "^\\d+[\\s-].+\\." + rawExtension + "$"
	fkAlbumName             = "albumName"
	fkArtistName            = "artistName"
	fkFieldName             = "field"
	fkSettings              = "settings"
	fkTrackName             = "trackName"
	genreFrame              = "TCON"
	rawExtension            = "mp3"
	titleFrame              = "TIT2"
	trackDiffUnreadTags     = "cannot determine differences, tags have not been read"
	trackDiffError          = "cannot determine differences, there was an error reading tags"
	trackFrame              = "TRCK"
	yearFrame               = "TYER"
)

// Track encapsulates data about a track in an album
type Track struct {
	path            string // full path to the file associated with the track, including the file itself
	name            string // name of the track, without the track number or file extension, e.g., "First Track"
	number          int    // number of the track
	containingAlbum *Album
	// these fields are populated when needed; acquisition is expensive
	TaggedTrackData
}

// String returns the track's path (implementation of Stringer interface)
func (t *Track) String() string {
	return t.path
}

// Path returns the track's full path
func (t *Track) Path() string {
	return t.path
}

// Directory returns the track's directory
func (t *Track) Directory() string {
	return filepath.Dir(t.path)
}

// FileName returns the track's full file name
func (t *Track) FileName() string {
	return filepath.Base(t.path)
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
		TaggedTrackData: t.TaggedTrackData,
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
	genre  string
	year   string
	track  int
	err    string
}

// NewTaggedTrackData creates a new instance of TaggedTrackData
func NewTaggedTrackData(albumFrame string, artistFrame string, titleFrame string, evaluatedNumberFrame int) *TaggedTrackData {
	return &TaggedTrackData{
		album:  albumFrame,
		artist: artistFrame,
		title:  titleFrame,
		track:  evaluatedNumberFrame,
		err:    "",
	}
}

var trackNameRegex *regexp.Regexp = regexp.MustCompile(defaultTrackNamePattern)

// BackupDirectory returns the path for this track
func (t *Track) BackupDirectory() string {
	return t.containingAlbum.BackupDirectory()
}

func (t *Track) needsTaggedData() bool {
	return t.track == 0 && !t.hasTagError()
}

func (t *Track) hasTagError() bool {
	return len(t.err) != 0
}

func toTrackNumber(s string) (i int, err error) {
	// this is more complicated than I wanted, because some mp3 rippers produce
	// track numbers like "12/14", meaning 12th track of 14
	if len(s) == 0 {
		err = fmt.Errorf(internal.ERROR_ZERO_LENGTH)
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
				err = fmt.Errorf(internal.ERROR_DOES_NOT_BEGIN_WITH_DIGIT)
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
	t.TaggedTrackData = *d
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
	hasError           bool
	noTags             bool
	numberingConflict  bool
	trackNameConflict  bool
	albumNameConflict  bool
	artistNameConflict bool
	genreConflict      bool
	yearConflict       bool
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
	return s.numberingConflict || s.trackNameConflict || s.albumNameConflict || s.artistNameConflict || s.genreConflict || s.yearConflict
}

// HasGenreConflict returs true if there is conflict between the track's album's
// genre and the value of the track's TCON frame
func (s taggedTrackState) HasGenreConflict() bool {
	return s.genreConflict
}

// HasYearConflict returs true if there is conflict between the track's album's
// year and the value of the track's TYER frame
func (s taggedTrackState) HasYearConflict() bool {
	return s.yearConflict
}

// AnalyzeIssues determines whether there are problems with the track's
// frame-based values.
func (t *Track) AnalyzeIssues() taggedTrackState {
	if t.hasTagError() {
		return taggedTrackState{hasError: true}
	}
	switch t.track {
	case 0:
		return taggedTrackState{noTags: true}
	default:
		return taggedTrackState{
			numberingConflict:  t.track != t.number,
			trackNameConflict:  !isComparable(nameTagPair{name: t.name, tag: t.title}),
			albumNameConflict:  t.containingAlbum.canonicalTitle != t.album,
			artistNameConflict: !isComparable(nameTagPair{name: t.containingAlbum.RecordingArtistName(), tag: t.artist}),
			genreConflict:      t.genre != t.containingAlbum.genre,
			yearConflict:       t.year != t.containingAlbum.year,
		}
	}
}

// FindDifferences returns strings describing the problems found by calling
// AnalyzeIssues.
func (t *Track) FindDifferences() []string {
	s := t.AnalyzeIssues()
	if s.hasError {
		return []string{trackDiffError}
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
			fmt.Sprintf("album %q does not agree with album tag %q", t.containingAlbum.canonicalTitle, t.album))
	}
	if s.HasArtistNameConflict() {
		differences = append(differences,
			fmt.Sprintf("artist %q does not agree with artist tag %q", t.containingAlbum.RecordingArtistName(), t.artist))
	}
	if s.HasGenreConflict() {
		differences = append(differences,
			fmt.Sprintf("genre %q does not agree with album genre %q", t.genre, t.containingAlbum.genre))
	}
	if s.HasYearConflict() {
		differences = append(differences,
			fmt.Sprintf("year %q does not agree with album year %q", t.year, t.containingAlbum.year))
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

var stdFrames = []string{albumFrame, artistFrame, genreFrame, titleFrame, trackFrame, yearFrame}

// RawReadTags reads the tag from an MP3 file and collects interesting frame
// values.
func RawReadTags(path string) (d *TaggedTrackData) {
	d = &TaggedTrackData{}
	var tag *id3v2.Tag
	var err error
	if tag, err = id3v2.Open(path, id3v2.Options{Parse: true, ParseFrames: stdFrames}); err != nil {
		d.err = fmt.Sprintf("%v", err)
		return
	}
	defer tag.Close()
	if trackNumber, err := toTrackNumber(tag.GetTextFrame(trackFrame).Text); err != nil {
		d.err = fmt.Sprintf("%v", err)
	} else {
		d.album = removeLeadingBOMs(tag.Album())
		d.artist = removeLeadingBOMs(tag.Artist())
		d.genre = removeLeadingBOMs(tag.Genre())
		d.title = removeLeadingBOMs(tag.Title())
		d.track = trackNumber
		d.year = removeLeadingBOMs(tag.Year())
	}
	return
}

// EditTags rewrites tag frames to match file-based values and saves (re-writes)
// the associated MP3 file.
func (t *Track) EditTags() error {
	a := t.AnalyzeIssues()
	if !a.HasTaggingConflicts() {
		return fmt.Errorf(internal.ERROR_EDIT_UNNECESSARY)
	}
	tag, err := id3v2.Open(t.path, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()
	tag.SetDefaultEncoding(id3v2.EncodingUTF8)
	if a.HasAlbumNameConflict() {
		tag.SetAlbum(t.containingAlbum.canonicalTitle)
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
	if a.HasGenreConflict() {
		tag.SetGenre(t.containingAlbum.genre)
	}
	if a.HasYearConflict() {
		tag.SetYear(t.containingAlbum.year)
	}
	return tag.Save()
}

// use of semaphores nicely documented here:
// https://gist.github.com/repejota/ed9070d57c23102d50c94e1a126b2f5b

type empty struct{}

var semaphores = make(chan empty, 20) // 20 is a typical limit for open files

func (t *Track) readTags(reader func(string) *TaggedTrackData) {
	if t.needsTaggedData() {
		semaphores <- empty{} // block while full
		go func() {
			defer func() {
				<-semaphores // read to release a slot
			}()
			t.SetTags(reader(t.path))
		}()
	}
}

// UpdateTracks reads the MP3 tags for all the associated tracks.
func UpdateTracks(o internal.OutputBus, artists []*Artist, reader func(string) *TaggedTrackData) {
	for _, artist := range artists {
		for _, album := range artist.Albums() {
			for _, track := range album.Tracks() {
				track.readTags(reader)
			}
		}
	}
	waitForSemaphoresDrained()
	processAlbumRelatedFrames(o, artists)
	reportTrackErrors(o, artists)
}

func processAlbumRelatedFrames(o internal.OutputBus, artists []*Artist) {
	for _, artist := range artists {
		for _, album := range artist.Albums() {
			genres := make(map[string]int)
			years := make(map[string]int)
			albumTitles := make(map[string]int)
			for _, track := range album.Tracks() {
				genre := strings.ToLower(track.genre)
				if len(genre) > 0 && !strings.HasPrefix(genre, "unknown") {
					genres[track.genre]++
				}
				if len(track.year) != 0 {
					years[track.year]++
				}
				if isComparable(nameTagPair{name: album.name, tag: track.album}) {
					albumTitles[track.album]++
				}
			}
			if chosenGenre, ok := pickKey(genres); !ok {
				o.LogWriter().Warn(internal.LW_AMBIGUOUS_VALUE, map[string]interface{}{
					fkFieldName:  "genre",
					fkSettings:   genres,
					fkAlbumName:  album.Name(),
					fkArtistName: artist.Name(),
				})
			} else {
				album.genre = chosenGenre
			}
			if chosenYear, ok := pickKey(years); !ok {
				o.LogWriter().Warn(internal.LW_AMBIGUOUS_VALUE, map[string]interface{}{
					fkFieldName:  "year",
					fkSettings:   years,
					fkAlbumName:  album.Name(),
					fkArtistName: artist.Name(),
				})
			} else {
				album.year = chosenYear
			}
			if chosenAlbumTitle, ok := pickKey(albumTitles); !ok {
				o.LogWriter().Warn(internal.LW_AMBIGUOUS_VALUE, map[string]interface{}{
					fkFieldName:  "album title",
					fkSettings:   albumTitles,
					fkAlbumName:  album.Name(),
					fkArtistName: artist.Name(),
				})
			} else {
				if len(chosenAlbumTitle) != 0 {
					album.canonicalTitle = chosenAlbumTitle
				}
			}
		}
	}
}

func pickKey(m map[string]int) (s string, ok bool) {
	// add up the total votes, divide by 2, force rounding up
	if len(m) == 0 {
		ok = true
		return
	}
	total := 0
	for _, v := range m {
		total += v
	}
	majority := 1 + (total / 2)
	// look for the one entry that equals or exceeds the majority vote
	for k, v := range m {
		if v >= majority {
			s = k
			ok = true
			return
		}
	}
	return
}

func reportTrackErrors(o internal.OutputBus, artists []*Artist) {
	for _, artist := range artists {
		for _, album := range artist.Albums() {
			for _, track := range album.Tracks() {
				if track.hasTagError() {
					fmt.Fprintf(o.ErrorWriter(), internal.USER_TAG_ERROR, track.name, album.name, artist.name, track.err)
					o.LogWriter().Warn(internal.LW_TAG_ERROR, map[string]interface{}{
						fkTrackName:       track.name,
						fkAlbumName:       album.name,
						fkArtistName:      artist.name,
						internal.FK_ERROR: track.err,
					})
				}
			}
		}
	}
}

func waitForSemaphoresDrained() {
	for len(semaphores) != 0 {
		time.Sleep(1 * time.Microsecond)
	}
}

// ParseTrackNameForTesting parses a name into its simple form (no leading track
// number, no file extension); it is for testing only
func ParseTrackNameForTesting(name string) (simpleName string, trackNumber int) {
	simpleName, trackNumber, _ = parseTrackName(nil, name, nil, defaultFileExtension)
	return
}

func parseTrackName(o internal.OutputBus, name string, album *Album, ext string) (simpleName string, trackNumber int, valid bool) {
	if !trackNameRegex.MatchString(name) {
		o.LogWriter().Warn(internal.LW_INVALID_TRACK_NAME, map[string]interface{}{
			fkTrackName:  name,
			fkAlbumName:  album.name,
			fkArtistName: album.RecordingArtistName(),
		})
		fmt.Fprintf(o.ErrorWriter(), internal.USER_TRACK_NAME_GARBLED, name, album.name, album.RecordingArtistName())
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
