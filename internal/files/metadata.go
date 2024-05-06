package files

import (
	"bytes"
	"strings"

	"github.com/bogem/id3v2/v2"
)

// SourceType identifies the source of a particular form of metadata
type SourceType int

const (
	UndefinedSource SourceType = iota
	ID3V1
	ID3V2
	TotalSources
)

var (
	nameComparators = map[SourceType]func(*ComparableStrings) bool{
		ID3V1: Id3v1NameDiffers,
		ID3V2: Id3v2NameDiffers,
	}
	genreComparators = map[SourceType]func(*ComparableStrings) bool{
		ID3V1: Id3v1GenreDiffers,
		ID3V2: Id3v2GenreDiffers,
	}
	metadataUpdaters = map[SourceType]func(tM *TrackMetadata, path string,
		src SourceType) error{
		ID3V1: updateID3V1Metadata,
		ID3V2: UpdateID3V2Metadata,
	}
	sourceTypes = []SourceType{ID3V1, ID3V2}
)

func (sT SourceType) Name() string {
	switch sT {
	case ID3V1:
		return "ID3V1"
	case ID3V2:
		return "ID3V2"
	case TotalSources:
		return "total"
	default:
		return "undefined"
	}
}

type TrackMetadata struct {
	albumName         []string
	artistName        []string
	primarySource     SourceType
	errorCause        []string
	genre             []string
	musicCDIdentifier id3v2.UnknownFrame
	trackName         []string
	trackNumber       []int
	year              []string
	// these fields are set by the various xDiffers methods
	correctedAlbumName         []string
	correctedArtistName        []string
	correctedGenre             []string
	correctedMusicCDIdentifier id3v2.UnknownFrame
	correctedTrackName         []string
	correctedTrackNumber       []int
	correctedYear              []string
	requiresEdit               []bool
}

func (tm *TrackMetadata) SetAlbumName(src SourceType, s string) {
	tm.albumName[src] = s
}

func (tm *TrackMetadata) SetArtistName(src SourceType, s string) {
	tm.artistName[src] = s
}

func (tm *TrackMetadata) SetErrorCause(src SourceType, s string) {
	tm.errorCause[src] = s
}

func (tm *TrackMetadata) SetGenre(src SourceType, s string) {
	tm.genre[src] = s
}

func (tm *TrackMetadata) SetTrackName(src SourceType, s string) {
	tm.trackName[src] = s
}

func (tm *TrackMetadata) SetTrackNumber(src SourceType, i int) {
	tm.trackNumber[src] = i
}

func (tm *TrackMetadata) SetYear(src SourceType, s string) {
	tm.year[src] = s
}

