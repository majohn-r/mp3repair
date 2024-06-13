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
	trackMetadataUpdaters = map[SourceType]func(tm *TrackMetadata, path string) error{
		ID3V1: updateID3V1TrackMetadata,
		ID3V2: updateID3V2TrackMetadata,
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

type metadataValue interface {
	string | int | id3v2.UnknownFrame
}

type CorrectableValue[V metadataValue] struct {
	original   V
	correction V
}

func (cv CorrectableValue[V]) Original() V {
	return cv.original
}

func (cv CorrectableValue[V]) Correction() V {
	return cv.correction
}

type commonMetadata struct {
	artistName   CorrectableValue[string]
	albumName    CorrectableValue[string]
	albumGenre   CorrectableValue[string]
	albumYear    CorrectableValue[string]
	trackName    CorrectableValue[string]
	trackNumber  CorrectableValue[int]
	errorCause   string
	requiresEdit bool
}

type TrackMetadata struct {
	data              map[SourceType]*commonMetadata
	musicCDIdentifier CorrectableValue[id3v2.UnknownFrame]
	canonicalSource   SourceType
}

func NewTrackMetadata() *TrackMetadata {
	return &TrackMetadata{
		data:            map[SourceType]*commonMetadata{},
		canonicalSource: UndefinedSource,
	}
}

func isValidSource(src SourceType) bool {
	switch src {
	case ID3V1:
		return true
	case ID3V2:
		return true
	default:
		return false
	}
}

func (tm *TrackMetadata) commonMetadata(src SourceType) *commonMetadata {
	if !isValidSource(src) {
		return &commonMetadata{}
	}
	data, dataExists := tm.data[src]
	if !dataExists {
		data = &commonMetadata{}
		tm.data[src] = data
	}
	return data
}

func (tm *TrackMetadata) SetArtistName(src SourceType, name string) {
	tm.commonMetadata(src).artistName.original = name
}

func (tm *TrackMetadata) CorrectArtistName(src SourceType, name string) {
	tm.commonMetadata(src).artistName.correction = name
}

func (tm *TrackMetadata) ArtistName(src SourceType) CorrectableValue[string] {
	return tm.commonMetadata(src).artistName
}

func (tm *TrackMetadata) CanonicalArtistName() string {
	return tm.ArtistName(tm.canonicalSource).original
}

func (tm *TrackMetadata) ArtistNameDiffers(nameFromFile string) (differs bool) {
	for _, src := range sourceTypes {
		comparison := &ComparableStrings{
			External: nameFromFile,
			Metadata: tm.ArtistName(src).original,
		}
		if tm.ErrorCause(src) == "" && nameComparators[src](comparison) {
			differs = true
			tm.SetEditRequired(src)
			tm.CorrectArtistName(src, nameFromFile)
		}
	}
	return
}

func (tm *TrackMetadata) CanonicalArtistNameMatches(artistNameFromFile string) bool {
	comparison := &ComparableStrings{
		External: artistNameFromFile,
		Metadata: tm.CanonicalArtistName(),
	}
	comparator, exists := nameComparators[tm.canonicalSource]
	if !exists {
		return false
	}
	return !comparator(comparison)
}

func (tm *TrackMetadata) SetAlbumName(src SourceType, name string) {
	tm.commonMetadata(src).albumName.original = name
}

func (tm *TrackMetadata) CorrectAlbumName(src SourceType, name string) {
	tm.commonMetadata(src).albumName.correction = name
}

func (tm *TrackMetadata) AlbumName(src SourceType) CorrectableValue[string] {
	return tm.commonMetadata(src).albumName
}

func (tm *TrackMetadata) CanonicalAlbumName() string {
	return tm.AlbumName(tm.canonicalSource).original
}

func (tm *TrackMetadata) AlbumNameDiffers(nameFromFile string) (differs bool) {
	for _, src := range sourceTypes {
		comparison := &ComparableStrings{
			External: nameFromFile,
			Metadata: tm.AlbumName(src).original,
		}
		if tm.ErrorCause(src) == "" && nameComparators[src](comparison) {
			differs = true
			tm.SetEditRequired(src)
			tm.CorrectAlbumName(src, nameFromFile)
		}
	}
	return
}

func (tm *TrackMetadata) CanonicalAlbumNameMatches(nameFromFile string) bool {
	comparison := &ComparableStrings{
		External: nameFromFile,
		Metadata: tm.CanonicalAlbumName(),
	}
	comparator, exists := nameComparators[tm.canonicalSource]
	if !exists {
		return false
	}
	return !comparator(comparison)
}

func (tm *TrackMetadata) SetAlbumGenre(src SourceType, name string) {
	tm.commonMetadata(src).albumGenre.original = name
}

func (tm *TrackMetadata) CorrectAlbumGenre(src SourceType, name string) {
	tm.commonMetadata(src).albumGenre.correction = name
}

func (tm *TrackMetadata) AlbumGenre(src SourceType) CorrectableValue[string] {
	return tm.commonMetadata(src).albumGenre
}

func (tm *TrackMetadata) CanonicalAlbumGenre() string {
	return tm.AlbumGenre(tm.canonicalSource).original
}

func (tm *TrackMetadata) AlbumGenreDiffers(canonicalAlbumGenre string) (differs bool) {
	for _, src := range sourceTypes {
		comparison := &ComparableStrings{
			External: canonicalAlbumGenre,
			Metadata: tm.AlbumGenre(src).original,
		}
		if tm.ErrorCause(src) == "" && genreComparators[src](comparison) {
			differs = true
			tm.SetEditRequired(src)
			tm.CorrectAlbumGenre(src, canonicalAlbumGenre)
		}
	}
	return
}

func (tm *TrackMetadata) SetAlbumYear(src SourceType, name string) {
	tm.commonMetadata(src).albumYear.original = name
}

func (tm *TrackMetadata) CorrectAlbumYear(src SourceType, name string) {
	tm.commonMetadata(src).albumYear.correction = name
}

func (tm *TrackMetadata) AlbumYear(src SourceType) CorrectableValue[string] {
	return tm.commonMetadata(src).albumYear
}

func (tm *TrackMetadata) CanonicalAlbumYear() string {
	return tm.AlbumYear(tm.canonicalSource).original
}

func (tm *TrackMetadata) AlbumYearDiffers(canonicalAlbumYear string) (differs bool) {
	for _, src := range sourceTypes {
		if tm.ErrorCause(src) == "" && !YearsMatch(tm.AlbumYear(src).original, canonicalAlbumYear) {
			differs = true
			tm.SetEditRequired(src)
			tm.CorrectAlbumYear(src, canonicalAlbumYear)
		}
	}
	return
}

func (tm *TrackMetadata) SetTrackName(src SourceType, name string) {
	tm.commonMetadata(src).trackName.original = name
}

func (tm *TrackMetadata) CorrectTrackName(src SourceType, name string) {
	tm.commonMetadata(src).trackName.correction = name
}

func (tm *TrackMetadata) TrackName(src SourceType) CorrectableValue[string] {
	return tm.commonMetadata(src).trackName
}

func (tm *TrackMetadata) CanonicalTrackName() string {
	return tm.TrackName(tm.canonicalSource).original
}

func (tm *TrackMetadata) TrackNameDiffers(nameFromFile string) (differs bool) {
	for _, src := range sourceTypes {
		comparison := &ComparableStrings{
			External: nameFromFile,
			Metadata: tm.TrackName(src).original,
		}
		if tm.ErrorCause(src) == "" && nameComparators[src](comparison) {
			differs = true
			tm.SetEditRequired(src)
			tm.CorrectTrackName(src, nameFromFile)
		}
	}
	return
}

func (tm *TrackMetadata) SetTrackNumber(src SourceType, number int) {
	tm.commonMetadata(src).trackNumber.original = number
}

func (tm *TrackMetadata) CorrectTrackNumber(src SourceType, number int) {
	tm.commonMetadata(src).trackNumber.correction = number
}

func (tm *TrackMetadata) TrackNumber(src SourceType) CorrectableValue[int] {
	return tm.commonMetadata(src).trackNumber
}

func (tm *TrackMetadata) CanonicalTrackNumber() int {
	return tm.TrackNumber(tm.canonicalSource).original
}

// does the track number from metadata differ from the track number acquired
// from the track's file name? If so, make a note of it
func (tm *TrackMetadata) TrackNumberDiffers(trackNumberFromFileName int) (differs bool) {
	for _, src := range sourceTypes {
		if tm.ErrorCause(src) == "" && tm.TrackNumber(src).original != trackNumberFromFileName {
			differs = true
			tm.SetEditRequired(src)
			tm.CorrectTrackNumber(src, trackNumberFromFileName)
		}
	}
	return
}

func (tm *TrackMetadata) SetErrorCause(src SourceType, cause string) {
	tm.commonMetadata(src).errorCause = cause
}

func (tm *TrackMetadata) ErrorCause(src SourceType) string {
	return tm.commonMetadata(src).errorCause
}

func (tm *TrackMetadata) SetEditRequired(src SourceType) {
	tm.commonMetadata(src).requiresEdit = true
}

func (tm *TrackMetadata) EditRequired(src SourceType) bool {
	return tm.commonMetadata(src).requiresEdit
}

func (tm *TrackMetadata) SetCDIdentifier(body []byte) {
	tm.musicCDIdentifier.original = id3v2.UnknownFrame{Body: body}
}

func (tm *TrackMetadata) CorrectCDIdentifier(body []byte) {
	tm.musicCDIdentifier.correction = id3v2.UnknownFrame{Body: body}
}

func (tm *TrackMetadata) CDIdentifier() CorrectableValue[id3v2.UnknownFrame] {
	return tm.musicCDIdentifier
}

func (tm *TrackMetadata) CanonicalCDIdentifier() id3v2.UnknownFrame {
	return tm.musicCDIdentifier.original
}

func (tm *TrackMetadata) CDIdentifierDiffers(canonicalCDIdentifier id3v2.UnknownFrame) (differs bool) {
	if tm.ErrorCause(ID3V2) == "" && !bytes.Equal(tm.CDIdentifier().original.Body, canonicalCDIdentifier.Body) {
		differs = true
		tm.SetEditRequired(ID3V2)
		tm.CorrectCDIdentifier(canonicalCDIdentifier.Body)
	}
	return
}

func (tm *TrackMetadata) SetCanonicalSource(src SourceType) {
	if isValidSource(src) {
		tm.canonicalSource = src
	}
}

func (tm *TrackMetadata) CanonicalSource() SourceType {
	return tm.canonicalSource
}

func (tm *TrackMetadata) SetID3v2Values(d *Id3v2Metadata) {
	tm.SetArtistName(ID3V2, d.ArtistName)
	tm.SetAlbumName(ID3V2, d.AlbumTitle)
	tm.SetAlbumGenre(ID3V2, d.Genre)
	tm.SetAlbumYear(ID3V2, d.Year)
	tm.SetTrackName(ID3V2, d.TrackName)
	tm.SetTrackNumber(ID3V2, d.TrackNumber)
	tm.SetCDIdentifier(d.MusicCDIdentifier.Body)
}

func (tm *TrackMetadata) SetID3v1Values(v1 *Id3v1Metadata) {
	tm.SetArtistName(ID3V1, v1.Artist())
	tm.SetAlbumName(ID3V1, v1.Album())
	if genre, genreFound := v1.Genre(); genreFound {
		tm.SetAlbumGenre(ID3V1, genre)
	}
	tm.SetAlbumYear(ID3V1, v1.Year())
	tm.SetTrackName(ID3V1, v1.Title())
	if track, trackValid := v1.Track(); trackValid {
		tm.SetTrackNumber(ID3V1, track)
	}
}

func (tm *TrackMetadata) IsValid() bool {
	return isValidSource(tm.canonicalSource)
}

func InitializeMetadata(path string) *TrackMetadata {
	id3v1Metadata, id3v1Err := InternalReadID3V1Metadata(path, FileReader)
	id3v2Metadata := RawReadID3V2Metadata(path)
	tm := NewTrackMetadata()
	switch {
	case id3v1Err != nil && id3v2Metadata.Err != nil:
		tm.SetErrorCause(ID3V1, id3v1Err.Error())
		tm.SetErrorCause(ID3V2, id3v2Metadata.Err.Error())
	case id3v1Err != nil:
		tm.SetErrorCause(ID3V1, id3v1Err.Error())
		tm.SetID3v2Values(id3v2Metadata)
		tm.SetCanonicalSource(ID3V2)
	case id3v2Metadata.Err != nil:
		tm.SetErrorCause(ID3V2, id3v2Metadata.Err.Error())
		tm.SetID3v1Values(id3v1Metadata)
		tm.SetCanonicalSource(ID3V1)
	default:
		tm.SetID3v2Values(id3v2Metadata)
		tm.SetID3v1Values(id3v1Metadata)
		tm.SetCanonicalSource(ID3V2)
	}
	return tm
}

func (tm *TrackMetadata) ErrorCauses() []string {
	errCauses := []string{}
	for _, src := range sourceTypes {
		if cause := tm.commonMetadata(src).errorCause; cause != "" {
			errCauses = append(errCauses, cause)
		}
	}
	return errCauses
}

func (tm *TrackMetadata) Update(path string) (e []error) {
	for _, source := range sourceTypes {
		if updateErr := trackMetadataUpdaters[source](tm, path); updateErr != nil {
			e = append(e, updateErr)
		}
	}
	return
}

type ComparableStrings struct {
	External string
	Metadata string
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
