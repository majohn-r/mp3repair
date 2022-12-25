package files

import (
	"fmt"
	"io/fs"
	"mp3/internal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bogem/id3v2/v2"
	"github.com/cheggaaa/pb/v3"
	"github.com/majohn-r/output"
)

const (
	rawExtension            = "mp3"
	defaultFileExtension    = "." + rawExtension
	defaultTrackNamePattern = "^\\d+[\\s-].+\\." + rawExtension + "$"

	mcdiFrame  = "MCDI"
	trackFrame = "TRCK"
)

// Track encapsulates data about a track in an album.
type Track struct {
	path            string // full path to the file associated with the track, including the file itself
	name            string // name of the track, without the track number or file extension, e.g., "First Track"
	number          int    // number of the track
	containingAlbum *Album
	// this is read from the file only when needed; file i/o is expensive
	tM *trackMetadata
}

// String returns the track's full path (implementation of Stringer interface).
func (t *Track) String() string {
	return t.path
}

// Path returns the track's full path.
func (t *Track) Path() string {
	return t.path
}

// Directory returns the directory containing the track file - in other words,
// its Album directory
func (t *Track) Directory() string {
	return filepath.Dir(t.path)
}

// FileName returns the track's full file name, minus its containing directory.
func (t *Track) FileName() string {
	return filepath.Base(t.path)
}

// Name returns the name of the track without its extension and track number.
func (t *Track) Name() string {
	return t.name
}

// Number returns the track's number as defined by its filename.
func (t *Track) Number() int {
	return t.number
}

func copyTrack(t *Track, a *Album) *Track {
	return &Track{
		path:            t.path,
		name:            t.name,
		number:          t.number,
		tM:              t.tM,
		containingAlbum: a, // do not use source track's album!
	}
}

func newTrackFromFile(a *Album, f fs.DirEntry, simpleName string, trackNumber int) *Track {
	return NewTrack(a, f.Name(), simpleName, trackNumber)
}

// NewTrack creates a new instance of Track without (expensive) tag data.
func NewTrack(a *Album, fullName, simpleName string, trackNumber int) *Track {
	return &Track{
		path:            a.subDirectory(fullName),
		name:            simpleName,
		number:          trackNumber,
		containingAlbum: a,
	}
}

// Tracks is used for sorting tracks spanning albums and artists.
type Tracks []*Track

// Len returns the number of *Track instances.
func (t Tracks) Len() int {
	return len(t)
}

// Less returns true if the first track's artist comes before the second track's
// artist. If the tracks are from the same artist, then it returns true if the
// first track's album comes before the second track's album. If the tracks come
// from the same artist and album, then it returns true if the first track's
// track number comes before the second track's track number.
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
		}
		return album1.Name() < album2.Name()
	}
	return artist1 < artist2
}

// Swap swaps two tracks.
func (t Tracks) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

var trackNameRegex *regexp.Regexp = regexp.MustCompile(defaultTrackNamePattern)

// BackupDirectory returns the path of the backup directory for this track.
func (t *Track) BackupDirectory() string {
	return t.containingAlbum.BackupDirectory()
}

func (t *Track) needsMetadata() bool {
	return t.tM == nil
}

func (t *Track) hasTagError() bool {
	return t.tM != nil && len(t.tM.errors()) != 0
}

// SetMetadata sets metadata read from the file and is public so it can be
// called from unit tests.
func (t *Track) SetMetadata(tM *trackMetadata) {
	t.tM = tM
}

// MetadataState contains information about metadata problems
type MetadataState struct {
	hasError           bool
	noMetadata         bool
	numberingConflict  bool
	trackNameConflict  bool
	albumNameConflict  bool
	artistNameConflict bool
	genreConflict      bool
	yearConflict       bool
	mcdiConflict       bool
}

// HasNumberingConflict returns true if there is a conflict between the track
// number (as derived from the track's file name) and the value of the track's
// ID3V2 TRCK frame.
func (m MetadataState) HasNumberingConflict() bool {
	return m.numberingConflict
}