func (tm *TrackMetadata) WithAlbumNames(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.albumName[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithArtistNames(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.artistName[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithPrimarySource(t SourceType) *TrackMetadata {
	tm.primarySource = t
	return tm
}

func (tm *TrackMetadata) WithErrorCauses(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.errorCause[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithGenres(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.genre[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithMusicCDIdentifier(b []byte) *TrackMetadata {
	tm.musicCDIdentifier = id3v2.UnknownFrame{Body: b}
	return tm
}

func (tm *TrackMetadata) WithTrackNames(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.trackName[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithTrackNumbers(k []int) *TrackMetadata {
	for i := range min(len(k), int(TotalSources)) {
		tm.trackNumber[i] = k[i]
	}
	return tm
}

func (tm *TrackMetadata) WithYears(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.year[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithCorrectedAlbumNames(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.correctedAlbumName[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithCorrectedArtistNames(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.correctedArtistName[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithCorrectedGenres(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.correctedGenre[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithCorrectedMusicCDIdentifier(b []byte) *TrackMetadata {
	tm.correctedMusicCDIdentifier = id3v2.UnknownFrame{Body: b}
	return tm
}

func (tm *TrackMetadata) WithCorrectedTrackNames(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.correctedTrackName[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithCorrectedTrackNumbers(k []int) *TrackMetadata {
	for i := range min(len(k), int(TotalSources)) {
		tm.correctedTrackNumber[i] = k[i]
	}
	return tm
}

func (tm *TrackMetadata) WithCorrectedYears(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.correctedYear[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithRequiresEdits(b []bool) *TrackMetadata {
	for i := range min(len(b), int(TotalSources)) {
		tm.requiresEdit[i] = b[i]
	}
	return tm
}

func NewTrackMetadata() *TrackMetadata {
	return &TrackMetadata{
		albumName:            make([]string, TotalSources),
		artistName:           make([]string, TotalSources),
		trackName:            make([]string, TotalSources),
		genre:                make([]string, TotalSources),
		year:                 make([]string, TotalSources),
		trackNumber:          make([]int, TotalSources),
		errorCause:           make([]string, TotalSources),
		correctedAlbumName:   make([]string, TotalSources),
		correctedArtistName:  make([]string, TotalSources),
		correctedTrackName:   make([]string, TotalSources),
		correctedGenre:       make([]string, TotalSources),
		correctedYear:        make([]string, TotalSources),
		correctedTrackNumber: make([]int, TotalSources),
		requiresEdit:         make([]bool, TotalSources),
	}
}

func ReadRawMetadata(path string) *TrackMetadata {
	v1, id3v1Err := InternalReadID3V1Metadata(path, FileReader)
	d := RawReadID3V2Metadata(path)
	tM := NewTrackMetadata()
	switch {
	case id3v1Err != nil && d.err != nil:
		tM.errorCause[ID3V1] = id3v1Err.Error()
		tM.errorCause[ID3V2] = d.err.Error()
	case id3v1Err != nil:
		tM.errorCause[ID3V1] = id3v1Err.Error()
		tM.SetID3v2Values(d)
		tM.primarySource = ID3V2
	case d.err != nil:
		tM.errorCause[ID3V2] = d.err.Error()
		tM.SetID3v1Values(v1)
		tM.primarySource = ID3V1
	default:
		tM.SetID3v2Values(d)
		tM.SetID3v1Values(v1)
		tM.primarySource = ID3V2
	}
	return tM
}

func (tM *TrackMetadata) SetID3v2Values(d *Id3v2Metadata) {
	i := ID3V2
	tM.albumName[i] = d.albumName
	tM.artistName[i] = d.artistName
	tM.trackName[i] = d.trackName
	tM.genre[i] = d.genre
	tM.year[i] = d.year
	tM.trackNumber[i] = d.trackNumber
	tM.musicCDIdentifier = d.musicCDIdentifier
}

func (tM *TrackMetadata) SetID3v1Values(v1 *Id3v1Metadata) {
	index := ID3V1
	tM.albumName[index] = v1.Album()
	tM.artistName[index] = v1.Artist()
	tM.trackName[index] = v1.Title()
	if genre, ok := v1.Genre(); ok {
		tM.genre[index] = genre
	}
	tM.year[index] = v1.Year()
	if track, ok := v1.Track(); ok {
		tM.trackNumber[index] = track
	}
}

func (tM *TrackMetadata) IsValid() bool {
	return tM.primarySource == ID3V1 || tM.primarySource == ID3V2
}

func (tM *TrackMetadata) CanonicalArtist() string {
	return tM.artistName[tM.primarySource]
}

func (tM *TrackMetadata) CanonicalAlbum() string {
	return tM.albumName[tM.primarySource]
}

func (tM *TrackMetadata) CanonicalGenre() string {
	return tM.genre[tM.primarySource]
}

func (tM *TrackMetadata) CanonicalYear() string {
	return tM.year[tM.primarySource]
}

func (tM *TrackMetadata) CanonicalMusicCDIdentifier() id3v2.UnknownFrame {
	return tM.musicCDIdentifier
}

func (tM *TrackMetadata) ErrorCauses() []string {
	errCauses := make([]string, 0, len(tM.errorCause))
	for _, e := range tM.errorCause {
		if e != "" {
			errCauses = append(errCauses, e)
		}
	}
	return errCauses
}

type ComparableStrings struct {
	external string
	metadata string
}

func (cs *ComparableStrings) External() string {
	return cs.external
}

func (cs *ComparableStrings) Metadata() string {
	return cs.metadata
}

func (cs *ComparableStrings) WithMetadata(s string) *ComparableStrings {
	cs.metadata = s
	return cs
}

func (cs *ComparableStrings) WithExternal(s string) *ComparableStrings {
	cs.external = s
	return cs
}

func NewComparableStrings() *ComparableStrings {
	return &ComparableStrings{}
}

func (tM *TrackMetadata) TrackDiffers(track int) (differs bool) {
	for _, sT := range sourceTypes {
		if tM.errorCause[sT] == "" && tM.trackNumber[sT] != track {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedTrackNumber[sT] = track
		}
	}
	return
}

func (tM *TrackMetadata) TrackTitleDiffers(title string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := &ComparableStrings{external: title, metadata: tM.trackName[sT]}
		if tM.errorCause[sT] == "" && nameComparators[sT](comparison) {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedTrackName[sT] = title
		}
	}
	return
}

func (tM *TrackMetadata) AlbumTitleDiffers(albumTitle string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := &ComparableStrings{external: albumTitle, metadata: tM.albumName[sT]}
		if tM.errorCause[sT] == "" && nameComparators[sT](comparison) {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedAlbumName[sT] = albumTitle
		}
	}
	return
}

func (tM *TrackMetadata) ArtistNameDiffers(artistName string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := &ComparableStrings{external: artistName, metadata: tM.artistName[sT]}
		if tM.errorCause[sT] == "" && nameComparators[sT](comparison) {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedArtistName[sT] = artistName
		}
	}
	return
}

func (tM *TrackMetadata) GenreDiffers(genre string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := &ComparableStrings{external: genre, metadata: tM.genre[sT]}
		if tM.errorCause[sT] == "" && genreComparators[sT](comparison) {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedGenre[sT] = genre
		}
	}
	return
}

func (tM *TrackMetadata) YearDiffers(year string) (differs bool) {
	for _, sT := range sourceTypes {
		if tM.errorCause[sT] == "" && !YearsMatch(tM.year[sT], year) {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedYear[sT] = year
		}
	}
	return
}

// YearsMatch compares a year field, as recorded in metadata, against the
// album's canonical year; returns true if they are considered equal
func YearsMatch(metadataYear, albumYear string) bool {
	switch {
	case len(metadataYear) < len(albumYear):
		if metadataYear == "" {
			return false
		}
		return strings.HasPrefix(albumYear, metadataYear)
	case len(metadataYear) > len(albumYear):
		if albumYear == "" {
			return false
		}
		return strings.HasPrefix(metadataYear, albumYear)
	default:
		return metadataYear == albumYear
	}
}

func (tM *TrackMetadata) MCDIDiffers(f id3v2.UnknownFrame) (differs bool) {
	if tM.errorCause[ID3V2] == "" && !bytes.Equal(tM.musicCDIdentifier.Body, f.Body) {
		differs = true
		tM.requiresEdit[ID3V2] = true
		tM.correctedMusicCDIdentifier = f
	}
	return
}

func (tM *TrackMetadata) CanonicalAlbumTitleMatches(albumTitle string) bool {
	comparison := &ComparableStrings{external: albumTitle, metadata: tM.CanonicalAlbum()}
	return !nameComparators[tM.primarySource](comparison)
}

func (tM *TrackMetadata) CanonicalArtistNameMatches(artistName string) bool {
	comparison := &ComparableStrings{external: artistName, metadata: tM.CanonicalArtist()}
	return !nameComparators[tM.primarySource](comparison)
}

func updateMetadata(tM *TrackMetadata, path string) (e []error) {
	for _, source := range sourceTypes {
		if err := metadataUpdaters[source](tM, path, source); err != nil {
			e = append(e, err)
		}
	}
	return
}
