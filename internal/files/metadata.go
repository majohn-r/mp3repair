package files

import (
	"bytes"
	"strings"

	"github.com/bogem/id3v2/v2"
)

// sourceType identifies the source of a particular form of metadata
type sourceType int

const (
	undefinedSource sourceType = iota
	ID3V1
	ID3V2
	totalSources
)

var (
	nameComparators = map[sourceType]func(*comparableStrings) bool{
		ID3V1: id3v1NameDiffers,
		ID3V2: id3v2NameDiffers,
	}
	genreComparators = map[sourceType]func(*comparableStrings) bool{
		ID3V1: id3v1GenreDiffers,
		ID3V2: id3v2GenreDiffers,
	}
	trackMetadataUpdaters = map[sourceType]func(tm *TrackMetadata, path string) error{
		ID3V1: updateID3V1TrackMetadata,
		ID3V2: updateID3V2TrackMetadata,
	}
	sourceTypes = []sourceType{ID3V1, ID3V2}
)

func (sT sourceType) Name() string {
	switch sT {
	case ID3V1:
		return "ID3V1"
	case ID3V2:
		return "ID3V2"
	case totalSources:
		return "total"
	default:
		return "undefined"
	}
}

type metadataValue interface {
	string | int | id3v2.UnknownFrame
}

type correctableValue[V metadataValue] struct {
	original   V
	correction V
}

func (cv correctableValue[V]) correctedValue() V {
	return cv.correction
}

type commonMetadata struct {
	artistName   correctableValue[string]
	albumName    correctableValue[string]
	albumGenre   correctableValue[string]
	albumYear    correctableValue[string]
	trackName    correctableValue[string]
	trackNumber  correctableValue[int]
	errorCause   string
	requiresEdit bool
}

type TrackMetadata struct {
	data              map[sourceType]*commonMetadata
	musicCDIdentifier correctableValue[id3v2.UnknownFrame]
	canonicalSrc      sourceType
}

func newTrackMetadata() *TrackMetadata {
	return &TrackMetadata{
		data:         map[sourceType]*commonMetadata{},
		canonicalSrc: undefinedSource,
	}
}

type TrackMetadataMaker struct {
	Artist       string
	Album        string
	Genre        string
	Year         string
	TrackName    string
	TrackNumber  int
	CDIdentifier []byte
	Source       sourceType
}

func (maker *TrackMetadataMaker) Make() *TrackMetadata {
	tm := newTrackMetadata()
	for _, src := range []sourceType{ID3V1, ID3V2} {
		tm.setArtistName(src, maker.Artist)
		tm.setAlbumName(src, maker.Album)
		tm.setAlbumGenre(src, maker.Genre)
		tm.setAlbumYear(src, maker.Year)
		tm.setTrackName(src, maker.TrackName)
		tm.setTrackNumber(src, maker.TrackNumber)
	}
	tm.setCDIdentifier(maker.CDIdentifier)
	tm.setCanonicalSource(maker.Source)
	return tm
}

func isValidSource(src sourceType) bool {
	switch src {
	case ID3V1:
		return true
	case ID3V2:
		return true
	default:
		return false
	}
}

