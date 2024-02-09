package files

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bogem/id3v2/v2"
	"github.com/cheggaaa/pb/v3"
	cmd_toolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

const (
	rawExtension            = "mp3"
	defaultFileExtension    = "." + rawExtension
	defaultTrackNamePattern = "^\\d+[\\s-].+\\." + rawExtension + "$"

	mcdiFrame  = "MCDI"
	trackFrame = "TRCK"
)

var (
	openFiles         = make(chan empty, 20) // 20 is a typical limit for open files
	frameDescriptions = map[string]string{
		"TCOM": "Composer",
		"TEXT": "Lyricist",
		"TIT3": "Subtitle",
		"TKEY": "Key",
		"TPE2": "Orchestra/Band",
		"TPE3": "Conductor",
	}
	ErrNoEditNeeded = fmt.Errorf("no edit required")
	trackNameRegex  = regexp.MustCompile(defaultTrackNamePattern)
)

// Track encapsulates data about a track in an album.
type Track struct {
	album    *Album
	fullPath string         // full path to the file associated with the track, including the file itself
	metadata *TrackMetadata // read from the file only when needed; file i/o is expensive
	name     string         // name of the track, without the track number or file extension, e.g., "First Track"
	number   int            // number of the track
}

func (t *Track) GetMetadata() *TrackMetadata {
	return t.metadata
}

func (t *Track) Album() *Album {
	return t.album
}

func (t *Track) WithNumber(i int) *Track {
	t.number = i
	return t
}

func (t *Track) WithAlbum(a *Album) *Track {
	t.album = a
	return t
}

func (t *Track) WithFullPath(s string) *Track {
	t.fullPath = s
	return t
}

func (t *Track) WithMetadata(tm *TrackMetadata) *Track {
	t.metadata = tm
	return t
}

func (t *Track) WithName(s string) *Track {
	t.name = s
	return t
}

func NewEmptyTrack() *Track {
	return &Track{}
}

// String returns the track's full path (implementation of Stringer interface).
func (t *Track) String() string {
	return t.fullPath
}

// Path returns the track's full path.
func (t *Track) Path() string {
	return t.fullPath
}

// Directory returns the directory containing the track file - in other words,
// its Album directory
func (t *Track) Directory() string {
	return filepath.Dir(t.fullPath)
}

// FileName returns the track's full file name, minus its containing directory.
func (t *Track) FileName() string {
	return filepath.Base(t.fullPath)
}

// CommonName returns the name of the track without its extension and track
// number.
func (t *Track) CommonName() string {
	return t.name
}

// Number returns the track's number as defined by its filename.
func (t *Track) Number() int {
	return t.number
}

func (t *Track) Copy(a *Album) *Track {
	return &Track{
		fullPath: t.fullPath,
		name:     t.name,
		number:   t.number,
		metadata: t.metadata,
		album:    a, // do not use source track's album!
	}
}

// NewTrack creates a new instance of Track without (expensive) metadata.
func NewTrack(a *Album, fullName, simpleName string, trackNumber int) *Track {
	return &Track{
		fullPath: a.subDirectory(fullName),
		name:     simpleName,
		number:   trackNumber,
		album:    a,
	}
}

// Tracks is used for sorting tracks spanning albums and artists.
type Tracks []*Track

// Len returns the number of *Track instances.
func (ts Tracks) Len() int {
	return len(ts)
}

