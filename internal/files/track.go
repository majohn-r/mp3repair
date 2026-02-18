/*
Copyright © 2026 Marc Johnson (marc.johnson27591@gmail.com)
*/
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
	"github.com/majohn-r/output"
)

const (
	rawExtension            = "mp3"
	defaultTrackNamePattern = "^\\d+[\\s-].+\\." + rawExtension + "$"

	mcdiFrame  = "MCDI"
	trackFrame = "TRCK"
)

var (
	frameDescriptions = map[string]string{
		// list is from https://id3.org/id3v2.3.0#Declared_ID3v2_frames
		"AENC": "Audio encryption",
		"APIC": "Attached picture",
		"COMM": "Comments",
		"COMR": "Commercial frame",
		"ENCR": "Encryption method registration",
		"EQUA": "Equalization",
		"ETCO": "Event timing codes",
		"GEOB": "General encapsulated object",
		"GRID": "Group identification registration",
		"IPLS": "Involved people list",
		"LINK": "Linked information",
		"MCDI": "Music CD identifier",
		"MLLT": "MPEG location lookup table",
		"OWNE": "Ownership frame",
		"PRIV": "Private frame",
		"PCNT": "Play counter",
		"POPM": "Popularimeter",
		"POSS": "Position synchronisation frame",
		"RBUF": "Recommended buffer size",
		"RVAD": "Relative volume adjustment",
		"RVRB": "Reverb",
		"SYLT": "Synchronized lyric/text",
		"SYTC": "Synchronized tempo codes",
		"TALB": "Album/Movie/Show title",
		"TBPM": "BPM (beats per minute)",
		"TCOM": "Composer",
		"TCON": "Content type (genre)",
		"TCOP": "Copyright message",
		"TDAT": "Date",
		"TDLY": "Playlist delay",
		"TENC": "Encoded by",
		"TEXT": "Lyricist/Text writer",
		"TFLT": "File type",
		"TIME": "Time",
		"TIT1": "Content group description",
		"TIT2": "Title/songname/content description",
		"TIT3": "Subtitle/Description refinement",
		"TKEY": "Initial key",
		"TLAN": "Language(s)",
		"TLEN": "Length",
		"TMED": "Media type",
		"TOAL": "Original album/movie/show title",
		"TOFN": "Original filename",
		"TOLY": "Original lyricist(s)/text writer(s)",
		"TOPE": "Original artist(s)/performer(s)",
		"TORY": "Original release year",
		"TOWN": "File owner/licensee",
		"TPE1": "Lead performer(s)/Soloist(s)",
		"TPE2": "Band/orchestra/accompaniment",
		"TPE3": "Conductor/performer refinement",
		"TPE4": "Interpreted, remixed, or otherwise modified by",
		"TPOS": "Part of a set",
		"TPUB": "Publisher",
		"TRCK": "Track number/Position in set",
		"TRDA": "Recording dates",
		"TRSN": "Internet radio station name",
		"TRSO": "Internet radio station owner",
		"TSIZ": "Size (bytes)",
		"TSRC": "ISRC (international standard recording code)",
		"TSSE": "Software/Hardware and settings used for encoding",
		"TYER": "Year",
		"TXXX": "User defined text information frame",
		"UFID": "Unique file identifier",
		"USER": "Terms of use",
		"USLT": "Unsychronized lyric/text transcription",
		"WCOM": "Commercial information",
		"WCOP": "Copyright/Legal information",
		"WOAF": "Official audio file webpage",
		"WOAR": "Official artist/performer webpage",
		"WOAS": "Official audio source webpage",
		"WORS": "Official internet radio station homepage",
		"WPAY": "Payment",
		"WPUB": "Publishers official webpage",
		"WXXX": "User defined URL link frame",
	}
	errNoEditNeeded = fmt.Errorf("no edit required")
	trackNameRegex  = regexp.MustCompile(defaultTrackNamePattern)
)

