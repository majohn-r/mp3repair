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
	cmdtoolkit "github.com/majohn-r/cmd-toolkit"
	"github.com/majohn-r/output"
)

const (
	rawExtension            = "mp3"
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
	errNoEditNeeded = fmt.Errorf("no edit required")
	trackNameRegex  = regexp.MustCompile(defaultTrackNamePattern)
)

// Track encapsulates data about a track on an album.
type Track struct {
	Album *Album
	// full path to the file associated with the track, including the file itself
	FilePath string
	// read from the file only when needed; file i/o is expensive
	Metadata *TrackMetadata
	// name of the track, without its number or file extension, e.g., "First Track"
	SimpleName string
	// number of the track
	Number int
}

// String returns the track's full path (implementation of Stringer interface).
func (t *Track) String() string {
	return t.FilePath
}

// Directory returns the directory containing the track file - in other words,
// its Album directory
func (t *Track) Directory() string {
	return filepath.Dir(t.FilePath)
}

// FileName returns the track's full file name, minus its containing directory.
func (t *Track) FileName() string {
	return filepath.Base(t.FilePath)
}

func (t *Track) Copy(a *Album) *Track {
	return &Track{
		FilePath:   t.FilePath,
		SimpleName: t.SimpleName,
		Number:     t.Number,
		Metadata:   t.Metadata,
		Album:      a, // do not use source track's album!
	}
}

type TrackMaker struct {
	Album      *Album
	FileName   string // just the name of the track file, no parent directories
	SimpleName string // the track's name minus its extension and track number
	Number     int
}

func (ti TrackMaker) NewTrack() *Track {
	return &Track{
		FilePath:   ti.Album.subDirectory(ti.FileName),
		SimpleName: ti.SimpleName,
		Number:     ti.Number,
		Album:      ti.Album,
	}
}

type tracks []*Track

// Len returns the number of *Track instances.
func (ts tracks) Len() int {
	return len(ts)
}

// Less returns true if the first track's artist comes before the second track's
// artist. If the tracks are from the same artist, then it returns true if the
// first track's album comes before the second track's album. If the tracks come
// from the same artist and album, then it returns true if the first track's
// track number comes before the second track's track number.
func (ts tracks) Less(i, j int) bool {
	track1 := ts[i]
	track2 := ts[j]
	album1 := track1.Album
	album2 := track2.Album
	artist1 := album1.RecordingArtistName()
	artist2 := album2.RecordingArtistName()
	// compare artist name first
	if artist1 == artist2 {
		// artist names are the same ... try the album name next
		if album1.Title == album2.Title {
			// and album names are the same ... go by track number
			return track1.Number < track2.Number
		}
		return album1.Title < album2.Title
	}
	return artist1 < artist2
}

// Swap swaps two tracks.
func (ts tracks) Swap(i, j int) {
	ts[i], ts[j] = ts[j], ts[i]
}

func (t *Track) needsMetadata() bool {
	return t.Metadata == nil
}

func (t *Track) hasMetadataError() bool {
	return t.Metadata != nil && len(t.Metadata.ErrorCauses()) != 0
}

func (t *Track) SetMetadata(tm *TrackMetadata) {
	t.Metadata = tm
}

// MetadataState contains information about metadata problems
type MetadataState struct {
	// errors occurred reading both ID3V1 and ID3V2 metadata
	corruptMetadata bool
	// no attempt has been made to read metadata
	noMetadata bool
	// an attempt was made to read metadata, but there was no ID3V1 metadata found
	missingID3V1 bool
	// an attempt was made to read metadata, but there was no ID3V2 metadata found
	missingID3V2 bool
	// various conflicts
	numberingConflict  bool
	trackNameConflict  bool
	albumNameConflict  bool
	artistNameConflict bool
	genreConflict      bool
	yearConflict       bool
	mcdiConflict       bool
}