// Less returns true if the first track's artist comes before the second track's
// artist. If the tracks are from the same artist, then it returns true if the
// first track's album comes before the second track's album. If the tracks come
// from the same artist and album, then it returns true if the first track's
// track number comes before the second track's track number.
func (ts Tracks) Less(i, j int) bool {
	track1 := ts[i]
	track2 := ts[j]
	album1 := track1.album
	album2 := track2.album
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
func (ts Tracks) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

func (t *Track) NeedsMetadata() bool {
	return t.metadata == nil
}

func (t *Track) HasMetadataError() bool {
	return t.metadata != nil && len(t.metadata.ErrorCauses()) != 0
}

func (t *Track) SetMetadata(tM *TrackMetadata) {
	t.metadata = tM
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
// number (as derived from the track's file name) and the value of any of the
// track's track number metadata.
func (m MetadataState) HasNumberingConflict() bool {
	return m.numberingConflict
}

// HasTrackNameConflict returns true if there is a conflict between the track
// name (as derived from the track's file name) and the value of any of the
// track's track name metadata.
func (m MetadataState) HasTrackNameConflict() bool {
	return m.trackNameConflict
}

// HasAlbumNameConflict returns true if there is a conflict between the name of
// the album the track is associated with and the value of any of the track's
// album name metadata.
func (m MetadataState) HasAlbumNameConflict() bool {
	return m.albumNameConflict
}

// HasArtistNameConflict returns true if there is a conflict between the track's
// recording artist and the value of any of the track's artist name metadata.
func (m MetadataState) HasArtistNameConflict() bool {
	return m.artistNameConflict
}

// HasConflicts returns true if there are any conflicts between the any of the
// track's metadata and their corresponding file-based values.
func (m MetadataState) HasConflicts() bool {
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
// album's genre and the value of any of the track's genre metadata.
func (m MetadataState) HasGenreConflict() bool {
	return m.genreConflict
}

// HasYearConflict returns true if there is conflict between the track's album's
// year and the value of any of the track's year metadata.
func (m MetadataState) HasYearConflict() bool {
	return m.yearConflict
}

// ReconcileMetadata determines whether there are problems with the track's
// metadata.
func (t *Track) ReconcileMetadata() MetadataState {
	if t.metadata == nil {
		return MetadataState{noMetadata: true}
	}
	if !t.metadata.IsValid() {
		return MetadataState{hasError: true}
	}
	return MetadataState{
		numberingConflict:  t.metadata.TrackDiffers(t.number),
		trackNameConflict:  t.metadata.TrackTitleDiffers(t.name),
		albumNameConflict:  t.metadata.AlbumTitleDiffers(t.album.canonicalTitle),
		artistNameConflict: t.metadata.ArtistNameDiffers(t.album.artist.canonicalName),
		genreConflict:      t.metadata.GenreDiffers(t.album.canonicalGenre),
		yearConflict:       t.metadata.YearDiffers(t.album.canonicalYear),
		mcdiConflict:       t.metadata.MCDIDiffers(t.album.musicCDIdentifier),
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
	if !s.HasConflicts() {
		return nil
	}
	var diffs []string
	if s.HasNumberingConflict() {
		diffs = append(diffs,
			fmt.Sprintf("metadata does not agree with track number %d", t.number))
	}
	if s.HasTrackNameConflict() {
		diffs = append(diffs,
			fmt.Sprintf("metadata does not agree with track name %q", t.name))
	}
	if s.HasAlbumNameConflict() {
		diffs = append(diffs,
			fmt.Sprintf("metadata does not agree with album name %q", t.album.canonicalTitle))
	}
	if s.HasArtistNameConflict() {
		diffs = append(diffs,
			fmt.Sprintf("metadata does not agree with artist name %q", t.album.artist.canonicalName))
	}
	if s.HasGenreConflict() {
		diffs = append(diffs,
			fmt.Sprintf("metadata does not agree with album genre %q", t.album.canonicalGenre))
	}
	if s.HasYearConflict() {
		diffs = append(diffs,
			fmt.Sprintf("metadata does not agree with album year %q", t.album.canonicalYear))
	}
	if s.HasMCDIConflict() {
		diffs = append(diffs,
			fmt.Sprintf("metadata does not agree with the MCDI frame %q", string(t.album.musicCDIdentifier.Body)))
	}
	sort.Strings(diffs)
	return diffs
}

// UpdateMetadata verifies that a track's metadata needs to be edited and then
// performs that work
func (t *Track) UpdateMetadata() (e []error) {
	if !t.ReconcileMetadata().HasConflicts() {
		e = append(e, ErrNoEditNeeded)
	} else {
		e = append(e, updateMetadata(t.metadata, t.fullPath)...)
	}
	return
}

// use of semaphores nicely documented here:
// https://gist.github.com/repejota/ed9070d57c23102d50c94e1a126b2f5b

type empty struct{}

func (t *Track) LoadMetadata(bar *pb.ProgressBar) {
	if t.NeedsMetadata() {
		openFiles <- empty{} // block while full
		go func() {
			defer func() {
				bar.Increment()
				<-openFiles // read to release a slot
			}()
			t.SetMetadata(ReadRawMetadata(t.fullPath))
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
	bar := pb.New(count).SetWriter(GetBestWriter(o)).SetTemplateString(t).Start()
	for _, artist := range artists {
		for _, album := range artist.Albums() {
			for _, track := range album.Tracks() {
				track.LoadMetadata(bar)
			}
		}
	}
	WaitForFilesClosed()
	bar.Finish()
	ProcessAlbumMetadata(o, artists)
	ProcessArtistMetadata(o, artists)
	reportAllTrackErrors(o, artists)
}

func GetBestWriter(o output.Bus) io.Writer {
	// preferred: error output, then console output, then no output at all
	switch {
	case o.IsErrorTTY():
		return o.ErrorWriter()
	case o.IsConsoleTTY():
		return o.ConsoleWriter()
	default:
		return output.NilWriter{}
	}
}

func ProcessArtistMetadata(o output.Bus, artists []*Artist) {
	for _, artist := range artists {
		recordedArtistNames := make(map[string]int)
		for _, album := range artist.Albums() {
			for _, track := range album.Tracks() {
				if track.metadata != nil && track.metadata.IsValid() && track.metadata.CanonicalArtistNameMatches(artist.fileName) {
					recordedArtistNames[track.metadata.CanonicalArtist()]++
				}
			}
		}
		if canonicalName, ok := CanonicalChoice(recordedArtistNames); !ok {
			reportAmbiguousChoices(o, "artist name", artist.Name(), recordedArtistNames)
			logAmbiguousValue(o, map[string]any{
				"field":      "artist name",
				"settings":   recordedArtistNames,
				"artistName": artist.Name(),
			})
		} else if canonicalName != "" {
			artist.canonicalName = canonicalName
		}
	}
}

func reportAmbiguousChoices(o output.Bus, subject, context string, choices map[string]int) {
	o.WriteCanonicalError("There are multiple %s fields for %q, and there is no unambiguously preferred choice; candidates are %v", subject, context, encodeChoices(choices))
}

func logAmbiguousValue(o output.Bus, m map[string]any) {
	o.Log(output.Error, "no value has a majority of instances", m)
}

func ProcessAlbumMetadata(o output.Bus, artists []*Artist) {
	for _, ar := range artists {
		for _, al := range ar.Albums() {
			recordedMCDIs := make(map[string]int)
			recordedMCDIFrames := make(map[string]id3v2.UnknownFrame)
			recordedGenres := make(map[string]int)
			recordedYears := make(map[string]int)
			recordedAlbumTitles := make(map[string]int)
			for _, t := range al.Tracks() {
				if t.metadata == nil || !t.metadata.IsValid() {
					continue
				}
				genre := strings.ToLower(t.metadata.CanonicalGenre())
				if genre != "" && !strings.HasPrefix(genre, "unknown") {
					recordedGenres[t.metadata.CanonicalGenre()]++
				}
				if t.metadata.CanonicalYear() != "" {
					recordedYears[t.metadata.CanonicalYear()]++
				}
				if t.metadata.CanonicalAlbumTitleMatches(al.title) {
					recordedAlbumTitles[t.metadata.CanonicalAlbum()]++
				}
				mcdiKey := string(t.metadata.CanonicalMusicCDIdentifier().Body)
				recordedMCDIs[mcdiKey]++
				recordedMCDIFrames[mcdiKey] = t.metadata.CanonicalMusicCDIdentifier()
			}
			if canonicalGenre, ok := CanonicalChoice(recordedGenres); !ok {
				reportAmbiguousChoices(o, "genre", fmt.Sprintf("%s by %s", al.Name(), ar.Name()), recordedGenres)
				logAmbiguousValue(o, map[string]any{
					"field":      "genre",
					"settings":   recordedGenres,
					"albumName":  al.Name(),
					"artistName": ar.Name(),
				})
			} else {
				al.canonicalGenre = canonicalGenre
			}
			if canonicalYear, ok := CanonicalChoice(recordedYears); !ok {
				reportAmbiguousChoices(o, "year", fmt.Sprintf("%s by %s", al.Name(), ar.Name()), recordedYears)
				logAmbiguousValue(o, map[string]any{
					"field":      "year",
					"settings":   recordedYears,
					"albumName":  al.Name(),
					"artistName": ar.Name(),
				})
			} else {
				al.canonicalYear = canonicalYear
			}
			if canonicalAlbumTitle, ok := CanonicalChoice(recordedAlbumTitles); !ok {
				reportAmbiguousChoices(o, "album title", fmt.Sprintf("%s by %s", al.Name(), ar.Name()), recordedAlbumTitles)
				logAmbiguousValue(o, map[string]any{
					"field":      "album title",
					"settings":   recordedAlbumTitles,
					"albumName":  al.Name(),
					"artistName": ar.Name(),
				})
			} else if canonicalAlbumTitle != "" {
				al.canonicalTitle = canonicalAlbumTitle
			}
			if canonicalMCDI, ok := CanonicalChoice(recordedMCDIs); !ok {
				reportAmbiguousChoices(o, "MCDI frame", fmt.Sprintf("%s by %s", al.Name(), ar.Name()), recordedMCDIs)
				logAmbiguousValue(o, map[string]any{
					"field":      "mcdi frame",
					"settings":   recordedMCDIs,
					"albumName":  al.Name(),
					"artistName": ar.Name(),
				})
			} else {
				al.musicCDIdentifier = recordedMCDIFrames[canonicalMCDI]
			}
		}
	}
}

func encodeChoices(m map[string]int) string {
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

func CanonicalChoice(m map[string]int) (s string, ok bool) {
	if len(m) == 0 {
		ok = true
		return
	}
	total := 0
	for _, v := range m {
		total += v
	}
	// add up the total votes, divide by 2, force rounding up
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

// ReportMetadataReadError outputs a problem reading the metadata as an error
// and as a log record
func (t *Track) ReportMetadataReadError(o output.Bus, sT SourceType, e string) {
	name := sT.Name()
	o.Log(output.Error, "metadata read error", map[string]any{
		"metadata": name,
		"track":    t.String(),
		"error":    e,
	})
}

func reportAllTrackErrors(o output.Bus, artists []*Artist) {
	for _, ar := range artists {
		for _, al := range ar.Albums() {
			for _, t := range al.Tracks() {
				t.ReportMetadataErrors(o)
			}
		}
	}
}

func (t *Track) ReportMetadataErrors(o output.Bus) {
	if t.HasMetadataError() {
		for _, sT := range []SourceType{ID3V1, ID3V2} {
			if metadata := t.metadata; metadata != nil {
				if e := metadata.ErrCause[sT]; e != "" {
					t.ReportMetadataReadError(o, sT, e)
				}
			}
		}
	}
}

func WaitForFilesClosed() {
	for len(openFiles) != 0 {
		time.Sleep(1 * time.Microsecond)
	}
}

func ParseTrackName(o output.Bus, name string, album *Album, ext string) (commonName string, trackNumber int, valid bool) {
	if !trackNameRegex.MatchString(name) {
		o.Log(output.Error, "the track name cannot be parsed", map[string]any{
			"trackName":  name,
			"albumName":  album.title,
			"artistName": album.RecordingArtistName(),
		})
		o.WriteCanonicalError("The track %q on album %q by artist %q cannot be parsed", name, album.title, album.RecordingArtistName())
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
			commonName = strings.TrimSuffix(string(runes[i:]), ext)
			break
		}
	}
	valid = true
	return
}

// AlbumPath returns the path of the track's album.
func (t *Track) AlbumPath() string {
	if t.album == nil {
		return ""
	}
	return t.album.Path()
}

// AlbumName returns the name of the track's album.
func (t *Track) AlbumName() string {
	if t.album == nil {
		return ""
	}
	return t.album.title
}

// RecordingArtist returns the name of the artist on whose album this track
// appears.
func (t *Track) RecordingArtist() string {
	if t.album == nil {
		return ""
	}
	return t.album.RecordingArtistName()
}

// CopyFile copies the track file to a specified destination path.
func (t *Track) CopyFile(destination string) error {
	return cmd_toolkit.CopyFile(t.fullPath, destination)
}

// ID3V1Diagnostics returns the ID3V1 tag contents, if any; a missing ID3V1 tag
// (e.g., the input file is too short to have an ID3V1 tag), or an invalid ID3V1
// tag (IsValid() is false), returns a non-nil error
func (t *Track) ID3V1Diagnostics() ([]string, error) {
	return ReadID3v1Metadata(t.fullPath)
}

// ID3V2Diagnostics returns ID3V2 tag data - the ID3V2 version, its encoding,
// and a slice of all the frames in the tag.
func (t *Track) ID3V2Diagnostics() (version byte, encoding string, frames []string, e error) {
	version, encoding, frames, _, e = ReadID3V2Metadata(t.fullPath)
	return
}

// Details returns relevant details about the track
func (t *Track) Details() (map[string]string, error) {
	if _, _, _, frames, err := ReadID3V2Metadata(t.fullPath); err != nil {
		return nil, err
	} else {
		m := map[string]string{}
		// only include known frames
		for _, frame := range frames {
			if value, ok := frameDescriptions[frame.Name]; ok {
				m[value] = frame.Value
			}
		}
		return m, nil
	}
}