// Track encapsulates data about a track on an album.
type Track struct {
	album *Album
	// full path to the file associated with the track, including the file itself
	filePath string
	// read from the file only when needed; file i/o is expensive
	metadata *TrackMetadata
	// name of the track, without its number or file extension, e.g., "First Track"
	simpleName string
	// number of the track
	number int
}

// FrameDescription returns a description of a frame based on the frame's name
func FrameDescription(name string) string {
	if description, descriptionFound := frameDescriptions[name]; descriptionFound {
		return description
	}
	return "No description found"
}

// Number returns the track's number
func (t *Track) Number() int { return t.number }

// Name returns the track's name; contrasted with the track's file name, this name does not
// include the track number or the file extension
func (t *Track) Name() string { return t.simpleName }

// Path returns the track's full file path, including the track file name
func (t *Track) Path() string { return t.filePath }

// SortTracks sorts a slice of *Track
func SortTracks(tracks []*Track) {
	sort.Slice(tracks, func(i, j int) bool {
		if tracks[i].simpleName == tracks[j].simpleName {
			album1 := tracks[i].album
			album2 := tracks[j].album
			if album1.title == album2.title {
				return album1.RecordingArtistName() < album2.RecordingArtistName()
			}
			return album1.title < album2.title
		}
		return tracks[i].simpleName < tracks[j].simpleName
	})
}

// String returns the track's full path (implementation of Stringer interface).
func (t *Track) String() string {
	return t.filePath
}

// Directory returns the directory containing the track file - in other words,
// its Album directory
func (t *Track) Directory() string {
	return filepath.Dir(t.filePath)
}

// FileName returns the track's full file name, minus its containing directory.
func (t *Track) FileName() string {
	return filepath.Base(t.filePath)
}

// Copy copies a track and optionally associates the copy with a new album
func (t *Track) Copy(a *Album, addToAlbum bool) *Track {
	t2 := &Track{
		filePath:   t.filePath,
		simpleName: t.simpleName,
		number:     t.number,
		metadata:   t.metadata,
		album:      a, // do not use source track's album!
	}
	if addToAlbum {
		a.addTrack(t2)
	}
	return t2
}

type TrackMaker struct {
	Album      *Album
	FileName   string // just the name of the track file, no parent directories
	SimpleName string // the track's name minus its extension and track number
	Number     int
	Metadata   *TrackMetadata
}

// NewTrack instantiates a new Track and optionally associates it with its album
func (ti TrackMaker) NewTrack(addToAlbum bool) *Track {
	t := &Track{
		filePath:   ti.Album.subDirectory(ti.FileName),
		simpleName: ti.SimpleName,
		number:     ti.Number,
		album:      ti.Album,
		metadata:   ti.Metadata,
	}
	if addToAlbum {
		ti.Album.addTrack(t)
	}
	return t
}

func (t *Track) needsMetadata() bool {
	return t.metadata == nil
}