// HasTrackNameConflict returns true if there is a conflict between the track
// name (as derived from the track's file name) and the value of the track's
// ID3V2 TIT2 frame.
func (m MetadataState) HasTrackNameConflict() bool {
	return m.trackNameConflict
}

// HasAlbumNameConflict returns true if there is a conflict between the name of
// the album the track is associated with and the value of the track's ID3V2
// TALB frame.
func (m MetadataState) HasAlbumNameConflict() bool {
	return m.albumNameConflict
}

// HasArtistNameConflict returns true if there is a conflict between the track's
// recording artist and the value of the track's ID3V2 TPE1 frame.
func (m MetadataState) HasArtistNameConflict() bool {
	return m.artistNameConflict
}

// HasTaggingConflicts returns true if there are any conflicts between the
// track's ID3V2 frame values and their corresponding file-based values.
func (m MetadataState) HasTaggingConflicts() bool {
	return m.numberingConflict ||
		m.trackNameConflict ||
		m.albumNameConflict ||
		m.artistNameConflict ||
		m.genreConflict ||
		m.yearConflict ||
		m.mcdiConflict
}

// HasMCDIConflict returns true if there is conflict between the track's album's
// music CD identifier and the value of the track's ID3V2 MCDI frame.
func (m MetadataState) HasMCDIConflict() bool {
	return m.mcdiConflict
}

// HasGenreConflict returns true if there is conflict between the track's
// album's genre and the value of the track's ID3V2 TCON frame.
func (m MetadataState) HasGenreConflict() bool {
	return m.genreConflict
}

// HasYearConflict returns true if there is conflict between the track's album's
// year and the value of the track's ID3V2 TYER frame.
func (m MetadataState) HasYearConflict() bool {
	return m.yearConflict
}

// ReconcileMetadata determines whether there are problems with the track's
// metadata.
func (t *Track) ReconcileMetadata() MetadataState {
	if t.tM == nil {
		return MetadataState{noMetadata: true}
	}
	if !t.tM.isValid() {
		return MetadataState{hasError: true}
	}
	return MetadataState{
		numberingConflict:  t.tM.trackDiffers(t.number),
		trackNameConflict:  t.tM.trackTitleDiffers(t.name),
		albumNameConflict:  t.tM.albumTitleDiffers(t.containingAlbum.canonicalTitle),
		artistNameConflict: t.tM.artistNameDiffers(t.containingAlbum.recordingArtist.canonicalName),
		genreConflict:      t.tM.genreDiffers(t.containingAlbum.canonicalGenre),
		yearConflict:       t.tM.yearDiffers(t.containingAlbum.canonicalYear),
		mcdiConflict:       t.tM.mcdiDiffers(t.containingAlbum.musicCDIdentifier),
	}
}

// ReportMetadataProblems returns a slice of strings describing the problems
// found by calling ReconcileMetadata().
func (t *Track) ReportMetadataProblems() []string {
	s := t.ReconcileMetadata()
	if s.hasError {
		return []string{"differences cannot be determined: there was an error reading metadata"}
	}
	if s.noMetadata {
		return []string{"differences cannot be determined: metadata has not been read"}
	}
	if !s.HasTaggingConflicts() {
		return nil
	}
	var differences []string
	if s.HasNumberingConflict() {
		differences = append(differences,
			fmt.Sprintf("metadata does not agree with track number %d", t.number))
	}
	if s.HasTrackNameConflict() {
		differences = append(differences,
			fmt.Sprintf("metadata does not agree with track name %q", t.name))
	}
	if s.HasAlbumNameConflict() {
		differences = append(differences,
			fmt.Sprintf("metadata does not agree with album name %q", t.containingAlbum.canonicalTitle))
	}
	if s.HasArtistNameConflict() {
		differences = append(differences,
			fmt.Sprintf("metadata does not agree with artist name %q", t.containingAlbum.recordingArtist.canonicalName))
	}
	if s.HasGenreConflict() {
		differences = append(differences,
			fmt.Sprintf("metadata does not agree with album genre %q", t.containingAlbum.canonicalGenre))
	}
	if s.HasYearConflict() {
		differences = append(differences,
			fmt.Sprintf("metadata does not agree with album year %q", t.containingAlbum.canonicalYear))
	}
	if s.HasMCDIConflict() {
		differences = append(differences,
			fmt.Sprintf("metadata does not agree with the MCDI frame %q", string(t.containingAlbum.musicCDIdentifier.Body)))
	}
	sort.Strings(differences)
	return differences
}