func (tm *TrackMetadata) commonMetadata(src sourceType) *commonMetadata {
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

func (tm *TrackMetadata) setArtistName(src sourceType, name string) {
	tm.commonMetadata(src).artistName.original = name
}

func (tm *TrackMetadata) correctArtistName(src sourceType, name string) {
	tm.commonMetadata(src).artistName.correction = name
}

func (tm *TrackMetadata) artistName(src sourceType) correctableValue[string] {
	return tm.commonMetadata(src).artistName
}

func (tm *TrackMetadata) canonicalArtistName() string {
	return tm.artistName(tm.canonicalSrc).original
}

func (tm *TrackMetadata) artistNameDiffers(nameFromFile string) (differs bool) {
	for _, src := range sourceTypes {
		comparison := &comparableStrings{
			external: nameFromFile,
			metadata: tm.artistName(src).original,
		}
		if tm.errorCause(src) == "" && nameComparators[src](comparison) {
			differs = true
			tm.setEditRequired(src)
			tm.correctArtistName(src, nameFromFile)
		}
	}
	return
}

func (tm *TrackMetadata) canonicalArtistNameMatches(artistNameFromFile string) bool {
	comparison := &comparableStrings{
		external: artistNameFromFile,
		metadata: tm.canonicalArtistName(),
	}
	comparator, exists := nameComparators[tm.canonicalSrc]
	if !exists {
		return false
	}
	return !comparator(comparison)
}

func (tm *TrackMetadata) setAlbumName(src sourceType, name string) {
	tm.commonMetadata(src).albumName.original = name
}

func (tm *TrackMetadata) correctAlbumName(src sourceType, name string) {
	tm.commonMetadata(src).albumName.correction = name
}

func (tm *TrackMetadata) albumName(src sourceType) correctableValue[string] {
	return tm.commonMetadata(src).albumName
}

func (tm *TrackMetadata) canonicalAlbumName() string {
	return tm.albumName(tm.canonicalSrc).original
}

func (tm *TrackMetadata) albumNameDiffers(nameFromFile string) (differs bool) {
	for _, src := range sourceTypes {
		comparison := &comparableStrings{
			external: nameFromFile,
			metadata: tm.albumName(src).original,
		}
		if tm.errorCause(src) == "" && nameComparators[src](comparison) {
			differs = true
			tm.setEditRequired(src)
			tm.correctAlbumName(src, nameFromFile)
		}
	}
	return
}

func (tm *TrackMetadata) canonicalAlbumNameMatches(nameFromFile string) bool {
	comparison := &comparableStrings{
		external: nameFromFile,
		metadata: tm.canonicalAlbumName(),
	}
	comparator, exists := nameComparators[tm.canonicalSrc]
	if !exists {
		return false
	}
	return !comparator(comparison)
}

func (tm *TrackMetadata) setAlbumGenre(src sourceType, name string) {
	tm.commonMetadata(src).albumGenre.original = name
}

func (tm *TrackMetadata) correctAlbumGenre(src sourceType, name string) {
	tm.commonMetadata(src).albumGenre.correction = name
}

func (tm *TrackMetadata) albumGenre(src sourceType) correctableValue[string] {
	return tm.commonMetadata(src).albumGenre
}

func (tm *TrackMetadata) canonicalAlbumGenre() string {
	return tm.albumGenre(tm.canonicalSrc).original
}

func (tm *TrackMetadata) albumGenreDiffers(canonicalAlbumGenre string) (differs bool) {
	for _, src := range sourceTypes {
		comparison := &comparableStrings{
			external: canonicalAlbumGenre,
			metadata: tm.albumGenre(src).original,
		}
		if tm.errorCause(src) == "" && genreComparators[src](comparison) {
			differs = true
			tm.setEditRequired(src)
			tm.correctAlbumGenre(src, canonicalAlbumGenre)
		}
	}
	return
}

func (tm *TrackMetadata) setAlbumYear(src sourceType, name string) {
	tm.commonMetadata(src).albumYear.original = name
}

func (tm *TrackMetadata) correctAlbumYear(src sourceType, name string) {
	tm.commonMetadata(src).albumYear.correction = name
}

func (tm *TrackMetadata) albumYear(src sourceType) correctableValue[string] {
	return tm.commonMetadata(src).albumYear
}

func (tm *TrackMetadata) canonicalAlbumYear() string {
	return tm.albumYear(tm.canonicalSrc).original
}

func (tm *TrackMetadata) albumYearDiffers(canonicalAlbumYear string) (differs bool) {
	for _, src := range sourceTypes {
		if tm.errorCause(src) == "" && !yearsMatch(tm.albumYear(src).original, canonicalAlbumYear) {
			differs = true
			tm.setEditRequired(src)
			tm.correctAlbumYear(src, canonicalAlbumYear)
		}
	}
	return
}

func (tm *TrackMetadata) setTrackName(src sourceType, name string) {
	tm.commonMetadata(src).trackName.original = name
}

func (tm *TrackMetadata) correctTrackName(src sourceType, name string) {
	tm.commonMetadata(src).trackName.correction = name
}

func (tm *TrackMetadata) trackName(src sourceType) correctableValue[string] {
	return tm.commonMetadata(src).trackName
}

func (tm *TrackMetadata) trackNameDiffers(nameFromFile string) (differs bool) {
	for _, src := range sourceTypes {
		comparison := &comparableStrings{
			external: nameFromFile,
			metadata: tm.trackName(src).original,
		}
		if tm.errorCause(src) == "" && nameComparators[src](comparison) {
			differs = true
			tm.setEditRequired(src)
			tm.correctTrackName(src, nameFromFile)
		}
	}
	return
}

func (tm *TrackMetadata) setTrackNumber(src sourceType, number int) {
	tm.commonMetadata(src).trackNumber.original = number
}

func (tm *TrackMetadata) correctTrackNumber(src sourceType, number int) {
	tm.commonMetadata(src).trackNumber.correction = number
}

func (tm *TrackMetadata) trackNumber(src sourceType) correctableValue[int] {
	return tm.commonMetadata(src).trackNumber
}

func (tm *TrackMetadata) trackNumberDiffers(trackNumberFromFileName int) (differs bool) {
	for _, src := range sourceTypes {
		if tm.errorCause(src) == "" && tm.trackNumber(src).original != trackNumberFromFileName {
			differs = true
			tm.setEditRequired(src)
			tm.correctTrackNumber(src, trackNumberFromFileName)
		}
	}
	return
}

func (tm *TrackMetadata) setErrorCause(src sourceType, cause string) {
	tm.commonMetadata(src).errorCause = cause
}

func (tm *TrackMetadata) errorCause(src sourceType) string {
	return tm.commonMetadata(src).errorCause
}

func (tm *TrackMetadata) setEditRequired(src sourceType) {
	tm.commonMetadata(src).requiresEdit = true
}

func (tm *TrackMetadata) editRequired(src sourceType) bool {
	return tm.commonMetadata(src).requiresEdit
}

func (tm *TrackMetadata) setCDIdentifier(body []byte) {
	tm.musicCDIdentifier.original = id3v2.UnknownFrame{Body: body}
}

func (tm *TrackMetadata) correctCDIdentifier(body []byte) {
	tm.musicCDIdentifier.correction = id3v2.UnknownFrame{Body: body}
}

func (tm *TrackMetadata) cdIdentifier() correctableValue[id3v2.UnknownFrame] {
	return tm.musicCDIdentifier
}

func (tm *TrackMetadata) canonicalCDIdentifier() id3v2.UnknownFrame {
	return tm.musicCDIdentifier.original
}

func (tm *TrackMetadata) cdIdentifierDiffers(canonicalCDIdentifier id3v2.UnknownFrame) (differs bool) {
	if tm.errorCause(ID3V2) == "" && !bytes.Equal(tm.cdIdentifier().original.Body, canonicalCDIdentifier.Body) {
		differs = true
		tm.setEditRequired(ID3V2)
		tm.correctCDIdentifier(canonicalCDIdentifier.Body)
	}
	return
}

func (tm *TrackMetadata) setCanonicalSource(src sourceType) {
	if isValidSource(src) {
		tm.canonicalSrc = src
	}
}

func (tm *TrackMetadata) setID3v2Values(d *id3v2Metadata) {
	tm.setArtistName(ID3V2, d.artistName)
	tm.setAlbumName(ID3V2, d.albumTitle)
	tm.setAlbumGenre(ID3V2, d.genre)
	tm.setAlbumYear(ID3V2, d.year)
	tm.setTrackName(ID3V2, d.trackName)
	tm.setTrackNumber(ID3V2, d.trackNumber)
	tm.setCDIdentifier(d.musicCDIdentifier.Body)
}

func (tm *TrackMetadata) setID3v1Values(v1 *id3v1Metadata) {
	tm.setArtistName(ID3V1, v1.artist())
	tm.setAlbumName(ID3V1, v1.album())
	if genre, genreFound := v1.genre(); genreFound {
		tm.setAlbumGenre(ID3V1, genre)
	}
	tm.setAlbumYear(ID3V1, v1.year())
	tm.setTrackName(ID3V1, v1.title())
	if track, trackValid := v1.track(); trackValid {
		tm.setTrackNumber(ID3V1, track)
	}
}

func (tm *TrackMetadata) IsValid() bool {
	return isValidSource(tm.canonicalSrc)
}

func initializeMetadata(path string) *TrackMetadata {
	id3v1Metadata, id3v1Err := internalReadID3V1Metadata(path, fileReader)
	id3v2Metadata := rawReadID3V2Metadata(path)
	tm := newTrackMetadata()
	switch {
	case id3v1Err != nil && id3v2Metadata.err != nil:
		tm.setErrorCause(ID3V1, id3v1Err.Error())
		tm.setErrorCause(ID3V2, id3v2Metadata.err.Error())
	case id3v1Err != nil:
		tm.setErrorCause(ID3V1, id3v1Err.Error())
		tm.setID3v2Values(id3v2Metadata)
		tm.setCanonicalSource(ID3V2)
	case id3v2Metadata.err != nil:
		tm.setErrorCause(ID3V2, id3v2Metadata.err.Error())
		tm.setID3v1Values(id3v1Metadata)
		tm.setCanonicalSource(ID3V1)
	default:
		tm.setID3v2Values(id3v2Metadata)
		tm.setID3v1Values(id3v1Metadata)
		tm.setCanonicalSource(ID3V2)
	}
	return tm
}

func (tm *TrackMetadata) errorCauses() []string {
	errCauses := make([]string, 0, len(sourceTypes))
	for _, src := range sourceTypes {
		if cause := tm.commonMetadata(src).errorCause; cause != "" {
			errCauses = append(errCauses, cause)
		}
	}
	return errCauses
}

func (tm *TrackMetadata) update(path string) (e []error) {
	for _, source := range sourceTypes {
		if updateErr := trackMetadataUpdaters[source](tm, path); updateErr != nil {
			e = append(e, updateErr)
		}
	}
	return
}

type comparableStrings struct {
	external string
	metadata string
}

func yearsMatch(metadataYear, albumYear string) bool {
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