func (t *Track) hasMetadataError() bool {
	return t.metadata != nil && len(t.metadata.errorCauses()) != 0
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

func (m MetadataState) hasConflicts() bool {
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
	if t.metadata == nil {
		return MetadataState{noMetadata: true}
	}
	mS := MetadataState{}
	metadataErrors := t.metadata.errorCauses()
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
	if !t.metadata.IsValid() {
		mS.corruptMetadata = true
		return mS
	}
	mS.numberingConflict = t.metadata.trackNumberDiffers(t.number)
	mS.trackNameConflict = t.metadata.trackNameDiffers(t.simpleName)
	mS.albumNameConflict = t.metadata.albumNameDiffers(t.album.canonicalTitle)
	mS.artistNameConflict = t.metadata.artistNameDiffers(t.album.recordingArtist.canonicalName())
	mS.genreConflict = t.metadata.albumGenreDiffers(t.album.genre)
	mS.yearConflict = t.metadata.albumYearDiffers(t.album.year)
	mS.mcdiConflict = t.metadata.cdIdentifierDiffers(t.album.cdIdentifier)
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
	if !s.hasConflicts() {
		return nil
	}
	// 13: 2 each for
	// - track numbering conflict
	// - track name conflict
	// - album name conflict
	// - artist name conflict
	// - album year conflict
	// - album genre conflict
	// and 1 for
	// - MCDI conflict
	diffs := make([]string, 0, 13)
	if s.HasNumberingConflict() {
		for _, src := range sourceTypes {
			if t.metadata.trackNumber(src).differenceExists {
				diffs = append(diffs,
					fmt.Sprintf("%s metadata [%v] does not agree with track number %d", src.String(),
						t.metadata.trackNumber(src).original, t.number))
			}
		}
	}
	if s.HasTrackNameConflict() {
		for _, src := range sourceTypes {
			if t.metadata.trackName(src).differenceExists {
				diffs = append(diffs,
					fmt.Sprintf("%s metadata [%v] does not agree with track name %q", src.String(),
						t.metadata.trackName(src).original, t.simpleName))
			}
		}
	}
	if s.HasAlbumNameConflict() {
		for _, src := range sourceTypes {
			if t.metadata.albumName(src).differenceExists {
				diffs = append(diffs,
					fmt.Sprintf("%s metadata [%v] does not agree with album name %q", src.String(),
						t.metadata.albumName(src).original, t.album.canonicalTitle))
			}
		}
	}
	if s.HasArtistNameConflict() {
		for _, src := range sourceTypes {
			if t.metadata.artistName(src).differenceExists {
				diffs = append(diffs,
					fmt.Sprintf("%s metadata [%v] does not agree with artist name %q", src.String(),
						t.metadata.artistName(src).original, t.album.recordingArtist.canonicalName()))
			}
		}
	}
	if s.HasGenreConflict() {
		for _, src := range sourceTypes {
			if t.metadata.albumGenre(src).differenceExists {
				diffs = append(diffs,
					fmt.Sprintf("%s metadata [%v] does not agree with album genre %q", src.String(),
						t.metadata.albumGenre(src).original, t.album.genre))
			}
		}
	}
	if s.HasYearConflict() {
		for _, src := range sourceTypes {
			if t.metadata.albumYear(src).differenceExists {
				diffs = append(diffs,
					fmt.Sprintf("%s metadata [%v] does not agree with album year %q", src.String(),
						t.metadata.albumYear(src).original, t.album.year))
			}
		}
	}
	if s.HasMCDIConflict() {
		diffs = append(diffs,
			fmt.Sprintf("ID3V2 metadata [%v] does not agree with the MCDI frame %q",
				t.metadata.cdIdentifier().original.Body, string(t.album.cdIdentifier.Body)))
	}
	sort.Strings(diffs)
	return diffs
}

// UpdateMetadata verifies that a track's metadata needs to be edited and then
// performs that work
func (t *Track) UpdateMetadata() (e []error) {
	if !t.ReconcileMetadata().hasConflicts() {
		e = append(e, errNoEditNeeded)
		return
	}
	e = append(e, t.metadata.update(t.filePath)...)
	return
}

// use of semaphores nicely documented here:
// https://gist.github.com/repejota/ed9070d57c23102d50c94e1a126b2f5b

type empty struct{}

func (t *Track) loadMetadata(openFiles chan empty, bar *pb.ProgressBar) {
	if t.needsMetadata() {
		openFiles <- empty{} // block while full
		go func() {
			defer func() {
				bar.Increment()
				<-openFiles // read to release a slot
			}()
			t.metadata = initializeMetadata(t.filePath)
		}()
	}
}

// ReadMetadata reads the metadata for all the artists' tracks.
func ReadMetadata(o output.Bus, artists []*Artist, fileLimit int) {
	// count the tracks
	count := 0
	for _, artist := range artists {
		for _, album := range artist.Albums() {
			count += len(album.tracks)
		}
	}
	o.ErrorPrintln("Reading track metadata.")
	openFiles := make(chan empty, fileLimit)
	// derived from the Default ProgressBarTemplate used by the progress bar,
	// following guidance in the ElementSpeed definition to change the output to
	// display the speed in tracks per second
	t := `{{with string . "prefix"}}{{.}} {{end}}{{counters . }} {{bar . }}` +
		` {{percent . }} {{speed . "%s tracks per second"}}{{with string . "suffix"}}` +
		` {{.}}{{end}}`
	bar := pb.New(count).SetWriter(progressWriter(o)).SetTemplateString(t).Start()
	for _, artist := range artists {
		for _, album := range artist.Albums() {
			for _, track := range album.tracks {
				track.loadMetadata(openFiles, bar)
			}
		}
	}
	waitForFilesClosed(openFiles)
	bar.Finish()
	processAlbumMetadata(o, artists)
	processArtistMetadata(o, artists)
	reportAllTrackErrors(o, artists)
}

func progressWriter(o output.Bus) io.Writer {
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

func processArtistMetadata(o output.Bus, artists []*Artist) {
	for _, artist := range artists {
		recordedArtistNames := make(map[string]int)
		for _, album := range artist.Albums() {
			for _, track := range album.tracks {
				if track.metadata != nil && track.metadata.IsValid() &&
					track.metadata.canonicalArtistNameMatches(artist.Name()) {
					recordedArtistNames[track.metadata.canonicalArtistName()]++
				}
			}
		}
		canonicalName, choiceSelected := canonicalChoice(recordedArtistNames)
		if !choiceSelected {
			reportAmbiguousChoices(o, "artist name", artist.Name(), recordedArtistNames)
			logAmbiguousValue(o, map[string]any{
				"field":      "artist name",
				"settings":   recordedArtistNames,
				"artistName": artist.Name(),
			})
			continue
		}
		if canonicalName != "" {
			artist.sharedName = canonicalName
		}
	}
}

func reportAmbiguousChoices(o output.Bus, subject, context string, choices map[string]int) {
	o.ErrorPrintf(
		"There are multiple %s fields for %q, and there is no unambiguously preferred choice; candidates are %v.\n",
		subject,
		context,
		encodeChoices(choices),
	)
}

func logAmbiguousValue(o output.Bus, m map[string]any) {
	o.Log(output.Error, "no value has a majority of instances", m)
}

func processAlbumMetadata(o output.Bus, artists []*Artist) {
	for _, ar := range artists {
		for _, al := range ar.Albums() {
			recordedMCDIs := make(map[string]int)
			recordedMCDIFrames := make(map[string]id3v2.UnknownFrame)
			recordedGenres := make(map[string]int)
			recordedYears := make(map[string]int)
			recordedAlbumTitles := make(map[string]int)
			for _, t := range al.tracks {
				if t.metadata == nil || !t.metadata.IsValid() {
					continue
				}
				genre := strings.ToLower(t.metadata.canonicalAlbumGenre())
				if genre != "" && !strings.HasPrefix(genre, "unknown") {
					recordedGenres[t.metadata.canonicalAlbumGenre()]++
				}
				if t.metadata.canonicalAlbumYear() != "" {
					recordedYears[t.metadata.canonicalAlbumYear()]++
				}
				if t.metadata.canonicalAlbumNameMatches(al.title) {
					recordedAlbumTitles[t.metadata.canonicalAlbumName()]++
				}
				mcdiKey := string(t.metadata.canonicalCDIdentifier().Body)
				recordedMCDIs[mcdiKey]++
				recordedMCDIFrames[mcdiKey] = t.metadata.canonicalCDIdentifier()
			}
			canonicalGenre, genreSelected := canonicalChoice(recordedGenres)
			switch {
			case genreSelected:
				al.genre = canonicalGenre
			default:
				reportAmbiguousChoices(o, "genre",
					fmt.Sprintf("%s by %s", al.title, ar.Name()), recordedGenres)
				logAmbiguousValue(o, map[string]any{
					"field":      "genre",
					"settings":   recordedGenres,
					"albumName":  al.title,
					"artistName": ar.Name(),
				})
			}
			canonicalYear, yearSelected := canonicalChoice(recordedYears)
			switch {
			case yearSelected:
				al.year = canonicalYear
			default:
				reportAmbiguousChoices(o, "year",
					fmt.Sprintf("%s by %s", al.title, ar.Name()), recordedYears)
				logAmbiguousValue(o, map[string]any{
					"field":      "year",
					"settings":   recordedYears,
					"albumName":  al.title,
					"artistName": ar.Name(),
				})
			}
			canonicalAlbumTitle, albumTitleSelected := canonicalChoice(recordedAlbumTitles)
			switch {
			case albumTitleSelected:
				if canonicalAlbumTitle != "" {
					al.canonicalTitle = canonicalAlbumTitle
				}
			default:
				reportAmbiguousChoices(o, "album title",
					fmt.Sprintf("%s by %s", al.title, ar.Name()), recordedAlbumTitles)
				logAmbiguousValue(o, map[string]any{
					"field":      "album title",
					"settings":   recordedAlbumTitles,
					"albumName":  al.title,
					"artistName": ar.Name(),
				})
			}
			canonicalMCDI, MCDISelected := canonicalChoice(recordedMCDIs)
			switch {
			case MCDISelected:
				al.cdIdentifier = recordedMCDIFrames[canonicalMCDI]
			default:
				reportAmbiguousChoices(o, "MCDI frame",
					fmt.Sprintf("%s by %s", al.title, ar.Name()), recordedMCDIs)
				logAmbiguousValue(o, map[string]any{
					"field":      "mcdi frame",
					"settings":   recordedMCDIs,
					"albumName":  al.title,
					"artistName": ar.Name(),
				})
			}
		}
	}
}

func encodeChoices(m map[string]int) string {
	values := make([]string, 0, len(m))
	for k, count := range m {
		switch count {
		case 1:
			values = append(values, fmt.Sprintf("%q: 1 instance", k))
		default:
			values = append(values, fmt.Sprintf("%q: %d instances", k, count))
		}
	}
	sort.Strings(values)
	return fmt.Sprintf("{%s}", strings.Join(values, ", "))
}

func canonicalChoice(m map[string]int) (value string, selected bool) {
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
func (t *Track) ReportMetadataReadError(o output.Bus, sT sourceType, e string) {
	name := sT.name()
	o.Log(output.Error, "metadata read error", map[string]any{
		"metadata": name,
		"track":    t.String(),
		"error":    e,
	})
}

func reportAllTrackErrors(o output.Bus, artists []*Artist) {
	for _, ar := range artists {
		for _, al := range ar.Albums() {
			for _, t := range al.tracks {
				t.reportMetadataErrors(o)
			}
		}
	}
}

func (t *Track) reportMetadataErrors(o output.Bus) {
	if t.hasMetadataError() {
		for _, src := range sourceTypes {
			if metadata := t.metadata; metadata != nil {
				if e := metadata.errorCause(src); e != "" {
					t.ReportMetadataReadError(o, src, e)
				}
			}
		}
	}
}

func waitForFilesClosed(openFiles chan empty) {
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
			"albumName":  parser.Album.title,
			"artistName": parser.Album.RecordingArtistName(),
		})
		o.ErrorPrintf(
			"The track %q on album %q by artist %q cannot be parsed.\n",
			parser.FileName,
			parser.Album.title,
			parser.Album.RecordingArtistName(),
		)
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

// ID3V1Diagnostics returns the ID3V1 tag contents, if any; a missing ID3V1 tag
// (e.g., the input file is too short to have an ID3V1 tag), or an invalid ID3V1
// tag (IsValid() is false), returns a non-nil error
func (t *Track) ID3V1Diagnostics() ([]string, error) {
	return readID3v1Metadata(t.filePath)
}

// ID3V2Diagnostics returns ID3V2 tag data - the ID3V2 version, its encoding,
// and a slice of all the frames in the tag.
func (t *Track) ID3V2Diagnostics() (*ID3V2Info, error) {
	return readID3V2Metadata(t.filePath)
}
