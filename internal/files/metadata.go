package files

import (
	"bytes"

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
	metadataUpdaters = map[SourceType]func(tM *TrackMetadata, path string, src SourceType) error{
		ID3V1: updateID3V1Metadata,
		ID3V2: updateID3V2Metadata,
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

// outside of unit tests
// TODO: make fields private
type TrackMetadata struct {
	Album             []string
	Artist            []string
	CanonicalType     SourceType
	ErrCause          []string
	Genre             []string
	MusicCDIdentifier id3v2.UnknownFrame
	Title             []string
	Track             []int
	Year              []string
	// these fields are set by the various xDiffers methods
	CorrectedAlbum             []string
	CorrectedArtist            []string
	CorrectedGenre             []string
	CorrectedMusicCDIdentifier id3v2.UnknownFrame
	CorrectedTitle             []string
	CorrectedTrack             []int
	CorrectedYear              []string
	RequiresEdit               []bool
}

func (tm *TrackMetadata) WithTrack(k []int) *TrackMetadata {
	for i := range min(len(k), int(TotalSources)) {
		tm.Track[i] = k[i]
	}
	return tm
}

func (tm *TrackMetadata) WithYear(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.Year[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithCorrectedAlbum(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.CorrectedAlbum[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithCorrectedArtist(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.CorrectedArtist[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithCorrectedGenre(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.CorrectedGenre[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithCorrectedMusicCDIdentifier(b []byte) *TrackMetadata {
	tm.CorrectedMusicCDIdentifier = id3v2.UnknownFrame{Body: b}
	return tm
}

func (tm *TrackMetadata) WithCorrectedTitle(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.CorrectedTitle[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithCorrectedTrack(k []int) *TrackMetadata {
	for i := range min(len(k), int(TotalSources)) {
		tm.CorrectedTrack[i] = k[i]
	}
	return tm
}

func (tm *TrackMetadata) WithCorrectedYear(s []string) *TrackMetadata {
	for i := range min(len(s), int(TotalSources)) {
		tm.CorrectedYear[i] = s[i]
	}
	return tm
}

func (tm *TrackMetadata) WithRequiresEdit(b []bool) *TrackMetadata {
	for i := range min(len(b), int(TotalSources)) {
		tm.RequiresEdit[i] = b[i]
	}
	return tm
}

func NewTrackMetadata() *TrackMetadata {
	return &TrackMetadata{
		Album:           make([]string, TotalSources),
		Artist:          make([]string, TotalSources),
		Title:           make([]string, TotalSources),
		Genre:           make([]string, TotalSources),
		Year:            make([]string, TotalSources),
		Track:           make([]int, TotalSources),
		ErrCause:        make([]string, TotalSources),
		CorrectedAlbum:  make([]string, TotalSources),
		CorrectedArtist: make([]string, TotalSources),
		CorrectedTitle:  make([]string, TotalSources),
		CorrectedGenre:  make([]string, TotalSources),
		CorrectedYear:   make([]string, TotalSources),
		CorrectedTrack:  make([]int, TotalSources),
		RequiresEdit:    make([]bool, TotalSources),
	}
}

func ReadRawMetadata(path string) *TrackMetadata {
	v1, id3v1Err := InternalReadID3V1Metadata(path, FileReader)
	d := RawReadID3V2Metadata(path)
	tM := NewTrackMetadata()
	switch {
	case id3v1Err != nil && d.err != nil:
		tM.ErrCause[ID3V1] = id3v1Err.Error()
		tM.ErrCause[ID3V2] = d.err.Error()
	case id3v1Err != nil:
		tM.ErrCause[ID3V1] = id3v1Err.Error()
		tM.SetID3v2Values(d)
		tM.CanonicalType = ID3V2
	case d.err != nil:
		tM.ErrCause[ID3V2] = d.err.Error()
		tM.SetID3v1Values(v1)
		tM.CanonicalType = ID3V1
	default:
		tM.SetID3v2Values(d)
		tM.SetID3v1Values(v1)
		tM.CanonicalType = ID3V2
	}
	return tM
}

func (tM *TrackMetadata) SetID3v2Values(d *Id3v2Metadata) {
	i := ID3V2
	tM.Album[i] = d.albumName
	tM.Artist[i] = d.artistName
	tM.Title[i] = d.trackName
	tM.Genre[i] = d.genre
	tM.Year[i] = d.year
	tM.Track[i] = d.trackNumber
	tM.MusicCDIdentifier = d.musicCDIdentifier
}

func (tM *TrackMetadata) SetID3v1Values(v1 *Id3v1Metadata) {
	index := ID3V1
	tM.Album[index] = v1.Album()
	tM.Artist[index] = v1.Artist()
	tM.Title[index] = v1.Title()
	if genre, ok := v1.Genre(); ok {
		tM.Genre[index] = genre
	}
	tM.Year[index] = v1.Year()
	if track, ok := v1.Track(); ok {
		tM.Track[index] = track
	}
}

func (tM *TrackMetadata) IsValid() bool {
	return tM.CanonicalType == ID3V1 || tM.CanonicalType == ID3V2
}

func (tM *TrackMetadata) CanonicalArtist() string {
	return tM.Artist[tM.CanonicalType]
}

func (tM *TrackMetadata) CanonicalAlbum() string {
	return tM.Album[tM.CanonicalType]
}

func (tM *TrackMetadata) CanonicalGenre() string {
	return tM.Genre[tM.CanonicalType]
}

func (tM *TrackMetadata) CanonicalYear() string {
	return tM.Year[tM.CanonicalType]
}

func (tM *TrackMetadata) CanonicalMusicCDIdentifier() id3v2.UnknownFrame {
	return tM.MusicCDIdentifier
}

func (tM *TrackMetadata) ErrorCauses() []string {
	errCauses := []string{}
	for _, e := range tM.ErrCause {
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
		if tM.ErrCause[sT] == "" && tM.Track[sT] != track {
			differs = true
			tM.RequiresEdit[sT] = true
			tM.CorrectedTrack[sT] = track
		}
	}
	return
}

func (tM *TrackMetadata) TrackTitleDiffers(title string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := &ComparableStrings{external: title, metadata: tM.Title[sT]}
		if tM.ErrCause[sT] == "" && nameComparators[sT](comparison) {
			differs = true
			tM.RequiresEdit[sT] = true
			tM.CorrectedTitle[sT] = title
		}
	}
	return
}

func (tM *TrackMetadata) AlbumTitleDiffers(albumTitle string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := &ComparableStrings{external: albumTitle, metadata: tM.Album[sT]}
		if tM.ErrCause[sT] == "" && nameComparators[sT](comparison) {
			differs = true
			tM.RequiresEdit[sT] = true
			tM.CorrectedAlbum[sT] = albumTitle
		}
	}
	return
}

func (tM *TrackMetadata) ArtistNameDiffers(artistName string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := &ComparableStrings{external: artistName, metadata: tM.Artist[sT]}
		if tM.ErrCause[sT] == "" && nameComparators[sT](comparison) {
			differs = true
			tM.RequiresEdit[sT] = true
			tM.CorrectedArtist[sT] = artistName
		}
	}
	return
}

func (tM *TrackMetadata) GenreDiffers(genre string) (differs bool) {
	for _, sT := range sourceTypes {
		comparison := &ComparableStrings{external: genre, metadata: tM.Genre[sT]}
		if tM.ErrCause[sT] == "" && genreComparators[sT](comparison) {
			differs = true
			tM.RequiresEdit[sT] = true
			tM.CorrectedGenre[sT] = genre
		}
	}
	return
}

func (tM *TrackMetadata) YearDiffers(year string) (differs bool) {
	for _, sT := range sourceTypes {
		if tM.ErrCause[sT] == "" && tM.Year[sT] != year {
			differs = true
			tM.RequiresEdit[sT] = true
			tM.CorrectedYear[sT] = year
		}
	}
	return
}

func (tM *TrackMetadata) MCDIDiffers(f id3v2.UnknownFrame) (differs bool) {
	if tM.ErrCause[ID3V2] == "" && !bytes.Equal(tM.MusicCDIdentifier.Body, f.Body) {
		differs = true
		tM.RequiresEdit[ID3V2] = true
		tM.CorrectedMusicCDIdentifier = f
	}
	return
}

func (tM *TrackMetadata) CanonicalAlbumTitleMatches(albumTitle string) bool {
	comparison := &ComparableStrings{external: albumTitle, metadata: tM.CanonicalAlbum()}
	return !nameComparators[tM.CanonicalType](comparison)
}

func (tM *TrackMetadata) CanonicalArtistNameMatches(artistName string) bool {
	comparison := &ComparableStrings{external: artistName, metadata: tM.CanonicalArtist()}
	return !nameComparators[tM.CanonicalType](comparison)
}

func updateMetadata(tM *TrackMetadata, path string) (e []error) {
	for _, source := range sourceTypes {
		if err := metadataUpdaters[source](tM, path, source); err != nil {
			e = append(e, err)
		}
	}
	return
}
