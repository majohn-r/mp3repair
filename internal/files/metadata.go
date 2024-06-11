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
	metadataUpdaters = map[SourceType]func(tM *TrackMetadataV1, path string) error{
		ID3V1: updateID3V1Metadata,
		ID3V2: UpdateID3V2Metadata,
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
			external: nameFromFile,
			metadata: tm.ArtistName(src).original,
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
		external: artistNameFromFile,
		metadata: tm.CanonicalArtistName(),
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
			external: nameFromFile,
			metadata: tm.AlbumName(src).original,
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
		external: nameFromFile,
		metadata: tm.CanonicalAlbumName(),
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
			external: canonicalAlbumGenre,
			metadata: tm.AlbumGenre(src).original,
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
			external: nameFromFile,
			metadata: tm.TrackName(src).original,
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

type TrackMetadataV1 struct {
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

func (tm *TrackMetadataV1) SetAlbumName(src SourceType, s string) {
	tm.albumName[src] = s
}

func (tm *TrackMetadataV1) SetArtistName(src SourceType, s string) {
	tm.artistName[src] = s
}

func (tm *TrackMetadataV1) SetErrorCause(src SourceType, s string) {
	tm.errorCause[src] = s
}

func (tm *TrackMetadataV1) SetGenre(src SourceType, s string) {
	tm.genre[src] = s
}

func (tm *TrackMetadataV1) SetTrackName(src SourceType, s string) {
	tm.trackName[src] = s
}

func (tm *TrackMetadataV1) SetTrackNumber(src SourceType, i int) {
	tm.trackNumber[src] = i
}

func (tm *TrackMetadataV1) SetYear(src SourceType, s string) {
	tm.year[src] = s
}

func (tm *TrackMetadataV1) WithAlbumNames(s []string) *TrackMetadataV1 {
	for i := range min(len(s), int(TotalSources)) {
		tm.albumName[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadataV1) WithArtistNames(s []string) *TrackMetadataV1 {
	for i := range min(len(s), int(TotalSources)) {
		tm.artistName[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadataV1) WithPrimarySource(t SourceType) *TrackMetadataV1 {
	tm.primarySource = t
	return tm
}

func (tm *TrackMetadataV1) WithErrorCauses(s []string) *TrackMetadataV1 {
	for i := range min(len(s), int(TotalSources)) {
		tm.errorCause[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadataV1) WithGenres(s []string) *TrackMetadataV1 {
	for i := range min(len(s), int(TotalSources)) {
		tm.genre[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadataV1) WithMusicCDIdentifier(b []byte) *TrackMetadataV1 {
	tm.musicCDIdentifier = id3v2.UnknownFrame{Body: b}
	return tm
}

func (tm *TrackMetadataV1) WithTrackNames(s []string) *TrackMetadataV1 {
	for i := range min(len(s), int(TotalSources)) {
		tm.trackName[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadataV1) WithTrackNumbers(k []int) *TrackMetadataV1 {
	for i := range min(len(k), int(TotalSources)) {
		tm.trackNumber[i] = k[i]
	}
	return tm
}

func (tm *TrackMetadataV1) WithYears(s []string) *TrackMetadataV1 {
	for i := range min(len(s), int(TotalSources)) {
		tm.year[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadataV1) WithCorrectedAlbumNames(s []string) *TrackMetadataV1 {
	for i := range min(len(s), int(TotalSources)) {
		tm.correctedAlbumName[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadataV1) WithCorrectedArtistNames(s []string) *TrackMetadataV1 {
	for i := range min(len(s), int(TotalSources)) {
		tm.correctedArtistName[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadataV1) WithCorrectedGenres(s []string) *TrackMetadataV1 {
	for i := range min(len(s), int(TotalSources)) {
		tm.correctedGenre[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadataV1) WithCorrectedMusicCDIdentifier(b []byte) *TrackMetadataV1 {
	tm.correctedMusicCDIdentifier = id3v2.UnknownFrame{Body: b}
	return tm
}

func (tm *TrackMetadataV1) WithCorrectedTrackNames(s []string) *TrackMetadataV1 {
	for i := range min(len(s), int(TotalSources)) {
		tm.correctedTrackName[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadataV1) WithCorrectedTrackNumbers(k []int) *TrackMetadataV1 {
	for i := range min(len(k), int(TotalSources)) {
		tm.correctedTrackNumber[i] = k[i]
	}
	return tm
}

func (tm *TrackMetadataV1) WithCorrectedYears(s []string) *TrackMetadataV1 {
	for i := range min(len(s), int(TotalSources)) {
		tm.correctedYear[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadataV1) WithRequiresEdits(b []bool) *TrackMetadataV1 {
	for i := range min(len(b), int(TotalSources)) {
		tm.requiresEdit[i] = b[i]
	}
	return tm
}

func NewTrackMetadataV1() *TrackMetadataV1 {
	return &TrackMetadataV1{
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

func ReadRawMetadata(path string) *TrackMetadataV1 {
	id3v1Metadata, id3v1Err := InternalReadID3V1Metadata(path, FileReader)
	id3v2Metadata := RawReadID3V2Metadata(path)
	tM := NewTrackMetadataV1()
	switch {
	case id3v1Err != nil && id3v2Metadata.Err != nil:
		tM.errorCause[ID3V1] = id3v1Err.Error()
		tM.errorCause[ID3V2] = id3v2Metadata.Err.Error()
	case id3v1Err != nil:
		tM.errorCause[ID3V1] = id3v1Err.Error()
		tM.SetID3v2Values(id3v2Metadata)
		tM.primarySource = ID3V2
	case id3v2Metadata.Err != nil:
		tM.errorCause[ID3V2] = id3v2Metadata.Err.Error()
		tM.SetID3v1Values(id3v1Metadata)
		tM.primarySource = ID3V1
	default:
		tM.SetID3v2Values(id3v2Metadata)
		tM.SetID3v1Values(id3v1Metadata)
		tM.primarySource = ID3V2
	}
	return tM
}

func (tM *TrackMetadataV1) SetID3v2Values(d *Id3v2Metadata) {
	i := ID3V2
	tM.albumName[i] = d.AlbumTitle
	tM.artistName[i] = d.ArtistName
	tM.trackName[i] = d.TrackName
	tM.genre[i] = d.Genre
	tM.year[i] = d.Year
	tM.trackNumber[i] = d.TrackNumber
	tM.musicCDIdentifier = d.MusicCDIdentifier
}

func (tM *TrackMetadataV1) SetID3v1Values(v1 *Id3v1Metadata) {
	index := ID3V1
	tM.albumName[index] = v1.Album()
	tM.artistName[index] = v1.Artist()
	tM.trackName[index] = v1.Title()
	if genre, genreFound := v1.Genre(); genreFound {
		tM.genre[index] = genre
	}
	tM.year[index] = v1.Year()
	if track, trackValid := v1.Track(); trackValid {
		tM.trackNumber[index] = track
	}
}

func (tM *TrackMetadataV1) IsValid() bool {
	return tM.primarySource == ID3V1 || tM.primarySource == ID3V2
}

func (tM *TrackMetadataV1) CanonicalArtist() string {
	return tM.artistName[tM.primarySource]
}

func (tM *TrackMetadataV1) CanonicalAlbum() string {
	return tM.albumName[tM.primarySource]
}

func (tM *TrackMetadataV1) CanonicalGenre() string {
	return tM.genre[tM.primarySource]
}

func (tM *TrackMetadataV1) CanonicalYear() string {
	return tM.year[tM.primarySource]
}

func (tM *TrackMetadataV1) CanonicalMusicCDIdentifier() id3v2.UnknownFrame {
	return tM.musicCDIdentifier
}

func (tM *TrackMetadataV1) ErrorCauses() []string {
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

// does the track number from metadata differ from the track number acquired
// from the track's file name? If so, make a note of it
func (tM *TrackMetadataV1) TrackNumberDiffers(trackFromFileName int) (differs bool) {
	for _, sT := range sourceTypes {
		if tM.errorCause[sT] == "" && tM.trackNumber[sT] != trackFromFileName {
			differs = true
			tM.requiresEdit[sT] = true
			tM.correctedTrackNumber[sT] = trackFromFileName
		}
	}
	return
}

func (tM *TrackMetadataV1) TrackTitleDiffers(title string) (differs bool) {
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

func (tM *TrackMetadataV1) AlbumTitleDiffers(albumTitle string) (differs bool) {
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

func (tM *TrackMetadataV1) ArtistNameDiffers(artistName string) (differs bool) {
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

func (tM *TrackMetadataV1) GenreDiffers(genre string) (differs bool) {
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

func (tM *TrackMetadataV1) YearDiffers(year string) (differs bool) {
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

func (tM *TrackMetadataV1) MCDIDiffers(f id3v2.UnknownFrame) (differs bool) {
	if tM.errorCause[ID3V2] == "" && !bytes.Equal(tM.musicCDIdentifier.Body, f.Body) {
		differs = true
		tM.requiresEdit[ID3V2] = true
		tM.correctedMusicCDIdentifier = f
	}
	return
}

func (tM *TrackMetadataV1) CanonicalAlbumTitleMatches(albumTitle string) bool {
	comparison := &ComparableStrings{external: albumTitle, metadata: tM.CanonicalAlbum()}
	return !nameComparators[tM.primarySource](comparison)
}

func (tM *TrackMetadataV1) CanonicalArtistNameMatches(artistName string) bool {
	comparison := &ComparableStrings{external: artistName, metadata: tM.CanonicalArtist()}
	return !nameComparators[tM.primarySource](comparison)
}

func updateMetadata(tM *TrackMetadataV1, path string) (e []error) {
	for _, source := range sourceTypes {
		if updateErr := metadataUpdaters[source](tM, path); updateErr != nil {
			e = append(e, updateErr)
		}
	}
	return
}