var noEditNeededError = fmt.Errorf("no edit required")

// EditTags verifies that a track's tags need to be edited and then performs
// that work
func (t *Track) EditTags() (e []error) {
	if !t.ReconcileMetadata().HasTaggingConflicts() {
		e = append(e, noEditNeededError)
	} else {
		e = append(e, editTags(t)...)
	}
	return
}

// use of semaphores nicely documented here:
// https://gist.github.com/repejota/ed9070d57c23102d50c94e1a126b2f5b

type empty struct{}

var semaphores = make(chan empty, 20) // 20 is a typical limit for open files

func (t *Track) readTags(bar *pb.ProgressBar) {
	if t.needsMetadata() {
		semaphores <- empty{} // block while full
		go func() {
			defer func() {
				bar.Increment()
				<-semaphores // read to release a slot
			}()
			t.SetMetadata(readMetadata(t.path))
		}()
	}
}

// ReadMetadata reads the metadata for all the artists' tracks.
func ReadMetadata(o output.Bus, artists []*Artist) {
	// count the tracks
	count := 0
	for _, artist := range artists {
		for _, album := range artist.Albums() {
			count += len(album.Tracks())
		}
	}
	o.WriteCanonicalError("Reading track metadata")
	// derived from the Default ProgressBarTemplate used by the progress bar,
	// following guidance in the ElementSpeed definition to change the output to
	// display the speed in tracks per second
	t := `{{with string . "prefix"}}{{.}} {{end}}{{counters . }} {{bar . }} {{percent . }} {{speed . "%s tracks per second"}}{{with string . "suffix"}} {{.}}{{end}}`
	bar := pb.New(count).SetTemplateString(t).Start()
	for _, artist := range artists {
		for _, album := range artist.Albums() {
			for _, track := range album.Tracks() {
				track.readTags(bar)
			}
		}
	}
	waitForSemaphoresDrained()
	bar.Finish()
	processAlbumMetadata(o, artists)
	processArtistMetadata(o, artists)
	reportAllTrackErrors(o, artists)
}

func processArtistMetadata(o output.Bus, artists []*Artist) {
	for _, artist := range artists {
		names := make(map[string]int)
		for _, album := range artist.Albums() {
			for _, track := range album.Tracks() {
				if track.tM != nil && track.tM.isValid() && track.tM.canonicalArtistNameMatches(artist.name) {
					names[track.tM.canonicalArtist()]++
				}
			}
		}
		if chosenName, ok := pickKey(names); !ok {
			reportAmbiguousChoices(o, "artist name", artist.Name(), names)
			logAmbiguousValue(o, map[string]any{
				"field":      "artist name",
				"settings":   names,
				"artistName": artist.Name(),
			})
		} else if len(chosenName) > 0 {
			artist.canonicalName = chosenName
		}
	}
}

func reportAmbiguousChoices(o output.Bus, subject, context string, choices map[string]int) {
	o.WriteCanonicalError("There are multiple %s fields for %q, and there is no unambiguously preferred choice; candidates are %v", subject, context, friendlyEncode(choices))
}

func logAmbiguousValue(o output.Bus, m map[string]any) {
	o.Log(output.Error, "no value has a majority of instances", m)
}