// HasNumberingConflict returns true if there is a conflict between the track
// number (as derived from the track's file name) and the value of the
// track's track number metadata.
func (m MetadataState) HasNumberingConflict() bool {
	return m.numberingConflict
}

// HasTrackNameConflict returns true if there is a conflict between the track
// name (as derived from the track's file name) and the value of the
// track's track name metadata.
func (m MetadataState) HasTrackNameConflict() bool {
	return m.trackNameConflict
}

// HasAlbumNameConflict returns true if there is a conflict between the name of
// the album the track is associated with and the value of the track's
// album name metadata.
func (m MetadataState) HasAlbumNameConflict() bool {
	return m.albumNameConflict
}

// HasArtistNameConflict returns true if there is a conflict between the track's
// recording artist and the value of the track's artist name metadata.
func (m MetadataState) HasArtistNameConflict() bool {
	return m.artistNameConflict
}

// HasConflicts returns true if there are any conflicts between the
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
// album's genre and the value of the track's genre metadata.
func (m MetadataState) HasGenreConflict() bool {
	return m.genreConflict
}

// HasYearConflict returns true if there is conflict between the track's album's
// year and the value of the track's year metadata.
func (m MetadataState) HasYearConflict() bool {
	return m.yearConflict
}

// ReconcileMetadata determines whether there are problems with the track's
// metadata.
func (t *Track) ReconcileMetadata() MetadataState {
	if t.Metadata == nil {
		return MetadataState{noMetadata: true}
	}
	mS := MetadataState{}
	metadataErrors := t.Metadata.ErrorCauses()
	if len(metadataErrors) != 0 {
		for _, e := range metadataErrors {
			switch e {
			case errNoID3V1MetadataFound.Error():
				mS.missingID3V1 = true
			case errNoID3V2MetadataFound.Error():
				mS.missingID3V2 = true
			}
		}
		if mS.missingID3V1 && mS.missingID3V2 {
			return mS
		}
	}
	if !t.Metadata.IsValid() {
		mS.corruptMetadata = true
		return mS
	}
	mS.numberingConflict = t.Metadata.TrackNumberDiffers(t.Number)
	mS.trackNameConflict = t.Metadata.TrackNameDiffers(t.SimpleName)
	mS.albumNameConflict = t.Metadata.AlbumNameDiffers(t.Album.CanonicalTitle)
	mS.artistNameConflict = t.Metadata.ArtistNameDiffers(t.Album.RecordingArtist.CanonicalName)
	mS.genreConflict = t.Metadata.AlbumGenreDiffers(t.Album.CanonicalGenre)
	mS.yearConflict = t.Metadata.AlbumYearDiffers(t.Album.CanonicalYear)
	mS.mcdiConflict = t.Metadata.CDIdentifierDiffers(t.Album.MusicCDIdentifier)
	return mS
}

// ReportMetadataProblems returns a slice of strings describing the problems
// found by calling ReconcileMetadata().
func (t *Track) ReportMetadataProblems() []string {
	s := t.ReconcileMetadata()
	if s.corruptMetadata {
		return []string{
			"differences cannot be determined: track metadata may be corrupted"}
	}
	if s.missingID3V1 && s.missingID3V2 {
		return []string{"differences cannot be determined: the track file contains no metadata"}
	}
	if s.noMetadata {
		return []string{"differences cannot be determined: metadata has not been read"}
	}
	if !s.HasConflicts() {
		return nil
	}
	// 7: 1 each for
	// - track numbering conflict
	// - track name conflict
	// - album name conflict
	// - artist name conflict
	// - album year conflict
	// - album genre conflict
	// - MCDI conflict
	diffs := make([]string, 0, 7)
	if s.HasNumberingConflict() {
		diffs = append(diffs,
			fmt.Sprintf("metadata does not agree with track number %d", t.Number))
	}
	if s.HasTrackNameConflict() {
		diffs = append(diffs,
			fmt.Sprintf("metadata does not agree with track name %q", t.SimpleName))
	}
	if s.HasAlbumNameConflict() {
		diffs = append(diffs,
			fmt.Sprintf("metadata does not agree with album name %q",
				t.Album.CanonicalTitle))
	}
	if s.HasArtistNameConflict() {
		diffs = append(diffs,
			fmt.Sprintf("metadata does not agree with artist name %q",
				t.Album.RecordingArtist.CanonicalName))
	}
	if s.HasGenreConflict() {
		diffs = append(diffs,
			fmt.Sprintf("metadata does not agree with album genre %q",
				t.Album.CanonicalGenre))
	}
	if s.HasYearConflict() {
		diffs = append(diffs,
			fmt.Sprintf("metadata does not agree with album year %q",
				t.Album.CanonicalYear))
	}
	if s.HasMCDIConflict() {
		diffs = append(diffs,
			fmt.Sprintf("metadata does not agree with the MCDI frame %q",
				string(t.Album.MusicCDIdentifier.Body)))
	}
	sort.Strings(diffs)
	return diffs
}