func processAlbumMetadata(o output.Bus, artists []*Artist) {
	for _, artist := range artists {
		for _, album := range artist.Albums() {
			mcdis := make(map[string]int)
			mcdiFrames := make(map[string]id3v2.UnknownFrame)
			genres := make(map[string]int)
			years := make(map[string]int)
			albumTitles := make(map[string]int)
			for _, track := range album.Tracks() {
				if track.tM == nil || !track.tM.isValid() {
					continue
				}
				genre := strings.ToLower(track.tM.canonicalGenre())
				if len(genre) > 0 && !strings.HasPrefix(genre, "unknown") {
					genres[track.tM.canonicalGenre()]++
				}
				if track.tM.canonicalYear() != "" {
					years[track.tM.canonicalYear()]++
				}
				if track.tM.canonicalAlbumTitleMatches(album.name) {
					albumTitles[track.tM.canonicalAlbum()]++
				}
				mcdiKey := string(track.tM.canonicalMusicCDIdentifier().Body)
				mcdis[mcdiKey]++
				mcdiFrames[mcdiKey] = track.tM.canonicalMusicCDIdentifier()
			}
			if chosenGenre, ok := pickKey(genres); !ok {
				reportAmbiguousChoices(o, "genre", fmt.Sprintf("%s by %s", album.Name(), artist.Name()), genres)
				logAmbiguousValue(o, map[string]any{
					"field":      "genre",
					"settings":   genres,
					"albumName":  album.Name(),
					"artistName": artist.Name(),
				})
			} else {
				album.canonicalGenre = chosenGenre
			}
			if chosenYear, ok := pickKey(years); !ok {
				reportAmbiguousChoices(o, "year", fmt.Sprintf("%s by %s", album.Name(), artist.Name()), years)
				logAmbiguousValue(o, map[string]any{
					"field":      "year",
					"settings":   years,
					"albumName":  album.Name(),
					"artistName": artist.Name(),
				})
			} else {
				album.canonicalYear = chosenYear
			}
			if chosenAlbumTitle, ok := pickKey(albumTitles); !ok {
				reportAmbiguousChoices(o, "album title", fmt.Sprintf("%s by %s", album.Name(), artist.Name()), albumTitles)
				logAmbiguousValue(o, map[string]any{
					"field":      "album title",
					"settings":   albumTitles,
					"albumName":  album.Name(),
					"artistName": artist.Name(),
				})
			} else if chosenAlbumTitle != "" {
				album.canonicalTitle = chosenAlbumTitle
			}
			if chosenMCDI, ok := pickKey(mcdis); !ok {
				reportAmbiguousChoices(o, "MCDI frame", fmt.Sprintf("%s by %s", album.Name(), artist.Name()), mcdis)
				logAmbiguousValue(o, map[string]any{
					"field":      "mcdi frame",
					"settings":   mcdis,
					"albumName":  album.Name(),
					"artistName": artist.Name(),
				})
			} else {
				album.musicCDIdentifier = mcdiFrames[chosenMCDI]
			}
		}
	}
}

func friendlyEncode(m map[string]int) string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var values []string
	for _, k := range keys {
		count := m[k]
		if count == 1 {
			values = append(values, fmt.Sprintf("%q: 1 instance", k))
		} else {
			values = append(values, fmt.Sprintf("%q: %d instances", k, count))
		}
	}
	return fmt.Sprintf("{%s}", strings.Join(values, ", "))
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

var (
	reportTagError = map[sourceType]func(output.Bus, *Track, error){
		id3v1Source: ReportID3V1TagError,
		id3v2Source: ReportID3V2TagError,
	}
)

// ReportID3V1TagError outputs a problem reading the ID3V1 tag as an error and
// as a log record
func ReportID3V1TagError(o output.Bus, t *Track, e error) {
	o.WriteCanonicalError("An error occurred when trying to read ID3V1 tag information for track %q on album %q by artist %q: %q", t.Name(), t.AlbumName(), t.RecordingArtist(), e.Error())
	o.Log(output.Error, "id3v1 tag error", map[string]any{
		"track": t.String(),
		"error": e,
	})
}