// UpdateMetadata verifies that a track's metadata needs to be edited and then
// performs that work
func (t *Track) UpdateMetadata() (e []error) {
	if !t.ReconcileMetadata().HasConflicts() {
		e = append(e, errNoEditNeeded)
		return
	}
	e = append(e, t.Metadata.Update(t.FilePath)...)
	return
}

// use of semaphores nicely documented here:
// https://gist.github.com/repejota/ed9070d57c23102d50c94e1a126b2f5b

type empty struct{}

func (t *Track) LoadMetadata(bar *pb.ProgressBar) {
	if t.needsMetadata() {
		openFiles <- empty{} // block while full
		go func() {
			defer func() {
				bar.Increment()
				<-openFiles // read to release a slot
			}()
			t.SetMetadata(InitializeMetadata(t.FilePath))
		}()
	}
}

// ReadMetadata reads the metadata for all the artists' tracks.
func ReadMetadata(o output.Bus, artists []*Artist) {
	// count the tracks
	count := 0
	for _, artist := range artists {
		for _, album := range artist.Albums {
			count += len(album.Tracks)
		}
	}
	o.WriteCanonicalError("Reading track metadata")
	// derived from the Default ProgressBarTemplate used by the progress bar,
	// following guidance in the ElementSpeed definition to change the output to
	// display the speed in tracks per second
	t := `{{with string . "prefix"}}{{.}} {{end}}{{counters . }} {{bar . }}` +
		` {{percent . }} {{speed . "%s tracks per second"}}{{with string . "suffix"}}` +
		` {{.}}{{end}}`
	bar := pb.New(count).SetWriter(ProgressWriter(o)).SetTemplateString(t).Start()
	for _, artist := range artists {
		for _, album := range artist.Albums {
			for _, track := range album.Tracks {
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

func ProgressWriter(o output.Bus) io.Writer {
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
		for _, album := range artist.Albums {
			for _, track := range album.Tracks {
				if track.Metadata != nil && track.Metadata.IsValid() &&
					track.Metadata.CanonicalArtistNameMatches(artist.Name) {
					recordedArtistNames[track.Metadata.CanonicalArtistName()]++
				}
			}
		}
		canonicalName, choiceSelected := CanonicalChoice(recordedArtistNames)
		if !choiceSelected {
			reportAmbiguousChoices(o, "artist name", artist.Name, recordedArtistNames)
			logAmbiguousValue(o, map[string]any{
				"field":      "artist name",
				"settings":   recordedArtistNames,
				"artistName": artist.Name,
			})
			continue
		}
		if canonicalName != "" {
			artist.CanonicalName = canonicalName
		}
	}
}

func reportAmbiguousChoices(o output.Bus, subject, context string, choices map[string]int) {
	o.WriteCanonicalError("There are multiple %s fields for %q,"+
		" and there is no unambiguously preferred choice; candidates are %v", subject, context,
		encodeChoices(choices))
}

func logAmbiguousValue(o output.Bus, m map[string]any) {
	o.Log(output.Error, "no value has a majority of instances", m)
}

func ProcessAlbumMetadata(o output.Bus, artists []*Artist) {
	for _, ar := range artists {
		for _, al := range ar.Albums {
			recordedMCDIs := make(map[string]int)
			recordedMCDIFrames := make(map[string]id3v2.UnknownFrame)
			recordedGenres := make(map[string]int)
			recordedYears := make(map[string]int)
			recordedAlbumTitles := make(map[string]int)
			for _, t := range al.Tracks {
				if t.Metadata == nil || !t.Metadata.IsValid() {
					continue
				}
				genre := strings.ToLower(t.Metadata.CanonicalAlbumGenre())
				if genre != "" && !strings.HasPrefix(genre, "unknown") {
					recordedGenres[t.Metadata.CanonicalAlbumGenre()]++
				}
				if t.Metadata.CanonicalAlbumYear() != "" {
					recordedYears[t.Metadata.CanonicalAlbumYear()]++
				}
				if t.Metadata.CanonicalAlbumNameMatches(al.Title) {
					recordedAlbumTitles[t.Metadata.CanonicalAlbumName()]++
				}
				mcdiKey := string(t.Metadata.CanonicalCDIdentifier().Body)
				recordedMCDIs[mcdiKey]++
				recordedMCDIFrames[mcdiKey] = t.Metadata.CanonicalCDIdentifier()
			}
			canonicalGenre, genreSelected := CanonicalChoice(recordedGenres)
			switch {
			case genreSelected:
				al.CanonicalGenre = canonicalGenre
			default:
				reportAmbiguousChoices(o, "genre",
					fmt.Sprintf("%s by %s", al.Title, ar.Name), recordedGenres)
				logAmbiguousValue(o, map[string]any{
					"field":      "genre",
					"settings":   recordedGenres,
					"albumName":  al.Title,
					"artistName": ar.Name,
				})
			}
			canonicalYear, yearSelected := CanonicalChoice(recordedYears)
			switch {
			case yearSelected:
				al.CanonicalYear = canonicalYear
			default:
				reportAmbiguousChoices(o, "year",
					fmt.Sprintf("%s by %s", al.Title, ar.Name), recordedYears)
				logAmbiguousValue(o, map[string]any{
					"field":      "year",
					"settings":   recordedYears,
					"albumName":  al.Title,
					"artistName": ar.Name,
				})
			}
			canonicalAlbumTitle, albumTitleSelected := CanonicalChoice(recordedAlbumTitles)
			switch {
			case albumTitleSelected:
				if canonicalAlbumTitle != "" {
					al.CanonicalTitle = canonicalAlbumTitle
				}
			default:
				reportAmbiguousChoices(o, "album title",
					fmt.Sprintf("%s by %s", al.Title, ar.Name), recordedAlbumTitles)
				logAmbiguousValue(o, map[string]any{
					"field":      "album title",
					"settings":   recordedAlbumTitles,
					"albumName":  al.Title,
					"artistName": ar.Name,
				})
			}
			canonicalMCDI, MCDISelected := CanonicalChoice(recordedMCDIs)
			switch {
			case MCDISelected:
				al.MusicCDIdentifier = recordedMCDIFrames[canonicalMCDI]
			default:
				reportAmbiguousChoices(o, "MCDI frame",
					fmt.Sprintf("%s by %s", al.Title, ar.Name), recordedMCDIs)
				logAmbiguousValue(o, map[string]any{
					"field":      "mcdi frame",
					"settings":   recordedMCDIs,
					"albumName":  al.Title,
					"artistName": ar.Name,
				})
			}
		}
	}
}

func encodeChoices(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	values := make([]string, 0, len(m))
	for _, k := range keys {
		count := m[k]
		switch count {
		case 1:
			values = append(values, fmt.Sprintf("%q: 1 instance", k))
		default:
			values = append(values, fmt.Sprintf("%q: %d instances", k, count))
		}
	}
	return fmt.Sprintf("{%s}", strings.Join(values, ", "))
}

func CanonicalChoice(m map[string]int) (value string, selected bool) {
	if len(m) == 0 {
		selected = true
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
			value = k
			selected = true
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
		for _, al := range ar.Albums {
			for _, t := range al.Tracks {
				t.ReportMetadataErrors(o)
			}
		}
	}
}

func (t *Track) ReportMetadataErrors(o output.Bus) {
	if t.hasMetadataError() {
		for _, src := range []SourceType{ID3V1, ID3V2} {
			if metadata := t.Metadata; metadata != nil {
				if e := metadata.ErrorCause(src); e != "" {
					t.ReportMetadataReadError(o, src, e)
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

type TrackNameParser struct {
	FileName  string
	Album     *Album
	Extension string
}

type ParsedTrackName struct {
	SimpleName string
	Number     int
}

func (parser TrackNameParser) Parse(o output.Bus) (*ParsedTrackName, bool) {
	if !trackNameRegex.MatchString(parser.FileName) {
		o.Log(output.Error, "the track name cannot be parsed", map[string]any{
			"trackName":  parser.FileName,
			"albumName":  parser.Album.Title,
			"artistName": parser.Album.RecordingArtistName(),
		})
		o.WriteCanonicalError("The track %q on album %q by artist %q cannot be parsed",
			parser.FileName, parser.Album.Title, parser.Album.RecordingArtistName())
		return nil, false
	}
	name := &ParsedTrackName{}
	wantDigit := true
	runes := []rune(parser.FileName)
	for i, r := range runes {
		if !wantDigit {
			name.SimpleName = strings.TrimSuffix(string(runes[i:]), parser.Extension)
			break
		}
		switch {
		case r >= '0' && r <= '9':
			name.Number *= 10
			name.Number += int(r - '0')
		default:
			wantDigit = false
		}
	}
	return name, true
}

// AlbumPath returns the path of the track's album.
func (t *Track) AlbumPath() string {
	if t.Album == nil {
		return ""
	}
	return t.Album.FilePath
}

// AlbumName returns the name of the track's album.
func (t *Track) AlbumName() string {
	if t.Album == nil {
		return ""
	}
	return t.Album.Title
}

// RecordingArtist returns the name of the artist on whose album this track
// appears.
func (t *Track) RecordingArtist() string {
	if t.Album == nil {
		return ""
	}
	return t.Album.RecordingArtistName()
}

// CopyFile copies the track file to a specified destination path.
func (t *Track) CopyFile(destination string) error {
	return cmdtoolkit.CopyFile(t.FilePath, destination)
}

// ID3V1Diagnostics returns the ID3V1 tag contents, if any; a missing ID3V1 tag
// (e.g., the input file is too short to have an ID3V1 tag), or an invalid ID3V1
// tag (IsValid() is false), returns a non-nil error
func (t *Track) ID3V1Diagnostics() ([]string, error) {
	return readID3v1Metadata(t.FilePath)
}

// ID3V2Diagnostics returns ID3V2 tag data - the ID3V2 version, its encoding,
// and a slice of all the frames in the tag.
func (t *Track) ID3V2Diagnostics() (*ID3V2Info, error) {
	return readID3V2Metadata(t.FilePath)
}

// Details returns relevant details about the track
func (t *Track) Details() (map[string]string, error) {
	info, readErr := readID3V2Metadata(t.FilePath)
	if readErr != nil {
		return nil, readErr
	}
	m := map[string]string{}
	// only include known frames
	for _, frame := range info.RawFrames {
		if value, descriptionFound := frameDescriptions[frame.name]; descriptionFound {
			m[value] = frame.value
		}
	}
	return m, nil
}