// ReportID3V2TagError outputs a problem reading the ID3V2 tag as an error and
// as a log record
func ReportID3V2TagError(o output.Bus, t *Track, e error) {
	o.WriteCanonicalError("An error occurred when trying to read ID3V2 tag information for track %q on album %q by artist %q: %q", t.Name(), t.AlbumName(), t.RecordingArtist(), e.Error())
	o.Log(output.Error, "id3v2 tag error", map[string]any{
		"track": t.String(),
		"error": e,
	})
}
func reportAllTrackErrors(o output.Bus, artists []*Artist) {
	for _, artist := range artists {
		for _, album := range artist.Albums() {
			for _, track := range album.Tracks() {
				reportTrackErrors(o, track)
			}
		}
	}
}

func reportTrackErrors(o output.Bus, track *Track) {
	if track.hasTagError() {
		for _, source := range []sourceType{id3v1Source, id3v2Source} {
			e := track.tM.err[source]
			if e != nil {
				reportTagError[source](o, track, e)
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
// number, no file extension); it is for testing only and assumes that the input
// name is well-formed.
func ParseTrackNameForTesting(name string) (simpleName string, trackNumber int) {
	simpleName, trackNumber, _ = parseTrackName(nil, name, nil, defaultFileExtension)
	return
}

func parseTrackName(o output.Bus, name string, album *Album, ext string) (simpleName string, trackNumber int, valid bool) {
	if !trackNameRegex.MatchString(name) {
		o.Log(output.Error, "the track name cannot be parsed", map[string]any{
			"trackName":  name,
			"albumName":  album.name,
			"artistName": album.RecordingArtistName(),
		})
		o.WriteCanonicalError("The track %q on album %q by artist %q cannot be parsed", name, album.name, album.RecordingArtistName())
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

// AlbumPath returns the path of the track's album.
func (t *Track) AlbumPath() string {
	if t.containingAlbum == nil {
		return ""
	}
	return t.containingAlbum.path
}

// AlbumName returns the name of the track's album.
func (t *Track) AlbumName() string {
	if t.containingAlbum == nil {
		return ""
	}
	return t.containingAlbum.name
}

// RecordingArtist returns the name of the artist on whose album this track
// appears.
func (t *Track) RecordingArtist() string {
	if t.containingAlbum == nil {
		return ""
	}
	return t.containingAlbum.RecordingArtistName()
}

// Copy copies the track file to a specified destination path.
func (t *Track) Copy(destination string) error {
	return internal.CopyFile(t.path, destination)
}

// ID3V1Diagnostics returns the ID3V1 tag contents, if any; a missing ID3V1 tag
// (e.g., the input file is too short to have an ID3V1 tag), or an invalid ID3V1
// tag (isValid() is false), returns a non-nil error
func (t *Track) ID3V1Diagnostics() ([]string, error) {
	return readID3v1Metadata(t.path)
}

// ID3V2Diagnostics returns ID3V2 tag data - the ID3V2 version, its encoding,
// and a slice of all the frames in the tag.
func (t *Track) ID3V2Diagnostics() (version byte, encoding string, frames []string, e error) {
	version, encoding, frames, _, e = readID3V2Metadata(t.path)
	return
}

var frameToName = map[string]string{
	"TCOM": "Composer",
	"TEXT": "Lyricist",
	"TIT3": "Subtitle",
	"TKEY": "Key",
	"TPE2": "Orchestra/Band",
	"TPE3": "Conductor",
}

// Details returns relevant details about the track
func (t *Track) Details() (map[string]string, error) {
	_, _, _, frames, err := readID3V2Metadata(t.path)
	if err != nil {
		return nil, err
	}
	m := map[string]string{}
	for _, frame := range frames {
		if value, ok := frameToName[frame.name]; ok {
			m[value] = frame.value
		}
	}
	return m, nil
}
